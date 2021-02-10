// btcgw is an API Server that handles anchor CRUD for BBc-1 Ledger Subsystem.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ebiiim/btcgw/api"
	"github.com/ebiiim/btcgw/auth"
	"github.com/ebiiim/btcgw/btc"
	"github.com/ebiiim/btcgw/gw"
	"github.com/ebiiim/btcgw/model"
	"github.com/ebiiim/btcgw/store"
	"github.com/ebiiim/btcgw/util"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/httprate"
	_ "gocloud.dev/docstore/mongodocstore"
)

var (
	cliPath = util.GetEnvOr("BITCOIN_CLI_PATH", "./bitcoin-cli")
	btcNet  = model.BTCNet(uint8(util.MustAtoi(util.GetEnvOr("BITCOIN_NETWORK", "3")))) // model.BTCTestnet3
	rpcAddr = util.GetEnvOr("BITCOIND_ADDR", "")
	rpcPort = util.GetEnvOr("BITCOIND_PORT", "")
	rpcUser = util.GetEnvOr("BITCOIND_RPC_USER", "")
	rpcPW   = util.GetEnvOr("BITCOIND_RPC_PASSWORD", "")
)

const (
	dbName      = "btcgw"
	anchorTable = "anchors"
	anchorKey   = "cid"
	authTable   = "apikeys"
	authKey     = "key"
)

func useMongoDBAtlas() {
	// Please set environment variables first.
	// e.g. `set -a; source .env; set +a;`
	var (
		mongoUser = os.Getenv("MONGO_USER")
		mongoPW   = os.Getenv("MONGO_PASSWORD")
		mongoHost = os.Getenv("MONGO_HOSTNAME")
	)
	const (
		mongoEnv      = "MONGO_SERVER_URL"
		mongoAtlasFmt = "mongodb+srv://%s:%s@%s"
	)
	mongoAtlas := fmt.Sprintf(mongoAtlasFmt, mongoUser, mongoPW, mongoHost)
	if err := os.Setenv(mongoEnv, mongoAtlas); err != nil {
		panic(err)
	}
}

func mongoStore() string {
	return fmt.Sprintf("mongo://%s/%s?id_field=%s", dbName, anchorTable, anchorKey)
}

func mongoAuthenticator() string {
	return fmt.Sprintf("mongo://%s/%s?id_field=%s", dbName, authTable, authKey)
}

func main() {
	var port = flag.Int("port", 8080, "HTTP port")
	var dev = flag.Bool("dev", false, "Use AnchorVersion 255 and prettify HTTP response body")
	flag.Parse()

	if *dev {
		fmt.Println("Development Environment")
		// Set AnchorVersion to test.
		model.XAnchorVersion(255)
		api.PrettifyResponseJSON = true
	} else {
		fmt.Println("")
		fmt.Println("==================================")
		fmt.Println("===== Production Environment =====")
		fmt.Println("==================================")
		fmt.Println("")
	}

	useMongoDBAtlas()
	// Setup Gateway.
	var err error
	btcCLI := btc.NewBitcoinCLI(cliPath, btcNet, rpcAddr, rpcPort, rpcUser, rpcPW)
	docStore := store.NewDocstore(mongoStore())
	if err = docStore.Open(); err != nil {
		log.Println(err)
		return
	}
	gwImpl := gw.NewGatewayImpl(model.BTCTestnet3, btcCLI, docStore)
	// Setup Authenticator.
	var a auth.Authenticator
	a = auth.MustNewDocstoreAuth(mongoAuthenticator())
	// a = &auth.SpecialAuth{}

	// Setup GatewayService.
	gwService := api.NewGatewayService(gwImpl, a)
	defer func() {
		if cErr := gwService.Close(); cErr != nil {
			log.Printf("%v (captured err: %v)", cErr, err)
		}
	}()

	// Setup Chi.
	r := chi.NewRouter()
	r.Use(middleware.RealIP)                     // use this only if you have a trusted reverse proxy
	r.Use(httprate.LimitByIP(60, 1*time.Minute)) // returns 429
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Heartbeat("/healthz"))
	r.Use(gwService.OAPIValidator())
	api.HandlerFromMux(gwService, r)

	// Serve.
	addr := fmt.Sprintf("0.0.0.0:%d", *port)
	fmt.Printf("Listening on: http://%s\n", addr)
	s := &http.Server{
		Handler: r,
		Addr:    addr,
	}
	go func() {
		fmt.Println(s.ListenAndServe())
	}()

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	fmt.Printf("Signal %s received. Shutting down...\n", sig)
	ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()
	if err := s.Shutdown(ctx); err != nil {
		log.Printf("Graceful shutdown failed: %v\n", err)
	}
}

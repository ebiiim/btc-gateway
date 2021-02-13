// apikey is an API Server that handles API Key creation and deletion for btcgw.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ebiiim/btcgw/api"
	"github.com/ebiiim/btcgw/auth"
	"github.com/ebiiim/btcgw/util"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	_ "gocloud.dev/docstore/mongodocstore"
)

var (
	port = util.GetEnvIntOr("PORT", 8081)
)

const (
	dbName    = "btcgw"
	authTable = "apikeys"
	authKey   = "key"
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

func mongoAuthenticator() string {
	return fmt.Sprintf("mongo://%s/%s?id_field=%s", dbName, authTable, authKey)
}

func main() {

	fmt.Println("")
	fmt.Println("==================================")
	fmt.Println("===== Production Environment =====")
	fmt.Println("==================================")
	fmt.Println("")

	useMongoDBAtlas()

	var err error

	// Setup APIKeyService.
	// It's allowed to open the same DocstoreAuth twice, and it makes APIKeyService.Close() successful.
	docAuth := auth.MustNewDocstoreAuth(mongoAuthenticator())
	akService := api.NewAPIKeyService(docAuth)
	defer func() {
		if cErr := akService.Close(); cErr != nil {
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
	r.Use(akService.OAPIValidator())
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"POST", "DELETE", "HEAD", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"*"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	api.APIKeyHandlerFromMux(akService, r)

	// Serve.
	addr := fmt.Sprintf("0.0.0.0:%d", port)
	fmt.Printf("Listening on: http://%s\n", addr)
	s := &http.Server{
		Handler: r,
		Addr:    addr,
	}

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		log.Println(s.ListenAndServe())
		sigCh <- syscall.SIGTERM
	}()
	sig := <-sigCh
	fmt.Printf("Signal <%s> received. Shutting down...\n", sig)
	ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()
	if err := s.Shutdown(ctx); err != nil {
		log.Printf("Graceful shutdown failed: %v\n", err)
	}
}

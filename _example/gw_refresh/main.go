package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ebiiim/btc-gateway/btc"
	"github.com/ebiiim/btc-gateway/gw"
	"github.com/ebiiim/btc-gateway/model"
	"github.com/ebiiim/btc-gateway/store"
	"github.com/ebiiim/btc-gateway/util"

	_ "gocloud.dev/docstore/mongodocstore"
)

func init() {
	// Set AnchorVersion to test.
	model.XAnchorVersion(255)
}

var (
	cliPath = util.GetEnvOr("BITCOIN_CLI_PATH", "../../bitcoin-cli")
	btcNet  = model.BTCNet(uint8(util.MustAtoi(util.GetEnvOr("BITCOIN_NETWORK", "3")))) // model.BTCTestnet3
	rpcAddr = util.GetEnvOr("BITCOIND_ADDR", "")
	rpcPort = util.GetEnvOr("BITCOIND_PORT", "")
	rpcUser = util.GetEnvOr("BITCOIND_RPC_USER", "")
	rpcPW   = util.GetEnvOr("BITCOIND_RPC_PASSWORD", "")
)

const (
	dbName    = "btcgw"
	tableName = "anchors"
	key       = "cid"
)

func useMongoDBAtlas() string {
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
		log.Println(err)
		return ""
	}
	return fmt.Sprintf("mongo://%s/%s?id_field=%s", dbName, tableName, key)
}

func getAnchor(txid string) *model.AnchorRecord {
	var b btc.BTC
	xCLI := btc.NewBitcoinCLI(cliPath, btcNet, rpcAddr, rpcPort, rpcUser, rpcPW)
	b = xCLI
	ctx, cancelFunc := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelFunc()
	btctx := util.MustDecodeHexString(txid)
	ar, err := b.GetAnchor(ctx, btctx)
	if err != nil {
		log.Fatal(err)
	}
	return ar
}

func main() {
	// Setup BTC.
	var b btc.BTC
	xCLI := btc.NewBitcoinCLI(cliPath, btcNet, rpcAddr, rpcPort, rpcUser, rpcPW)
	b = xCLI

	// Setup Store.
	var conn string
	conn = useMongoDBAtlas()
	var st store.Store
	docs := store.NewDocstore(conn)
	st = docs
	var err error
	if err = docs.Open(); err != nil {
		log.Println(err)
		return
	}
	defer func() {
		if cErr := docs.Close(); cErr != nil {
			log.Printf("%v (captured err: %v)", cErr, err)
		}
	}()

	// Setup Gateway.
	var g gw.Gateway
	gImpl := gw.NewGatewayImpl(model.BTCTestnet3, b, st)
	g = gImpl

	// Get the record and update its Note.
	ctx, cancelFunc := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelFunc()
	dom32 := util.MustDecodeHexString("456789abc0ef0123456089abcdef0023456789a0cdef0123406789abcde00123")
	tx32 := util.MustDecodeHexString("56789abcd0f0123456709abcdef0103456789ab0def0123450789abcdef01235")
	ar, err := g.GetRecord(ctx, dom32, tx32)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(ar)
	s := fmt.Sprintf("Updated at: %s", time.Now().Format(time.RFC3339))
	if err := g.RefreshRecord(ctx, dom32, tx32, nil, &s); err != nil {
		log.Println(err)
		return
	}
	ar2, err := g.GetRecord(ctx, dom32, tx32)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(ar2)
}

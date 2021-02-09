package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ebiiim/btcgw/btc"
	"github.com/ebiiim/btcgw/gw"
	"github.com/ebiiim/btcgw/model"
	"github.com/ebiiim/btcgw/store"
	"github.com/ebiiim/btcgw/util"

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

	// Define Anchor data.
	dom32 := util.MustDecodeHexString("456789abc0ef0123456089abcdef0023456789a0cdef0123406789abcde00123")
	tx32 := util.MustDecodeHexString("56789abcd0f0123456709abcdef0103456789ab0def0123450789abcdef01235")

	// Set UTXO and Bitcoin Address for PutAnchor.
	utxo := util.MustDecodeHexString("6928e1c6478d1f55ed1a5d86e1ab24669a14f777b879bbb25c746543810bf916")
	btcAddr := "tb1qhexc7d0fzex7lrzw3l0j2dmvhgegt02ckfdzjr"
	xCLI.XPrepareAnchor(utxo, btcAddr)

	// Anchor a BBc-1 transaction in Bitcoin block chain.
	ctx, cancelFunc := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelFunc()
	txID, err := g.RegisterTransaction(ctx, dom32, tx32)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Printf("%x\n", txID) // cfb3b1082976d42374e8561b21226595add8ae3d37cf9fb7b7a78055cade8a4c
}

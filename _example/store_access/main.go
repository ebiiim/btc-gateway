package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ebiiim/btc-gateway/btc"
	"github.com/ebiiim/btc-gateway/model"
	"github.com/ebiiim/btc-gateway/store"
	"github.com/ebiiim/btc-gateway/util"

	_ "gocloud.dev/docstore/awsdynamodb"
	_ "gocloud.dev/docstore/memdocstore"
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

func useMemdocstore() string {
	var (
		dbName = fmt.Sprintf("%s_%s", dbName, tableName)
		dbFile = dbName + ".db"
	)
	return fmt.Sprintf("mem://%s/%s?filename=%s", dbName, key, dbFile)
}

func useDynamoDB() string {
	// Assumes that AWS CLI default profile exists.
	// Please create a table in DynamoDB first.
	//
	// store-create-dynamodb:
	//   aws dynamodb create-table \
	//     --table-name anchors.btcgw \
	//     --attribute-definitions \
	//       AttributeName=cid,AttributeType=S \
	//     --key-schema AttributeName=cid,KeyType=HASH \
	//     --provisioned-throughput ReadCapacityUnits=1,WriteCapacityUnits=1
	//
	// store-delete-dynamodb:
	// 	aws dynamodb delete-table --table-name anchors.btcgw
	//
	return fmt.Sprintf("dynamodb://%s.%s?partition_key=%s", tableName, dbName, key)
}

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
		log.Fatal(err)
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
	// Setup Store.
	var conn string
	conn = useMemdocstore()
	// conn = useDynamoDB()
	// conn = useMongoDBAtlas()
	var st store.Store
	docs := store.NewDocstore(conn)
	st = docs
	var err error
	defer func() {
		if cErr := docs.Close(); cErr != nil {
			log.Printf("%v (captured err: %v)", cErr, err)
		}
	}()

	// Prepare anchor to put.
	ar := getAnchor("6928e1c6478d1f55ed1a5d86e1ab24669a14f777b879bbb25c746543810bf916")
	ar.BBc1DomainName = "testDom"
	ar.Note = "hello world"

	// Put the anchor in Store, and then Get it.
	ctx, cancelFunc := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelFunc()
	if err = st.Put(ctx, ar); err != nil {
		log.Println(err)
		return
	}
	dom32 := util.MustDecodeHexString("456789abc0ef0123456089abcdef0023456789a0cdef0123406789abcde00123")
	tx32 := util.MustDecodeHexString("56789abcd0f0123456709abcdef0103456789ab0def0123450789abcdef01234")
	ar2, err := st.Get(ctx, dom32, tx32)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Printf("%s\n", ar2)
}

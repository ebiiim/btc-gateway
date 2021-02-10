// utxoadd is a CLI tool to add UTXOs to DocstoreWallet.
package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ebiiim/btcgw/btc"

	_ "gocloud.dev/docstore/mongodocstore"
)

const (
	dbName    = "btcgw"
	tableName = "utxos"
	key       = "addr"
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

func mongoWallet() string {
	return fmt.Sprintf("mongo://%s/%s?id_field=%s", dbName, tableName, key)
}

var usageFmt = "Usage: %s [addr] [utxo]\n"

func do() int {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usageFmt, os.Args[0])
	}
	if len(os.Args) < 3 {
		flag.Usage()
		return 1
	}

	// Check args.
	addr := os.Args[1]
	txid, err := hex.DecodeString(os.Args[2])
	if err != nil {
		log.Println(err)
		return 1
	}

	// Open Wallet.
	useMongoDBAtlas()
	d := btc.MustNewDocstoreWallet(mongoWallet(), addr)
	defer func() {
		if cErr := d.Close(); cErr != nil {
			log.Printf("%v (captured err: %v)", cErr, err)
		}
	}()

	// Do.
	if err := d.AddUTXO(txid, addr); err != nil {
		log.Println(err)
		return 2
	}

	return 0
}

func main() {
	os.Exit(do())
}

// apikey-cli is an CLI tool to manage API Keys.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ebiiim/btcgw/auth"

	_ "gocloud.dev/docstore/mongodocstore"
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

var usageFmt1 = "Usage: %s cmd [args]...\n"
var usageFmt2 = fmt.Sprintf(`cmd:
  %s
args:
`, cmdCreate)
var usageFmt3 = fmt.Sprintf(`cmd:
  %s
args:
`, cmdDelete)

const (
	cmdCreate = "create"
	cmdDelete = "delete"
)

func do() int {
	createCmd := flag.NewFlagSet(cmdCreate, flag.ExitOnError)
	domID := createCmd.String("domain", "", `BBc-1 Domain ID in hex string (required if admin=false)`)
	gAdmin := createCmd.Bool("admin", false, "Global Administrator (default: false)")
	note := createCmd.String("note", "", `Note (optional)`)

	deleteCmd := flag.NewFlagSet(cmdDelete, flag.ExitOnError)
	apiKey := deleteCmd.String("apikey", "", "API Key (required)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usageFmt1, os.Args[0])
		fmt.Fprintf(os.Stderr, usageFmt2)
		createCmd.PrintDefaults()
		fmt.Fprintf(os.Stderr, usageFmt3)
		deleteCmd.PrintDefaults()
	}

	if len(os.Args) < 2 {
		flag.Usage()
		return 1
	}

	useMongoDBAtlas()
	a := auth.MustNewDocstoreAuth(mongoAuthenticator())
	defer a.Close()
	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()

	switch os.Args[1] {
	default:
		flag.Usage()
		return 2
	case cmdCreate:
		createCmd.Parse(os.Args[2:])
		if !(*gAdmin) && (*domID) == "" {
			flag.Usage()
			return 3
		}
		a, err := a.Generate(ctx, *domID, *gAdmin, *note)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v", err)
			return 4
		}
		fmt.Println(a.Key)
	case cmdDelete:
		deleteCmd.Parse(os.Args[2:])
		err := a.Delete(ctx, *apiKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v", err)
			return 5
		}
	}
	return 0
}

func main() {
	os.Exit(do())
}

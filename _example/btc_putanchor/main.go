package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ebiiim/btcgw/btc"
	"github.com/ebiiim/btcgw/model"
	"github.com/ebiiim/btcgw/util"
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

func main() {
	var b btc.BTC
	xCLI := btc.NewBitcoinCLI(cliPath, btcNet, rpcAddr, rpcPort, rpcUser, rpcPW)
	b = xCLI

	// Define Anchor data.
	ts := time.Unix(1612449628, 0)
	dom32 := util.MustDecodeHexString("456789abc0ef0123456089abcdef0023456789a0cdef0123406789abcde00123")
	tx32 := util.MustDecodeHexString("56789abcd0f0123456709abcdef0103456789ab0def0123450789abcdef01234")
	anc := model.NewAnchor(model.BTCTestnet3, ts, dom32, tx32)

	// Set UTXO and Bitcoin Address for PutAnchor.
	utxo := util.MustDecodeHexString("57511f74c3836c0d4d62a6183fa54e600372e1aed5b5be2f78ef5b766a314a5d")
	btcAddr := "tb1qhexc7d0fzex7lrzw3l0j2dmvhgegt02ckfdzjr"
	xCLI.XSetUTXO(utxo, btcAddr)

	// Put an Anchor in Bitcoin block chain.
	ctx, cancelFunc := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelFunc()
	txID, err := b.PutAnchor(ctx, anc)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%x\n", txID) // 6928e1c6478d1f55ed1a5d86e1ab24669a14f777b879bbb25c746543810bf916
}

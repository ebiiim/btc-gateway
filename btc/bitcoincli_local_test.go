// +build localTest

package btc_test

import (
	"context"
	"encoding/hex"
	"testing"
	"time"

	"github.com/ebiiim/btc-gateway/btc"
	"github.com/ebiiim/btc-gateway/model"
)

const (
	path = "../bitcoin-cli" // Put bitcoin-cli in project root.
	net  = model.BTCTestnet3
	addr = ""
	port = ""
	user = ""
	pw   = ""
)

func prepCLI(t *testing.T) (*btc.BitcoinCLI, context.Context, context.CancelFunc) {
	t.Helper()
	b := btc.NewBitcoinCLI(path, net, addr, port, user, pw)
	ctx := context.Background()
	ctx, cancelFunc := context.WithTimeout(ctx, 5*time.Second)
	return b, ctx, cancelFunc
}

func TestPing(t *testing.T) {
	b, ctx, cf := prepCLI(t)
	defer cf()
	if err := b.Ping(ctx); err != nil {
		t.Error(err)
	}
}

func TestGetBalance(t *testing.T) {
	b, ctx, cf := prepCLI(t)
	defer cf()
	bal, err := b.GetBalance(ctx)
	if err != nil {
		t.Error(err)
	}
	t.Log(bal)
}

func TestCreateRawTransactionForAnchor(t *testing.T) {
	const (
		txid     = "c7ace9d33c00b870e183f7dc929d3887efe257317a0d24810b2ee91fd08c6535"
		recvAddr = "tb1qhexc7d0fzex7lrzw3l0j2dmvhgegt02ckfdzjr"
		unspent  = "0.01168624"
		fee      = 10000
		opRet    = "7468697320697320612070656e0a" // "this is a pen"
		want     = "020000000135658cd01fe92e0b81240d7a3157e2ef87389d92dcf783e170b8003cd3e9acc70000000000ffffffff02e0ad110000000000160014be4d8f35e9164def8c4e8fdf25376cba3285bd580000000000000000106a0e7468697320697320612070656e0a00000000"
	)
	b, ctx, cf := prepCLI(t)
	defer cf()
	bT, _ := hex.DecodeString(txid)
	bOpRet, _ := hex.DecodeString(opRet)
	bs, err := b.CreateRawTransactionForAnchor(ctx, bT, unspent, recvAddr, fee, bOpRet)
	if err != nil {
		t.Error(err)
		t.Skip()
	}
	if got := hex.EncodeToString(bs); got != want {
		t.Errorf("want %s but got %s", want, got)
	}
}

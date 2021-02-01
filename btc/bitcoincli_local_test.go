// +build localTest

package btc_test

import (
	"context"
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

func TestPing(t *testing.T) {
	b := btc.NewBitcoinCLI(path, net, addr, port, user, pw)
	ctx := context.Background()
	ctx, cancelFunc := context.WithTimeout(ctx, 5*time.Second)
	defer cancelFunc()
	if err := b.Ping(ctx); err != nil {
		t.Error(err)
	}
}

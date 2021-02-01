package btc_test

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/ebiiim/btc-gateway/btc"
	"github.com/ebiiim/btc-gateway/model"
)

const (
	path1 = "/home/foo/bar/bitcoin-cli"
	addr1 = "192.168.0.1"
	port1 = "12345"
	user1 = "taro"
	pw1   = "super_strong_password"

	txid1       = "57511f74c3836c0d4d62a6183fa54e600372e1aed5b5be2f78ef5b766a314a5d"
	recvAddr1   = "tb1qhexc7d0fzex7lrzw3l0j2dmvhgegt02ckfdzjr"
	recvAmount1 = "0.01158624"
)

func TestCalcFee(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		bal        string
		feeSatoshi uint
		result     string
	}{
		{"normal", recvAmount1, 20_000, "0.01138624"},
		{"big_fee", recvAmount1, 123_456, "0.01035168"},
		{"big_amo", "12345.12345678", 20_000, "12345.12325678"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			bal, err := btc.CalcFee(c.bal, c.feeSatoshi)
			if err != nil {
				t.Error(err)
				t.Skip()
			}
			if bal != c.result {
				t.Errorf("got %s but want %s", bal, c.result)
			}
		})
	}
}

func TestNewBitcoinCLI(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		binPath     string
		btcNet      model.BTCNet
		rpcAddr     string
		rpcPort     string
		rpcUser     string
		rpcPassword string
	}{
		{"all", path1, model.BTCMainnet, addr1, port1, user1, pw1},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			b := btc.NewBitcoinCLI(c.binPath, c.btcNet, c.rpcAddr, c.rpcPort, c.rpcUser, c.rpcPassword)
			if b.BinPath() != c.binPath {
				t.Error("wrong binPath")
			}
			if b.BTCNet() != c.btcNet {
				t.Error("wrong btcNet")
			}
			if b.RPCAddr() != c.rpcAddr {
				t.Error("wrong rpcAddr")
			}
			if b.RPCPort() != c.rpcPort {
				t.Error("wrong rpcPort")
			}
			if b.RPCUser() != c.rpcUser {
				t.Error("wrong rpcUser")
			}
			if b.RPCPassword() != c.rpcPassword {
				t.Error("wrong rpcPassword")
			}
		})
	}
}

func TestBitcoinCLI_ConnArgs(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		btcNet      model.BTCNet
		rpcAddr     string
		rpcPort     string
		rpcUser     string
		rpcPassword string
		want        []string
	}{
		{"all_mainnet", model.BTCMainnet, addr1, port1, user1, pw1, []string{"-chain=main", "-rpcconnect=" + addr1, "-rpcport=" + port1, "-rpcuser=" + user1, "-rpcpassword=" + pw1}},
		{"all_testnet3", model.BTCTestnet3, addr1, port1, user1, pw1, []string{"-chain=test", "-rpcconnect=" + addr1, "-rpcport=" + port1, "-rpcuser=" + user1, "-rpcpassword=" + pw1}},
		{"no_addr", model.BTCMainnet, "", port1, user1, pw1, []string{"-chain=main", "-rpcport=" + port1, "-rpcuser=" + user1, "-rpcpassword=" + pw1}},
		{"no_port", model.BTCMainnet, addr1, "", user1, pw1, []string{"-chain=main", "-rpcconnect=" + addr1, "-rpcuser=" + user1, "-rpcpassword=" + pw1}},
		{"no_user", model.BTCMainnet, addr1, port1, "", pw1, []string{"-chain=main", "-rpcconnect=" + addr1, "-rpcport=" + port1, "-rpcpassword=" + pw1}},
		{"no_pw", model.BTCMainnet, addr1, port1, user1, "", []string{"-chain=main", "-rpcconnect=" + addr1, "-rpcport=" + port1, "-rpcuser=" + user1}},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			b := btc.NewBitcoinCLI(path1, c.btcNet, c.rpcAddr, c.rpcPort, c.rpcUser, c.rpcPassword)
			s := b.ConnArgs()
			if !reflect.DeepEqual(s, c.want) {
				t.Errorf("got %+v but want %+v", s, c.want)
			}
		})
	}
}

func TestBitcoinCLI_Run_DryRun(t *testing.T) {
	btc.DryRun(true)

	t.Parallel()
	cases := []struct {
		name        string
		args        []string
		fullcommand string
	}{
		{"normal", []string{"ABC"}, fmt.Sprintf("%s -chain=test ABC", path1)},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			b := btc.NewBitcoinCLI(path1, model.BTCTestnet3, "", "", "", "")
			ctx := context.Background()
			stdout, stderr, err := b.Run(ctx, c.args)
			if !errors.Is(err, btc.ErrDryRun) {
				t.Errorf("unexpected err %+v (stdout=%v, stderr=%v)", err, stdout.String(), stderr.String())
				t.Skip()
			}
			if err.Error() != c.fullcommand {
				t.Errorf("got %+v but want %+v", err.Error(), c.fullcommand)
			}
		})
	}
}

func TestBitcoinCLI_Ping_DryRun(t *testing.T) {
	btc.DryRun(true)

	t.Parallel()
	cases := []struct {
		name        string
		fullcommand string
	}{
		{"normal", fmt.Sprintf("%s -chain=test ping", path1)},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			b := btc.NewBitcoinCLI(path1, model.BTCTestnet3, "", "", "", "")
			ctx := context.Background()
			err := b.Ping(ctx)
			if !errors.Is(err, btc.ErrDryRun) {
				t.Errorf("unexpected err %+v", err)
				t.Skip()
			}
			if err.Error() != c.fullcommand {
				t.Errorf("got %+v but want %+v", err.Error(), c.fullcommand)
			}
		})
	}
}

func TestBitcoinCLI_GetBalance_DryRun(t *testing.T) {
	btc.DryRun(true)

	t.Parallel()
	cases := []struct {
		name        string
		fullcommand string
	}{
		{"normal", fmt.Sprintf("%s -chain=test getbalance", path1)},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			b := btc.NewBitcoinCLI(path1, model.BTCTestnet3, "", "", "", "")
			ctx := context.Background()
			_, err := b.GetBalance(ctx)
			if !errors.Is(err, btc.ErrDryRun) {
				t.Errorf("unexpected err %+v", err)
				t.Skip()
			}
			if err.Error() != c.fullcommand {
				t.Errorf("got %+v but want %+v", err.Error(), c.fullcommand)
			}
		})
	}
}

func TestBitcoinCLI_GetTransaction_DryRun(t *testing.T) {
	btc.DryRun(true)

	t.Parallel()
	cases := []struct {
		name        string
		txid        string
		fullcommand string
	}{
		{"normal", txid1, fmt.Sprintf("%s -chain=test gettransaction %s", path1, txid1)},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			b := btc.NewBitcoinCLI(path1, model.BTCTestnet3, "", "", "", "")
			ctx := context.Background()
			bT, _ := hex.DecodeString(c.txid)
			_, err := b.GetTransaction(ctx, bT)
			if !errors.Is(err, btc.ErrDryRun) {
				t.Errorf("unexpected err %+v", err)
				t.Skip()
			}
			if err.Error() != c.fullcommand {
				t.Errorf("got %+v but want %+v", err.Error(), c.fullcommand)
			}
		})
	}
}

func TestBitcoinCLI_ParseTransactionReceived(t *testing.T) {
	// TODO
}

func TestBitcoinCLI_CreateRawTransactionForAnchor(t *testing.T) {
	btc.DryRun(true)
	// TODO
}

func TestBitcoinCLI_SignRawTransactionWithWallet(t *testing.T) {
	btc.DryRun(true)
	// TODO
}

func TestBitcoinCLI_ParseSignRawTransactionWithWallet(t *testing.T) {
	// TODO
}

func TestBitcoinCLI_SendRawTransaction(t *testing.T) {
	// TODO
}

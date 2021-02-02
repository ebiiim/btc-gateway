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

	txid1        = "57511f74c3836c0d4d62a6183fa54e600372e1aed5b5be2f78ef5b766a314a5d"
	recvAddr1    = "tb1qhexc7d0fzex7lrzw3l0j2dmvhgegt02ckfdzjr"
	recvAmount1  = "0.01158624"
	opRet1       = "7468697320697320612070656e0a" // "this is a pen"
	rawTx1       = "020000000135658cd01fe92e0b81240d7a3157e2ef87389d92dcf783e170b8003cd3e9acc70000000000ffffffff02e0ad110000000000160014be4d8f35e9164def8c4e8fdf25376cba3285bd580000000000000000106a0e7468697320697320612070656e0a00000000"
	signedRawTx1 = "0200000000010135658cd01fe92e0b81240d7a3157e2ef87389d92dcf783e170b8003cd3e9acc70000000000ffffffff02e0ad110000000000160014be4d8f35e9164def8c4e8fdf25376cba3285bd580000000000000000106a0e7468697320697320612070656e0a0247304402207081f817c5cfe5579c44b770ce13fe8b4aff04a241a666e2ad8a6cdf2f88286e02202176b0ae03924adb869b4c17ae3ef1bee12ed0a0798e7673bfeeeb290d954eb501210201f52ea462e04534e2e5f9be72a4bddd6e5fe7a001bc8bdba8a8dad392222d5300000000"
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

	t.Parallel()
	cases := []struct {
		name        string
		txid        string
		bal         string
		toAddr      string
		fee         uint
		data        string
		recvAmo     string
		fullcommand string
	}{
		{"normal", txid1, "0.01168624", recvAddr1, 10000, opRet1, recvAmount1, fmt.Sprintf(`%s -chain=test createrawtransaction [{"txid": "%s", "vout": 0}] [{"%s": %s}, {"data": "%s"}]`, path1, txid1, recvAddr1, recvAmount1, opRet1)},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			b := btc.NewBitcoinCLI(path1, model.BTCTestnet3, "", "", "", "")
			ctx := context.Background()
			bT, _ := hex.DecodeString(c.txid)
			bOpRet, _ := hex.DecodeString(c.data)
			_, err := b.CreateRawTransactionForAnchor(ctx, bT, c.bal, c.toAddr, c.fee, bOpRet)
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

func TestBitcoinCLI_SignRawTransactionWithWallet(t *testing.T) {
	btc.DryRun(true)

	t.Parallel()
	cases := []struct {
		name        string
		rawTx       string
		fullcommand string
	}{
		{"normal", rawTx1, fmt.Sprintf("%s -chain=test signrawtransactionwithwallet %s", path1, rawTx1)},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			b := btc.NewBitcoinCLI(path1, model.BTCTestnet3, "", "", "", "")
			ctx := context.Background()
			bRawTx, _ := hex.DecodeString(c.rawTx)
			_, err := b.SignRawTransactionWithWallet(ctx, bRawTx)
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

func TestBitcoinCLI_ParseSignRawTransactionWithWallet(t *testing.T) {
	// TODO
}

func TestBitcoinCLI_SendRawTransaction(t *testing.T) {
	btc.DryRun(true)

	t.Parallel()
	cases := []struct {
		name        string
		signedRawTx string
		fullcommand string
	}{
		{"normal", signedRawTx1, fmt.Sprintf("%s -chain=test sendrawtransaction %s", path1, signedRawTx1)},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			b := btc.NewBitcoinCLI(path1, model.BTCTestnet3, "", "", "", "")
			ctx := context.Background()
			bSignedRawTx, _ := hex.DecodeString(c.signedRawTx)
			_, err := b.SendRawTransaction(ctx, bSignedRawTx)
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

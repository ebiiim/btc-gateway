// +build localTest

package btc_test

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/ebiiim/btc-gateway/btc"
	"github.com/ebiiim/btc-gateway/model"
	"github.com/ebiiim/btc-gateway/util"
)

var (
	lPath = util.GetEnvOr("BITCOIN_CLI_PATH", "../bitcoin-cli")
	lNet  = model.BTCNet(uint8(util.MustAtoi(util.GetEnvOr("BITCOIN_NETWORK", "3")))) // model.BTCTestnet3
	lAddr = util.GetEnvOr("BITCOIND_ADDR", "")
	lPort = util.GetEnvOr("BITCOIND_PORT", "")
	lUser = util.GetEnvOr("BITCOIND_RPC_USER", "")
	lPW   = util.GetEnvOr("BITCOIND_RPC_PASSWORD", "")
)

func prepCLI(t *testing.T) (*btc.BitcoinCLI, context.Context, context.CancelFunc) {
	t.Helper()
	b := btc.NewBitcoinCLI(lPath, lNet, lAddr, lPort, lUser, lPW)
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

const (
	lTxid1     = "c7ace9d33c00b870e183f7dc929d3887efe257317a0d24810b2ee91fd08c6535"
	lTxid2     = txid1
	lRecvAddr1 = recvAddr1
	lUnspent1  = "0.01168624"
	lFee1      = 10000
	lOpRet1    = opRet1
	lRawTx1    = rawTx1
	lSignedTx1 = signedRawTx1
	lDecRawTx1 = decRawTx1
	lGetTx1    = getTx1
)

func Test_Local_BitcoinCLI_CreateRawTransactionForAnchor(t *testing.T) {
	const (
		txid     = lTxid1
		recvAddr = lRecvAddr1
		unspent  = lUnspent1
		fee      = lFee1
		opRet    = lOpRet1
		want     = lRawTx1
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

func Test_Local_BitcoinCLI_SignRawTransactionWithWallet_Error_FailedToSign(t *testing.T) {
	const (
		rawTx = lRawTx1
	)
	b, ctx, cf := prepCLI(t)
	defer cf()
	buf, err := b.SignRawTransactionWithWallet(ctx, util.MustDecodeHexString(rawTx))
	if err != nil {
		t.Error(err)
		t.Skip()
	}
	if _, err := btc.ParseSignRawTransactionWithWallet(buf); !errors.Is(err, btc.ErrFailedToSign) {
		t.Errorf("want %+v but got %+v", btc.ErrFailedToSign, err)
	}
}

func Test_Local_BitcoinCLI_SendRawTransaction_Error_AlreadyExistsORAlreadySpent(t *testing.T) {
	const (
		signedTx = lSignedTx1
	)
	b, ctx, cf := prepCLI(t)
	defer cf()
	_, err := b.SendRawTransaction(ctx, util.MustDecodeHexString(signedTx))
	if !(errors.Is(err, btc.ErrTxAlreadyExists) || errors.Is(err, btc.ErrTxAlreadySpent)) {
		t.Errorf("want %+v but got %+v", btc.ErrTxAlreadyExists, err)
	}
}

func Test_Local_BitcoinCLI_DecodeRawTransaction(t *testing.T) {
	const (
		signedTx = lSignedTx1
		decRawTx = lDecRawTx1
	)
	b, ctx, cf := prepCLI(t)
	defer cf()
	decoded, err := b.DecodeRawTransaction(ctx, util.MustDecodeHexString(signedTx))
	if err != nil {
		t.Error(err)
		t.Skip()
	}
	decoded = btc.RemoveCRLF(decoded)
	got := decoded.String()
	got = strings.ReplaceAll(got, " ", "")
	want := strings.ReplaceAll(decRawTx, " ", "")
	if !reflect.DeepEqual(got, want) {
		t.Errorf("want %+v but got %+v", want, got)
	}
}

func Test_Local_BitcoinCLI_GetTransaction(t *testing.T) {
	const (
		txid     = lTxid2
		recvAddr = lRecvAddr1
		wantTx   = lGetTx1
	)
	b, ctx, cf := prepCLI(t)
	defer cf()
	got, err := b.GetTransaction(ctx, util.MustDecodeHexString(txid))
	if err != nil {
		t.Error(err)
		t.Skip()
	}
	want := bytes.NewBufferString(wantTx)

	// The output changes so verify it by check specific fields.
	var gt, wt, gc, wc, gh, wh, gr, wr bytes.Buffer
	wg := io.MultiWriter(&gt, &gc, &gh, &gr)
	ww := io.MultiWriter(&wt, &wc, &wh, &wr)
	io.Copy(wg, got)
	io.Copy(ww, want)
	gotTime, err1 := btc.ParseTransactionTime(&gt)
	wantTime, err2 := btc.ParseTransactionTime(&wt)
	if err1 != nil || err2 != nil || gotTime != wantTime {
		t.Errorf("wrong time got=%+v (err=%+v) want=%+v (err=%+v)", gotTime, err1, wantTime, err2)
	}
	gotConfs, err1 := btc.ParseTransactionConfirmations(&gc)
	wantConfs, err2 := btc.ParseTransactionConfirmations(&wc)
	if err1 != nil || err2 != nil || gotConfs < wantConfs {
		t.Errorf("wrong confs got=%+v (err=%+v) want=%+v (err=%+v)", gotConfs, err1, wantConfs, err2)
	}
	gotHex, err1 := btc.ParseTransactionRawHex(&gh)
	wantHex, err2 := btc.ParseTransactionRawHex(&wh)
	if err1 != nil || err2 != nil || !reflect.DeepEqual(gotHex, wantHex) {
		t.Errorf("wrong hex got=%+v (err=%+v) want=%+v (err=%+v)", gotHex, err1, wantHex, err2)
	}
	gotAmo, err1 := btc.ParseTransactionReceived(&gr, recvAddr)
	wantAmo, err2 := btc.ParseTransactionReceived(&wr, recvAddr)
	if err1 != nil || err2 != nil || gotAmo != wantAmo {
		t.Errorf("wrong amount got=%+v (err=%+v) want=%+v (err=%+v)", gotAmo, err1, wantAmo, err2)
	}
}

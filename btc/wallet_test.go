package btc_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/ebiiim/btcgw/btc"
	"github.com/ebiiim/btcgw/util"

	_ "gocloud.dev/docstore/memdocstore"
)

//
// Do not parallelize DocstoreWallet tests because memdocstore is NOT thread-safe.
//

// Assumes tests will be run from package root.
var (
	testdb1 = "testdata/wallet1.db"
	conn1   = "mem://wallet_test/addr?filename=" + testdb1
	conn2   = "mem://wallet_test/addr"
)

var (
	waddr1 = "walletaddr0001"
	wtx1   = util.MustDecodeHexString("123456")
	wtx2   = util.MustDecodeHexString("654321")
	wtx3   = util.MustDecodeHexString("0123456789abcdef0123456789abcdef")
)

func TestDocstoreWallet_UTXOs(t *testing.T) {
	utxos1 := btc.MustNewDocstoreWallet(conn2, waddr1)

	// peek and dequeue from [] -> [] err
	txid, addr, err := utxos1.PeekNextUTXO()
	if txid != nil || addr != "" || !errors.Is(err, btc.ErrCouldNotGetNextUTXO) {
		t.Errorf("want empty but got txid=%v, addr=%v", txid, addr)
		t.Skip()
	}
	txid, addr, err = utxos1.NextUTXO()
	if txid != nil || addr != "" || !errors.Is(err, btc.ErrCouldNotGetNextUTXO) {
		t.Errorf("want empty but got txid=%v, addr=%v", txid, addr)
		t.Skip()
	}
	// enqueue utxo1 to [] -> [utxo1]
	if err := utxos1.AddUTXO(wtx1, waddr1); err != nil {
		t.Error(err)
		t.Skip()
	}
	// peek & dequeue from [utxo1] -> [] utxo1
	txid, addr, err = utxos1.PeekNextUTXO()
	if bytes.Compare(txid, wtx1) != 0 || addr != waddr1 || err != nil {
		t.Error("1")
		t.Skip()
	}
	txid, addr, err = utxos1.NextUTXO()
	if bytes.Compare(txid, wtx1) != 0 || addr != waddr1 || err != nil {
		t.Error("1")
		t.Skip()
	}
	// dequeue from [] -> [] err
	txid, addr, err = utxos1.NextUTXO()
	if txid != nil || addr != "" || !errors.Is(err, btc.ErrCouldNotGetNextUTXO) {
		t.Errorf("want empty but got txid=%v, addr=%v", txid, addr)
		t.Skip()
	}
	// enqueue utxo1 to [] -> [utxo1]
	// enqueue utxo2 to [] -> [utxo1, utxo2]
	// enqueue utxo3 to [] -> [utxo1, utxo2, utxo3]
	err1 := utxos1.AddUTXO(wtx1, waddr1)
	err2 := utxos1.AddUTXO(wtx2, waddr1)
	err3 := utxos1.AddUTXO(wtx3, waddr1)
	if err1 != nil || err2 != nil || err3 != nil {
		t.Error("2")
		t.Skip()
	}
	// dequeue from [utxo1, utxo2, utxo3] -> [utxo2, utxo3] utxo1
	txid, addr, err = utxos1.NextUTXO()
	if bytes.Compare(txid, wtx1) != 0 || addr != waddr1 || err != nil {
		t.Error("3")
		t.Skip()
	}
	// dequeue from [utxo2, utxo3] -> [utxo3] utxo2
	txid, addr, err = utxos1.NextUTXO()
	if bytes.Compare(txid, wtx2) != 0 || addr != waddr1 || err != nil {
		t.Error("4")
		t.Skip()
	}
	// dequeue from [utxo3] -> [] utxo3
	txid, addr, err = utxos1.NextUTXO()
	if bytes.Compare(txid, wtx3) != 0 || addr != waddr1 || err != nil {
		t.Error("5")
		t.Skip()
	}
	if err := utxos1.Close(); err != nil {
		t.Error(err)
		t.Skip()
	}
}

func TestNewDocstoreWallet(t *testing.T) {
	utxos1 := btc.MustNewDocstoreWallet(conn1, waddr1)
	// dequeue from [utxo1, utxo2, utxo3] -> [utxo2, utxo3] utxo1
	// dequeue from [utxo2, utxo3] -> [utxo3] utxo2
	// dequeue from [utxo3] -> [] utxo3
	// dequeue from [] -> [] err
	// enqueue utxo1 to [] -> [utxo1]
	// enqueue utxo2 to [utxo1] -> [utxo1, utxo2]
	// enqueue utxo3 to [utxo1, utxo2] -> [utxo1, utxo2, utxo3]
	txid, addr, err := utxos1.NextUTXO()
	if bytes.Compare(txid, wtx1) != 0 || addr != waddr1 || err != nil {
		t.Errorf("1 txid=%v, addr=%v, err=%v", txid, addr, err)
		t.Skip()
	}
	txid, addr, err = utxos1.NextUTXO()
	if bytes.Compare(txid, wtx2) != 0 || addr != waddr1 || err != nil {
		t.Error("2")
		t.Skip()
	}
	txid, addr, err = utxos1.NextUTXO()
	if bytes.Compare(txid, wtx3) != 0 || addr != waddr1 || err != nil {
		t.Error("3")
		t.Skip()
	}
	txid, addr, err = utxos1.NextUTXO()
	if txid != nil || addr != "" || !errors.Is(err, btc.ErrCouldNotGetNextUTXO) {
		t.Errorf("want empty but got txid=%v, addr=%v", txid, addr)
		t.Skip()
	}
	err1 := utxos1.AddUTXO(wtx1, waddr1)
	err2 := utxos1.AddUTXO(wtx2, waddr1)
	err3 := utxos1.AddUTXO(wtx3, waddr1)
	if err1 != nil || err2 != nil || err3 != nil {
		t.Error("4")
		t.Skip()
	}
	utxos2 := btc.MustNewDocstoreWallet(conn1, "XXX")
	if txid, addr, err := utxos2.NextUTXO(); txid != nil || addr != "" || !errors.Is(err, btc.ErrCouldNotGetNextUTXO) {
		t.Errorf("want empty but got txid=%v, addr=%v", txid, addr)
		t.Skip()
	}
	// No need to close utxos2 because it's same with utxos1.
	if err := utxos1.Close(); err != nil {
		t.Error(err)
		t.Skip()
	}
}

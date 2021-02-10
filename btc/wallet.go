package btc

import (
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"gocloud.dev/docstore"
)

// TODO: we can manage UTXO statelessly by using `bitcoin-cli listunspent`.

func init() {
	gob.Register(map[string]interface{}{})
	gob.Register([]interface{}{})
}

// Errors
var (
	ErrCouldNotOpenWalletStore  = errors.New("ErrCouldNotOpenWalletStore")
	ErrCouldNotCloseWalletStore = errors.New("ErrCouldNotCloseWalletStore")
	ErrCouldNotLoadWallet       = errors.New("ErrCouldNotLoadWallet")
	ErrCouldNotSaveWallet       = errors.New("ErrCouldNotSaveWallet")
	ErrCouldNotGetNextUTXO      = errors.New("ErrCouldNotGetNextUTXO")
	ErrCouldNotAddUTXO          = errors.New("ErrCouldNotAddUTXO")
)

type Wallet interface {
	NextUTXO() (txid []byte, addr string, err error)
	AddUTXO(txid []byte, addr string) error
	io.Closer
}

var _ Wallet = (*DocstoreWallet)(nil)

type utxo struct {
	TXID []byte
	Addr string
}

type utxosDoc struct {
	Addr  string `docstore:"addr"`
	UTXOs []utxo `docstore:"utxos"`
}

type DocstoreWallet struct {
	q    []utxo
	addr string

	conn string
	coll *docstore.Collection
}

func (w *DocstoreWallet) enqueue(v utxo) {
	w.q = append(w.q, v)
}

func (w *DocstoreWallet) dequeue() (utxo, error) {
	if len(w.q) == 0 {
		return utxo{}, errors.New("empty queue")
	}
	v := w.q[0]
	w.q = w.q[1:]
	return v, nil
}

func (w *DocstoreWallet) load() error {
	doc := &utxosDoc{
		Addr: w.addr,
	}
	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()
	if err := w.coll.Get(ctx, doc); err != nil {
		return fmt.Errorf("%w (%v)", ErrCouldNotLoadWallet, err)
	}
	w.q = doc.UTXOs
	return nil
}

func (w *DocstoreWallet) save() error {
	doc := &utxosDoc{
		Addr: w.addr,
		UTXOs:  w.q,
	}
	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()
	if err := w.coll.Put(ctx, doc); err != nil {
		return fmt.Errorf("%w (%v)", ErrCouldNotSaveWallet, err)
	}
	return nil
}

func (w *DocstoreWallet) open() error {
	coll, err := docstore.OpenCollection(context.Background(), w.conn)
	if err != nil {
		return fmt.Errorf("%w (%v)", ErrCouldNotOpenWalletStore, err)
	}
	w.coll = coll
	return nil
}

// MustNewDocstoreWallet initializes a DocstoreWallet,
// panics if failed to access datastore.
//
// Parameters:
//   - conn sets connection string.
//   - addr sets document name (uses Bitcoin addresses).
//
// DocstoreWallet puts all UTXOs in a single document specified by name.
// TODO: put all UTXOs on every save is too heavy...
func MustNewDocstoreWallet(conn string, addr string) *DocstoreWallet {
	w := &DocstoreWallet{
		q:      nil,
		addr: addr,
		conn:   conn,
		coll:   nil,
	}
	if err := w.open(); err != nil {
		panic(fmt.Sprintf("%v conn=%s", err, conn))
	}
	if err := w.load(); err != nil {
		log.Printf("MustNewDocstoreWallet: addr=%s not found, create a new doc\n", addr)
	}
	return w
}

func (w *DocstoreWallet) Close() error {
	if err := w.save(); err != nil {
		return fmt.Errorf("%w (%v)", ErrCouldNotCloseWalletStore, err)
	}
	if err := w.coll.Close(); err != nil {
		return fmt.Errorf("%w (%v)", ErrCouldNotCloseWalletStore, err)
	}
	return nil
}

func (w *DocstoreWallet) NextUTXO() (txid []byte, addr string, err error) {
	v, err := w.dequeue()
	if err != nil {
		return nil, "", fmt.Errorf("%w, %v", ErrCouldNotGetNextUTXO, err)
	}
	// Save on every NextUTXO call.
	if err := w.save(); err != nil {
		return nil, "", fmt.Errorf("%w, %v", ErrCouldNotGetNextUTXO, err)
	}
	return v.TXID, v.Addr, nil
}

func (w *DocstoreWallet) AddUTXO(txid []byte, addr string) error {
	w.enqueue(utxo{
		TXID: txid,
		Addr: addr,
	})
	// Save on every AddUTXO call.
	if err := w.save(); err != nil {
		return fmt.Errorf("%w, %v", ErrCouldNotAddUTXO, err)
	}
	return nil
}

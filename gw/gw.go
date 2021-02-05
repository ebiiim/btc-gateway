package gw

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ebiiim/btc-gateway/btc"
	"github.com/ebiiim/btc-gateway/model"
	"github.com/ebiiim/btc-gateway/store"
)

// Gateway provides features to register and verify BBc-1 transactions.
// This interface does not handle datastore closes.
type Gateway interface {
	// RegisterTransaction inserts an anchor into Bitcoin block chain
	// by sending a transaction, and returns its Bitcoin transaction ID.
	RegisterTransaction(ctx context.Context, domID, txID []byte) (btcTXID []byte, err error)
	// StoreTransaction retrieve a Bitcoin transaction,
	// and saves its AnchorRecord embedded in OP_RETURN in the datastore.
	StoreTransaction(ctx context.Context, btcTXID []byte) error
	// GetTransaction gets an AnchorRecord
	// specified by the given information from the datastore.
	GetTransaction(ctx context.Context, domID, txID []byte) (*model.AnchorRecord, error)
}

// Errors
var (
	ErrCouldNotPutAnchor        = errors.New("ErrCouldNotPutAnchor")
	ErrCouldNotStoreTransaction = errors.New("ErrCouldNotStoreTransaction")
	ErrCouldNotGetTransaction   = errors.New("ErrCouldNotGetTransaction")
)

type GatewayApp struct {
	BTCNet model.BTCNet
	BTC    btc.BTC
	Store  store.Store
}

// NewGatewayApp initializes a GatewayApp.
//
// Parameters:
//   - bn sets Bitcoin network to anchor.
//   - b sets btc.BTC.
//   - s sets store.Store.
//
// bn must be same as b.BTCNet.
func NewGatewayApp(bn model.BTCNet, b btc.BTC, s store.Store) *GatewayApp {
	g := &GatewayApp{
		BTCNet: bn,
		BTC:    b,
		Store:  s,
	}
	return g
}

var timeNow = time.Now

func (g *GatewayApp) RegisterTransaction(ctx context.Context, domID, txID []byte) (btcTXID []byte, err error) {
	a := model.NewAnchor(g.BTCNet, timeNow(), domID, txID)
	txid, err := g.BTC.PutAnchor(ctx, a)
	if err != nil {
		return nil, fmt.Errorf("%w (%v)", ErrCouldNotPutAnchor, err)
	}
	return txid, err
}

func (g *GatewayApp) StoreTransaction(ctx context.Context, btcTXID []byte) error {
	ar, err := g.BTC.GetAnchor(ctx, btcTXID)
	if err != nil {
		return fmt.Errorf("%w (%v)", ErrCouldNotStoreTransaction, err)
	}
	if err := g.Store.Put(ctx, ar); err != nil {
		return fmt.Errorf("%w (%v)", ErrCouldNotStoreTransaction, err)
	}
	return nil
}

func (g *GatewayApp) GetTransaction(ctx context.Context, domID, txID []byte) (*model.AnchorRecord, error) {
	ar, err := g.Store.Get(ctx, domID, txID)
	if err != nil {
		return nil, fmt.Errorf("%w (%v)", ErrCouldNotGetTransaction, err)
	}
	return ar, nil
}

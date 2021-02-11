/*
Package gw provides ability to anchor BBc-1 transactions to the Bitcoin block chain,
that hides the implementation of BTC (package btc)
and Store (package store) from applications.
*/
package gw

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/ebiiim/btcgw/btc"
	"github.com/ebiiim/btcgw/model"
	"github.com/ebiiim/btcgw/store"
)

// Gateway provides features to register and verify BBc-1 transactions.
type Gateway interface {
	// RegisterTransaction inserts an anchor into Bitcoin block chain
	// by sending a transaction, and returns its Bitcoin transaction ID.
	RegisterTransaction(ctx context.Context, domID, txID []byte) (btcTXID []byte, err error)

	// StoreRecord retrieves a Bitcoin transaction,
	// and saves its AnchorRecord embedded in OP_RETURN in the datastore.
	StoreRecord(ctx context.Context, btcTXID []byte) error

	// GetRecord gets an AnchorRecord
	// specified by the given information from the datastore.
	GetRecord(ctx context.Context, domID, txID []byte) (*model.AnchorRecord, error)

	// RefreshRecord update AnchorRecord specified by domID and txID.
	// Get it from datastore, update AnchorRecord.Confirmations by
	// retrieving and checking the Bitcoin transaction, and then put it into datastore.
	// In addition, changes AnchorRecord.BBc1DomainName or AnchorRecord.Note or both, if the given value is not nil.
	RefreshRecord(ctx context.Context, domID, txID []byte, pBBc1domName, pNote *string) error

	io.Closer
}

var _ Gateway = (*GatewayImpl)(nil)

// Errors
// TODO: Handle more error types: e.g. ErrInvalidFee
// TODO: Provide more error types. e.g. ErrNoConfirmedUTXO
var (
	ErrCouldNotPutAnchor     = errors.New("ErrCouldNotPutAnchor")
	ErrCouldNotStoreRecord   = errors.New("ErrCouldNotStoreRecord")
	ErrCouldNotGetRecord     = errors.New("ErrCouldNotGetRecord")
	ErrCouldNotRefreshRecord = errors.New("ErrCouldNotRefreshRecord")
	ErrCouldNotCloseStore    = errors.New("ErrCouldNotCloseStore")
)

type GatewayImpl struct {
	BTCNet model.BTCNet
	BTC    btc.BTC
	Wallet btc.Wallet
	Store  store.Store

	xBTCImpl *btc.BitcoinCLI
}

// NewGatewayImpl initializes a GatewayImpl.
//
// Parameters:
//   - bn sets Bitcoin network to anchor.
//   - b sets btc.BTC.
//   - w sets btc.Wallet.
//   - s sets store.Store.
//
// bn must be same as b.BTCNet.
// b must be *btc.BitcoinCLI for now.
func NewGatewayImpl(bn model.BTCNet, b btc.BTC, w btc.Wallet, s store.Store) *GatewayImpl {
	bImpl, ok := b.(*btc.BitcoinCLI)
	if !ok {
		panic("NewGatewayImpl: b must be *btc.BitcoinCLI for now")
	}
	g := &GatewayImpl{
		BTCNet:   bn,
		BTC:      b,
		Wallet:   w,
		Store:    s,
		xBTCImpl: bImpl,
	}
	return g
}

var timeNow = time.Now

func (g *GatewayImpl) RegisterTransaction(ctx context.Context, domID, txID []byte) (btcTXID []byte, err error) {
	a := model.NewAnchor(g.BTCNet, timeNow(), domID, txID)
	// Set UTXO if Wallet is set.
	if g.Wallet != nil {
		tx, addr, err := g.Wallet.PeekNextUTXO()
		if err != nil {
			return nil, fmt.Errorf("%w (%v)", ErrCouldNotPutAnchor, err)
		}
		g.xBTCImpl.XSetUTXO(tx, addr)
	}
	txid, err := g.BTC.PutAnchor(ctx, a)
	if err != nil {
		return nil, fmt.Errorf("%w (%v)", ErrCouldNotPutAnchor, err)
	}
	// Update Wallet if it is set.
	if g.Wallet != nil {
		if _, _, err := g.Wallet.NextUTXO(); err != nil {
			return nil, fmt.Errorf("%w (%v)", ErrCouldNotPutAnchor, err)
		}
		if err := g.Wallet.AddUTXO(g.xBTCImpl.XGetUTXO()); err != nil {
			return nil, fmt.Errorf("%w (%v)", ErrCouldNotPutAnchor, err)
		}
	}
	return txid, err
}

func (g *GatewayImpl) StoreRecord(ctx context.Context, btcTXID []byte) error {
	ar, err := g.BTC.GetAnchor(ctx, btcTXID)
	if err != nil {
		return fmt.Errorf("%w (%v)", ErrCouldNotStoreRecord, err)
	}
	if err := g.Store.Put(ctx, ar); err != nil {
		return fmt.Errorf("%w (%v)", ErrCouldNotStoreRecord, err)
	}
	return nil
}

func (g *GatewayImpl) GetRecord(ctx context.Context, domID, txID []byte) (*model.AnchorRecord, error) {
	ar, err := g.Store.Get(ctx, domID, txID)
	if err != nil {
		return nil, fmt.Errorf("%w (%v)", ErrCouldNotGetRecord, err)
	}
	return ar, nil
}

func (g *GatewayImpl) RefreshRecord(ctx context.Context, domID, txID []byte, pBBc1domName, pNote *string) error {
	oldAR, err := g.GetRecord(ctx, domID, txID)
	if err != nil {
		return fmt.Errorf("%w (%v)", ErrCouldNotRefreshRecord, err)
	}
	newAR, err := g.BTC.GetAnchor(ctx, oldAR.BTCTransactionID)
	if err != nil {
		return fmt.Errorf("%w (%v)", ErrCouldNotRefreshRecord, err)
	}
	if err := g.Store.UpdateConfirmations(ctx, domID, txID, newAR.Confirmations); err != nil {
		return fmt.Errorf("%w (%v)", ErrCouldNotRefreshRecord, err)
	}
	if pBBc1domName != nil {
		if err := g.Store.UpdateBBc1DomainName(ctx, domID, txID, *pBBc1domName); err != nil {
			return fmt.Errorf("%w (%v)", ErrCouldNotRefreshRecord, err)
		}
	}
	if pNote != nil {
		if err := g.Store.UpdateNote(ctx, domID, txID, *pNote); err != nil {
			return fmt.Errorf("%w (%v)", ErrCouldNotRefreshRecord, err)
		}
	}
	return nil
}

// Close closes g.Store.
func (g *GatewayImpl) Close() error {
	err := g.Store.Close()
	if err != nil {
		return fmt.Errorf("%w (%v)", ErrCouldNotCloseStore, err)
	}
	return nil
}

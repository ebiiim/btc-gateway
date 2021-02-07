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
}

var _ Gateway = (*GatewayImpl)(nil)

// Errors
// TODO: Provide more error types.
var (
	ErrCouldNotPutAnchor     = errors.New("ErrCouldNotPutAnchor")
	ErrCouldNotStoreRecord   = errors.New("ErrCouldNotStoreRecord")
	ErrCouldNotGetRecord     = errors.New("ErrCouldNotGetRecord")
	ErrCouldNotRefreshRecord = errors.New("ErrCouldNotRefreshRecord")
)

type GatewayImpl struct {
	BTCNet model.BTCNet
	BTC    btc.BTC
	Store  store.Store
}

// NewGatewayImpl initializes a GatewayImpl.
//
// Parameters:
//   - bn sets Bitcoin network to anchor.
//   - b sets btc.BTC.
//   - s sets store.Store.
//
// bn must be same as b.BTCNet.
func NewGatewayImpl(bn model.BTCNet, b btc.BTC, s store.Store) *GatewayImpl {
	g := &GatewayImpl{
		BTCNet: bn,
		BTC:    b,
		Store:  s,
	}
	return g
}

var timeNow = time.Now

func (g *GatewayImpl) RegisterTransaction(ctx context.Context, domID, txID []byte) (btcTXID []byte, err error) {
	a := model.NewAnchor(g.BTCNet, timeNow(), domID, txID)
	txid, err := g.BTC.PutAnchor(ctx, a)
	if err != nil {
		return nil, fmt.Errorf("%w (%v)", ErrCouldNotPutAnchor, err)
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

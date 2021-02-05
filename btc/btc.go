package btc

import (
	"context"

	"github.com/ebiiim/btc-gateway/model"
)

// BTC provides features to send and get anchor.
// This interface is not responsible for which Bitcoin wallet is used.
type BTC interface {
	// PutAnchor anchors the given Anchor by sending a Bitcoin transaction and returns its ID.
	PutAnchor(ctx context.Context, a *model.Anchor) ([]byte, error)
	// GetAnchor returns an AnchorRecord by searching the given Bitcoin transaction ID and parsing its data.
	GetAnchor(ctx context.Context, btctx []byte) (*model.AnchorRecord, error)
}

var _ BTC = (*BitcoinCLI)(nil)

package store

import (
	"context"

	"github.com/ebiiim/btc-gateway/model"
)

// Store provides features to store anchor data in a datastore.
// Anchor data (especially the Bitcoin transaction IDs) should be stored,
// as finding an Anchor needs walking through all Bitcoin blockchains and is time consuming.
type Store interface {
	// Put adds or replaces an AnchorRecord in O(1) time.
	Put(ctx context.Context, r *model.AnchorRecord) error
	// Get returns the AnchorRecord specified by bbc1dom and bbc1tx in O(1) time.
	Get(ctx context.Context, bbc1dom, bbc1tx []byte) (*model.AnchorRecord, error)
}

var _ Store = (*Docstore)(nil)

/*
Package btc provides ability to send and get anchors,
that brings an abstraction layer between the Gateway (package gw) and Bitcoin implementations.
*/
package btc

import (
	"context"
	"io"

	"github.com/ebiiim/btcgw/model"
)

// BTC provides features to send and get anchors.
// This interface is not responsible for which Bitcoin wallet is used.
type BTC interface {
	// PutAnchor anchors the given Anchor by sending a Bitcoin transaction and returns its ID.
	PutAnchor(ctx context.Context, a *model.Anchor) ([]byte, error)
	// GetAnchor returns an AnchorRecord by searching the given Bitcoin transaction ID and parsing its data.
	GetAnchor(ctx context.Context, btctx []byte) (*model.AnchorRecord, error)

	io.Closer
}

var _ BTC = (*BitcoinCLI)(nil)

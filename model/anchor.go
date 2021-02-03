package model

import "time"

// BTCNet represents a Bitcoin network.
type BTCNet uint8

// Bitcoin networks.
const (
	BTCTestnet3 = 3
	BTCTestnet4 = 4
	BTCMainnet  = 255
)

// String returns the name of n.
func (n BTCNet) String() string {
	switch n {
	default:
		return ""
	case BTCMainnet:
		return "Mainnet"
	case BTCTestnet3:
		return "Testnet3"
	case BTCTestnet4:
		return "Testnet4"
	}
}

// Anchor contains an anchor that can be encoded to OP_RETURN.
type Anchor struct {
	Version           uint8
	BTCNet            BTCNet
	Timestamp         time.Time
	BBc1DomainID      [32]byte
	BBc1TransactionID [32]byte
}

// validAnchorVersions contains valid anchor versions.
var validAnchorVersions map[uint8]struct{} = map[uint8]struct{}{
	1: {},
}

// anchorVersion specifies the version to be embeded by NewAnchor.
var anchorVersion uint8 = 1

// NewAnchor initializes an Anchor.
//
// Parameters:
//   - btcnet sets target Bitcoin network.
//   - timestamp sets time stamp.
//   - bbc1dom sets BBc-1 Domain ID.
//   - bbc1tx sets BBc-1 Transaction ID.
//
// Anchor.BBc1DomainID and Anchor.BBc1TransactionID are fixed at 32 bytes.
// If the given []byte is shorter than 32bytes, padding with 0.
// If the given []byte is longer than 32bytes, only use the first 32 bytes.
func NewAnchor(btcnet BTCNet, timestamp time.Time, bbc1dom, bbc1tx []byte) *Anchor {
	// Copy the first up to 32 bytes from bbc1dom and bbc1tx.
	var d, t [32]byte
	dlen := len(bbc1dom)
	if dlen > 32 {
		dlen = 32
	}
	tlen := len(bbc1tx)
	if tlen > 32 {
		tlen = 32
	}
	copy(d[:dlen], bbc1dom[:dlen])
	copy(t[:tlen], bbc1tx[:tlen])

	a := &Anchor{
		Version:           anchorVersion,
		BTCNet:            btcnet,
		Timestamp:         timestamp,
		BBc1DomainID:      d,
		BBc1TransactionID: t,
	}
	return a
}

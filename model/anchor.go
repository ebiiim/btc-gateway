package model

type BTCNet uint8

const (
	BTCTestnet3 = 3
	BTCTestnet4 = 4
	BTCMainnet  = 255
)

var BTCNetNames map[BTCNet]string = map[BTCNet]string{
	BTCMainnet:  "Mainnet",
	BTCTestnet3: "Testnet3",
	BTCTestnet4: "Testnet4",
}

type Anchor struct {
	Version           uint8
	BTCNet            BTCNet
	BBc1DomainID      [32]byte
	BBc1TransactionID [32]byte
}

var validAnchorVersions map[uint8]struct{} = map[uint8]struct{}{
	1: {},
}

var anchorVersion uint8 = 1

func NewAnchor(btcnet BTCNet, bbc1dom, bbc1tx []byte) *Anchor {
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
		BBc1DomainID:      d,
		BBc1TransactionID: t,
	}
	return a
}

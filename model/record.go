package model

import "time"

// AnchorRecord contains an Anchor and a Bitcoin transaction ID in which the Anchor is embedded.
// Some additional information from the Bitcoin transaction are also included.
// Some optional information NOT from the Bitcoin transaction can be added.
type AnchorRecord struct {
	Anchor           *Anchor
	BTCTransactionID []byte

	// Data from the Bitcoin transaction.
	Timestamp     time.Time
	Confirmations uint
	BTCAddr       string

	// Optional data NOT included in Bitcoin.
	BBc1DomainName string
	Note           string
}

// NewAnchorRecord initializes an AnchorRecord.
//
// Parameters:
//   - anchor sets the Anchor
//   - btctx sets the Bitcoin transaction ID in which anchor is embedded.
//   - ts sets the time in Bitcoin transaction.
//   - conf sets the number of confirmations.
//   - btcaddr sets the sender's Bitcoin address.
//   - bbc1domName sets the BBc-1 domain name.
//   - note sets a string for note.
//
// bbc1domName and note are not included in Bitcoin blockchain.
// They cannot be restored when the datastore is lost.
func NewAnchorRecord(anchor *Anchor, btctx []byte, ts time.Time, conf uint, btcaddr string, bbc1domName, note string) *AnchorRecord {
	r := &AnchorRecord{
		Anchor:           anchor,
		BTCTransactionID: btctx,
		Timestamp:        ts,
		Confirmations:    conf,
		BTCAddr:          btcaddr,
		BBc1DomainName:   bbc1domName,
		Note:             note,
	}
	return r
}

package model

import "time"

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

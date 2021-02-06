package model

import (
	"fmt"
	"time"
)

// AnchorRecord contains an Anchor and a Bitcoin transaction ID in which the Anchor is embedded.
// Some additional information from the Bitcoin transaction are also included.
// Some optional information NOT from the Bitcoin transaction can be added.
type AnchorRecord struct {
	Anchor           *Anchor
	BTCTransactionID []byte

	// Data from the Bitcoin transaction.
	TransactionTime time.Time
	Confirmations   uint

	// Optional data NOT included in Bitcoin.
	BBc1DomainName string
	Note           string
}

// String returns a human-readable expression for the AnchorRecord.
func (r *AnchorRecord) String() string {
	var s string
	s += "==========AnchorRecord==========\n"
	s += r.Anchor.String()
	s += "-------------Record-------------\n"
	s += fmt.Sprintf("   BTCTransactionID: %x\n", r.BTCTransactionID)
	s += fmt.Sprintf("    TransactionTime: %d | %s | 0x%016x\n", r.TransactionTime.Unix(), r.TransactionTime, r.TransactionTime.Unix())
	s += fmt.Sprintf("      Confirmations: %d\n", r.Confirmations)
	s += "------------Optional------------\n"
	s += fmt.Sprintf("     BBc1DomainName: %s\n", r.BBc1DomainName)
	s += fmt.Sprintf("               Note: %s\n", r.Note)
	s += "================================\n"
	return s
}

// NewAnchorRecord initializes an AnchorRecord.
//
// Parameters:
//   - anchor sets the Anchor
//   - btctx sets the Bitcoin transaction ID in which anchor is embedded.
//   - ts sets the time in Bitcoin transaction.
//   - conf sets the number of confirmations.
//   - bbc1domName sets the BBc-1 domain name.
//   - note sets a string for note.
//
// bbc1domName and note are not included in Bitcoin blockchain.
// They cannot be restored when the datastore is lost.
func NewAnchorRecord(anchor *Anchor, btctx []byte, ts time.Time, conf uint, bbc1domName, note string) *AnchorRecord {
	r := &AnchorRecord{
		Anchor:           anchor,
		BTCTransactionID: btctx,
		TransactionTime:  ts,
		Confirmations:    conf,
		BBc1DomainName:   bbc1domName,
		Note:             note,
	}
	return r
}

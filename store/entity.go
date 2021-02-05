package store

import (
	"time"

	"github.com/ebiiim/btc-gateway/model"
)

// AnchorEntity contains data equalavant to AnchorRecord,
// but focuses on placing data in the data store.
// In particular, ID is a 64 bytes key field that contains
// BBc-1 domain ID in the first 32 bytes,
// followed by transaction ID.
type AnchorEntity struct {
	ID               [64]byte  `docstore:"id"`
	AnchorVersion    uint8     `docstore:"anchorver"`
	BTCNet           uint8     `docstore:"btcnet"`
	AnchorTime       time.Time `docstore:"anchortime"`
	BTCTransactionID []byte    `docstore:"btctxid"`
	TransactionTime  time.Time `dosctore:"txtime"`
	Confirmations    uint      `docstore:"confirmations"`
	BBc1DomainName   string    `docstore:"bbc1dom,omitempty"`
	Note             string    `docstore:"note,omitempty"`
}

// NewAnchorEntity initializes an AnchorEntity from the given AnchorRecord.
func NewAnchorEntity(r *model.AnchorRecord) *AnchorEntity {
	var id [64]byte
	copy(id[0:32], r.Anchor.BBc1DomainID[:])
	copy(id[32:64], r.Anchor.BBc1TransactionID[:])
	e := &AnchorEntity{
		ID:               id,
		AnchorVersion:    r.Anchor.Version,
		BTCNet:           uint8(r.Anchor.BTCNet),
		AnchorTime:       r.Anchor.Timestamp,
		BTCTransactionID: r.BTCTransactionID,
		TransactionTime:  r.TransactionTime,
		Confirmations:    r.Confirmations,
		BBc1DomainName:   r.BBc1DomainName,
		Note:             r.Note,
	}
	return e
}

// AnchorRecord returns an AnchorRecord from the AnchorEntity.
func (e *AnchorEntity) AnchorRecord() *model.AnchorRecord {
	var domid, txid [32]byte
	copy(domid[0:32], e.ID[0:32])
	copy(txid[0:32], e.ID[32:64])
	a := &model.Anchor{
		Version:           e.AnchorVersion,
		BTCNet:            model.BTCNet(e.BTCNet),
		Timestamp:         e.AnchorTime,
		BBc1DomainID:      domid,
		BBc1TransactionID: txid,
	}
	r := &model.AnchorRecord{
		Anchor:           a,
		BTCTransactionID: e.BTCTransactionID,
		TransactionTime:  e.TransactionTime,
		Confirmations:    e.Confirmations,
		BBc1DomainName:   e.BBc1DomainName,
		Note:             e.Note,
	}
	return r
}

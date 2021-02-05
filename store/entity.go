package store

import (
	"encoding/hex"
	"time"

	"github.com/ebiiim/btc-gateway/model"
)

// AnchorEntity contains data equalavant to AnchorRecord,
// but focuses on placing data in the data store.
// In particular, CID is a key field that contains
// a combined string that starts with BBc-1 domain ID,
// followed by transaction ID.
type AnchorEntity struct {
	CID               string    `docstore:"cid"`
	BBc1DomainID      []byte    `docstore:"bbc1domid"`
	BBc1TransactionID []byte    `docstore:"bbc1txid"`
	AnchorVersion     uint8     `docstore:"anchorver"`
	BTCNet            uint8     `docstore:"btcnet"`
	AnchorTime        time.Time `docstore:"anchortime"`
	BTCTransactionID  []byte    `docstore:"btctxid"`
	TransactionTime   time.Time `docstore:"txtime"`
	Confirmations     uint      `docstore:"confirmations"`
	BBc1DomainName    string    `docstore:"bbc1dom,omitempty"`
	Note              string    `docstore:"note,omitempty"`
}

// NewAnchorEntity initializes an AnchorEntity from the given AnchorRecord.
func NewAnchorEntity(r *model.AnchorRecord) *AnchorEntity {
	cid := hex.EncodeToString(r.Anchor.BBc1DomainID[:]) + hex.EncodeToString(r.Anchor.BBc1TransactionID[:])
	e := &AnchorEntity{
		CID:               cid,
		BBc1DomainID:      r.Anchor.BBc1DomainID[:],
		BBc1TransactionID: r.Anchor.BBc1TransactionID[:],
		AnchorVersion:     r.Anchor.Version,
		BTCNet:            uint8(r.Anchor.BTCNet),
		AnchorTime:        r.Anchor.Timestamp,
		BTCTransactionID:  r.BTCTransactionID,
		TransactionTime:   r.TransactionTime,
		Confirmations:     r.Confirmations,
		BBc1DomainName:    r.BBc1DomainName,
		Note:              r.Note,
	}
	return e
}

// AnchorRecord returns an AnchorRecord from the AnchorEntity.
func (e *AnchorEntity) AnchorRecord() *model.AnchorRecord {
	var did, txid [32]byte
	copy(did[:], e.BBc1DomainID[0:len(e.BBc1DomainID)])
	copy(txid[:], e.BBc1TransactionID[0:len(e.BBc1TransactionID)])
	a := &model.Anchor{
		Version:           e.AnchorVersion,
		BTCNet:            model.BTCNet(e.BTCNet),
		Timestamp:         e.AnchorTime,
		BBc1DomainID:      did,
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

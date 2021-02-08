package store

import (
	"context"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ebiiim/btc-gateway/model"

	"gocloud.dev/docstore"
)

func init() {
	gob.Register([64]uint8{})
	gob.Register(time.Time{})
	gob.Register([]interface{}{})
}

// Errors
var (
	ErrFailedToOpen   = errors.New("ErrFailedToOpen")
	ErrFailedToClose  = errors.New("ErrFailedToClose")
	ErrFailedToGet    = errors.New("ErrFailedToGet")
	ErrFailedToPut    = errors.New("ErrFailedToPut")
	ErrFailedToUpdate = errors.New("ErrFailedToUpdate")
)

type Docstore struct {
	conn string
	coll *docstore.Collection

	// Checks whether Open is called.
	// For internal use only.
	once sync.Once
}

func NewDocstore(conn string) *Docstore {
	d := &Docstore{
		conn: conn,
		coll: nil,
	}
	return d
}

func (d *Docstore) open() error {
	coll, err := docstore.OpenCollection(context.Background(), d.conn)
	if err != nil {
		return fmt.Errorf("%w (%v)", ErrFailedToOpen, err)
	}
	d.coll = coll
	return nil
}

// Open opens d.coll once.
func (d *Docstore) Open() error {
	var oErr error
	d.once.Do(func() { oErr = d.open() })
	if oErr != nil {
		return oErr
	}
	return nil
}

// Close closes the Docstore.
func (d *Docstore) Close() error {
	if err := d.Open(); err != nil {
		return fmt.Errorf("%w (%v)", ErrFailedToClose, err)
	}
	if err := d.coll.Close(); err != nil {
		return fmt.Errorf("%w (%v)", ErrFailedToClose, err)
	}
	return nil
}

func (d *Docstore) PutEntity(ctx context.Context, e *AnchorEntity) error {
	if err := d.Open(); err != nil {
		return fmt.Errorf("%w (%v)", ErrFailedToPut, err)
	}
	if err := d.coll.Put(ctx, e); err != nil {
		return fmt.Errorf("%w (%v)", ErrFailedToPut, err)
	}
	return nil
}

func (d *Docstore) GetEntity(ctx context.Context, e *AnchorEntity) error {
	if err := d.Open(); err != nil {
		return fmt.Errorf("%w (%v)", ErrFailedToGet, err)
	}
	if err := d.coll.Get(ctx, e); err != nil {
		return fmt.Errorf("%w (%v)", ErrFailedToGet, err)
	}
	return nil
}

// UpdateEntity updates the AnchorEntity specified by e.CID.
// It updates Confirmations, BBc1DomainName, and Note only,
// as other data must not be changed.
func (d *Docstore) UpdateEntity(ctx context.Context, e *AnchorEntity, updateConfirmations, updateBBc1Dom, updateNote bool) error {
	if err := d.Open(); err != nil {
		return fmt.Errorf("%w (%v)", ErrFailedToUpdate, err)
	}
	mod := docstore.Mods{}
	if updateConfirmations {
		mod["confirmations"] = e.Confirmations
	}
	if updateBBc1Dom {
		mod["bbc1dom"] = e.BBc1DomainName
	}
	if updateNote {
		mod["note"] = e.Note
	}
	if err := d.coll.Update(ctx, e, mod); err != nil {
		return fmt.Errorf("%w (%v)", ErrFailedToUpdate, err)
	}
	return nil
}

func (d *Docstore) Put(ctx context.Context, r *model.AnchorRecord) error {
	e := NewAnchorEntity(r)
	if err := d.PutEntity(ctx, e); err != nil {
		return err
	}
	return nil
}

func (d *Docstore) Get(ctx context.Context, bbc1dom, bbc1tx []byte) (*model.AnchorRecord, error) {
	e := &AnchorEntity{
		CID: hex.EncodeToString(bbc1dom) + hex.EncodeToString(bbc1tx),
	}
	if err := d.GetEntity(ctx, e); err != nil {
		return nil, err
	}
	return e.AnchorRecord(), nil
}

func (d *Docstore) UpdateConfirmations(ctx context.Context, bbc1dom, bbc1tx []byte, confirmations uint) error {
	e := &AnchorEntity{
		CID:           hex.EncodeToString(bbc1dom) + hex.EncodeToString(bbc1tx),
		Confirmations: confirmations,
	}
	if err := d.UpdateEntity(ctx, e, true, false, false); err != nil {
		return err
	}
	return nil
}

func (d *Docstore) UpdateBBc1DomainName(ctx context.Context, bbc1dom, bbc1tx []byte, bbc1domName string) error {
	e := &AnchorEntity{
		CID:            hex.EncodeToString(bbc1dom) + hex.EncodeToString(bbc1tx),
		BBc1DomainName: bbc1domName,
	}
	if err := d.UpdateEntity(ctx, e, false, true, false); err != nil {
		return err
	}
	return nil
}

func (d *Docstore) UpdateNote(ctx context.Context, bbc1dom, bbc1tx []byte, note string) error {
	e := &AnchorEntity{
		CID:  hex.EncodeToString(bbc1dom) + hex.EncodeToString(bbc1tx),
		Note: note,
	}
	if err := d.UpdateEntity(ctx, e, false, false, true); err != nil {
		return err
	}
	return nil
}

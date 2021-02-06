package store

import (
	"context"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
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
	ErrFailedToOpen  = errors.New("ErrFailedToOpen")
	ErrFailedToClose = errors.New("ErrFailedToClose")
	ErrFailedToGet   = errors.New("ErrFailedToGet")
	ErrFailedToPut   = errors.New("ErrFailedToPut")
)

type Docstore struct {
	io.Closer
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

func (d *Docstore) Open() error {
	var oErr error
	d.once.Do(func() { oErr = d.open() })
	if oErr != nil {
		return oErr
	}
	return nil
}

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
		return fmt.Errorf("%w (%v)", ErrFailedToClose, err)
	}
	if err := d.coll.Put(ctx, e); err != nil {
		return fmt.Errorf("%w (%v)", ErrFailedToPut, err)
	}
	return nil
}

func (d *Docstore) GetEntity(ctx context.Context, e *AnchorEntity) error {
	if err := d.Open(); err != nil {
		return fmt.Errorf("%w (%v)", ErrFailedToClose, err)
	}
	if err := d.coll.Get(ctx, e); err != nil {
		return fmt.Errorf("%w (%v)", ErrFailedToGet, err)
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
	tmp := model.NewAnchor(255, time.Time{}, bbc1dom, bbc1tx)
	e := &AnchorEntity{
		CID: hex.EncodeToString(tmp.BBc1DomainID[:]) + hex.EncodeToString(tmp.BBc1TransactionID[:]),
	}
	if err := d.GetEntity(ctx, e); err != nil {
		return nil, err
	}
	return e.AnchorRecord(), nil
}

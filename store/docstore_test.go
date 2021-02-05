package store_test

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/ebiiim/btc-gateway/model"
	"github.com/ebiiim/btc-gateway/store"
	"github.com/ebiiim/btc-gateway/util"

	_ "gocloud.dev/docstore/memdocstore"
)

var (
	dom1     = util.MustConvert32B(util.MustDecodeHexString("456789abc0ef0123456089abcdef0023456789a0cdef0123406789abcde00123"))
	tx1      = util.MustConvert32B(util.MustDecodeHexString("56789abcd0f0123456709abcdef0103456789ab0def0123450789abcdef01234"))
	aeID1    = util.MustConvert64B(util.MustDecodeHexString("456789abc0ef0123456089abcdef0023456789a0cdef0123406789abcde0012356789abcd0f0123456709abcdef0103456789ab0def0123450789abcdef01234"))
	ts1      = time.Unix(1612449628, 0)
	btctx1   = util.MustDecodeHexString("6928e1c6478d1f55ed1a5d86e1ab24669a14f777b879bbb25c746543810bf916")
	txts1    = time.Unix(1612449916, 0)
	confirm1 = uint(239)
	a1       = &model.Anchor{
		Version:           255,
		BTCNet:            model.BTCTestnet3,
		Timestamp:         ts1,
		BBc1DomainID:      dom1,
		BBc1TransactionID: tx1,
	}
	ar1 = &model.AnchorRecord{
		Anchor:           a1,
		BTCTransactionID: btctx1,
		TransactionTime:  txts1,
		Confirmations:    confirm1,
		BBc1DomainName:   "testDom",
		Note:             "hello world",
	}
	ae1 = &store.AnchorEntity{
		ID:               aeID1,
		AnchorVersion:    255,
		BTCNet:           model.BTCTestnet3,
		AnchorTime:       ts1,
		BTCTransactionID: btctx1,
		TransactionTime:  txts1,
		Confirmations:    confirm1,
		BBc1DomainName:   "testDom",
		Note:             "hello world",
	}
)

func TestNewAnchorEntity(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input *model.AnchorRecord
		want  *store.AnchorEntity
	}{
		{"normal", ar1, ae1},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := store.NewAnchorEntity(ar1)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("got %+v but want %+v", got, c.want)
			}
		})
	}
}

func TestAnchorEntity_AnchorRecord(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input *store.AnchorEntity
		want  *model.AnchorRecord
	}{
		{"normal", ae1, ar1},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := ae1.AnchorRecord()
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("got %+v but want %+v", got, c.want)
			}
		})
	}
}

// Assumes tests will be run from package root.
var (
	testdb1 = "testdata/anchors1.db"
	conn1   = "mem://store_test_get/id?filename=" + testdb1
	testdb2 = "testdata/anchors2.db"
	conn2   = "mem://store_test_put/id?filename=" + testdb2
)

//
// Do not parallelize Docstore tests as memdocstore is NOT thread-safe.
//

func TestDocstore_GetEntity(t *testing.T) {
	cases := []struct {
		name string
		conn string
		id   [64]byte
		want *store.AnchorEntity
	}{
		{"normal", conn1, aeID1, ae1},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			docs := store.NewDocstore(c.conn)
			defer docs.Close() // Ignores error as no write access.

			ctx := context.Background()
			ctx, cancelFunc := context.WithTimeout(ctx, 5*time.Second)
			defer cancelFunc()

			got := &store.AnchorEntity{ID: c.id}
			if err := docs.GetEntity(ctx, got); err != nil {
				t.Error(err)
				t.Skip()
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("got %+v but want %+v", got, c.want)
			}
		})
	}
}

func TestDocstore_PutEntity(t *testing.T) {
	cases := []struct {
		name   string
		dbFile string
		conn   string
		id     [64]byte
		put    *store.AnchorEntity
	}{
		{"normal", testdb2, conn2, aeID1, ae1},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			docs := store.NewDocstore(c.conn)

			ctx := context.Background()
			ctx, cancelFunc := context.WithTimeout(ctx, 30*time.Second)
			defer cancelFunc()

			// put
			if err := docs.PutEntity(ctx, c.put); err != nil {
				t.Error(err)
			}
			// save
			if err := docs.Close(); err != nil {
				t.Error(err)
				os.Remove(c.dbFile)
				t.Skip()
			}
			// load & get
			docs2 := store.NewDocstore(c.conn)
			got := &store.AnchorEntity{ID: c.id}
			if err := docs2.GetEntity(ctx, got); err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual(got, c.put) {
				t.Errorf("got %+v but want %+v", got, c.put)
			}
			// cleanup
			docs2.Close()
			os.Remove(c.dbFile)
		})
	}
}

func TestDocstore_Get(t *testing.T) {
	cases := []struct {
		name  string
		conn  string
		domid []byte
		txid  []byte
		want  *model.AnchorRecord
	}{
		{"normal", conn1, dom1[:], tx1[:], ar1},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			docs := store.NewDocstore(c.conn)
			defer docs.Close() // Ignores error as no write access.

			ctx := context.Background()
			ctx, cancelFunc := context.WithTimeout(ctx, 5*time.Second)
			defer cancelFunc()

			got, err := docs.Get(ctx, c.domid, c.txid)
			if err != nil {
				t.Error(err)
				t.Skip()
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("got %+v but want %+v", got, c.want)
			}
		})
	}

}

func TestDocstore_Put(t *testing.T) {
	cases := []struct {
		name   string
		dbFile string
		conn   string
		domid  []byte
		txid   []byte
		want   *model.AnchorRecord
	}{
		{"normal", testdb2, conn2, dom1[:], tx1[:], ar1},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			docs := store.NewDocstore(c.conn)

			ctx := context.Background()
			ctx, cancelFunc := context.WithTimeout(ctx, 30*time.Second)
			defer cancelFunc()

			// put
			if err := docs.Put(ctx, c.want); err != nil {
				t.Error(err)
			}
			// save
			if err := docs.Close(); err != nil {
				t.Error(err)
				os.Remove(c.dbFile)
				t.Skip()
			}
			// load & get
			docs2 := store.NewDocstore(c.conn)
			got, err := docs2.Get(ctx, c.domid, c.txid)
			if err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("got %+v but want %+v", got, c.want)
			}
			// cleanup
			docs2.Close()
			os.Remove(c.dbFile)
		})
	}
}

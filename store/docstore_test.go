package store_test

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/ebiiim/btcgw/model"
	"github.com/ebiiim/btcgw/store"
	"github.com/ebiiim/btcgw/util"

	_ "gocloud.dev/docstore/memdocstore"
)

var (
	dom1     = util.MustDecodeHexString("456789abc0ef0123456089abcdef0023456789a0cdef0123406789abcde00123")
	tx1      = util.MustDecodeHexString("56789abcd0f0123456709abcdef0103456789ab0def0123450789abcdef01234")
	cid1     = "456789abc0ef0123456089abcdef0023456789a0cdef0123406789abcde0012356789abcd0f0123456709abcdef0103456789ab0def0123450789abcdef01234"
	ts1      = time.Unix(1612449628, 0)
	btctx1   = util.MustDecodeHexString("6928e1c6478d1f55ed1a5d86e1ab24669a14f777b879bbb25c746543810bf916")
	txts1    = time.Unix(1612449916, 0)
	confirm1 = uint(312)
	a1       = &model.Anchor{
		Version:           255,
		BTCNet:            model.BTCTestnet3,
		Timestamp:         ts1,
		BBc1DomainID:      util.MustConvert32B(dom1),
		BBc1TransactionID: util.MustConvert32B(tx1),
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
		CID:               cid1,
		BBc1DomainID:      dom1,
		BBc1TransactionID: tx1,
		AnchorVersion:     255,
		BTCNet:            model.BTCTestnet3,
		AnchorTime:        ts1,
		BTCTransactionID:  btctx1,
		TransactionTime:   txts1,
		Confirmations:     confirm1,
		BBc1DomainName:    "testDom",
		Note:              "hello world",
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
	conn1   = "mem://store_test_get/cid?filename=" + testdb1
	testdb2 = "testdata/anchors2.db"
	conn2   = "mem://store_test_put/cid?filename=" + testdb2
)

//
// Do not parallelize Docstore tests as memdocstore is NOT thread-safe.
//

func TestDocstore_GetEntity(t *testing.T) {
	cases := []struct {
		name string
		conn string
		cid  string
		want *store.AnchorEntity
	}{
		{"normal", conn1, cid1, ae1},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			docs := store.NewDocstore(c.conn)
			defer docs.Close() // Ignores error as no write access.

			ctx := context.Background()
			ctx, cancelFunc := context.WithTimeout(ctx, 5*time.Second)
			defer cancelFunc()

			got := &store.AnchorEntity{CID: c.cid}
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
		cid    string
		put    *store.AnchorEntity
	}{
		{"normal", testdb2, conn2, cid1, ae1},
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
			got := &store.AnchorEntity{CID: c.cid}
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

func TestDocstore_UpdateEntity(t *testing.T) {
	cases := []struct {
		name                string
		dbFile              string
		conn                string
		cid                 string
		original            *store.AnchorEntity
		updateConfirmations bool
		confirmations       uint
		updateBBc1Dom       bool
		bbc1Dom             string
		updateNote          bool
		note                string
	}{
		{"all", testdb2, conn2, cid1, ae1, true, 123, true, "my-domain", true, "yo"},
		{"confs_only", testdb2, conn2, cid1, ae1, true, 123, false, "", false, ""},
		{"dom_only", testdb2, conn2, cid1, ae1, false, 0, true, "my-domain", false, ""},
		{"note_only", testdb2, conn2, cid1, ae1, false, 0, false, "", true, "yo"},
		{"empty_dom_note", testdb2, conn2, cid1, ae1, false, 0, true, "", true, ""},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			docs := store.NewDocstore(c.conn)

			ctx := context.Background()
			ctx, cancelFunc := context.WithTimeout(ctx, 30*time.Second)
			defer cancelFunc()

			// put
			if err := docs.PutEntity(ctx, c.original); err != nil {
				t.Error(err)
			}
			// update
			ue := *c.original
			ue.Confirmations = c.confirmations
			ue.BBc1DomainName = c.bbc1Dom
			ue.Note = c.note
			if err := docs.UpdateEntity(ctx, &ue, c.updateConfirmations, c.updateBBc1Dom, c.updateNote); err != nil {
				t.Error(err)
			}
			// get
			got := &store.AnchorEntity{CID: c.cid}
			if err := docs.GetEntity(ctx, got); err != nil {
				t.Error(err)
			}
			// test
			if c.updateConfirmations && (got.Confirmations != c.confirmations) {
				t.Errorf("updateConfirmations: got %+v but want %+v", got.Confirmations, c.confirmations)
			}
			if !c.updateConfirmations && (got.Confirmations != c.original.Confirmations) {
				t.Errorf("!updateConfirmations: got %+v but want %+v", got.Confirmations, c.original.Confirmations)
			}
			if c.updateBBc1Dom && (got.BBc1DomainName != c.bbc1Dom) {
				t.Errorf("updateBBc1Dom: got %+v but want %+v", got.BBc1DomainName, c.bbc1Dom)
			}
			if !c.updateBBc1Dom && (got.BBc1DomainName != c.original.BBc1DomainName) {
				t.Errorf("!updateBBc1Dom: got %+v but want %+v", got.BBc1DomainName, c.original.BBc1DomainName)
			}
			if c.updateNote && (got.Note != c.note) {
				t.Errorf("updateNote: got %+v but want %+v", got.Note, c.note)
			}
			if !c.updateNote && (got.Note != c.original.Note) {
				t.Errorf("!updateNote got %+v but want %+v", got.Note, c.original.Note)
			}
			// cleanup
			docs.Close()
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
		{"normal", conn1, dom1, tx1, ar1},
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
		{"normal", testdb2, conn2, dom1, tx1, ar1},
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

func TestDocstore_UpdateConfirmations(t *testing.T) {
	cases := []struct {
		name     string
		dbFile   string
		conn     string
		domid    []byte
		txid     []byte
		original *model.AnchorRecord
		wantConf uint
	}{
		{"normal", testdb2, conn2, dom1, tx1, ar1, 123},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			docs := store.NewDocstore(c.conn)
			ctx := context.Background()
			ctx, cancelFunc := context.WithTimeout(ctx, 30*time.Second)
			defer cancelFunc()
			// put
			if err := docs.Put(ctx, c.original); err != nil {
				t.Error(err)
			}
			// update
			if err := docs.UpdateConfirmations(ctx, c.domid, c.txid, c.wantConf); err != nil {
				t.Error(err)
			}
			// get
			got, err := docs.Get(ctx, c.domid, c.txid)
			if err != nil {
				t.Error(err)
			}
			want := *c.original
			want.Confirmations = c.wantConf
			if !reflect.DeepEqual(got, &want) {
				t.Errorf("got %+v but want %+v", got, &want)
			}
			// cleanup
			docs.Close()
			os.Remove(c.dbFile)
		})
	}
}

func TestDocstore_UpdateBBc1DomainName(t *testing.T) {
	cases := []struct {
		name     string
		dbFile   string
		conn     string
		domid    []byte
		txid     []byte
		original *model.AnchorRecord
		wantDom  string
	}{
		{"normal", testdb2, conn2, dom1, tx1, ar1, "my-domain"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			docs := store.NewDocstore(c.conn)
			ctx := context.Background()
			ctx, cancelFunc := context.WithTimeout(ctx, 30*time.Second)
			defer cancelFunc()
			// put
			if err := docs.Put(ctx, c.original); err != nil {
				t.Error(err)
			}
			// update
			if err := docs.UpdateBBc1DomainName(ctx, c.domid, c.txid, c.wantDom); err != nil {
				t.Error(err)
			}
			// get
			got, err := docs.Get(ctx, c.domid, c.txid)
			if err != nil {
				t.Error(err)
			}
			want := *c.original
			want.BBc1DomainName = c.wantDom
			if !reflect.DeepEqual(got, &want) {
				t.Errorf("got %+v but want %+v", got, &want)
			}
			// cleanup
			docs.Close()
			os.Remove(c.dbFile)
		})
	}
}

func TestDocstore_UpdateNote(t *testing.T) {
	cases := []struct {
		name     string
		dbFile   string
		conn     string
		domid    []byte
		txid     []byte
		original *model.AnchorRecord
		wantNote string
	}{
		{"normal", testdb2, conn2, dom1, tx1, ar1, "yo"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			docs := store.NewDocstore(c.conn)
			ctx := context.Background()
			ctx, cancelFunc := context.WithTimeout(ctx, 30*time.Second)
			defer cancelFunc()
			// put
			if err := docs.Put(ctx, c.original); err != nil {
				t.Error(err)
			}
			// update
			if err := docs.UpdateNote(ctx, c.domid, c.txid, c.wantNote); err != nil {
				t.Error(err)
			}
			// get
			got, err := docs.Get(ctx, c.domid, c.txid)
			if err != nil {
				t.Error(err)
			}
			want := *c.original
			want.Note = c.wantNote
			if !reflect.DeepEqual(got, &want) {
				t.Errorf("got %+v but want %+v", got, &want)
			}
			// cleanup
			docs.Close()
			os.Remove(c.dbFile)
		})
	}
}

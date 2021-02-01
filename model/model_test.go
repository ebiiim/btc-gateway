package model_test

import (
	"encoding/hex"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/ebiiim/btc-gateway/model"
)

func mustDecodeHexString(s string) []byte {
	p, err := hex.DecodeString(s)
	if err != nil {
		panic("mustDecodeHexString")
	}
	return p
}

func mustConvert32B(p []byte) [32]byte {
	if len(p) != 32 {
		panic("mustConvert32B")
	}
	var b [32]byte
	copy(b[:], p[:])
	return b
}

func mustConvert80B(p []byte) [80]byte {
	if len(p) != 80 {
		panic("mustConvert80B")
	}
	var b [80]byte
	copy(b[:], p[:])
	return b
}

var (
	dom16      = mustDecodeHexString("23456789a0cdef0123406789abcde001")
	dom16a     = mustConvert32B(mustDecodeHexString("23456789a0cdef0123406789abcde00100000000000000000000000000000000"))
	tx16       = mustDecodeHexString("3456789ab0def0123450789abcdef012")
	tx16a      = mustConvert32B(mustDecodeHexString("3456789ab0def0123450789abcdef01200000000000000000000000000000000"))
	dom32      = mustDecodeHexString("456789abc0ef0123456089abcdef0023456789a0cdef0123406789abcde00123")
	dom32a     = mustConvert32B(dom32)
	tx32       = mustDecodeHexString("56789abcd0f0123456709abcdef0103456789ab0def0123450789abcdef01234")
	tx32a      = mustConvert32B(tx32)
	dom64      = mustDecodeHexString("6789abcde00123456780abcdef0120456789abc0ef0123456089abcdef0023456789a0cdef0123406789abcde00123456780abcdef0120456789abc0ef012345")
	dom64a     = mustConvert32B(mustDecodeHexString("6789abcde00123456780abcdef0120456789abc0ef0123456089abcdef002345"))
	tx64       = mustDecodeHexString("789abcdef01234567890bcdef0123056789abcd0f0123456709abcdef0103456789ab0def0123450789abcdef01234567890bcdef0123056789abcd0f0123456")
	tx64a      = mustConvert32B(mustDecodeHexString("789abcdef01234567890bcdef0123056789abcd0f0123456709abcdef0103456"))
	opRet32M   = mustConvert80B(mustDecodeHexString("4242633101ff00000000000000000000456789abc0ef0123456089abcdef0023456789a0cdef0123406789abcde0012356789abcd0f0123456709abcdef0103456789ab0def0123450789abcdef01234"))
	opRet64T3  = mustConvert80B(mustDecodeHexString("424263310103000000000000000000006789abcde00123456780abcdef0120456789abc0ef0123456089abcdef002345789abcdef01234567890bcdef0123056789abcd0f0123456709abcdef0103456"))
	opRet16T4  = mustConvert80B(mustDecodeHexString("4242633101040000000000000000000023456789a0cdef0123406789abcde001000000000000000000000000000000003456789ab0def0123450789abcdef01200000000000000000000000000000000"))
	InvalidSig = mustConvert80B(mustDecodeHexString("4242003101ff00000000000000000000456789abc0ef0123456089abcdef0023456789a0cdef0123406789abcde0012356789abcd0f0123456709abcdef0103456789ab0def0123450789abcdef01234"))
	InvalidVer = mustConvert80B(mustDecodeHexString("4242633100ff00000000000000000000456789abc0ef0123456089abcdef0023456789a0cdef0123406789abcde0012356789abcd0f0123456709abcdef0103456789ab0def0123450789abcdef01234"))
	InvalidNet = mustConvert80B(mustDecodeHexString("42426331010000000000000000000000456789abc0ef0123456089abcdef0023456789a0cdef0123406789abcde0012356789abcd0f0123456709abcdef0103456789ab0def0123450789abcdef01234"))
)

func TestNewAnchor(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		btcnet  model.BTCNet
		bbc1dom []byte
		bbc1tx  []byte
		want    *model.Anchor
	}{
		{"16bit", model.BTCMainnet, dom16, tx16, &model.Anchor{1, model.BTCMainnet, dom16a, tx16a}},
		{"32bit", model.BTCMainnet, dom32, tx32, &model.Anchor{1, model.BTCMainnet, dom32a, tx32a}},
		{"64bit", model.BTCMainnet, dom64, tx64, &model.Anchor{1, model.BTCMainnet, dom64a, tx64a}},
		{"testnet3", model.BTCTestnet3, dom32, tx32, &model.Anchor{1, model.BTCTestnet3, dom32a, tx32a}},
		{"testnet4", model.BTCTestnet4, dom32, tx32, &model.Anchor{1, model.BTCTestnet4, dom32a, tx32a}},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			a := model.NewAnchor(c.btcnet, c.bbc1dom, c.bbc1tx)
			if !reflect.DeepEqual(a, c.want) {
				t.Errorf("got %+v but want %+v", a, c.want)
			}
		})
	}
}

func TestEncodeOpReturn(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		input *model.Anchor
		want  [80]byte
	}{
		{"32bit_mainnet", model.NewAnchor(model.BTCMainnet, dom32, tx32), opRet32M},
		{"64bit_testnet3", model.NewAnchor(model.BTCTestnet3, dom64, tx64), opRet64T3},
		{"16bit_testnet4", model.NewAnchor(model.BTCTestnet4, dom16, tx16), opRet16T4},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			b := model.EncodeOpReturn(c.input)
			for i, v := range c.want {
				if v != b[i] {
					t.Errorf("idx=%d got=%x want=%x", i, b[i], v)
				}
			}
		})
	}
}

func TestDecodeOpReturn(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		input [80]byte
		want  *model.Anchor
	}{
		{"32bit_mainnet", opRet32M, model.NewAnchor(model.BTCMainnet, dom32, tx32)},
		{"64bit_testnet3", opRet64T3, model.NewAnchor(model.BTCTestnet3, dom64, tx64)},
		{"16bit_testnet4", opRet16T4, model.NewAnchor(model.BTCTestnet4, dom16, tx16)},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			a, err := model.DecodeOpReturn(c.input)
			if err != nil {
				t.Errorf("err %v", err)
			} else if !reflect.DeepEqual(a, c.want) {
				t.Errorf("got %+v but want %+v", a, c.want)
			}
		})
	}
}

func TestDecodeOpReturnErr(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		input [80]byte
		want  error
	}{
		{"invalid_signature", InvalidSig, model.ErrInvalidSignature},
		{"invalid_version", InvalidVer, model.ErrInvalidVersion},
		{"invalid_network", InvalidNet, model.ErrInvalidBTCNet},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if _, err := model.DecodeOpReturn(c.input); !errors.Is(err, c.want) {
				t.Errorf("got %+v but want %+v", err, c.want)
			}
		})
	}
}

var (
	normalAnchor = model.NewAnchor(model.BTCMainnet, dom32, tx32)
	btctx1       = mustDecodeHexString("57511f74c3836c0d4d62a6183fa54e600372e1aed5b5be2f78ef5b766a314a5d")
	btcaddr1     = "tb1qhexc7d0fzex7lrzw3l0j2dmvhgegt02ckfdzjr"
	ts1          = time.Unix(1611334493, 0)
	domName1     = "bbc1test"
	note1        = "hello world"
)

func TestNewAnchorRecord(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		anchor      *model.Anchor
		btctx       []byte
		ts          time.Time
		conf        uint
		btcaddr     string
		bbc1domName string
		note        string
		want        *model.AnchorRecord
	}{
		{"normal", normalAnchor, btctx1, ts1, 1500, btcaddr1, domName1, note1, &model.AnchorRecord{normalAnchor, btctx1, ts1, 1500, btcaddr1, domName1, note1}},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			r := model.NewAnchorRecord(c.anchor, c.btctx, c.ts, c.conf, c.btcaddr, c.bbc1domName, c.note)
			if !reflect.DeepEqual(r, c.want) {
				t.Errorf("got %+v but want %+v", r, c.want)
			}
		})
	}
}

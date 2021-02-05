package model_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/ebiiim/btc-gateway/model"
	"github.com/ebiiim/btc-gateway/util"
)

var (
	time1      = time.Unix(1612363134, 0)  // 2021-02-03T14:38:54Z
	time2      = time.Unix(11847472496, 0) // 2345-06-07T12:34:56Z
	dom16      = util.MustDecodeHexString("23456789a0cdef0123406789abcde001")
	dom16a     = util.MustConvert32B(util.MustDecodeHexString("23456789a0cdef0123406789abcde00100000000000000000000000000000000"))
	tx16       = util.MustDecodeHexString("3456789ab0def0123450789abcdef012")
	tx16a      = util.MustConvert32B(util.MustDecodeHexString("3456789ab0def0123450789abcdef01200000000000000000000000000000000"))
	dom32      = util.MustDecodeHexString("456789abc0ef0123456089abcdef0023456789a0cdef0123406789abcde00123")
	dom32a     = util.MustConvert32B(dom32)
	tx32       = util.MustDecodeHexString("56789abcd0f0123456709abcdef0103456789ab0def0123450789abcdef01234")
	tx32a      = util.MustConvert32B(tx32)
	dom64      = util.MustDecodeHexString("6789abcde00123456780abcdef0120456789abc0ef0123456089abcdef0023456789a0cdef0123406789abcde00123456780abcdef0120456789abc0ef012345")
	dom64a     = util.MustConvert32B(util.MustDecodeHexString("6789abcde00123456780abcdef0120456789abc0ef0123456089abcdef002345"))
	tx64       = util.MustDecodeHexString("789abcdef01234567890bcdef0123056789abcd0f0123456709abcdef0103456789ab0def0123450789abcdef01234567890bcdef0123056789abcd0f0123456")
	tx64a      = util.MustConvert32B(util.MustDecodeHexString("789abcdef01234567890bcdef0123056789abcd0f0123456709abcdef0103456"))
	opRet32M   = util.MustConvert80B(util.MustDecodeHexString("4242633101ff000000000000601ab57e456789abc0ef0123456089abcdef0023456789a0cdef0123406789abcde0012356789abcd0f0123456709abcdef0103456789ab0def0123450789abcdef01234"))
	opRet64T3  = util.MustConvert80B(util.MustDecodeHexString("424263310103000000000000601ab57e6789abcde00123456780abcdef0120456789abc0ef0123456089abcdef002345789abcdef01234567890bcdef0123056789abcd0f0123456709abcdef0103456"))
	opRet16T4  = util.MustConvert80B(util.MustDecodeHexString("424263310104000000000000601ab57e23456789a0cdef0123406789abcde001000000000000000000000000000000003456789ab0def0123450789abcdef01200000000000000000000000000000000"))
	InvalidSig = util.MustConvert80B(util.MustDecodeHexString("4242003101ff000000000000601ab57e456789abc0ef0123456089abcdef0023456789a0cdef0123406789abcde0012356789abcd0f0123456709abcdef0103456789ab0def0123450789abcdef01234"))
	InvalidVer = util.MustConvert80B(util.MustDecodeHexString("4242633100ff000000000000601ab57e456789abc0ef0123456089abcdef0023456789a0cdef0123406789abcde0012356789abcd0f0123456709abcdef0103456789ab0def0123450789abcdef01234"))
	InvalidNet = util.MustConvert80B(util.MustDecodeHexString("424263310100000000000000601ab57e456789abc0ef0123456089abcdef0023456789a0cdef0123406789abcde0012356789abcd0f0123456709abcdef0103456789ab0def0123450789abcdef01234"))
	o32Mtime34 = util.MustConvert80B(util.MustDecodeHexString("4242633101ff000000000002c22a1570456789abc0ef0123456089abcdef0023456789a0cdef0123406789abcde0012356789abcd0f0123456709abcdef0103456789ab0def0123450789abcdef01234"))
	o32MAnc255 = util.MustConvert80B(util.MustDecodeHexString("42426331ffff000000000000601ab57e456789abc0ef0123456089abcdef0023456789a0cdef0123406789abcde0012356789abcd0f0123456709abcdef0103456789ab0def0123450789abcdef01234"))
)

func TestNewAnchor(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		btcnet  model.BTCNet
		ts      time.Time
		bbc1dom []byte
		bbc1tx  []byte
		want    *model.Anchor
	}{
		{"16bit", model.BTCMainnet, time1, dom16, tx16, &model.Anchor{1, model.BTCMainnet, time1, dom16a, tx16a}},
		{"32bit", model.BTCMainnet, time1, dom32, tx32, &model.Anchor{1, model.BTCMainnet, time1, dom32a, tx32a}},
		{"64bit", model.BTCMainnet, time1, dom64, tx64, &model.Anchor{1, model.BTCMainnet, time1, dom64a, tx64a}},
		{"testnet3", model.BTCTestnet3, time1, dom32, tx32, &model.Anchor{1, model.BTCTestnet3, time1, dom32a, tx32a}},
		{"testnet4", model.BTCTestnet4, time1, dom32, tx32, &model.Anchor{1, model.BTCTestnet4, time1, dom32a, tx32a}},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			a := model.NewAnchor(c.btcnet, c.ts, c.bbc1dom, c.bbc1tx)
			if !reflect.DeepEqual(a, c.want) {
				t.Errorf("got %+v but want %+v", a, c.want)
			}
		})
	}
}

func TestEncodeOpReturn(t *testing.T) {
	// Do not parallelize as this test changes model.anchorVersion.
	// t.Parallel()
	cases := []struct {
		name  string
		input *model.Anchor
		want  [80]byte
	}{
		{"32bit_mainnet", model.NewAnchor(model.BTCMainnet, time1, dom32, tx32), opRet32M},
		{"64bit_testnet3", model.NewAnchor(model.BTCTestnet3, time1, dom64, tx64), opRet64T3},
		{"16bit_testnet4", model.NewAnchor(model.BTCTestnet4, time1, dom16, tx16), opRet16T4},
		{"32bit_mainnet_time34bit", model.NewAnchor(model.BTCMainnet, time2, dom32, tx32), o32Mtime34},
		{"anchor_version_255", func() *model.Anchor {
			model.XAnchorVersion(255)
			a := model.NewAnchor(model.BTCMainnet, time1, dom32, tx32)
			model.XAnchorVersion(1)
			return a
		}(), o32MAnc255},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			// t.Parallel()
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
		{"32bit_mainnet", opRet32M, model.NewAnchor(model.BTCMainnet, time1, dom32, tx32)},
		{"64bit_testnet3", opRet64T3, model.NewAnchor(model.BTCTestnet3, time1, dom64, tx64)},
		{"16bit_testnet4", opRet16T4, model.NewAnchor(model.BTCTestnet4, time1, dom16, tx16)},
		{"32bit_mainnet_time34bit", o32Mtime34, model.NewAnchor(model.BTCMainnet, time2, dom32, tx32)},
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
	normalAnchor = model.NewAnchor(model.BTCMainnet, time1, dom32, tx32)
	btctx1       = util.MustDecodeHexString("57511f74c3836c0d4d62a6183fa54e600372e1aed5b5be2f78ef5b766a314a5d")
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
		bbc1domName string
		note        string
		want        *model.AnchorRecord
	}{
		{"normal", normalAnchor, btctx1, ts1, 1500, domName1, note1, &model.AnchorRecord{normalAnchor, btctx1, ts1, 1500, domName1, note1}},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			r := model.NewAnchorRecord(c.anchor, c.btctx, c.ts, c.conf, c.bbc1domName, c.note)
			if !reflect.DeepEqual(r, c.want) {
				t.Errorf("got %+v but want %+v", r, c.want)
			}
		})
	}
}

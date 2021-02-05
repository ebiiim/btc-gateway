package util_test

import (
	"testing"

	"github.com/ebiiim/btc-gateway/util"
)

func Test(t *testing.T) {
	s32 := "012345678901234567890123456789ab012345678901234567890123456789ab"
	s64 := s32 + s32
	s80 := s64 + "0123456789ABCDEF0123456789ABCDEF"
	util.MustConvert32B(util.MustDecodeHexString(s32))
	util.MustConvert64B(util.MustDecodeHexString(s64))
	util.MustConvert80B(util.MustDecodeHexString(s80))
}

func TestMustDecodeHexString_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("panic needed")
		}
	}()

	util.MustDecodeHexString("X")
}

func TestMustConvert32B_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("panic needed")
		}
	}()

	util.MustConvert32B([]byte{123, 123})
}

func TestMustConvert64B_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("panic needed")
		}
	}()

	util.MustConvert64B([]byte{123, 123})
}

func TestMustConvert80B_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("panic needed")
		}
	}()

	util.MustConvert80B([]byte{123, 123})
}

package util_test

import (
	"os"
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
	util.MustAtoi("12345")
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

func TestMustAtoi_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("panic needed")
		}
	}()

	util.MustAtoi("A")
}

func TestGetEnvOr(t *testing.T) {
	key := "___TESTGETENVOR___"
	def := "___DEFAULTVALUE___"
	val := "___COOOOOLVALUE___"
	if err := os.Setenv(key, ""); err != nil {
		t.Error(err)
	}
	if s := util.GetEnvOr(key, def); s != def {
		t.Errorf("got %s but want %s", s, def)
	}
	if err := os.Setenv(key, val); err != nil {
		t.Error(err)
	}
	if s := util.GetEnvOr(key, def); s != val {
		t.Errorf("got %s but want %s", s, def)
	}
}

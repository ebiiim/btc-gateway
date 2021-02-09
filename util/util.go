// Package util provides some common utilities for testing.
package util

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
)

// MustDecodeHexString returns a decoded string or raises panic.
func MustDecodeHexString(s string) []byte {
	p, err := hex.DecodeString(s)
	if err != nil {
		panic(fmt.Sprintf("MustDecodeHexString: %s", s))
	}
	return p
}

// MustConvert32B returns a byte array or raises panic.
func MustConvert32B(p []byte) [32]byte {
	if len(p) != 32 {
		panic(fmt.Sprintf("MustConvert32B: %x", p))
	}
	var b [32]byte
	copy(b[:], p[:])
	return b
}

// MustConvert64B returns a byte array or raises panic.
func MustConvert64B(p []byte) [64]byte {
	if len(p) != 64 {
		panic(fmt.Sprintf("MustConvert64B: %x", p))
	}
	var b [64]byte
	copy(b[:], p[:])
	return b
}

// MustConvert80B returns a byte array or raises panic.
func MustConvert80B(p []byte) [80]byte {
	if len(p) != 80 {
		panic(fmt.Sprintf("MustConvert80B: %x", p))
	}
	var b [80]byte
	copy(b[:], p[:])
	return b
}

// MustAtoi returns an int or raises panic.
func MustAtoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(fmt.Sprintf("MustAtoi: %s", s))
	}
	return i
}

// GetEnvOr returns an environment variable specified by env,
// or returns defaultValue if the environment variable is not defined.
func GetEnvOr(env string, defaultValue string) string {
	v := os.Getenv(env)
	if v == "" {
		return defaultValue
	}
	return v
}

package model

import (
	"errors"
	"fmt"
	"time"
)

// Errors
var (
	ErrInvalidSignature = errors.New("ErrInvalidSignature")
	ErrInvalidVersion   = errors.New("ErrInvalidVersion")
	ErrInvalidBTCNet    = errors.New("ErrInvalidBTCNet")
)

// putUint64BE puts an uint64 value in a big endian byte array.
func putUint64BE(b *[8]byte, v uint64) {
	b[0] = byte(v >> 56)
	b[1] = byte(v >> 48)
	b[2] = byte(v >> 40)
	b[3] = byte(v >> 32)
	b[4] = byte(v >> 24)
	b[5] = byte(v >> 16)
	b[6] = byte(v >> 8)
	b[7] = byte(v)
}

// getUint64BE returns an uint64 value from a big endian byte array.
func getUint64BE(b *[8]byte) uint64 {
	var v uint64
	v += uint64(b[0]) << 56
	v += uint64(b[1]) << 48
	v += uint64(b[2]) << 40
	v += uint64(b[3]) << 32
	v += uint64(b[4]) << 24
	v += uint64(b[5]) << 16
	v += uint64(b[6]) << 8
	v += uint64(b[7])
	return v
}

// EncodeOpReturn encodes the given Anchor to OP_RETURN.
func EncodeOpReturn(a *Anchor) [80]byte {
	var opRet [80]byte

	// Set signature.
	opRet[0] = 0x42 // B
	opRet[1] = 0x42 // B
	opRet[2] = 0x63 // c
	opRet[3] = 0x31 // 1

	// Set Version and BTCNet.
	opRet[4] = a.Version
	opRet[5] = byte(a.BTCNet)

	// Set timestamp.
	var ts [8]byte
	putUint64BE(&ts, uint64(a.Timestamp.Unix()))
	copy(opRet[8:16], ts[0:8])

	// Set BBc1DomainID and BBc1TransactionID.
	copy(opRet[16:48], a.BBc1DomainID[:])
	copy(opRet[48:80], a.BBc1TransactionID[:])

	return opRet
}

// DecodeOpReturn decodes the given bytes array to Anchor.
func DecodeOpReturn(b [80]byte) (*Anchor, error) {
	// Check signature.
	if b[0] != 0x42 || b[1] != 0x42 || b[2] != 0x63 || b[3] != 0x31 {
		return nil, fmt.Errorf("%w (AnchorSignature: %s)", ErrInvalidSignature, b[0:4])
	}

	var a Anchor

	// Check Version and BTCNet.
	a.Version = b[4]
	if _, ok := validAnchorVersions[a.Version]; !ok {
		return nil, fmt.Errorf("%w (AnchorVersion: %v)", ErrInvalidVersion, a.Version)
	}
	a.BTCNet = BTCNet(b[5])
	if n := a.BTCNet.String(); n == "" {
		return nil, fmt.Errorf("%w (AnchorBTCNet: %v)", ErrInvalidBTCNet, a.BTCNet)
	}

	// Copy timestamp
	var ts [8]byte
	copy(ts[0:8], b[8:16])
	a.Timestamp = time.Unix(int64(getUint64BE(&ts)), 0)

	// Copy BBc1DomainID and BBc1TransactionID
	var d, t [32]byte
	copy(d[:], b[16:48])
	copy(t[:], b[48:80])
	a.BBc1DomainID = d
	a.BBc1TransactionID = t

	return &a, nil
}

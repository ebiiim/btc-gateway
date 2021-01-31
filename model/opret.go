package model

import (
	"errors"
	"fmt"
)

var (
	ErrCannotDecodeOpReturn = errors.New("ErrCannotDecodeOpReturn")
)

func EncodeOpReturn(a *Anchor) [80]byte {
	var opRet [80]byte
	opRet[0] = 0x42 // B
	opRet[1] = 0x42 // B
	opRet[2] = 0x63 // c
	opRet[3] = 0x31 // 1
	opRet[4] = a.Version
	opRet[5] = byte(a.BTCNet)
	copy(opRet[16:48], a.BBc1DomainID[:])
	copy(opRet[48:80], a.BBc1TransactionID[:])
	return opRet
}

func DecodeOpReturn(b [80]byte) (*Anchor, error) {
	// Check signature.
	if b[0] != 0x42 || b[1] != 0x42 || b[2] != 0x63 || b[3] != 0x31 {
		return nil, fmt.Errorf("%w (invalid signature %s)", ErrCannotDecodeOpReturn, b[0:4])
	}

	var a Anchor

	// Check Version and BTCNet.
	a.Version = b[4]
	if _, ok := validAnchorVersions[a.Version]; !ok {
		return nil, fmt.Errorf("%w (invalid anchor version %v)", ErrCannotDecodeOpReturn, a.Version)
	}
	a.BTCNet = BTCNet(b[5])
	if _, ok := BTCNetNames[a.BTCNet]; !ok {
		return nil, fmt.Errorf("%w (invalid bitcoin network %v)", ErrCannotDecodeOpReturn, a.BTCNet)
	}

	// Copy BBc1DomainID and BBc1TransactionID
	var d, t [32]byte
	copy(d[:], b[16:48])
	copy(t[:], b[48:80])
	a.BBc1DomainID = d
	a.BBc1TransactionID = t

	return &a, nil
}

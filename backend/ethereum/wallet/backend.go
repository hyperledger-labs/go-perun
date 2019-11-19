// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package wallet

import (
	"io"

	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	perun "perun.network/go-perun/wallet"
	"perun.network/go-perun/wire"
)

// Backend implements the utility interface defined in the wallet package.
type Backend struct{}

// SignatureLength length of a signature in byte.
// ref https://godoc.org/github.com/ethereum/go-ethereum/crypto/secp256k1#Sign
// ref https://github.com/ethereum/go-ethereum/blob/54b271a86dd748f3b0bcebeaf678dc34e0d6177a/crypto/signature_cgo.go#L66
const SignatureLength = 65

// compile-time check that the ethereum backend implements the perun backend
var _ perun.Backend = (*Backend)(nil)

// NewAddressFromString creates a new address from a string.
func (h *Backend) NewAddressFromString(s string) (perun.Address, error) {
	addr, err := common.NewMixedcaseAddressFromString(s)
	if err != nil {
		return nil, errors.Wrap(err, "parsing address from string")
	}
	return &Address{addr.Address()}, nil
}

// NewAddressFromBytes creates a new address from a byte array.
func (h *Backend) NewAddressFromBytes(data []byte) (perun.Address, error) {
	if len(data) != common.AddressLength {
		return nil, errors.Errorf("could not create address from bytes of length: %d", len(data))
	}
	return &Address{common.BytesToAddress(data)}, nil
}

func (h *Backend) DecodeAddress(r io.Reader) (perun.Address, error) {
	addr := new(Address)
	return addr, addr.Decode(r)
}

// DecodeSig reads a []byte with length of an ethereum signature
func (*Backend) DecodeSig(r io.Reader) (perun.Sig, error) {
	buf := make(perun.Sig, SignatureLength)
	return buf, wire.Decode(r, &buf)
}

// VerifySignature verifies if a signature was made by this account.
func (*Backend) VerifySignature(msg []byte, sig perun.Sig, a perun.Address) (bool, error) {
	hash := crypto.Keccak256(msg)
	pk, err := crypto.SigToPub(hash, sig)
	if err != nil {
		return false, err
	}
	addr := crypto.PubkeyToAddress(*pk)
	return a.Equals(&Address{addr}), nil
}

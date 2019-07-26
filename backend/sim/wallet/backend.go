// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"io"

	perun "perun.network/go-perun/wallet"
)

var curve = elliptic.P256()

// Backend implements the utility interface defined in the wallet package.
type Backend struct{}

// NewAddressFromString creates a new address from a string.
func (h *Backend) NewAddressFromString(s string) (perun.Address, error) {
	addr := Address(s)
	return &addr, nil
}

// NewAddressFromBytes creates a new address from a byte array.
func (h *Backend) NewAddressFromBytes(data []byte) (perun.Address, error) {
	addr := Address(data)
	return &addr, nil
}

// DecodeAddress decodes an address from an io.Reader.
func (h *Backend) DecodeAddress(r io.Reader) (perun.Address, error) {
	addr := new(Address)
	return addr, addr.Decode(r)
}

// VerifySignature verifies if a signature was made by this account.
func (h *Backend) VerifySignature(msg, sig []byte, a perun.Address) (bool, error) {
	pubKey := addressToPubKey(a.(*Address))
	r, s, err := deserializeSignature(sig)
	if err != nil {
		return false, err
	}
	return ecdsa.Verify(pubKey, hash(msg), r, s), nil
}

func hash(msg []byte) []byte {
	hash := sha256.Sum256(msg)
	return hash[:]
}

func addressToPubKey(addr *Address) *ecdsa.PublicKey {
	x, y := elliptic.Unmarshal(curve, *addr)
	return &ecdsa.PublicKey{
		X:     x,
		Y:     y,
		Curve: curve,
	}
}

func pubKeyToAddress(pub *ecdsa.PublicKey) Address {
	return elliptic.Marshal(curve, pub.X, pub.Y)
}

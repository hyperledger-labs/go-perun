// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package test // import "perun.network/go-perun/wallet/test"

import (
	"encoding/hex"
	"io"
	"math/rand"

	"github.com/pkg/errors"

	"perun.network/go-perun/wallet"
)

// The DefaultBackend generates random [32]byte addresses.
type DefaultBackend struct{}

var _ Backend = new(DefaultBackend)

func (DefaultBackend) NewRandomAddress(rng *rand.Rand) wallet.Address {
	var a Address
	rng.Read(a[:])
	return &a
}

// Address is a [32]byte implementing the wallet.Address interface.
type Address [32]byte

var _ wallet.Address = new(Address)

// Decode reads an object from a stream.
// If the stream fails, the underlying error is returned.
// Returns an error if the stream's data is invalid.
func (a *Address) Decode(r io.Reader) error {
	_, err := io.ReadFull(r, a[:])
	return err
}

// Encode writes an object to a stream.
// If the stream fails, the underyling error is returned.
func (a Address) Encode(w io.Writer) error {
	_, err := w.Write(a[:])
	return err
}

// Bytes should return the representation of the address as byte slice.
func (a Address) Bytes() []byte {
	return a[:]
}

// String returns the hex encoding of the address.
func (a Address) String() string {
	return hex.EncodeToString(a[:])
}

// Equals checks the equality of two addresses.
func (a Address) Equals(b wallet.Address) bool {
	return a == *b.(*Address)
}

// The DefaultWalletBackend is a wallet.Backend implementation that fits to the
// test.Address type. Use it in your tests by setting
// `wallet.SetBackend(new(test.DefaultWalletBackend))`
type DefaultWalletBackend struct{}

var _ wallet.Backend = new(DefaultWalletBackend)

// NewAddressFromString creates a new address from the natural string representation of this blockchain.
func (b DefaultWalletBackend) NewAddressFromString(s string) (wallet.Address, error) {
	var addr Address
	bytes, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	} else if len(bytes) != 32 {
		return nil, errors.New("Wrong length of hex string, must represent 32 bytes")
	}

	copy(addr[:], bytes)
	return &addr, nil
}

// NewAddressFromBytes creates a new address from a byte array.
func (b DefaultWalletBackend) NewAddressFromBytes(data []byte) (wallet.Address, error) {
	var addr Address
	if len(data) != 32 {
		return nil, errors.New("Wrong length of data, must be 32 bytes")
	}

	copy(addr[:], data)
	return &addr, nil
}

// DecodeAddress reads and decodes an address from an io.Writer
func (b DefaultWalletBackend) DecodeAddress(r io.Reader) (wallet.Address, error) {
	var a Address
	return &a, a.Decode(r)
}

// VerifySignature not implemented for wallet backend of wallet/test default
// backend.
func (b DefaultWalletBackend) VerifySignature(msg []byte, sign []byte, a wallet.Address) (bool, error) {
	panic("not implemented")
}

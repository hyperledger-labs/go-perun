// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package wallet // import "perun.network/go-perun/backend/sim/wallet"

import (
	"bufio"
	"bytes"
	"crypto/ecdsa"
	"io"
	"math/big"

	"github.com/pkg/errors"

	"perun.network/go-perun/log"
	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wire"
)

// Address represents a simulated address.
type Address ecdsa.PublicKey

// compile time check that we implement the perun Address interface
var _ wallet.Address = (*Address)(nil)

// NewRandomAddress creates a new address using the randomness
// provided by rng
func NewRandomAddress(rng io.Reader) *Address {
	privateKey, err := ecdsa.GenerateKey(curve, rng)

	if err != nil {
		log.Panicf("Creation of account failed with error", err)
	}

	return &Address{Curve: privateKey.Curve, X: privateKey.X, Y: privateKey.Y}
}

// Bytes converts this address to bytes.
func (a Address) Bytes() []byte {
	// Serialize the Address into a buffer and return the buffers bytes
	buff := new(bytes.Buffer)
	w := bufio.NewWriter(buff)
	if err := a.Encode(w); err != nil {
		log.Panic("address encode error ", err)
	}
	if err := w.Flush(); err != nil {
		log.Panic("bufio flush ", err)
	}

	return buff.Bytes()
}

// String converts this address to a string.
func (a Address) String() string {
	return string(a.Bytes())
}

// Equals checks the equality of two addresses.
func (a Address) Equals(addr wallet.Address) bool {
	acc, ok := addr.(*Address)
	if !ok {
		log.Panic("Passed non wrong address type to Equals")
	}

	return (a.X == acc.X) && (a.Y == acc.Y)
}

// Encode encodes this address into an io.Writer. Part of the
// go-perun/pkg/io.Serializable interface.
func (a *Address) Encode(w io.Writer) error {
	if err := (wire.BigInt{Int: a.X}.Encode(w)); err != nil {
		return errors.Wrap(err, "address encode error")
	}
	if err := (wire.BigInt{Int: a.Y}.Encode(w)); err != nil {
		return errors.Wrap(err, "address encode error")
	}
	// Dont sezialize the curve since its constant

	return nil
}

// Decode decodes an address from an io.Reader. Part of the
// go-perun/pkg/io.Serializable interface.
func (a *Address) Decode(r io.Reader) error {
	var X, Y wire.BigInt

	if err := X.Decode(r); err != nil {
		return errors.Wrap(err, "address decode error")
	}
	if err := Y.Decode(r); err != nil {
		return errors.Wrap(err, "address decode error")
	}

	a.X = new(big.Int).SetBytes(X.Bytes())
	a.Y = new(big.Int).SetBytes(Y.Bytes())
	a.Curve = curve

	return nil
}

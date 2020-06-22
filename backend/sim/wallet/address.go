// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package wallet

import (
	"bufio"
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"io"
	"math/big"

	"github.com/pkg/errors"

	"perun.network/go-perun/log"
	perunio "perun.network/go-perun/pkg/io"
	"perun.network/go-perun/wallet"
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

	return &Address{
		Curve: privateKey.Curve,
		X:     privateKey.X,
		Y:     privateKey.Y,
	}
}

// Bytes converts this address to bytes.
func (a *Address) Bytes() []byte {
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

// ByteArray converts an address into a 64-byte array. The returned array
// consists of two 32-byte chunks representing the public key's X and Y values.
func (a *Address) ByteArray() (data [64]byte) {
	xb := a.X.Bytes()
	yb := a.Y.Bytes()

	// Left-pad with 0 bytes.
	copy(data[32-len(xb):32], xb)
	copy(data[64-len(yb):64], yb)

	return data
}

// String converts this address to a human-readable string.
func (a *Address) String() string {
	// Encode the address directly instead of using Address.Bytes() because
	// * some addresses may have a very short encoding, e.g., the null address,
	// * the Address.Bytes() output may contain encoding information, e.g., the
	//   length.
	bs := make([]byte, 4)
	copy(bs, a.X.Bytes())

	return "0x" + hex.EncodeToString(bs)
}

// Equals checks the equality of two addresses.
func (a *Address) Equals(addr wallet.Address) bool {
	b, ok := addr.(*Address)
	if !ok {
		log.Panic("Equals(): wrong address type")
	}

	return (a.X.Cmp(b.X) == 0) && (a.Y.Cmp(b.Y) == 0)
}

// Encode encodes this address into an io.Writer. Part of the
// go-perun/pkg/io.Serializer interface.
func (a *Address) Encode(w io.Writer) error {
	data := a.ByteArray()
	return perunio.Encode(w, data[:])
}

// Decode decodes an address from an io.Reader. Part of the
// go-perun/pkg/io.Serializer interface.
func (a *Address) Decode(r io.Reader) error {
	data := make([]byte, 64)
	if err := perunio.Decode(r, &data); err != nil {
		return errors.WithMessage(err, "decoding address")
	}
	a.X = new(big.Int).SetBytes(data[:32])
	a.Y = new(big.Int).SetBytes(data[32:])
	a.Curve = curve

	return nil
}

// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package wallet

import (
	"io"
)

// backend is set to the global wallet backend. Must not be set directly but
// through importing the needed backend.
var backend Backend

// Backend provides useful methods for this blockchain.
type Backend interface {
	// DecodeAddress reads and decodes an address from an io.Writer
	DecodeAddress(io.Reader) (Address, error)

	// DecodeSig reads a signature from the provided stream. It is needed for
	// decoding of wire messages.
	DecodeSig(io.Reader) (Sig, error)

	// VerifySignature verifies if this signature was signed by this address.
	// It should return an error iff the signature or message are malformed.
	// If the signature does not match the address it should return false, nil
	VerifySignature(msg []byte, sign Sig, a Address) (bool, error)
}

// SetBackend sets the global wallet backend. Must not be called directly but
// through importing the needed backend.
func SetBackend(b Backend) {
	if backend != nil {
		panic("wallet backend already set")
	}
	backend = b
}

// DecodeAddress calls DecodeAddress of the current backend
func DecodeAddress(r io.Reader) (Address, error) {
	return backend.DecodeAddress(r)
}

// DecodeSig calls DecodeSig of the current backend
func DecodeSig(r io.Reader) (Sig, error) {
	return backend.DecodeSig(r)
}

// VerifySignature calls VerifySignature of the current backend
func VerifySignature(msg []byte, sign Sig, a Address) (bool, error) {
	return backend.VerifySignature(msg, sign, a)
}

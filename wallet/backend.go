// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package wallet

import "io"

// Backend provides useful methods for this blockchain.
type Backend interface {
	// NewAddressFromString creates a new address from the natural string representation of this blockchain.
	NewAddressFromString(s string) (Address, error)

	// NewAddressFromBytes creates a new address from a byte array.
	NewAddressFromBytes(data []byte) (Address, error)

	// DecodeAddress reads and decodes an address from an io.Writer
	DecodeAddress(io.Reader) (Address, error)

	// VerifySignature verifies if this signature was signed by this address.
	// It should return an error iff the signature or message are malformed.
	// If the signature does not match the address it should return false, nil
	VerifySignature(msg, sign []byte, a Address) (bool, error)
}

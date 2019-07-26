// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package wallet

import (
	"bytes"
	"io"

	"github.com/pkg/errors"

	perun "perun.network/go-perun/wallet"
)

// AddressLength is the maximum length for an address.
const AddressLength = 65

// compile time check that we implement the perun Address interface
var _ perun.Address = (*Address)(nil)

// Address represents a simulated address.
type Address []byte

// Bytes converts this address to bytes.
func (a Address) Bytes() []byte {
	return a
}

// String converts this address to a string.
func (a Address) String() string {
	return string(a)
}

// Equals checks the equality of two addresses.
func (a Address) Equals(addr perun.Address) bool {
	return bytes.Equal(a, addr.Bytes())
}

// Encode encodes this address into a io.Writer. Part of the
// go-perun/pkg/io.Serializable interface.
func (a Address) Encode(w io.Writer) error {
	_, err := w.Write(a)
	return err
}

// Decode decodes an address from a io.Reader. Part of the
// go-perun/pkg/io.Serializable interface.
func (a *Address) Decode(r io.Reader) error {
	*a = make([]byte, AddressLength)
	_, err := io.ReadFull(r, *a)
	return errors.Wrap(err, "error decoding address")
}

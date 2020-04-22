// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package wallet

import (
	"bytes"
	"io"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"

	"perun.network/go-perun/wallet"
)

// compile time check that we implement the perun Address interface
var _ wallet.Address = (*Address)(nil)

// Address represents an ethereum address as a perun address.
type Address common.Address

// Bytes returns the address as a byte slice.
func (a *Address) Bytes() []byte {
	return (*common.Address)(a).Bytes()
}

// Encode encodes this address into a io.Writer. Part of the
// go-perun/pkg/io.Serializer interface.
func (a *Address) Encode(w io.Writer) error {
	_, err := w.Write(a.Bytes())
	return err
}

// Decode decodes an address from a io.Reader. Part of the
// go-perun/pkg/io.Serializer interface.
func (a *Address) Decode(r io.Reader) error {
	buf := make([]byte, common.AddressLength)
	_, err := io.ReadFull(r, buf)
	(*common.Address)(a).SetBytes(buf)
	return errors.Wrap(err, "error decoding address")
}

// String converts this address to a string.
func (a *Address) String() string {
	return (*common.Address)(a).String()
}

// Equals checks the equality of two addresses.
func (a *Address) Equals(addr wallet.Address) bool {
	_, ok := addr.(*Address)
	if !ok {
		panic("comparing ethereum address to address of different type")
	}
	return bytes.Equal(a.Bytes(), addr.Bytes())
}

// AsEthAddr is a helper function to convert an address interface back into an
// ethereum address.
func AsEthAddr(a wallet.Address) common.Address {
	return common.Address(*a.(*Address))
}

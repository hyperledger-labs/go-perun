// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package wallet

import (
	"io"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"

	perun "perun.network/go-perun/wallet"
)

// compile time check that we implement the perun Address interface
var _ perun.Address = (*Address)(nil)

// Address represents an ethereum address as a perun address.
type Address struct {
	common.Address
}

// Bytes returns the address as a byte slice.
func (a *Address) Bytes() []byte {
	return a.Address.Bytes()
}

// Encode encodes this address into a io.Writer. Part of the
// go-perun/pkg/io.Serializable interface.
func (a *Address) Encode(w io.Writer) error {
	_, err := w.Write(a.Address.Bytes())
	return err
}

// Decode decodes an address from a io.Reader. Part of the
// go-perun/pkg/io.Serializable interface.
func (a *Address) Decode(r io.Reader) error {
	buf := make([]byte, common.AddressLength)
	n, err := io.ReadAtLeast(r, buf, common.AddressLength)
	if err != nil {
		return errors.Wrap(err, "error decoding address")
	}
	if n != common.AddressLength {
		return errors.Errorf("error decoding address: read %d bytes", n)
	}
	a.Address.SetBytes(buf)
	return nil
}

// String converts this address to a string.
func (a *Address) String() string {
	return a.Address.String()
}

// Equals checks the equality of two addresses.
func (a *Address) Equals(addr perun.Address) bool {
	ethaddr, ok := addr.(*Address)
	if !ok {
		panic("comparing ethereum address to address of different type")
	}
	return [common.AddressLength]byte(a.Address) == [common.AddressLength]byte(ethaddr.Address)
}

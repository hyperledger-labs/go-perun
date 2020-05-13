// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package wallet

import (
	"fmt"
	stdio "io"

	"github.com/pkg/errors"

	"perun.network/go-perun/pkg/io"
	"perun.network/go-perun/wire"
)

// Address represents a identifier used in a cryptocurrency.
// It is dependent on the currency and needs to be implemented for every blockchain.
type Address interface {
	io.Serializer
	// Bytes should return the representation of the address as byte slice.
	Bytes() []byte
	// String converts this address to a string
	fmt.Stringer
	// Equals checks the equality of two addresses
	Equals(Address) bool
}

// IndexOfAddr returns the index of the given address in the address slice,
// or -1 if it is not part of the slice.
func IndexOfAddr(addrs []Address, addr Address) int {
	for i, a := range addrs {
		if addr.Equals(a) {
			return i
		}
	}

	return -1
}

// Addresses is a helper type for encoding and decoding address slices in
// situations where the length of the slice is known.
type Addresses []Address

// AddressesWithLen is a helper type for encoding and decoding address slices
// of unknown length.
type AddressesWithLen []Address

// addressSliceLen is needed to break the import cycle with channel. It should
// be the same as channel.Index.
type addressSliceLen = uint16

// AddressDec is a helper type to decode single wallet addresses.
type AddressDec struct {
	Addr *Address
}

// Encode encodes a wallet address slice, the length of which is known to the
// following decode operation.
func (a Addresses) Encode(w stdio.Writer) error {
	for i, addr := range a {
		if err := addr.Encode(w); err != nil {
			return errors.WithMessagef(err, "encoding %d-th address", i)
		}
	}

	return nil
}

// Encode encodes a wallet address slice, the length of which is unknown to the
// following decode operation.
func (a AddressesWithLen) Encode(w stdio.Writer) error {
	return wire.Encode(w,
		addressSliceLen(len(a)),
		(Addresses)(a))
}

// Decode decodes a wallet address slice of known length. The slice has to be
// allocated to the correct size already.
func (a Addresses) Decode(r stdio.Reader) (err error) {
	for i := range a {
		a[i], err = DecodeAddress(r)
		if err != nil {
			return errors.WithMessagef(err, "decoding %d-th address", i)
		}
	}
	return nil
}

// Decode decodes a wallet address slice of unknown length.
func (a *AddressesWithLen) Decode(r stdio.Reader) (err error) {
	var parts addressSliceLen
	if err = wire.Decode(r, &parts); err != nil {
		return errors.WithMessage(err, "decoding count")
	}

	*a = make(AddressesWithLen, parts)
	return (*Addresses)(a).Decode(r)
}

// Decode decodes a single wallet address.
func (a AddressDec) Decode(r stdio.Reader) (err error) {
	*a.Addr, err = DecodeAddress(r)
	return err
}

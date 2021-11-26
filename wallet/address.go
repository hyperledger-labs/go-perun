// Copyright 2019 - See NOTICE file for copyright holders.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package wallet

import (
	"bytes"
	"encoding"
	"fmt"
	stdio "io"
	"strings"

	"github.com/pkg/errors"

	"perun.network/go-perun/pkg/io"
)

// Address represents a identifier used in a cryptocurrency.
// It is dependent on the currency and needs to be implemented for every blockchain.
type Address interface {
	// BinaryMarshaler marshals the blockchain specific address to binary
	// format (a byte array).
	encoding.BinaryMarshaler
	// BinaryUnmarshaler unmarshals the blockchain specific address from
	// binary format (a byte array).
	encoding.BinaryUnmarshaler

	// String converts this address to a string.
	fmt.Stringer
	// Equal returns wether the two addresses are equal. The implementation
	// must be equivalent to checking `Address.Cmp(Address) == 0`.
	Equal(Address) bool
	// Cmp compares the byte representation of two addresses. For `a.Cmp(b)`
	// returns -1 if a < b, 0 if a == b, 1 if a > b.
	Cmp(Address) int
}

// IndexOfAddr returns the index of the given address in the address slice,
// or -1 if it is not part of the slice.
func IndexOfAddr(addrs []Address, addr Address) int {
	for i, a := range addrs {
		if addr.Equal(a) {
			return i
		}
	}

	return -1
}

// AddressPredicate is a function for filtering Addresses.
type AddressPredicate = func(Address) bool

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

// AddrKey is a non-human readable representation of an `Address`.
// It can be compared and therefore used as a key in a map.
type AddrKey string

// Encode encodes a wallet address slice, the length of which is known to the
// following decode operation.
func (a Addresses) Encode(w stdio.Writer) error {
	for i, addr := range a {
		if err := io.Encode(w, addr); err != nil {
			return errors.WithMessagef(err, "encoding %d-th address", i)
		}
	}

	return nil
}

// Encode encodes a wallet address slice, the length of which is unknown to the
// following decode operation.
func (a AddressesWithLen) Encode(w stdio.Writer) error {
	return io.Encode(w,
		addressSliceLen(len(a)),
		(Addresses)(a))
}

// Decode decodes a wallet address slice of known length. The slice has to be
// allocated to the correct size already.
func (a Addresses) Decode(r stdio.Reader) (err error) {
	for i := range a {
		a[i] = NewAddress()
		err = io.Decode(r, a[i])
		if err != nil {
			return errors.WithMessagef(err, "decoding %d-th address", i)
		}
	}
	return nil
}

// Decode decodes a wallet address slice of unknown length.
func (a *AddressesWithLen) Decode(r stdio.Reader) (err error) {
	var parts addressSliceLen
	if err = io.Decode(r, &parts); err != nil {
		return errors.WithMessage(err, "decoding count")
	}

	*a = make(AddressesWithLen, parts)
	return (*Addresses)(a).Decode(r)
}

// Decode decodes a single wallet address.
func (a AddressDec) Decode(r stdio.Reader) (err error) {
	*a.Addr = NewAddress()
	err = io.Decode(r, *a.Addr)
	return err
}

// Key returns the `AddrKey` corresponding to the passed `Address`.
// The `Address` can be retrieved with `FromKey`.
// Panics when the `Address` can't be encoded.
func Key(a Address) AddrKey {
	var buff strings.Builder
	if err := io.Encode(&buff, a); err != nil {
		panic("Could not encode address in AddrKey: " + err.Error())
	}
	return AddrKey(buff.String())
}

// FromKey returns the `Address` corresponding to the passed `AddrKey`
// created by `Key`.
// Panics when the `Address` can't be decoded.
func FromKey(k AddrKey) Address {
	a := NewAddress()
	err := io.Decode(bytes.NewBuffer([]byte(k)), a)
	if err != nil {
		panic("Could not decode address in FromKey: " + err.Error())
	}
	return a
}

// Equal Returns whether the passed `Address` has the same key as the
// receiving `AddrKey`.
func (k AddrKey) Equal(a Address) bool {
	return k == Key(a)
}

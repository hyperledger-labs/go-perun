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

package wire

import (
	"encoding"
	stdio "io"
	"strings"

	"github.com/pkg/errors"
	"perun.network/go-perun/wire/perunio"
)

var (
	_ perunio.Serializer = (*Addresses)(nil)
	_ perunio.Serializer = (*AddressesWithLen)(nil)
)

// Address is a Perun node's network address, which is used as a permanent
// identity within the Perun peer-to-peer network. For now, it is based on type
// wallet.Address.
type Address interface {
	// BinaryMarshaler marshals the address to binary.
	encoding.BinaryMarshaler
	// BinaryUnmarshaler unmarshals an address from binary.
	encoding.BinaryUnmarshaler
	// Equal returns wether the two addresses are equal.
	Equal(Address) bool
	// Cmp compares the byte representation of two addresses. For `a.Cmp(b)`
	// returns -1 if a < b, 0 if a == b, 1 if a > b.
	Cmp(Address) int
	// Verify verifies a message signature.
	// It returns an error if the signature is invalid.
	Verify(msg []byte, sig []byte) error
}

// Addresses is a helper type for encoding and decoding address slices in
// situations where the length of the slice is known.
type Addresses []Address

// AddressesWithLen is a helper type for encoding and decoding address slices
// of unknown length.
type AddressesWithLen []Address

type addressSliceLen = uint16

// Encode encodes wire addresses.
func (a Addresses) Encode(w stdio.Writer) error {
	for i, addr := range a {
		if err := perunio.Encode(w, addr); err != nil {
			return errors.WithMessagef(err, "encoding %d-th address", i)
		}
	}

	return nil
}

// Encode encodes wire addresses with length.
func (a AddressesWithLen) Encode(w stdio.Writer) error {
	return perunio.Encode(w,
		addressSliceLen(len(a)),
		(Addresses)(a))
}

// Decode decodes wallet addresses.
func (a Addresses) Decode(r stdio.Reader) error {
	for i := range a {
		a[i] = NewAddress()
		err := perunio.Decode(r, a[i])
		if err != nil {
			return errors.WithMessagef(err, "decoding %d-th address", i)
		}
	}
	return nil
}

// Decode decodes a wallet address slice of unknown length.
func (a *AddressesWithLen) Decode(r stdio.Reader) error {
	var n addressSliceLen
	if err := perunio.Decode(r, &n); err != nil {
		return errors.WithMessage(err, "decoding count")
	}

	*a = make(AddressesWithLen, n)
	return (*Addresses)(a).Decode(r)
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

// AddrKey is a non-human readable representation of an `Address`.
// It can be compared and therefore used as a key in a map.
type AddrKey string

// Key returns the `AddrKey` corresponding to the passed `Address`.
// The `Address` can be retrieved with `FromKey`.
// Panics when the `Address` can't be encoded.
func Key(a Address) AddrKey {
	var buff strings.Builder
	if err := perunio.Encode(&buff, a); err != nil {
		panic("Could not encode address in AddrKey: " + err.Error())
	}
	return AddrKey(buff.String())
}

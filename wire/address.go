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
	"sort"
	"strings"

	"perun.network/go-perun/wallet"

	"github.com/pkg/errors"
	"perun.network/go-perun/wire/perunio"
)

var (
	_ perunio.Serializer = (*AddressDecMap)(nil)
	_ perunio.Serializer = (*AddressMapArray)(nil)
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

// AddressMapArray is a helper type for encoding and decoding address maps.
type AddressMapArray []map[wallet.BackendID]Address

// AddressDecMap is a helper type for encoding and decoding arrays of address maps.
type AddressDecMap map[wallet.BackendID]Address

// Encode encodes first the length of the map,
// then all Addresses and their key in the map.
func (a AddressDecMap) Encode(w stdio.Writer) error {
	length := int32(len(a))
	if err := perunio.Encode(w, length); err != nil {
		return errors.WithMessage(err, "encoding map length")
	}
	for i, addr := range a {
		if err := perunio.Encode(w, int32(i)); err != nil {
			return errors.WithMessage(err, "encoding map index")
		}
		if err := perunio.Encode(w, addr); err != nil {
			return errors.WithMessagef(err, "encoding %d-th address map entry", i)
		}
	}
	return nil
}

// Encode encodes first the length of the array,
// then all AddressDecMaps in the array.
func (a AddressMapArray) Encode(w stdio.Writer) error {
	length := int32(len(a))
	if err := perunio.Encode(w, length); err != nil {
		return errors.WithMessage(err, "encoding array length")
	}
	for i, addr := range a {
		if err := perunio.Encode(w, AddressDecMap(addr)); err != nil {
			return errors.WithMessagef(err, "encoding %d-th address array entry", i)
		}
	}
	return nil
}

// Decode decodes the map length first, then all Addresses and their key in the map.
func (a *AddressDecMap) Decode(r stdio.Reader) (err error) {
	var mapLen int32
	if err := perunio.Decode(r, &mapLen); err != nil {
		return errors.WithMessage(err, "decoding map length")
	}
	*a = make(map[wallet.BackendID]Address, mapLen)
	for i := 0; i < int(mapLen); i++ {
		var idx int32
		if err := perunio.Decode(r, &idx); err != nil {
			return errors.WithMessage(err, "decoding map index")
		}
		addr := NewAddress()
		if err := perunio.Decode(r, addr); err != nil {
			return errors.WithMessagef(err, "decoding %d-th address map entry", i)
		}
		(*a)[wallet.BackendID(idx)] = addr
	}
	return nil
}

// Decode decodes the array length first, then all AddressDecMaps in the array.
// Decode decodes the array length first, then all AddressDecMaps in the array.
func (a *AddressMapArray) Decode(r stdio.Reader) (err error) {
	var mapLen int32
	if err := perunio.Decode(r, &mapLen); err != nil {
		return errors.WithMessage(err, "decoding array length")
	}
	*a = make([]map[wallet.BackendID]Address, mapLen)
	for i := 0; i < int(mapLen); i++ {
		if err := perunio.Decode(r, (*AddressDecMap)(&(*a)[i])); err != nil {
			return errors.WithMessagef(err, "decoding %d-th address map entry", i)
		}
	}
	return nil
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

// IndexOfAddrs returns the index of the given address in the address slice,
// or -1 if it is not part of the slice.
func IndexOfAddrs(addrs []map[wallet.BackendID]Address, addr map[wallet.BackendID]Address) int {
	for i, a := range addrs {
		if addrEqual(a, addr) {
			return i
		}
	}

	return -1
}

func addrEqual(a, b map[wallet.BackendID]Address) bool {
	if len(a) != len(b) {
		return false
	}
	for i, addr := range a {
		if !addr.Equal(b[i]) {
			return false
		}
	}
	return true
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

// Keys returns the `AddrKey` corresponding to the passed `map[int]Address`.
func Keys(addressMap map[wallet.BackendID]Address) AddrKey {
	var indexes []int //nolint:prealloc
	for i := range addressMap {
		indexes = append(indexes, int(i))
	}
	sort.Ints(indexes)

	keyParts := make([]string, len(indexes))
	for i, index := range indexes {
		key := Key(addressMap[wallet.BackendID(index)])
		keyParts[i] = string(key)
	}
	return AddrKey(strings.Join(keyParts, "|"))
}

// AddressMapfromAccountMap converts a map of accounts to a map of addresses.
func AddressMapfromAccountMap(accs map[wallet.BackendID]Account) map[wallet.BackendID]Address {
	addresses := make(map[wallet.BackendID]Address)
	for id, a := range accs {
		addresses[id] = a.Address()
	}
	return addresses
}

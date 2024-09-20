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

	"perun.network/go-perun/log"
	"perun.network/go-perun/wire/perunio"
)

// BackendID is a unique identifier for a backend.
type BackendID int

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
	// BackendID returns the id of the backend that created this address.
	BackendID() BackendID
}

func (a *BackendID) Equal(b BackendID) bool {
	return *a == b
}

// IndexOfAddr returns the index of the given address in the address slice,
// or -1 if it is not part of the slice.
func IndexOfAddr(addrs []map[BackendID]Address, addr Address) int {
	for i, as := range addrs {
		for _, a := range as {
			if a.Equal(addr) {
				return i
			}
		}
	}

	return -1
}

// IndexOfAddr returns the index of the given address in the address slice,
// or -1 if it is not part of the slice.
func IndexOfAddrs(addrs []map[BackendID]Address, addr map[BackendID]Address) int {
	for i, as := range addrs {
		for j, a := range as {
			if a.Equal(addr[j]) {
				return i
			}
		}
	}

	return -1
}

// CloneAddress returns a clone of an Address using its binary marshaling
// implementation. It panics if an error occurs during binary (un)marshaling.
func CloneAddress(a Address) Address {
	data, err := a.MarshalBinary()
	if err != nil {
		log.WithError(err).Panic("error binary-marshaling Address")
	}

	clone := NewAddress(a.BackendID())
	if err := clone.UnmarshalBinary(data); err != nil {
		log.WithError(err).Panic("error binary-unmarshaling Address")
	}
	return clone
}

// CloneAddresses returns a clone of a slice of Addresses using their binary
// marshaling implementation. It panics if an error occurs during binary
// (un)marshaling.
func CloneAddresses(as []Address) []Address {
	clones := make([]Address, 0, len(as))
	for _, a := range as {
		clones = append(clones, CloneAddress(a))
	}
	return clones
}

// CloneAddressesMap returns a clone of a map of Addresses using their binary
// marshaling implementation. It panics if an error occurs during binary
// (un)marshaling.
func CloneAddressesMap(as map[BackendID]Address) map[BackendID]Address {
	clones := make(map[BackendID]Address)
	for i, a := range as {
		clones[i] = CloneAddress(a)
	}
	return clones
}

// AddressMapArray is a helper type for encoding and decoding arrays of address maps.
type AddressMapArray struct {
	Addr []map[BackendID]Address
}

// AddressDecMap is a helper type for encoding and decoding address maps.
type AddressDecMap map[BackendID]Address

// AddrKey is a non-human readable representation of an `Address`.
// It can be compared and therefore used as a key in a map.
type AddrKey string

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
	length := int32(len(a.Addr))
	if err := perunio.Encode(w, length); err != nil {
		return errors.WithMessage(err, "encoding array length")
	}
	for i, addr := range a.Addr {
		if err := perunio.Encode(w, (*AddressDecMap)(&addr)); err != nil {
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
	*a = make(map[BackendID]Address, mapLen)
	for i := 0; i < int(mapLen); i++ {
		var idx int32
		if err := perunio.Decode(r, &idx); err != nil {
			return errors.WithMessage(err, "decoding map index")
		}
		addr := NewAddress(BackendID(idx))
		if err := perunio.Decode(r, addr); err != nil {
			return errors.WithMessagef(err, "decoding %d-th address map entry", i)
		}
		(*a)[BackendID(idx)] = addr
	}
	return nil
}

// Decode decodes the array length first, then all AddressDecMaps in the array.
func (a *AddressMapArray) Decode(r stdio.Reader) (err error) {
	var mapLen int32
	if err := perunio.Decode(r, &mapLen); err != nil {
		return errors.WithMessage(err, "decoding array length")
	}
	a.Addr = make([]map[BackendID]Address, mapLen)
	for i := 0; i < int(mapLen); i++ {
		if err := perunio.Decode(r, (*AddressDecMap)(&a.Addr[i])); err != nil {
			return errors.WithMessagef(err, "decoding %d-th address map entry", i)
		}
	}
	return nil
}

// Key returns the `AddrKey` corresponding to the passed `Address`.
// The `Address` can be retrieved with `FromKey`.
// Panics when the `Address` can't be encoded.
func Key(a Address) AddrKey {
	var buff strings.Builder
	if err := perunio.Encode(&buff, uint32(a.BackendID())); err != nil {
		panic("Could not encode id in AddrKey: " + err.Error())
	}
	if err := perunio.Encode(&buff, a); err != nil {
		panic("Could not encode address in AddrKey: " + err.Error())
	}
	return AddrKey(buff.String())
}

// FromKey returns the `Address` corresponding to the passed `AddrKey`
// created by `Key`.
// Panics when the `Address` can't be decoded.
func FromKey(k AddrKey) Address {
	buff := bytes.NewBuffer([]byte(k))
	var id uint32
	err := perunio.Decode(buff, &id)
	a := NewAddress(BackendID(int(id)))
	err = perunio.Decode(buff, a)
	if err != nil {
		panic("Could not decode address in FromKey: " + err.Error())
	}
	return a
}

// Equal Returns whether the passed `Address` has the same key as the
// receiving `AddrKey`.
func (k AddrKey) Equal(a Address) bool {
	key := Key(a)
	return k == key
}

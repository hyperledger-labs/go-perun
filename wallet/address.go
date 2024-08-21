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
	"encoding/binary"
	"fmt"
	stdio "io"
	"strings"

	"github.com/pkg/errors"

	"perun.network/go-perun/log"
	"perun.network/go-perun/wire/perunio"
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
	// BackendID returns the id of the backend that created this address.
	BackendID() int
}

// IndexOfAddr returns the index of the given address in the address slice,
// or -1 if it is not part of the slice.
func IndexOfAddr(addrs []map[int]Address, addr map[int]Address) int {
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
func CloneAddressesMap(as map[int]Address) map[int]Address {
	clones := make(map[int]Address)
	for i, a := range as {
		clones[i] = CloneAddress(a)
	}
	return clones
}

// AddressMapArray is a helper type for encoding and decoding arrays of address maps.
type AddressMapArray struct {
	Addr []map[int]Address
}

// AddressDecMap is a helper type for encoding and decoding address maps.
type AddressDecMap map[int]Address

// AddrKey is a non-human readable representation of an `Address`.
// It can be compared and therefore used as a key in a map.
type AddrKey struct {
	key string
}

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
	*a = make(map[int]Address, mapLen)
	for i := 0; i < int(mapLen); i++ {
		var idx int32
		if err := perunio.Decode(r, &idx); err != nil {
			return errors.WithMessage(err, "decoding map index")
		}
		addr := NewAddress(int(idx))
		if err := perunio.Decode(r, addr); err != nil {
			return errors.WithMessagef(err, "decoding %d-th address map entry", i)
		}
		(*a)[int(idx)] = addr
	}
	return nil
}

// Decode decodes the array length first, then all AddressDecMaps in the array.
func (a *AddressMapArray) Decode(r stdio.Reader) (err error) {
	var mapLen int32
	if err := perunio.Decode(r, &mapLen); err != nil {
		return errors.WithMessage(err, "decoding array length")
	}
	a.Addr = make([]map[int]Address, mapLen)
	for i := 0; i < int(mapLen); i++ {
		if err := perunio.Decode(r, (*AddressDecMap)(&a.Addr[i])); err != nil {
			return errors.WithMessagef(err, "decoding %d-th address map entry", i)
		}
	}
	return nil
}

// Key returns the `AddrKey` corresponding to the passed `map[int]Address`.
// The `Address` can be retrieved with `FromKey`.
// Returns an error if the `map[int]Address` can't be encoded.
func Key(a map[int]Address) AddrKey {
	var buff strings.Builder
	// Encode the number of elements in the map first.
	length := int32(len(a)) // Using int32 to encode the length
	err := binary.Write(&buff, binary.BigEndian, length)
	if err != nil {
		log.Panic("could not encode map length in Key: ", err)

	}
	// Iterate over the map and encode each key-value pair.
	for id, addr := range a {
		if err := binary.Write(&buff, binary.BigEndian, int32(id)); err != nil {
			log.Panicf("could not encode map length in AddrKey: " + err.Error())
		}
		if err := perunio.Encode(&buff, addr); err != nil {
			log.Panicf("could not encode map[int]Address in AddrKey: " + err.Error())
		}
	}
	return AddrKey{buff.String()}
}

// FromKey returns the `map[int]Address` corresponding to the passed `AddrKey`
// created by `Key`.
// Returns an error if the `map[int]Address` can't be decoded.
func FromKey(k AddrKey) map[int]Address {
	buff := bytes.NewBuffer([]byte(k.key))
	var numElements int32

	// Manually decode the number of elements in the map.
	if err := binary.Read(buff, binary.BigEndian, &numElements); err != nil {
		log.Panicf("could not decode map length in FromKey: " + err.Error())
	}
	a := make(map[int]Address, numElements)
	// Decode each key-value pair and insert them into the map.
	for i := 0; i < int(numElements); i++ {
		var id int32
		if err := binary.Read(buff, binary.BigEndian, &id); err != nil {
			log.Panicf("could not decode map length in FromKey: " + err.Error())
		}
		addr := NewAddress(int(id))
		if err := perunio.Decode(buff, addr); err != nil {
			log.Panicf("could not decode map[int]Address in FromKey: " + err.Error())
		}
		a[int(id)] = addr
	}
	return a
}

// Equal Returns whether the passed `Address` has the same key as the
// receiving `AddrKey`.
func (k AddrKey) Equal(a map[int]Address) bool {
	key := Key(a)
	return k == key
}

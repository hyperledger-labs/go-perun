// Copyright 2022 - See NOTICE file for copyright holders.
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

package simple

import (
	"bytes"
	"math/rand"

	"perun.network/go-perun/wire"
)

// Address is a wire address.
type Address string

var _ wire.Address = NewAddress("")

// NewAddress returns a new address.
func NewAddress(host string) *Address {
	a := Address(host)
	return &a
}

// MarshalBinary marshals the address to binary.
func (a Address) MarshalBinary() ([]byte, error) {
	buf := make([]byte, len(a))
	copy(buf, []byte(a))
	return buf, nil
}

// UnmarshalBinary unmarshals an address from binary.
func (a *Address) UnmarshalBinary(data []byte) error {
	buf := make([]byte, len(data))
	copy(buf, data)
	*a = Address(buf)
	return nil
}

// Equal returns whether the two addresses are equal.
func (a Address) Equal(b wire.Address) bool {
	bTyped, ok := b.(*Address)
	if !ok {
		return false
	}
	return a == *bTyped
}

// Cmp compares the byte representation of two addresses. For `a.Cmp(b)`
// returns -1 if a < b, 0 if a == b, 1 if a > b.
func (a Address) Cmp(b wire.Address) int {
	bTyped, ok := b.(*Address)
	if !ok {
		panic("wrong type")
	}
	return bytes.Compare([]byte(a), []byte(*bTyped))
}

// NewRandomAddress returns a new random peer address.
func NewRandomAddress(rng *rand.Rand) *Address {
	const addrLen = 32
	l := rng.Intn(addrLen)
	d := make([]byte, l)
	if _, err := rng.Read(d); err != nil {
		panic(err)
	}

	a := Address(d)
	return &a
}

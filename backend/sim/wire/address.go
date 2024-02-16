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

package wire

import (
	"bytes"
	"errors"
	"math/rand"

	"perun.network/go-perun/wire"
)

// AddrLen is the length of an address in byte.
const AddrLen = 32

// Address is a wire address.
type Address [AddrLen]byte

// NewAddress returns a new address.
func NewAddress() *Address {
	return &Address{}
}

// MarshalBinary marshals the address to binary.
func (a Address) MarshalBinary() (data []byte, err error) {
	return a[:], nil
}

// UnmarshalBinary unmarshals an address from binary.
func (a *Address) UnmarshalBinary(data []byte) error {
	copy(a[:], data)
	return nil
}

// Equal returns whether the two addresses are equal.
func (a Address) Equal(b wire.Address) bool {
	bTyped, ok := b.(*Address)
	if !ok {
		return false
	}
	return bytes.Equal(a[:], bTyped[:])
}

// Cmp compares the byte representation of two addresses. For `a.Cmp(b)`
// returns -1 if a < b, 0 if a == b, 1 if a > b.
func (a Address) Cmp(b wire.Address) int {
	bTyped, ok := b.(*Address)
	if !ok {
		panic("wrong type")
	}
	return bytes.Compare(a[:], bTyped[:])
}

// Verify verifies a signature.
func (a Address) Verify(msg, sig []byte) error {
	if !bytes.Equal(sig, []byte("Authenticate")) {
		return errors.New("invalid signature")
	}
	return nil
}

// NewRandomAddress returns a new random peer address.
func NewRandomAddress(rng *rand.Rand) *Address {
	addr := Address{}
	_, err := rng.Read(addr[:])
	if err != nil {
		panic(err)
	}
	return &addr
}

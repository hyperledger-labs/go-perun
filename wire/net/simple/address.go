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
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"math/rand"

	"perun.network/go-perun/wire"
)

// Address is a wire address.
type Address struct {
	Name      string
	PublicKey *rsa.PublicKey // Public key for verifying signatures

}

var _ wire.Address = (*Address)(nil)

// NewAddress returns a new address.
func NewAddress(host string) *Address {
	return &Address{
		Name:      host,
		PublicKey: nil,
	}
}

// MarshalBinary marshals the address to binary.
func (a *Address) MarshalBinary() ([]byte, error) {
	// Initialize a buffer to hold the binary data
	var buf bytes.Buffer

	// Encode the length of the name string and the name itself
	nameLen := uint16(len(a.Name))
	if err := binary.Write(&buf, binary.BigEndian, nameLen); err != nil {
		return nil, err
	}
	if _, err := buf.WriteString(a.Name); err != nil {
		return nil, err
	}

	// If the public key is not nil, encode it using gob
	if a.PublicKey != nil {
		enc := gob.NewEncoder(&buf)
		if err := enc.Encode(a.PublicKey); err != nil {
			return nil, err
		}
	}

	// Return the binary representation
	return buf.Bytes(), nil
}

// UnmarshalBinary unmarshals an address from binary.
func (a *Address) UnmarshalBinary(data []byte) error {
	// Initialize a buffer with the binary data
	buf := bytes.NewReader(data)

	// Decode the length of the name string
	var nameLen uint16
	if err := binary.Read(buf, binary.BigEndian, &nameLen); err != nil {
		return err
	}

	// Read the name string from the buffer
	nameBytes := make([]byte, nameLen)
	if _, err := buf.Read(nameBytes); err != nil {
		return err
	}
	a.Name = string(nameBytes)

	// If there's remaining data, decode the public key using gob
	if buf.Len() > 0 {
		dec := gob.NewDecoder(buf)
		if err := dec.Decode(&a.PublicKey); err != nil {
			return err
		}
	}

	return nil
}

// Equal returns whether the two addresses are equal.
func (a *Address) Equal(b wire.Address) bool {
	bTyped, ok := b.(*Address)
	if !ok {
		return false
	}
	if a.PublicKey == nil {
		return a.Name == bTyped.Name && bTyped.PublicKey == nil
	}

	return a.Name == bTyped.Name && a.PublicKey.Equal(bTyped.PublicKey)
}

// Cmp compares the byte representation of two addresses.
func (a *Address) Cmp(b wire.Address) int {
	bTyped, ok := b.(*Address)
	if !ok {
		panic("wrong type")
	}
	if cmp := bytes.Compare([]byte(a.Name), []byte(bTyped.Name)); cmp != 0 {
		return cmp
	}

	bytesA, err := a.MarshalBinary()
	if err != nil {
		panic(err)
	}
	bytesB, err := bTyped.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return bytes.Compare(bytesA, bytesB)
}

// NewRandomAddress returns a new random peer address.
func NewRandomAddress(rng *rand.Rand) *Address {
	const addrLen = 32
	l := rng.Intn(addrLen)
	d := make([]byte, l)
	if _, err := rng.Read(d); err != nil {
		panic(err)
	}

	a := &Address{
		Name: string(d),
	}
	return a
}

// Verify verifies a message signature.
func (a *Address) Verify(msg []byte, sig []byte) error {
	hashed := sha256.Sum256(msg)
	err := rsa.VerifyPKCS1v15(a.PublicKey, crypto.SHA256, hashed[:], sig)
	if err != nil {
		return err
	}
	return nil
}

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
	"errors"
	"io"
)

// backend is set to the global wallet backend. Must not be set directly but
// through importing the needed backend.
var backend map[int]Backend

// Backend provides useful methods for this blockchain.
type Backend interface {
	// NewAddress returns a variable of type Address, which can be used
	// for unmarshalling an address from its binary representation.
	NewAddress() Address

	// DecodeSig reads a signature from the provided stream. It is needed for
	// decoding of wire messages.
	DecodeSig(io.Reader) (Sig, error)

	// VerifySignature verifies if this signature was signed by this address.
	// It should return an error iff the signature or message are malformed.
	// If the signature does not match the address it should return false, nil
	VerifySignature(msg []byte, sign Sig, a Address) (bool, error)
}

// SetBackend sets the global wallet backend. Must not be called directly but
// through importing the needed backend.
func SetBackend(b Backend, id int) {
	if backend == nil {
		backend = make(map[int]Backend)

	}
	if backend[id] != nil {
		panic("wallet backend already set")
	}
	backend[id] = b
}

// NewAddress returns a variable of type Address, which can be used
// for unmarshalling an address from its binary representation.
func NewAddress(id int) Address {
	return backend[id].NewAddress()
}

// DecodeSig calls DecodeSig of all Backends and returns an error if none return a valid signature.
func DecodeSig(r io.Reader) (Sig, error) {
	var errs []error
	for _, b := range backend {
		sig, err := b.DecodeSig(r)
		if err == nil {
			return sig, nil
		} else {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		// Join all errors into a single error message.
		return nil, errors.Join(errs...)
	}

	return nil, errors.New("no valid signature found")
}

// VerifySignature calls VerifySignature of the current backend.
func VerifySignature(msg []byte, sign Sig, a map[int]Address) (bool, error) {
	var errs []error
	for _, addr := range a {
		v, err := backend[addr.BackendID()].VerifySignature(msg, sign, addr)
		if err == nil && v {
			return v, nil
		} else {
			errs = append(errs, err)
		}
	}
	return false, errors.Join(errs...)
}

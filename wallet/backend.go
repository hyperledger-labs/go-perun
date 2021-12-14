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

// backend is set to the global wallet backend. Must not be set directly but
// through importing the needed backend.
var backend Backend

// Backend provides useful methods for this blockchain.
type Backend interface {
	// NewAddress returns a variable of type Address, which can be used
	// for unmarshalling an address from its binary representation.
	NewAddress() Address

	// NewSig returns a variable of type Sig, which can be used for
	// unmarshalling a signature from its binary representation.
	NewSig() Sig
}

// SetBackend sets the global wallet backend. Must not be called directly but
// through importing the needed backend.
func SetBackend(b Backend) {
	if backend != nil {
		panic("wallet backend already set")
	}
	backend = b
}

// NewAddress returns a variable of type Address, which can be used
// for unmarshalling an address from its binary representation.
func NewAddress() Address {
	return backend.NewAddress()
}

// NewSig returns a variable of type Sig, which can be used for unmarshalling a
// signature from its binary representation.
func NewSig() Sig {
	return backend.NewSig()
}

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
	"crypto/elliptic"

	"perun.network/go-perun/wallet"
)

var curve = elliptic.P256()

// Backend implements the utility interface defined in the wallet package.
type Backend struct{}

var _ wallet.Backend = new(Backend)

// NewAddress returns a variable of type Address, which can be used
// for unmarshalling an address from its binary representation.
func (b *Backend) NewAddress() wallet.Address {
	addr := Address{}
	return &addr
}

// NewSig returns a variable of type Sig, which can be used for unmarshalling a
// signature from its binary representation.
func (*Backend) NewSig() wallet.Sig {
	return &Sig{}
}

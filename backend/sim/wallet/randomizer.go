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
	"math/rand"

	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wallet/test"
)

// Randomizer provides random addresses and accounts.
type Randomizer struct{ Wallet }

var _ test.Randomizer = (*Randomizer)(nil)

func newRandomizer() *Randomizer { return &Randomizer{*NewWallet()} }

// NewRandomAddress creates a new random simulated address.
func (*Randomizer) NewRandomAddress(rng *rand.Rand) wallet.Address {
	return NewRandomAddress(rng)
}

// RandomWallet returns a fixed wallet that can be used to generate random
// accounts.
func (r *Randomizer) RandomWallet() test.Wallet {
	return r
}

// NewWallet returns a new, empty Wallet.
func (r *Randomizer) NewWallet() test.Wallet {
	return NewWallet()
}

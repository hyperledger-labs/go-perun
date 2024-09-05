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

package test

import (
	"math/rand"
	"perun.network/go-perun/wallet"

	"perun.network/go-perun/wire"
)

type (
	// NewRandomAddressFunc is a address randomizer function.
	NewRandomAddressFunc = func(*rand.Rand) wire.Address
	// NewRandomAccountFunc is a account randomizer function.
	NewRandomAccountFunc = func(*rand.Rand) wire.Account
)

var (
	newRandomAddress NewRandomAddressFunc
	newRandomAccount NewRandomAccountFunc
)

// SetNewRandomAddress sets the address randomizer function.
func SetNewRandomAddress(f NewRandomAddressFunc) {
	newRandomAddress = f
}

// SetNewRandomAccount sets the account randomizer function.
func SetNewRandomAccount(f NewRandomAccountFunc) {
	newRandomAccount = f
}

// NewRandomAddress returns a new random address.
func NewRandomAddress(rng *rand.Rand) map[wallet.BackendID]wire.Address {
	return map[wallet.BackendID]wire.Address{0: newRandomAddress(rng)}
}

// NewRandomAccount returns a new random account.
func NewRandomAccount(rng *rand.Rand) wire.Account {
	return newRandomAccount(rng)
}

// NewRandomAccountMap returns a new random account.
func NewRandomAccountMap(rng *rand.Rand) map[wallet.BackendID]wire.Account {
	return map[wallet.BackendID]wire.Account{0: newRandomAccount(rng)}
}

// NewRandomAddresses returns a slice of random peer addresses.
func NewRandomAddresses(rng *rand.Rand, n int) []map[wallet.BackendID]wire.Address {
	addresses := make([]map[wallet.BackendID]wire.Address, n)
	for i := range addresses {
		addresses[i] = NewRandomAddress(rng)
	}
	return addresses
}

// NewRandomAddressesMap returns a slice of random peer addresses.
func NewRandomAddressesMap(rng *rand.Rand, n int) []map[wallet.BackendID]wire.Address {
	addresses := make([]map[wallet.BackendID]wire.Address, n)
	for i := range addresses {
		addresses[i] = NewRandomAddress(rng)
	}
	return addresses
}

// NewRandomEnvelope returns an envelope around message m with random sender and
// recipient generated using randomness from rng.
func NewRandomEnvelope(rng *rand.Rand, m wire.Msg) *wire.Envelope {
	return &wire.Envelope{
		Sender:    NewRandomAddress(rng),
		Recipient: NewRandomAddress(rng),
		Msg:       m,
	}
}

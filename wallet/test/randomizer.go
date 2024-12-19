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
)

type (
	// Randomizer is a wallet testing backend. It should support the generation
	// of random addresses and accounts.
	Randomizer interface {
		// NewRandomAddress should return a new random address generated from the
		// passed rng.
		NewRandomAddress(*rand.Rand) wallet.Address

		// RandomWallet should return a fixed random wallet that is part of the
		// randomizer's state. It will be used to generate accounts with
		// NewRandomAccount.
		RandomWallet() Wallet

		// NewWallet should return a fresh, temporary Wallet that doesn't hold any
		// accounts yet.
		NewWallet() Wallet
	}

	// A Wallet is an extension of a wallet.Wallet to also generate random
	// accounts in test settings.
	Wallet interface {
		wallet.Wallet

		// NewRandomAccount should return an account generated from the passed rng.
		// The account should be stored and unlocked in the Wallet.
		NewRandomAccount(*rand.Rand) wallet.Account
	}
)

// randomizer is the currently set wallet testing randomizer. It is initially set to
// the default randomizer.
var randomizer map[wallet.BackendID]Randomizer

// SetRandomizer sets the wallet randomizer. It may be set multiple times.
func SetRandomizer(b Randomizer, bID wallet.BackendID) {
	if randomizer == nil {
		randomizer = make(map[wallet.BackendID]Randomizer)
	}
	if randomizer[bID] != nil {
		panic("wallet/test randomizer already set")
	}
	randomizer[bID] = b
}

// NewRandomAddress returns a new random address by calling the currently set
// wallet randomizer.
func NewRandomAddress(rng *rand.Rand, bID wallet.BackendID) wallet.Address {
	return randomizer[bID].NewRandomAddress(rng)
}

// NewRandomAddresses returns a new random address by calling the currently set
// wallet randomizer.
func NewRandomAddresses(rng *rand.Rand, bID wallet.BackendID) map[wallet.BackendID]wallet.Address {
	return map[wallet.BackendID]wallet.Address{bID: randomizer[bID].NewRandomAddress(rng)}
}

// RandomWallet returns the randomizer backend's wallet. All accounts created
// with NewRandomAccount can be found in this wallet.
func RandomWallet(bID wallet.BackendID) Wallet {
	return randomizer[bID].RandomWallet()
}

// NewRandomAccount returns a new random account by calling the currently set
// wallet randomizer. The account is generated from the randomizer wallet
// available via RandomWallet. It should already be unlocked.
func NewRandomAccount(rng *rand.Rand, bID wallet.BackendID) wallet.Account {
	return randomizer[bID].RandomWallet().NewRandomAccount(rng)
}

// NewWallet returns a fresh, temporary Wallet for testing purposes that doesn't
// hold any accounts yet. New random accounts can be generated using method
// NewRandomAccount.
func NewWallet(bID wallet.BackendID) Wallet {
	return randomizer[bID].NewWallet()
}

// NewRandomAccounts returns a slice of new random accounts
// by calling NewRandomAccount.
func NewRandomAccounts(rng *rand.Rand, n int, bID wallet.BackendID) ([]map[wallet.BackendID]wallet.Account, []map[wallet.BackendID]wallet.Address) {
	accs := make([]map[wallet.BackendID]wallet.Account, n)
	addrs := make([]map[wallet.BackendID]wallet.Address, n)
	for i := range accs {
		accs[i] = map[wallet.BackendID]wallet.Account{bID: NewRandomAccount(rng, bID)}
		addrs[i] = map[wallet.BackendID]wallet.Address{bID: accs[i][bID].Address()}
	}
	return accs, addrs
}

// NewRandomAccountMapSlice returns a slice of new random accounts map
// by calling NewRandomAccount.
func NewRandomAccountMapSlice(rng *rand.Rand, bID wallet.BackendID, n int) []map[wallet.BackendID]wallet.Account {
	accs := make([]map[wallet.BackendID]wallet.Account, n)
	for i := range accs {
		accs[i] = map[wallet.BackendID]wallet.Account{bID: NewRandomAccount(rng, bID)}
	}
	return accs
}

// NewRandomAccountMap returns a slice of new random accounts
// by calling NewRandomAccount.
func NewRandomAccountMap(rng *rand.Rand, bID wallet.BackendID) map[wallet.BackendID]wallet.Account {
	accs := make(map[wallet.BackendID]wallet.Account)
	accs[bID] = NewRandomAccount(rng, bID)
	return accs
}

// NewRandomAddressArray returns a slice of new random addresses.
func NewRandomAddressArray(rng *rand.Rand, n int, bID wallet.BackendID) []wallet.Address {
	addrs := make([]wallet.Address, n)
	for i := range addrs {
		addrs[i] = NewRandomAddress(rng, bID)
	}
	return addrs
}

// NewRandomAddressesMap returns a slice of new random address maps.
func NewRandomAddressesMap(rng *rand.Rand, n int, bID wallet.BackendID) []map[wallet.BackendID]wallet.Address {
	addrs := make([]map[wallet.BackendID]wallet.Address, n)
	for i := range addrs {
		addrs[i] = NewRandomAddresses(rng, bID)
	}
	return addrs
}

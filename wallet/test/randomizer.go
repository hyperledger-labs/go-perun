// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package test // import "perun.network/go-perun/wallet/test"

import (
	"math/rand"

	"perun.network/go-perun/wallet"
)

// Randomizer is the wallet testing backend. It currently supports the generation
// of random addresses.
type Randomizer interface {
	NewRandomAddress(*rand.Rand) wallet.Address
	NewRandomAccount(*rand.Rand) wallet.Account
}

// randomizer is the currently set wallet testing randomizer. It is initially set to
// the default randomizer.
var randomizer Randomizer

// SetRandomizer sets the wallet randomizer. It may be set multiple times.
func SetRandomizer(b Randomizer) {
	if randomizer != nil {
		panic("wallet/test randomizer already set")
	}
	randomizer = b
}

// NewRandomAddress returns a new random address by calling the currently set
// wallet randomizer.
func NewRandomAddress(rng *rand.Rand) wallet.Address {
	return randomizer.NewRandomAddress(rng)
}

// NewRandomAccount returns a new random account by calling the currently set
// wallet randomizer.
func NewRandomAccount(rng *rand.Rand) wallet.Account {
	return randomizer.NewRandomAccount(rng)
}

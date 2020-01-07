// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package wallet

import (
	"math/rand"

	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wallet/test"
)

// Randomizer provides random addresses and accounts.
type Randomizer struct{}

var _ test.Randomizer = &Randomizer{}

// NewRandomAddress creates a new random simulated address.
func (Randomizer) NewRandomAddress(rng *rand.Rand) wallet.Address {
	return NewRandomAddress(rng)
}

// NewRandomAccount creates a new random simulated account.
func (Randomizer) NewRandomAccount(rng *rand.Rand) wallet.Account {
	return NewRandomAccount(rng)
}

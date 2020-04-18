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

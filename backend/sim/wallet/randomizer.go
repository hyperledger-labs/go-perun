// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package wallet

import (
	"math/rand"

	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wallet/test"
)

type Randomizer struct{}

var _ test.Randomizer = &Randomizer{}

func (Randomizer) NewRandomAddress(rng *rand.Rand) wallet.Address {
	return NewRandomAddress(rng)
}

func (Randomizer) NewRandomAccount(rng *rand.Rand) wallet.Account {
	return NewRandomAccount(rng)
}

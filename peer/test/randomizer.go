// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package test

import (
	"math/rand"

	"perun.network/go-perun/peer"
	"perun.network/go-perun/wallet/test"
)

// NewRandomAddress returns a new random peer address. Currently still a stub
// until the crypto for peer addresses is decided.
func NewRandomAddress(rng *rand.Rand) peer.Address {
	return test.NewRandomAddress(rng)
}

// NewRandomAddresses returns a slice of random peer addresses.
func NewRandomAddresses(rng *rand.Rand, n int) []peer.Address {
	return test.NewRandomAddresses(rng, n)
}

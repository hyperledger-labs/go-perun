// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package test

import (
	"math/rand"

	"perun.network/go-perun/wallet/test"
	"perun.network/go-perun/wire"
)

// NewRandomAddress returns a new random peer address. Currently still a stub
// until the crypto for peer addresses is decided.
func NewRandomAddress(rng *rand.Rand) wire.Address {
	return test.NewRandomAddress(rng)
}

// NewRandomAddresses returns a slice of random peer addresses.
func NewRandomAddresses(rng *rand.Rand, n int) []wire.Address {
	return test.NewRandomAddresses(rng, n)
}

// NewRandomEnvelope returns an envelope around message m with random sender and
// recipient generated using randomness from rng.
func NewRandomEnvelope(rng *rand.Rand, m wire.Msg) *wire.Envelope {
	return &wire.Envelope{
		Sender:    test.NewRandomAddress(rng),
		Recipient: test.NewRandomAddress(rng),
		Msg:       m,
	}
}

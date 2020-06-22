// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package test

import (
	"math/rand"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet/test"
)

// MockAppRandomizer implements the appRandomizer interface.
type MockAppRandomizer struct {
}

// NewRandomApp creates a new MockApp with a random address.
func (MockAppRandomizer) NewRandomApp(rng *rand.Rand) channel.App {
	return channel.NewMockApp(test.NewRandomAddress(rng))
}

// NewRandomData creates a new MockOp with a random operation.
func (MockAppRandomizer) NewRandomData(rng *rand.Rand) channel.Data {
	return channel.NewMockOp(channel.MockOp(rng.Uint64()))
}

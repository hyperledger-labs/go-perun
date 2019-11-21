// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package test // import "perun.network/go-perun/channel/test"

import (
	"math/rand"

	"perun.network/go-perun/channel"
	wallettest "perun.network/go-perun/wallet/test"
)

type TestBackend struct{}

var _ AppRandomizer = &TestBackend{}

func (TestBackend) NewRandomApp(rng *rand.Rand) channel.App {
	return NewNoApp(wallettest.NewRandomAddress(rng))
}

func (TestBackend) NewRandomData(rng *rand.Rand) channel.Data {
	return newRandomNoAppData(rng)
}

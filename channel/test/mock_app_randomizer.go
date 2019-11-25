// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package test // import "perun.network/go-perun/channel/test"

import (
	"math/rand"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet/test"
)

type MockAppRandomizer struct {
}

func (MockAppRandomizer) NewRandomApp(rng *rand.Rand) channel.App {
	return channel.NewMockApp(test.NewRandomAddress(rng))
}

func (MockAppRandomizer) NewRandomData(rng *rand.Rand) channel.Data {
	return channel.NewMockOp(channel.MockOp(rng.Uint64()))
}

// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package payment

import (
	"math/rand"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/test"
)

type Randomizer struct{}

var _ test.AppRandomizer = (*Randomizer)(nil)

// NewRandomApp always returns a payment app with the same address. Currently,
// one payment address has to be set at program startup.
func (*Randomizer) NewRandomApp(*rand.Rand) channel.App {
	return &App{AppDef()}
}

func (*Randomizer) NewRandomData(*rand.Rand) channel.Data {
	return new(NoData)
}

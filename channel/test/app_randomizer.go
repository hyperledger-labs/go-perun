// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package test // import "perun.network/go-perun/channel/test"

import (
	"math/rand"

	"perun.network/go-perun/channel"
)

type AppRandomizer interface {
	NewRandomApp(*rand.Rand) channel.App
	NewRandomData(*rand.Rand) channel.Data
}

var appRandomizer AppRandomizer = &MockAppRandomizer{}

func SetAppRandomizer(r AppRandomizer) {
	appRandomizer = r
}

func NewRandomApp(rng *rand.Rand) channel.App {
	return appRandomizer.NewRandomApp(rng)
}

func NewRandomData(rng *rand.Rand) channel.Data {
	return appRandomizer.NewRandomData(rng)
}

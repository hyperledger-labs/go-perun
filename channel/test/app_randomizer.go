// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package test // import "perun.network/go-perun/channel/test"

import (
	"math/rand"

	"perun.network/go-perun/channel"
)

// The AppRandomizer interface provides functionality to create RandomApps and
// RandomData which is useful for testing.
type AppRandomizer interface {
	NewRandomApp(*rand.Rand) channel.App
	NewRandomData(*rand.Rand) channel.Data
}

var appRandomizer AppRandomizer = &MockAppRandomizer{}

// isAppRandomizerSet whether the appRandomizer was already set with `SetAppRandomizer`
var isAppRandomizerSet bool

// SetAppRandomizer sets the global appRandomizer.
func SetAppRandomizer(r AppRandomizer) {
	if r == nil || isAppRandomizerSet {
		panic("app randomizer already set or nil argument")
	}
	isAppRandomizerSet = true
	appRandomizer = r
}

// NewRandomApp creates a new random channel.App.
func NewRandomApp(rng *rand.Rand) channel.App {
	return appRandomizer.NewRandomApp(rng)
}

// NewRandomData creates new random data for an app.
func NewRandomData(rng *rand.Rand) channel.Data {
	return appRandomizer.NewRandomData(rng)
}

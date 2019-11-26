// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package test

import (
	"math/rand"

	"perun.network/go-perun/channel"
)

type Randomizer interface {
	NewRandomAsset(*rand.Rand) channel.Asset
}

var randomizer Randomizer

func SetRandomizer(r Randomizer) {
	randomizer = r
}

func NewRandomAsset(rng *rand.Rand) channel.Asset {
	return randomizer.NewRandomAsset(rng)
}

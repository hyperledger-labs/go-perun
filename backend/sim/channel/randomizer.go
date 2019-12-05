// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel

import (
	"math/rand"

	"perun.network/go-perun/channel"
)

type randomizer struct {
}

func (randomizer) NewRandomAsset(rng *rand.Rand) channel.Asset {
	return NewRandomAsset(rng)
}

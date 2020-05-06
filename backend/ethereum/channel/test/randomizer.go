// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package test

import (
	"math/rand"

	ethchannel "perun.network/go-perun/backend/ethereum/channel"
	ethwtest "perun.network/go-perun/backend/ethereum/wallet/test"
	"perun.network/go-perun/channel"
)

type randomizer struct{}

func (randomizer) NewRandomAsset(rng *rand.Rand) channel.Asset {
	return NewRandomAsset(rng)
}

// NewRandomAsset returns a new random ethereum Asset
func NewRandomAsset(rng *rand.Rand) *ethchannel.Asset {
	asset := ethwtest.NewRandomAddress(rng)
	return &asset
}

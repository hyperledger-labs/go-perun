// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package channel

import (
	"math/rand"

	"perun.network/go-perun/backend/ethereum/wallet"
	"perun.network/go-perun/backend/ethereum/wallet/test"
	"perun.network/go-perun/channel"
)

// Asset is an Ethereum asset
type Asset = wallet.Address

var _ channel.Asset = new(Asset)

// NewRandomAsset returns a new random sim Asset
func NewRandomAsset(rng *rand.Rand) *Asset {
	asset := test.NewRandomAddress(rng)
	return &asset
}

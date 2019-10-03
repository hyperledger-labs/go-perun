// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel // import "perun.network/go-perun/backend/sim/channel"

import (
	"math/rand"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/test"
	perun "perun.network/go-perun/pkg/io"
	wallet "perun.network/go-perun/wallet/test"
)

func init() {
	channel.SetBackend(new(backend))
	// The following is only needed for testing
	testbackend := (test.TestBackend{
		NewRandomAssetFunc: func(rng *rand.Rand) perun.Serializable {
			return NewRandomAsset(rng)
		},
		NewRandomAppFunc: func(rng *rand.Rand) channel.App {
			return NewNoApp(wallet.NewRandomAddress(rng))
		},
		NewRandomDataFunc: func(rng *rand.Rand) channel.Data {
			return NewRandomNoAppData(rng)
		}})

	test.SetBackend(testbackend)
}

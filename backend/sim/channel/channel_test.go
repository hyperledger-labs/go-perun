// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel // import "perun.network/go-perun/backend/sim/channel"

import (
	"math/rand"
	"testing"

	"perun.network/go-perun/backend/sim/wallet"
	"perun.network/go-perun/channel/test"
	chtest "perun.network/go-perun/channel/test"
	pkgtest "perun.network/go-perun/pkg/io/test"
	perun "perun.network/go-perun/wallet"
)

func TestGenericTests(t *testing.T) {
	setup := newChannelSetup()
	test.GenericBackendTest(t, setup)
}

func TestAsset(t *testing.T) {
	rng := rand.New(rand.NewSource(1337))
	asset := NewRandomAsset(rng)
	pkgtest.GenericSerializableTest(t, asset)
}

func TestNoAppData(t *testing.T) {
	rng := rand.New(rand.NewSource(1337))
	data := NewRandomNoAppData(rng)
	pkgtest.GenericSerializableTest(t, data)
}

func newChannelSetup() *test.Setup {
	rng := rand.New(rand.NewSource(1337))

	app := NewNoApp(wallet.NewRandomAddress(rng))
	app2 := NewNoApp(wallet.NewRandomAddress(rng))

	params := chtest.NewRandomParams(rng, app)
	params2 := chtest.NewRandomParams(rng, app2)

	state := chtest.NewRandomState(rng, params)
	state2 := chtest.NewRandomState(rng, params2)
	state2.IsFinal = !state.IsFinal

	stateParamFields := make(map[string]bool)
	stateParamFields["App"] = true
	stateParamFields["ChallengeDuration"] = true

	return &test.Setup{
		Params:        params,
		Params2:       params2,
		State:         state,
		State2:        state2,
		Account:       wallet.NewRandomAccount(rng),
		RandomAddress: func() perun.Address { return wallet.NewRandomAddress(rng) },
	}
}

// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package channel // import "perun.network/go-perun/backend/sim/channel"

import (
	"math/rand"
	"testing"

	"perun.network/go-perun/backend/sim/wallet"
	"perun.network/go-perun/channel/test"
	chtest "perun.network/go-perun/channel/test"
	perun "perun.network/go-perun/wallet"
)

func TestGenericTests(t *testing.T) {
	setup := newChannelSetup()
	test.GenericBackendTest(t, setup)
}

func newChannelSetup() *test.Setup {
	rng := rand.New(rand.NewSource(1337))

	app := chtest.NewRandomApp(rng)
	app2 := chtest.NewRandomApp(rng)

	params := chtest.NewRandomParams(rng, app.Def())
	params2 := chtest.NewRandomParams(rng, app2.Def())

	state := chtest.NewRandomState(rng, params)
	state2 := chtest.NewRandomState(rng, params2)
	state2.IsFinal = !state.IsFinal

	return &test.Setup{
		Params:        params,
		Params2:       params2,
		State:         state,
		State2:        state2,
		Account:       wallet.NewRandomAccount(rng),
		RandomAddress: func() perun.Address { return wallet.NewRandomAddress(rng) },
	}
}

// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package channel

import (
	"math/rand"
	"testing"

	chtest "perun.network/go-perun/channel/test"
	"perun.network/go-perun/wallet"
	wtest "perun.network/go-perun/wallet/test"
)

func TestGenericTests(t *testing.T) {
	setup := newChannelSetup()
	chtest.GenericBackendTest(t, setup)
}

func newChannelSetup() *chtest.Setup {
	rng := rand.New(rand.NewSource(1337))

	params, state := chtest.NewRandomParamsAndState(rng, chtest.WithNumLocked(int(rng.Int31n(4)+1)))
	params2, state2 := chtest.NewRandomParamsAndState(rng, chtest.WithIsFinal(!state.IsFinal), chtest.WithNumLocked(int(rng.Int31n(4)+1)))

	return &chtest.Setup{
		Params:        params,
		Params2:       params2,
		State:         state,
		State2:        state2,
		Account:       wtest.NewRandomAccount(rng),
		RandomAddress: func() wallet.Address { return wtest.NewRandomAddress(rng) },
	}
}

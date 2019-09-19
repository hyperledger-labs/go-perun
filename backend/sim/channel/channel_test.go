// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel // import "perun.network/go-perun/backend/sim/channel"

import (
	"math/big"
	"math/rand"
	"testing"

	"perun.network/go-perun/backend/sim/wallet"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/test"
	"perun.network/go-perun/log"
	"perun.network/go-perun/pkg/io"
	perun "perun.network/go-perun/wallet"
)

func TestGenericTests(t *testing.T) {
	setup := newChannelSetup()
	test.GenericBackendTest(t, setup)
}

func newRandomAllocation(rng *rand.Rand, params *channel.Params) *channel.Allocation {
	assets := make([]io.Serializable, 10)
	for i := 0; i < len(assets); i++ {
		assets[i] = newRandomAsset(rng)
	}

	ofparts := make([][]channel.Bal, len(params.Parts))
	for i := 0; i < len(ofparts); i++ {
		ofparts[i] = make([]channel.Bal, len(assets))
		for j := 0; j < len(ofparts); j++ {
			ofparts[i][j] = channel.Bal(big.NewInt(rng.Int63()))
		}
	}

	locked := make([]channel.SubAlloc, 10)
	for i := 0; i < len(locked); i++ {
		bals := make([]channel.Bal, len(assets))
		for j := 0; j < len(bals); j++ {
			bals[j] = channel.Bal(big.NewInt(rng.Int63()))
		}

		var ID channel.ID
		rng.Read(ID[:])
		locked[i] = channel.SubAlloc{ID: ID, Bals: bals}
	}

	return &channel.Allocation{Assets: assets, OfParts: ofparts, Locked: locked}
}

func newRandomParams(rng *rand.Rand, app channel.App) *channel.Params {
	var challengeDuration = rng.Uint64()
	parts := make([]perun.Address, 10)
	for i := 0; i < len(parts); i++ {
		parts[i] = wallet.NewRandomAddress(rng)
	}
	nonce := big.NewInt(rng.Int63())

	params, err := channel.NewParams(challengeDuration, parts, app, nonce)
	if err != nil {
		log.Panic("NewParams failed ", err)
	}
	return params
}

func newRandomState(rng *rand.Rand, p *channel.Params) *channel.State {
	return &channel.State{
		ID:         p.ID(),
		Version:    rng.Uint64(),
		Allocation: *newRandomAllocation(rng, p),
		Data:       newDummyData(rng.Int31n(2) == 0),
		IsFinal:    (rng.Int31n(2) == 0),
	}
}

func newChannelSetup() *test.Setup {
	rng := rand.New(rand.NewSource(1337))

	app := newNoApp(wallet.NewRandomAddress(rng))
	app2 := newNoApp(wallet.NewRandomAddress(rng))

	params := newRandomParams(rng, app)
	params2 := newRandomParams(rng, app2)

	state := newRandomState(rng, params)
	state2 := newRandomState(rng, params2)
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

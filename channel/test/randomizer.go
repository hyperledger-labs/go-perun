// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package test

import (
	"fmt"
	"log"
	"math/big"
	"math/rand"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wallet/test"
	wallettest "perun.network/go-perun/wallet/test"
)

// The Randomizer interface provides the ability to create random assets.
// This is useful for testing.
type Randomizer interface {
	NewRandomAsset(*rand.Rand) channel.Asset
}

var randomizer Randomizer

// SetRandomizer sets the global Randomizer variable.
func SetRandomizer(r Randomizer) {
	if randomizer != nil {
		panic("channel/test randomizer already set")
	}
	randomizer = r
}

// NewRandomAsset creates a new random channel.Asset.
func NewRandomAsset(rng *rand.Rand) channel.Asset {
	return randomizer.NewRandomAsset(rng)
}

// NewRandomAssets creates a new slice of n random channel.Asset
func NewRandomAssets(rng *rand.Rand, n int) []channel.Asset {
	as := make([]channel.Asset, n)
	for i := 0; i < n; i++ {
		as[i] = NewRandomAsset(rng)
	}
	return as
}

// NewRandomAllocation creates a new random allocation.
func NewRandomAllocation(rng *rand.Rand, numParts int) *channel.Allocation {
	if numParts > channel.MaxNumParts {
		panic(fmt.Sprintf(
			"Expected at most %d participants, got %d",
			channel.MaxNumParts, numParts))
	}

	assets := NewRandomAssets(rng, int(rng.Int31n(9))+2)

	balances := NewRandomBalances(rng, len(assets), numParts)

	locked := make([]channel.SubAlloc, rng.Int31n(9)+2)
	for i := 0; i < len(locked); i++ {
		locked[i] = *NewRandomSubAlloc(rng, len(assets))
	}

	return &channel.Allocation{Assets: assets, Balances: balances, Locked: locked}
}

// NewRandomSubAlloc creates a new random suballocation.
func NewRandomSubAlloc(rng *rand.Rand, size int) *channel.SubAlloc {
	return &channel.SubAlloc{ID: NewRandomChannelID(rng), Bals: NewRandomBals(rng, size)}
}

// NewRandomParams creates new random channel.Params.
func NewRandomParams(rng *rand.Rand, appDef wallet.Address) *channel.Params {
	return NewRandomParamsNumParts(rng, appDef, int(rng.Int31n(5))+2)
}

// NewRandomParamsNumParts creates new random channel.Params with n Parts.
func NewRandomParamsNumParts(rng *rand.Rand, appDef wallet.Address, n int) *channel.Params {
	var challengeDuration = rng.Uint64()
	parts := make([]wallet.Address, n)
	for i := 0; i < len(parts); i++ {
		parts[i] = wallettest.NewRandomAddress(rng)
	}
	nonce := big.NewInt(int64(rng.Uint32()))

	params, err := channel.NewParams(challengeDuration, parts, appDef, nonce)
	if err != nil {
		panic(err)
	}
	return params
}

// NewRandomState creates a new random state.
func NewRandomState(rng *rand.Rand, p *channel.Params) *channel.State {
	return &channel.State{
		ID:         p.ID(),
		Version:    rng.Uint64(),
		App:        p.App,
		Allocation: *NewRandomAllocation(rng, len(p.Parts)),
		Data:       NewRandomData(rng),
		IsFinal:    (rng.Int31n(2) == 0),
	}
}

// NewRandomChannelID creates a new random channel.ID.
func NewRandomChannelID(rng *rand.Rand) (id channel.ID) {
	if _, err := rng.Read(id[:]); err != nil {
		log.Panic("could not read from rng")
	}
	return
}

// NewRandomBal creates a new random balance.
func NewRandomBal(rng *rand.Rand) channel.Bal {
	return channel.Bal(big.NewInt(rng.Int63()))
}

// NewRandomBals creates new random balances.
func NewRandomBals(rng *rand.Rand, size int) []channel.Bal {
	bals := make([]channel.Bal, size)
	for i := 0; i < size; i++ {
		bals[i] = NewRandomBal(rng)
	}
	return bals
}

//NewRandomBalances creates an slice containing *numAssets* random Balances slices, each one *numParties* long
func NewRandomBalances(rng *rand.Rand, numAssets int, numParties int) [][]channel.Bal {
	balances := make([][]channel.Bal, numAssets)
	for i := range balances {
		balances[i] = NewRandomBals(rng, numParties)
	}
	return balances
}

// NewRandomTransaction generates a new random Transaction with numParts participants.
// sigMask defines which signatures are generated. It must have size numParts.
// If an entry is false, nil is set as signature for that index.
func NewRandomTransaction(rng *rand.Rand, numParts int, sigMask []bool) *channel.Transaction {
	app := NewRandomApp(rng)
	params := NewRandomParamsNumParts(rng, app.Def(), numParts)
	accs, addrs := test.NewRandomAccounts(rng, len(params.Parts))
	params.Parts = addrs
	state := NewRandomState(rng, params)

	sigs := make([]wallet.Sig, len(params.Parts))
	var err error
	for i, choice := range sigMask {
		if !choice {
			sigs[i] = nil
		} else {
			sigs[i], err = channel.Sign(accs[i], params, state)
		}
		if err != nil {
			panic(err)
		}
	}
	return &channel.Transaction{
		State: state,
		Sigs:  sigs,
	}
}

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

	assets := make([]channel.Asset, rng.Int31n(9)+2)
	for i := 0; i < len(assets); i++ {
		assets[i] = NewRandomAsset(rng)
	}

	ofparts := make([][]channel.Bal, numParts)
	for i := 0; i < len(ofparts); i++ {
		ofparts[i] = NewRandomBals(rng, len(assets))
	}

	locked := make([]channel.SubAlloc, rng.Int31n(9)+2)
	for i := 0; i < len(locked); i++ {
		locked[i] = *NewRandomSubAlloc(rng, len(assets))
	}

	return &channel.Allocation{Assets: assets, Balances: ofparts, Locked: locked}
}

// NewRandomSubAlloc creates a new random suballocation.
func NewRandomSubAlloc(rng *rand.Rand, size int) *channel.SubAlloc {
	return &channel.SubAlloc{ID: NewRandomChannelID(rng), Bals: NewRandomBals(rng, size)}
}

// NewRandomParams creates new random channel.Params.
func NewRandomParams(rng *rand.Rand, appDef wallet.Address) *channel.Params {
	var challengeDuration = rng.Uint64()
	parts := make([]wallet.Address, rng.Int31n(5)+2)
	for i := 0; i < len(parts); i++ {
		parts[i] = wallettest.NewRandomAddress(rng)
	}
	nonce := big.NewInt(int64(rng.Uint32()))

	params, err := channel.NewParams(challengeDuration, parts, appDef, nonce)
	if err != nil {
		log.Panic("NewParams failed ", err)
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

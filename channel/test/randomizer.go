// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package test

import (
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

// NewRandomAsset generates a new random `channel.Asset`.
func NewRandomAsset(rng *rand.Rand) channel.Asset {
	return randomizer.NewRandomAsset(rng)
}

// NewRandomAssets generates new random `channel.Asset`s.
// Options: `WithAssets` and `WithNumAssets`.
func NewRandomAssets(rng *rand.Rand, opts ...RandomOpt) []channel.Asset {
	opt := mergeRandomOpts(opts...)
	if assets := opt.Assets(); assets != nil {
		return assets
	}
	numAssets := opt.NumAssets(rng)
	as := make([]channel.Asset, numAssets)
	for i := range as {
		as[i] = NewRandomAsset(rng)
	}

	updateOpts(opts, WithAssets(as...))
	return as
}

// NewRandomAllocation generates a new random `channel.Allocation`.
// Options: all from `NewRandomAssets`, `NewRandomBalances` and `NewRandomLocked`.
func NewRandomAllocation(rng *rand.Rand, opts ...RandomOpt) *channel.Allocation {
	opt := mergeRandomOpts(opts...)

	if alloc := opt.Allocation(); alloc != nil {
		return alloc
	}

	assets := NewRandomAssets(rng, opt)
	bals := NewRandomBalances(rng, opt)
	locked := NewRandomLocked(rng, opt)

	alloc := &channel.Allocation{Assets: assets, Balances: bals, Locked: locked}
	updateOpts(opts, WithAllocation(alloc))
	return alloc
}

// NewRandomLocked generates new random `channel.SubAlloc`s.
// Options: `WithLocked`, `WithNumLocked` and all from `NewRandomLockedIDs`.
func NewRandomLocked(rng *rand.Rand, opts ...RandomOpt) []channel.SubAlloc {
	opt := mergeRandomOpts(opts...)

	if locked, valid := opt.Locked(); valid {
		return locked
	}

	ids := NewRandomLockedIDs(rng, opt)
	locked := make([]channel.SubAlloc, opt.NumLocked(rng))
	for i := range locked {
		locked[i] = *NewRandomSubAlloc(rng, opt, WithLockedID(ids[i]))
	}
	updateOpts(opts, WithLocked(locked...))
	return locked
}

// NewRandomLockedIDs generates new random `channel.ID`s used in `channel.SubAlloc`.
// Options: `WithLockedIDs` and `WithNumLocked`.
func NewRandomLockedIDs(rng *rand.Rand, opts ...RandomOpt) []channel.ID {
	opt := mergeRandomOpts(opts...)

	if ids := opt.LockedIDs(rng); ids != nil {
		return ids
	}

	numLockedIds := opt.NumLocked(rng)
	ids := make([]channel.ID, numLockedIds)
	for i := range ids {
		rng.Read(ids[i][:])
	}
	return ids
}

// NewRandomSubAlloc generates a new random `channel.SubAlloc`.
// Options: `WithLockedID`, `WithLockedBals` and all from `NewRandomBals`.
func NewRandomSubAlloc(rng *rand.Rand, opts ...RandomOpt) *channel.SubAlloc {
	opt := mergeRandomOpts(opts...)

	id := opt.LockedID(rng)
	var bals []channel.Bal
	if bals = opt.LockedBals(); bals == nil {
		bals = NewRandomBals(rng, opt.NumAssets(rng), opt)
	}

	return &channel.SubAlloc{ID: id, Bals: bals}
}

// NewRandomParamsAndState generates a new random `channel.Params` and `channel.State`.
// Options: all from `NewRandomParams` and `NewRandomState`.
func NewRandomParamsAndState(rng *rand.Rand, opts ...RandomOpt) (params *channel.Params, state *channel.State) {
	opt := mergeRandomOpts(opts...)

	params = NewRandomParams(rng, opt)
	state = NewRandomState(rng, WithParams(params), opt)

	return
}

// NewRandomParams generates a new random `channel.Params`.
// Options: `WithParams`, `WithNumParts`, `WithParts`, `WithFirstPart`, `WithNonce`, `WithChallengeDuration`
// and all from `NewRandomApp`.
func NewRandomParams(rng *rand.Rand, opts ...RandomOpt) *channel.Params {
	opt := mergeRandomOpts(opts...)
	if params := opt.Params(); params != nil {
		return params
	}
	numParts := opt.NumParts(rng)
	var parts []wallet.Address
	if parts = opt.Parts(); parts == nil {
		parts = make([]wallet.Address, numParts)
		for i := range parts {
			parts[i] = wallettest.NewRandomAddress(rng)
		}
	}
	if firstPart := opt.FirstPart(); firstPart != nil {
		parts[0] = firstPart
	}

	nonce := opt.Nonce(rng)
	challengeDuration := opt.ChallengeDuration(rng)
	app := NewRandomApp(rng, opt)

	params := channel.NewParamsUnsafe(challengeDuration, parts, app.Def(), nonce)
	updateOpts(opts, WithParams(params))
	return params
}

// NewRandomState generates a new random `channel.State`.
// Options: `WithState`, `WithVersion`, `WithIsFinal`
// and all from `NewRandomChannelID`, `NewRandomApp`, `NewRandomAllocation` and `NewRandomData`.
func NewRandomState(rng *rand.Rand, opts ...RandomOpt) (state *channel.State) {
	opt := mergeRandomOpts(opts...)
	if state := opt.State(); state != nil {
		return state
	}

	id := NewRandomChannelID(rng, opt)
	version := opt.Version(rng)
	app := NewRandomApp(rng, opt)
	alloc := NewRandomAllocation(rng, opt)
	data := NewRandomData(rng, opt)
	isFinal := opt.IsFinal(rng)

	state = &channel.State{
		ID:         id,
		Version:    version,
		App:        app,
		Allocation: *alloc,
		Data:       data,
		IsFinal:    isFinal,
	}
	updateOpts(opts, WithState(state))
	return
}

// NewRandomChannelID generates a new random `channel.ID`.
// Options: `WithID`.
func NewRandomChannelID(rng *rand.Rand, opts ...RandomOpt) (id channel.ID) {
	opt := mergeRandomOpts(opts...)

	if id, valid := opt.ID(); valid {
		return id
	}

	if _, err := rng.Read(id[:]); err != nil {
		log.Panic("could not read from rng")
	}
	return
}

// NewRandomBal generates a new random `channel.Bal`.
// Options: `WithBalancesRange`.
func NewRandomBal(rng *rand.Rand, opts ...RandomOpt) channel.Bal {
	opt := mergeRandomOpts(opts...)
	min, max := opt.BalancesRange()
	if min == nil {
		min = new(int64)
		*min = 0
	}
	if max == nil {
		max = new(int64)
		*max = (1 << 62)
	}

	return channel.Bal(big.NewInt(rng.Int63n(*max) + (*max - *min) + 1))
}

// NewRandomBals generates new random `channel.Bal`s.
// Options: all from `NewRandomBal`.
func NewRandomBals(rng *rand.Rand, numBals int, opts ...RandomOpt) []channel.Bal {
	opt := mergeRandomOpts(opts...)

	bals := make([]channel.Bal, numBals)
	for i := range bals {
		bals[i] = NewRandomBal(rng, opt)
	}
	return bals
}

// NewRandomBalances generates a new random `[][]channel.Bal`.
// Options: `WithBalances`, `WithNumAssets`, `WithNumParts`
// and all from `NewRandomBals`.
func NewRandomBalances(rng *rand.Rand, opts ...RandomOpt) [][]channel.Bal {
	opt := mergeRandomOpts(opts...)

	if balances := opt.Balances(); balances != nil {
		return balances
	}

	balances := make([][]channel.Bal, opt.NumAssets(rng))
	for i := range balances {
		balances[i] = NewRandomBals(rng, opt.NumParts(rng), opt)
	}

	updateOpts(opts, WithBalances(balances...))
	return balances
}

// NewRandomTransaction generates a new random `channel.Transaction`.
// `sigMask` defines which signatures are generated. `len(sigmask)` is
// assumed to be the number of participants.
// If an entry is false, nil is set as signature for that index.
// Options: all from `NewRandomParamsAndState`.
func NewRandomTransaction(rng *rand.Rand, sigMask []bool, opts ...RandomOpt) *channel.Transaction {
	opt := mergeRandomOpts(opts...)
	numParts := len(sigMask)
	accs, addrs := test.NewRandomAccounts(rng, numParts)
	params := NewRandomParams(rng, WithParts(addrs...), opt)
	state := NewRandomState(rng, WithID(params.ID()), WithNumParts(numParts), opt)

	sigs := make([]wallet.Sig, numParts)
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

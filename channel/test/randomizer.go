// Copyright 2019 - See NOTICE file for copyright holders.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

import (
	crand "crypto/rand"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"time"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wallet/test"
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

// NewRandomPhase generates a random channel machine phase.
func NewRandomPhase(rng *rand.Rand) channel.Phase {
	return channel.Phase(rng.Intn(channel.LastPhase + 1))
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

	return channel.NewSubAlloc(id, bals)
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
			parts[i] = test.NewRandomAddress(rng)
		}
	}
	if firstPart := opt.FirstPart(); firstPart != nil {
		parts[0] = firstPart
	}

	nonce := opt.Nonce(rng)
	challengeDuration := opt.ChallengeDuration(rng)
	app := NewRandomApp(rng, opt)

	params := channel.NewParamsUnsafe(challengeDuration, parts, app, nonce)
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
		// Use 1 here since 0 is nearly impossible anyway and many
		// test assume != 0.
		min = big.NewInt(1)
	}
	if max == nil {
		max = new(big.Int).Rsh(MaxBalance, 30) // 2^98
	}

	// rng(max - min + 1)
	bal, err := crand.Int(rng, new(big.Int).Add(new(big.Int).Sub(max, min), big.NewInt(1)))
	if err != nil {
		panic(fmt.Sprintf("Error creating random big.Int: %v", err))
	}

	// min + rng(max - min + 1)
	return new(big.Int).Add(min, bal)
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

// NewRandomBalances generates a new random `channel.Balances`.
// Options: `WithBalances`, `WithNumAssets`, `WithNumParts`
// and all from `NewRandomBals`.
func NewRandomBalances(rng *rand.Rand, opts ...RandomOpt) channel.Balances {
	opt := mergeRandomOpts(opts...)

	if balances := opt.Balances(); balances != nil {
		return balances
	}

	balances := make(channel.Balances, opt.NumAssets(rng))
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

// ShuffleBalances shuffles the balances of the participants per asset
// and returns it. The returned `Balance` has the same `Sum()` value.
func ShuffleBalances(rng *rand.Rand, b channel.Balances) channel.Balances {
	ret := b.Clone()
	for _, a := range ret {
		a := a
		rng.Shuffle(len(a), func(i, j int) {
			a[i], a[j] = a[j], a[i]
		})
	}
	return ret
}

// NewRandomTimeout creates a new random timeout object.
func NewRandomTimeout(rng *rand.Rand) channel.Timeout {
	return &channel.TimeTimeout{
		Time: time.Unix(rng.Int63(), rng.Int63()),
	}
}

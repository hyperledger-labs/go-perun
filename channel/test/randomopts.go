// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package test

import (
	"math/big"
	"math/rand"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"
)

// RandomOpt defines a map of options than can be passed
// to `NewRandomX` functions in order to alter their default behaviour.
// Should only be constructed by `WithX` functions.
type RandomOpt map[string]interface{}

// WithAllocation sets the `Allocation` that should be used.
// Also defines `WithAssets`, `WithBalances` and `WithLocked`.
func WithAllocation(alloc *channel.Allocation) RandomOpt {
	return RandomOpt{"allocation": alloc}.Append(WithAssets(alloc.Assets...), WithBalances(alloc.Balances...), WithLocked(alloc.Locked...))
}

// WithApp sets the `App` that should be used.
// Also defines `WithDef`.
func WithApp(app channel.App) RandomOpt {
	return RandomOpt{"app": app, "appDef": app.Def()}
}

// WithAppData sets the `AppData` that should be used.
func WithAppData(data channel.Data) RandomOpt {
	return RandomOpt{"appData": data}
}

// WithAppDef sets the `AppDef` that should be used.
func WithAppDef(appDef wallet.Address) RandomOpt {
	return RandomOpt{"appDef": appDef}
}

// WithAssets sets the `Assets` that should be used.
// Also sets `WithNumAssets`.
func WithAssets(assets ...channel.Asset) RandomOpt {
	return RandomOpt{"assets": assets, "numAssets": len(assets)}
}

// WithBalances sets the `Balances` that should be used in a generated Allocation.
// Also sets `WithNumAssets` and `WithNumParts` iff `balances` is not empty.
func WithBalances(balances ...[]channel.Bal) RandomOpt {
	opt := RandomOpt{"balances": balances, "numAssets": len(balances)}
	if len(balances) != 0 {
		opt["numParts"] = len(balances[0])
	}
	return opt
}

// WithBalancesInRange sets the range within which balances are randomly generated.
func WithBalancesInRange(min, max int64) RandomOpt {
	return RandomOpt{"balanceRange": []int64{min, max}}
}

// WithChallengeDuration sets the `ChallengeDuration` that should be used.
func WithChallengeDuration(d uint64) RandomOpt {
	return RandomOpt{"challengeDuration": d}
}

// WithFirstPart sets the first participant that should be used in randomly generated Params. Overrides `WithParts`.
func WithFirstPart(part wallet.Address) RandomOpt {
	return RandomOpt{"firstPart": part}
}

// WithID sets the channel ID that should be used.
func WithID(id channel.ID) RandomOpt {
	return RandomOpt{"id": id}
}

// WithIsFinal sets whether the generated State is final.
func WithIsFinal(isFinal bool) RandomOpt {
	return RandomOpt{"isFinal": isFinal}
}

// WithLocked sets the `Locked` sub-allocations in the generated Allocation.
// Also sets `WithNumLocked` and `WithNumAssets` iff `locked` is not empty.
func WithLocked(locked ...channel.SubAlloc) RandomOpt {
	opt := RandomOpt{"locked": locked, "numLocked": len(locked)}
	if len(locked) > 0 {
		opt["numAssets"] = len(locked[0].Bals)
	}
	return opt
}

// WithLockedBals causes exactly one sub-allocation with the given balances to be generated in the Allocation.
// Also sets `WithNumAssets` and `WithNumLocked` to 1.
func WithLockedBals(bals ...channel.Bal) RandomOpt {
	return RandomOpt{"lockedBals": bals, "numAssets": len(bals), "numLocked": 1}
}

// WithLockedID sets the channel id that should be used when generating a single sub-allocation with `NewRandomSubAlloc`.
func WithLockedID(id channel.ID) RandomOpt {
	return RandomOpt{"lockedId": id}
}

// WithLockedIDs sets the locked channel ids that should be used.
// Also sets `WithNumLocked`.
func WithLockedIDs(ids ...channel.ID) RandomOpt {
	return RandomOpt{"lockedIds": ids, "numLocked": len(ids)}
}

// WithNonce sets the `Nonce` that should be used.
func WithNonce(nonce *big.Int) RandomOpt {
	return RandomOpt{"nonce": nonce}
}

// WithNumAssets sets the `NumAssets` that should be used.
func WithNumAssets(numAssets int) RandomOpt {
	return RandomOpt{"numAssets": numAssets}
}

// WithNumLocked sets the `NumLocked` that should be used.
func WithNumLocked(numLocked int) RandomOpt {
	return RandomOpt{"numLocked": numLocked}
}

// WithNumParts sets the `NumParts` that should be used.
func WithNumParts(numParts int) RandomOpt {
	return RandomOpt{"numParts": numParts}
}

// WithState sets the `State` that should be used.
// Also sets `WithID`, `WithVersion`, `WithApp`, `WithAllocation`, `WithAppData` and `WithIsFinal`.
func WithState(state *channel.State) RandomOpt {
	return RandomOpt{"state": state}.Append(WithID(state.ID), WithVersion(state.Version), WithApp(state.App), WithAllocation(&state.Allocation), WithAppData(state.Data), WithIsFinal(state.IsFinal))
}

// WithParams sets the `Params` that should be used.
// Also sets `WithID`, `WithChallengeDuration`, `WithParts`, `WithApp` and `WithNonce`.
func WithParams(params *channel.Params) RandomOpt {
	return RandomOpt{"params": params}.Append(WithID(params.ID()), WithChallengeDuration(params.ChallengeDuration), WithParts(params.Parts...), WithApp(params.App), WithNonce(params.Nonce))
}

// WithParts sets the `Parts` that should be used when generating Params.
// Also sets `WithNumParts`.
func WithParts(parts ...wallet.Address) RandomOpt {
	return RandomOpt{"parts": parts, "numParts": len(parts)}
}

// WithVersion sets the `Version` that should be used when generating a State.
func WithVersion(version uint64) RandomOpt {
	return RandomOpt{"version": version}
}

// Append inserts all `opts` into the receiving object and returns the result.
// Overrides entries that occur more than once with the last occurrence.
func (o RandomOpt) Append(opts ...RandomOpt) RandomOpt {
	for _, opt := range opts {
		for k, v := range opt {
			o[k] = v
		}
	}
	return o
}

func mergeRandomOpts(opts ...RandomOpt) RandomOpt {
	return RandomOpt{"merged": true}.Append(opts...)
}

func updateOpts(opts []RandomOpt, newOpt RandomOpt) {
	for i := range opts {
		// skip non-merged
		if _, ok := opts[i]["merged"]; ok {
			for k, v := range newOpt {
				opts[i][k] = v
			}
		}
	}
}

// Allocation returns the `Allocation` value of the `RandomOpt`.
// If not present, returns nil.
func (o RandomOpt) Allocation() *channel.Allocation {
	if _, ok := o["alloc"]; !ok {
		return nil
	}
	return o["alloc"].(*channel.Allocation)
}

// App returns the `App` value of the `RandomOpt`.
// If not present, returns nil.
func (o RandomOpt) App() channel.App {
	if _, ok := o["app"]; !ok {
		return nil
	}
	return o["app"].(channel.App)
}

// AppData returns the `AppData` value of the `RandomOpt`.
// If not present, returns nil.
func (o RandomOpt) AppData() channel.Data {
	if _, ok := o["appData"]; !ok {
		return nil
	}
	return o["appData"].(channel.Data)
}

// AppDef returns the `AppDef` value of the `RandomOpt`.
// If not present, returns nil.
func (o RandomOpt) AppDef() wallet.Address {
	if _, ok := o["appDef"]; !ok {
		return nil
	}
	return o["appDef"].(wallet.Address)
}

// Assets returns the `Assets` value of the `RandomOpt`.
// If not present, returns nil.
func (o RandomOpt) Assets() []channel.Asset {
	if _, ok := o["assets"]; !ok {
		return nil
	}
	return o["assets"].([]channel.Asset)
}

// Balances returns the `Balances` value of the `RandomOpt`.
// If not present, returns nil.
func (o RandomOpt) Balances() [][]channel.Bal {
	if _, ok := o["balances"]; !ok {
		return nil
	}
	return o["balances"].([][]channel.Bal)
}

// BalancesRange returns the `BalancesRange` value of the `RandomOpt`.
// If not present, returns nil,nil.
func (o RandomOpt) BalancesRange() (min, max *int64) {
	if _, ok := o["balanceRange"]; !ok {
		return nil, nil
	}
	r := o["balanceRange"].([]int64)
	return &r[0], &r[1]
}

// ChallengeDuration returns the `ChallengeDuration` value of the `RandomOpt`.
// If not present, a random value is generated with `rng` as entropy source.
func (o RandomOpt) ChallengeDuration(rng *rand.Rand) uint64 {
	if _, ok := o["challengeDuration"]; !ok {
		o["challengeDuration"] = rng.Uint64()
	}
	return o["challengeDuration"].(uint64)
}

// ID returns the `ID` value of the `RandomOpt`.
// If not present, returns `false` as second argument.
func (o RandomOpt) ID() (id channel.ID, valid bool) {
	if _, ok := o["id"]; !ok {
		return channel.ID{}, false
	}
	return o["id"].(channel.ID), true
}

// FirstPart returns the `FirstPart` value of the `RandomOpt`.
// If not present, returns nil.
func (o RandomOpt) FirstPart() wallet.Address {
	if _, ok := o["firstPart"]; !ok {
		return nil
	}
	return o["firstPart"].(wallet.Address)
}

// IsFinal returns the `IsFinal` value of the `RandomOpt`.
// If not present, a random value is generated with `rng` as entropy source.
func (o RandomOpt) IsFinal(rng *rand.Rand) bool {
	if _, ok := o["isFinal"]; !ok {
		o["isFinal"] = (rng.Int31n(2) == 0)
	}
	return o["isFinal"].(bool)
}

// Locked returns the `Locked` value of the `RandomOpt`.
// If not present, returns `false` as second argument.
func (o RandomOpt) Locked() (locked []channel.SubAlloc, valid bool) {
	if _, ok := o["locked"]; !ok {
		return nil, false
	}
	return o["locked"].([]channel.SubAlloc), true
}

// LockedBals returns the `LockedBals` value of the `RandomOpt`.
// If not present, returns nil.
func (o RandomOpt) LockedBals() []channel.Bal {
	if _, ok := o["lockedBals"]; !ok {
		return nil
	}
	return o["lockedBals"].([]channel.Bal)
}

// LockedID returns the `LockedID` value of the `RandomOpt`.
// If not present, a random value is generated with `rng` as entropy source.
func (o RandomOpt) LockedID(rng *rand.Rand) channel.ID {
	if _, ok := o["lockedId"]; !ok {
		o["lockedId"] = NewRandomChannelID(rng)
	}
	return o["lockedId"].(channel.ID)
}

// LockedIDs returns the `LockedIDs` value of the `RandomOpt`.
// If not present, returns nil.
func (o RandomOpt) LockedIDs(rng *rand.Rand) (ids []channel.ID) {
	if _, ok := o["lockedIds"]; !ok {
		return nil
	}
	return o["lockedIds"].([]channel.ID)
}

// Nonce returns the `Nonce` value of the `RandomOpt`.
// If not present, a random value is generated with `rng` as entropy source.
func (o RandomOpt) Nonce(rng *rand.Rand) *big.Int {
	if _, ok := o["nonce"]; !ok {
		o["nonce"] = big.NewInt(rng.Int63())
	}
	return o["nonce"].(*big.Int)
}

// NumAssets returns the `NumAssets` value of the `RandomOpt`.
// If not present, a random value is generated with `rng` as entropy source.
func (o RandomOpt) NumAssets(rng *rand.Rand) int {
	if _, ok := o["numAssets"]; !ok {
		o["numAssets"] = int(rng.Int31n(10) + 1)
	}
	return o["numAssets"].(int)
}

// NumLocked returns the `NumLocked` value of the `RandomOpt`.
// If not present, returns 0.
func (o RandomOpt) NumLocked(rng *rand.Rand) int {
	if _, ok := o["numLocked"]; !ok {
		// We return 0 here because when no `WithLocked` or `WithNumLocked`
		// are given, 0 is assumed.
		o["numLocked"] = 0
	}
	return o["numLocked"].(int)
}

// NumParts returns the `NumParts` value of the `RandomOpt`.
// If not present, a random value between 2 and 11 is generated with `rng` as entropy source.
func (o RandomOpt) NumParts(rng *rand.Rand) int {
	if _, ok := o["numParts"]; !ok {
		o["numParts"] = int(rng.Int31n(10) + 2)
	}
	return o["numParts"].(int)
}

// State returns the `State` value of the `RandomOpt`.
// If not present, returns nil.
func (o RandomOpt) State() *channel.State {
	if _, ok := o["state"]; !ok {
		return nil
	}
	return o["state"].(*channel.State)
}

// Params returns the `Params` value of the `RandomOpt`.
// If not present, returns nil.
func (o RandomOpt) Params() *channel.Params {
	if _, ok := o["params"]; !ok {
		return nil
	}
	return o["params"].(*channel.Params)
}

// Parts returns the `Parts` value of the `RandomOpt`.
// If not present, returns nil.
func (o RandomOpt) Parts() []wallet.Address {
	if _, ok := o["parts"]; !ok {
		return nil
	}
	return o["parts"].([]wallet.Address)
}

// Version returns the `Version` value of the `RandomOpt`.
// If not present, a random value is generated with `rng` as entropy source.
func (o RandomOpt) Version(rng *rand.Rand) uint64 {
	if _, ok := o["version"]; !ok {
		o["version"] = rng.Uint64()
	}
	return o["version"].(uint64)
}

// Copyright 2025 - See NOTICE file for copyright holders.
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
	"errors"
	"fmt"
	"io"
	"math/big"
	"math/rand"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"
)

var (
	// MaxBalance is the maximum balance used for testing.
	// It is set to 2 ^ 128 - 1 since when 2 ^ 256 - 1 is used, the faucet
	// key is depleted.
	// The production limit can be found in `go-perun/channel.MaxBalance`.
	MaxBalance = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1)) //nolint:mnd

	// Highest balance that is returned by NewRandomBal. Set to MaxBalance / 2^30.
	maxRandomBalance = new(big.Int).Rsh(MaxBalance, 30) //nolint:mnd
)

const (
	// maximal number of assets that NumAssets returns.
	maxNumAssets = int32(10)
	// maximal number of participants that NumParts returns.
	maxNumParts = int32(10)
	// minimal number of participants that NumParts returns.
	minNumParts = int32(2)
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
	var appDef channel.AppID
	if !channel.IsNoApp(app) {
		appDef = app.Def()
	}
	return RandomOpt{"app": app, "appDef": appDef}
}

// WithBackend sets the `backend` that should be used.
func WithBackend(id wallet.BackendID) RandomOpt {
	return RandomOpt{"backend": id}
}

// WithBackendIDs sets the `backend` that should be used.
func WithBackendIDs(id []wallet.BackendID) RandomOpt {
	return RandomOpt{"backendIDs": id}
}

// WithoutApp configures a NoApp and NoData.
func WithoutApp() RandomOpt {
	return RandomOpt{"app": channel.NoApp(), "appDef": nil, "appData": channel.NoData()}
}

// WithAppData sets the `AppData` that should be used.
func WithAppData(data channel.Data) RandomOpt {
	return RandomOpt{"appData": data}
}

// WithAppRandomizer sets the `AppRandomizer` that should be used.
func WithAppRandomizer(randomizer AppRandomizer) RandomOpt {
	return RandomOpt{"appRandomizer": randomizer}
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

// WithBalancesInRange sets the range within which balances are randomly generated to [min, max].
func WithBalancesInRange(balanceMin, balanceMax channel.Bal) RandomOpt {
	return RandomOpt{"balanceRange": []channel.Bal{balanceMin, balanceMax}}
}

// WithChallengeDuration sets the `ChallengeDuration` that should be used.
func WithChallengeDuration(d uint64) RandomOpt {
	return RandomOpt{"challengeDuration": d}
}

// WithFirstPart sets the first participant that should be used in randomly generated Params. Overrides `WithParts`.
func WithFirstPart(part map[wallet.BackendID]wallet.Address) RandomOpt {
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
func WithNonce(nonce channel.Nonce) RandomOpt {
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
	return RandomOpt{"params": params}.Append(WithID(params.ID()), WithChallengeDuration(params.ChallengeDuration), WithParts(params.Parts), WithApp(params.App), WithNonce(params.Nonce))
}

// WithParts sets the `Parts` that should be used when generating Params.
// Also sets `WithNumParts`.
func WithParts(parts []map[wallet.BackendID]wallet.Address) RandomOpt {
	return RandomOpt{"parts": parts, "numParts": len(parts)}
}

// WithVersion sets the `Version` that should be used when generating a State.
func WithVersion(version uint64) RandomOpt {
	return RandomOpt{"version": version}
}

const (
	nameLedgerChannel  = "ledgerChannel"
	nameVirtualChannel = "virtualChannel"
)

// WithLedgerChannel sets the `LedgerChannel` attribute.
func WithLedgerChannel(ledger bool) RandomOpt {
	return RandomOpt{nameLedgerChannel: ledger}
}

// WithVirtualChannel sets the `VirtualChannel` attribute.
func WithVirtualChannel(b bool) RandomOpt {
	return RandomOpt{nameVirtualChannel: b}
}

// WithAux sets the `Aux` attribute.
func WithAux(aux channel.Aux) RandomOpt {
	return RandomOpt{"aux": aux}
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

// AppRandomizer returns the `AppRandomizer` value of the `RandomOpt`.
// If not present, returns the default appRandomizer.
func (o RandomOpt) AppRandomizer() AppRandomizer {
	if _, ok := o["appRandomizer"]; !ok {
		return appRandomizer
	}
	return o["appRandomizer"].(AppRandomizer)
}

// AppDef returns the `AppDef` value of the `RandomOpt`.
// If not present, returns nil.
func (o RandomOpt) AppDef() channel.AppID {
	if _, ok := o["appDef"]; !ok {
		return nil
	}
	return o["appDef"].(channel.AppID)
}

// Assets returns the `Assets` value of the `RandomOpt`.
// If not present, returns nil.
func (o RandomOpt) Assets() []channel.Asset {
	if _, ok := o["assets"]; !ok {
		return nil
	}
	return o["assets"].([]channel.Asset)
}

// Backend returns the `Backend` value of the `RandomOpt`.
func (o RandomOpt) Backend() (wallet.BackendID, error) {
	if _, ok := o["backend"]; !ok {
		return 0, errors.New("backend not set")
	}
	return o["backend"].(wallet.BackendID), nil
}

// BackendID returns the `BackendID` value  from `Allocation` of the `RandomOpt`.
func (o RandomOpt) BackendID() ([]wallet.BackendID, error) {
	if _, ok := o["backendIDs"]; !ok {
		return []wallet.BackendID{0}, errors.New("backend not set")
	}
	return o["backendIDs"].([]wallet.BackendID), nil
}

// Balances returns the `Balances` value of the `RandomOpt`.
// If not present, returns nil.
func (o RandomOpt) Balances() channel.Balances {
	if _, ok := o["balances"]; !ok {
		return nil
	}
	return o["balances"].([][]channel.Bal)
}

// BalancesRange returns the `BalancesRange` value of the `RandomOpt`.
// If not present, returns nil,nil.
func (o RandomOpt) BalancesRange() (balanceMin, balanceMax channel.Bal) {
	if _, ok := o["balanceRange"]; !ok {
		return nil, nil
	}
	r := o["balanceRange"].([]channel.Bal)
	return r[0], r[1]
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
func (o RandomOpt) FirstPart() map[wallet.BackendID]wallet.Address {
	if _, ok := o["firstPart"]; !ok {
		return nil
	}
	return o["firstPart"].(map[wallet.BackendID]wallet.Address)
}

// IsFinal returns the `IsFinal` value of the `RandomOpt`.
// If not present, a random value is generated with `rng` as entropy source.
func (o RandomOpt) IsFinal(rng *rand.Rand) bool {
	if _, ok := o["isFinal"]; !ok {
		o["isFinal"] = (rng.Int31n(2) == 0) //nolint:mnd
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
func (o RandomOpt) Nonce(rng io.Reader) channel.Nonce {
	if _, ok := o["nonce"]; !ok {
		n := make([]byte, channel.MaxNonceLen)
		if _, err := rng.Read(n); err != nil {
			panic(fmt.Sprintf("reading rnd: %v", err))
		}
		o["nonce"] = channel.NonceFromBytes(n)
	}
	return o["nonce"].(channel.Nonce)
}

// LedgerChannel returns the `LedgerChannel` value of the `RandomOpt`.
// If not present, a random value is generated with `rng` as entropy source.
func (o RandomOpt) LedgerChannel(rng io.Reader) bool {
	if _, ok := o[nameLedgerChannel]; !ok {
		a := make([]byte, 1)
		_, err := rng.Read(a)
		if err != nil {
			panic(err)
		}
		o[nameLedgerChannel] = a[0]%2 == 0 //nolint:mnd
	}
	return o[nameLedgerChannel].(bool)
}

// VirtualChannel returns the `VirtualChannel` value of the `RandomOpt`.
// If not present, a random value is generated with `rng` as entropy source.
func (o RandomOpt) VirtualChannel(rng io.Reader) bool {
	if _, ok := o[nameVirtualChannel]; !ok {
		a := make([]byte, 1)
		_, err := rng.Read(a)
		if err != nil {
			panic(err)
		}
		o[nameVirtualChannel] = a[0]%2 == 0 //nolint:mnd
	}
	return o[nameVirtualChannel].(bool)
}

// Aux returns the `Aux` value of the `RandomOpt`.
// If not present, a random value is generated with `rng` as entropy source.
func (o RandomOpt) Aux(rng io.Reader) channel.Aux {
	if _, ok := o["aux"]; !ok {
		a := make([]byte, channel.AuxMaxLen)
		_, err := rng.Read(a)
		if err != nil {
			panic(err)
		}
		o["aux"] = channel.Aux(a)
	}
	return o["aux"].(channel.Aux)
}

// NumAssets returns the `NumAssets` value of the `RandomOpt`.
// If not present, a random value is generated with `rng` as entropy source.
func (o RandomOpt) NumAssets(rng *rand.Rand) int {
	if _, ok := o["numAssets"]; !ok {
		o["numAssets"] = int(rng.Int31n(maxNumAssets) + 1)
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
		o["numParts"] = int(rng.Int31n(maxNumParts) + minNumParts)
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
func (o RandomOpt) Parts() []map[wallet.BackendID]wallet.Address {
	if _, ok := o["parts"]; !ok {
		return nil
	}
	return o["parts"].([]map[wallet.BackendID]wallet.Address)
}

// Version returns the `Version` value of the `RandomOpt`.
// If not present, a random value is generated with `rng` as entropy source.
func (o RandomOpt) Version(rng *rand.Rand) uint64 {
	if _, ok := o["version"]; !ok {
		o["version"] = rng.Uint64()
	}
	return o["version"].(uint64)
}

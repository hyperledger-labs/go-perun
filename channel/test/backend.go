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
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"
)

type addressCreator = func() wallet.Address

// Setup provides all objects needed for the generic channel tests.
type (
	Setup struct {
		// Params are the random parameters of `State`
		Params *channel.Params
		// Params2 are the parameters of `State2` and must differ in all fields from `Params`
		Params2 *channel.Params

		// State is a random state with parameters `Params`
		State *channel.State
		// State2 is a random state with parameters `Params2` and should differ in all fields from `State`
		State2 *channel.State

		// Account is a random account
		Account wallet.Account

		// RandomAddress returns a new random address
		RandomAddress addressCreator
	}

	// GenericTestOption can be used to control the behaviour of generic tests.
	GenericTestOption int
	// GenericTestOptions is a collection of GenericTestOption.
	GenericTestOptions map[GenericTestOption]bool
)

const (
	// IgnoreAssets ignores the Assets in tests that support it.
	IgnoreAssets GenericTestOption = iota
	// IgnoreApp ignores the App in tests that support it.
	IgnoreApp
)

// mergeTestOpts merges all passed options into one and returns it.
func mergeTestOpts(opts ...GenericTestOption) GenericTestOptions {
	ret := make(GenericTestOptions)
	for _, o := range opts {
		ret[o] = true
	}
	return ret
}

// GenericBackendTest tests the interface functions of the global channel.Backend with the passed test data.
func GenericBackendTest(t *testing.T, s *Setup, opts ...GenericTestOption) {
	t.Helper()
	require := require.New(t)
	ID := channel.CalcID(s.Params)
	require.Equal(ID, s.State.ID, "ChannelID(params) should match the States ID")
	require.Equal(ID, s.Params.ID(), "ChannelID(params) should match the Params ID")
	require.NotNil(s.State.Data, "State data can not be nil")
	require.NotNil(s.State2.Data, "State2 data can not be nil")

	t.Run("ChannelID", func(t *testing.T) {
		genericChannelIDTest(t, s)
	})

	t.Run("Sign", func(t *testing.T) {
		genericSignTest(t, s)
	})

	t.Run("Verify", func(t *testing.T) {
		genericVerifyTest(t, s, opts...)
	})
}

func genericChannelIDTest(t *testing.T, s *Setup) {
	t.Helper()
	require.NotNil(t, s.Params.Parts, "params.Parts can not be nil")
	assert.Panics(t, func() { channel.CalcID(nil) }, "ChannelID(nil) should panic")

	// Check that modifying the state changes the id
	for _, modParams := range buildModifiedParams(s.Params, s.Params2, s) {
		params := modParams
		ID := channel.CalcID(&params)
		assert.NotEqual(t, ID, s.State.ID, "Channel ids should differ")
	}
}

func genericSignTest(t *testing.T, s *Setup) {
	t.Helper()
	_, err := channel.Sign(s.Account, s.State)
	assert.NoError(t, err, "Sign should not return an error")
}

func genericVerifyTest(t *testing.T, s *Setup, opts ...GenericTestOption) {
	t.Helper()
	addr := s.Account.Address()
	require.Equal(t, channel.CalcID(s.Params), s.Params.ID(), "Invalid test params")
	sig, err := channel.Sign(s.Account, s.State)
	require.NoError(t, err, "Sign should not return an error")

	ok, err := channel.Verify(addr, s.State, sig)
	assert.NoError(t, err, "Verify should not return an error")
	assert.True(t, ok, "Verify should return true")

	for i, _modState := range buildModifiedStates(s.State, s.State2, append(opts, IgnoreApp)...) {
		modState := _modState
		ok, err = channel.Verify(addr, &modState, sig)
		assert.Falsef(t, ok, "Verify should return false: index %d", i)
		assert.NoError(t, err, "Verify should not return an error")
	}

	// Different address and same state and params
	for i := 0; i < 10; i++ {
		ok, err := channel.Verify(s.RandomAddress(), s.State, sig)
		assert.NoError(t, err, "Verify should not return an error")
		assert.False(t, ok, "Verify should return false")
	}
}

// buildModifiedParams returns a slice of Params that are different from `p1` assuming that `p2` differs in
// every member from `p1`.
func buildModifiedParams(p1, p2 *channel.Params, s *Setup) (ret []channel.Params) {
	// Modify params
	{
		// Modify complete Params
		{
			modParams := *p2
			ret = appendModParams(ret, modParams)
		}
		// Modify ChallengeDuration
		{
			modParams := *p1
			modParams.ChallengeDuration = p2.ChallengeDuration
			ret = appendModParams(ret, modParams)
		}
		// Modify Parts
		{
			// Modify complete Parts
			{
				modParams := *p1
				modParams.Parts = p2.Parts
				ret = appendModParams(ret, modParams)
			}
			// Modify Parts[0]
			{
				modParams := *p1
				modParams.Parts = make([]wallet.Address, len(p1.Parts))
				copy(modParams.Parts, p1.Parts)
				modParams.Parts[0] = s.RandomAddress()
				ret = appendModParams(ret, modParams)
			}
		}
		// Modify Nonce
		{
			modParams := *p1
			modParams.Nonce = p2.Nonce
			ret = appendModParams(ret, modParams)
		}
	}

	return
}

func appendModParams(a []channel.Params, modParams channel.Params) []channel.Params {
	p := channel.NewParamsUnsafe(
		modParams.ChallengeDuration,
		modParams.Parts,
		modParams.App,
		modParams.Nonce,
		modParams.LedgerChannel,
		modParams.VirtualChannel,
	)
	return append(a, *p)
}

// buildModifiedStates returns a slice of States that are different from `s1` assuming that `s2` differs in
// every member from `s1`.
// `modifyApp` indicates whether the app should also be changed or not. In some cases (signature) it is desirable
// not to modify it.
func buildModifiedStates(s1, s2 *channel.State, _opts ...GenericTestOption) (ret []channel.State) {
	opts := mergeTestOpts(_opts...)
	// Modify state
	{
		// Modify complete state
		{
			modState := s2.Clone()
			ret = append(ret, *modState)
		}
		// Modify ID
		{
			modState := s1.Clone()
			modState.ID = s2.ID
			ret = append(ret, *modState)
		}
		// Modify Version
		{
			modState := s1.Clone()
			modState.Version = s2.Version
			ret = append(ret, *modState)
		}
		// Modify App
		if !opts[IgnoreApp] {
			modState := s1.Clone()
			modState.App = s2.App
			ret = append(ret, *modState)
		}
		// Modify Allocation
		{
			// Modify complete Allocation
			{
				modState := s1.Clone()
				modState.Allocation = s2.Allocation
				ret = append(ret, *modState)
			}
			// Modify Assets
			if !opts[IgnoreAssets] {
				// Modify complete Assets
				{
					modState := s1.Clone()
					modState.Assets = s2.Assets
					modState = ensureConsistentBalances(modState)
					ret = append(ret, *modState)
				}
				// Modify Assets[0]
				{
					modState := s1.Clone()
					modState.Allocation.Assets[0] = s2.Allocation.Assets[0]
					ret = append(ret, *modState)
				}
			}
			// Modify Balances
			{
				// Modify complete Balances
				{
					modState := s1.Clone()
					modState.Balances = s2.Balances
					modState = ensureConsistentBalances(modState)
					ret = append(ret, *modState)
				}
				// Modify Balances[0]
				{
					modState := s1.Clone()
					modState.Allocation.Balances[0] = s2.Allocation.Balances[0]
					modState = ensureConsistentBalances(modState)
					ret = append(ret, *modState)
				}
				// Modify Balances[0][0]
				{
					modState := s1.Clone()
					modState.Allocation.Balances[0][0] = s2.Allocation.Balances[0][0]
					ret = append(ret, *modState)
				}
			}
			// Modify Locked
			if len(s1.Locked) > 0 || len(s2.Locked) > 0 {
				// Modify complete Locked
				{
					modState := s1.Clone()
					modState.Allocation.Locked = s2.Clone().Locked
					modState = ensureConsistentBalances(modState)
					ret = append(ret, *modState)
				}
				// Modify Locked[0].ID
				{
					modState := s1.Clone()
					modState.Allocation.Locked[0].ID = s2.Allocation.Locked[0].ID
					ret = append(ret, *modState)
				}
				// Modify Locked[0].Bals
				{
					modState := s1.Clone()
					modState.Allocation.Locked[0].Bals = s2.Locked[0].Bals
					modState = ensureConsistentBalances(modState)
					ret = append(ret, *modState)
				}
				// Modify Locked[0].Bals[0]
				{
					modState := s1.Clone()
					modState.Allocation.Locked[0].Bals[0] = s2.Allocation.Locked[0].Bals[0]
					ret = append(ret, *modState)
				}
			}
		}
		// Modify Data
		if !channel.IsNoData(s1.Data) || !channel.IsNoData(s2.Data) {
			modState := s1.Clone()
			modState.Data = s2.Data
			ret = append(ret, *modState)
		}
		// Modify IsFinal
		{
			modState := s1.Clone()
			modState.IsFinal = s2.IsFinal
			ret = append(ret, *modState)
		}
	}

	return
}

func ensureConsistentBalances(s *channel.State) *channel.State {
	_s := s.Clone()
	numAssets := len(_s.Assets)
	numParts := _s.NumParts()

	// Ensure Balances has correct length.
	// Ensure at least numAssets.
	for numAssets-len(_s.Balances) > 0 {
		assetBals := make([]channel.Bal, numParts)
		for i := range assetBals {
			assetBals[i] = big.NewInt(0)
		}
		_s.Balances = append(_s.Balances, assetBals)
	}
	// Ensure at most numAssets.
	_s.Balances = _s.Balances[:numAssets]

	// Ensure asset balances have correct length.
	for i, assetBals := range _s.Balances {
		_s.Balances[i] = ensureBalanceVectorLength(assetBals, numParts)
	}

	// Ensure locked balances have correct length.
	for i, subAlloc := range _s.Locked {
		_s.Locked[i].Bals = ensureBalanceVectorLength(subAlloc.Bals, numAssets)
	}

	return _s
}

func ensureBalanceVectorLength(bals []channel.Bal, l int) []channel.Bal {
	// Ensure at least numParts.
	for l-len(bals) > 0 {
		bals = append(bals, big.NewInt(0))
	}
	// Ensure at most numParts.
	bals = bals[:l]
	return bals
}

// GenericStateEqualTest tests the State.Equal function.
func GenericStateEqualTest(t *testing.T, s1, s2 *channel.State, opts ...GenericTestOption) {
	t.Helper()
	assert.NoError(t, s1.Equal(s1))
	assert.NoError(t, s2.Equal(s2))

	for _, differentState := range buildModifiedStates(s1, s2, opts...) {
		assert.Error(t, differentState.Equal(s1))
	}
}

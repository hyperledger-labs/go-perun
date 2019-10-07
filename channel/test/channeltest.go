// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package test // import "perun.network/go-perun/channel/test"

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"
)

type addressCreator = func() wallet.Address

// Setup provides all objects needed for the generic channel tests
type Setup struct {
	// Params are the random parameters of `State`
	Params *channel.Params
	// Params2 are the parameters of `State2` and must differ in all fields from `Params`
	Params2 *channel.Params

	// State is a random state with parameters `Params`
	State *channel.State
	// State2 is a random state with parameters `Params2` and must differ in all fields from `State`
	State2 *channel.State

	// Account is a random account
	Account wallet.Account

	// RandomAddress returns a new random address
	RandomAddress addressCreator
}

// GenericBackendTest tests the interface functions of the global channel.Backend with the passed test data.
func GenericBackendTest(t *testing.T, s *Setup) {
	require := require.New(t)
	ID := channel.ChannelID(s.Params)
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
		genericVerifyTest(t, s)
	})
}

func genericChannelIDTest(t *testing.T, s *Setup) {
	require.NotNil(t, s.Params.Parts, "params.Parts can not be nil")
	assert.Panics(t, func() { channel.ChannelID(nil) }, "ChannelID(nil) should panic")

	// Check that modifying the state changes the id
	for _, modParams := range buildModifiedParams(s.Params, s.Params2, s) {
		ID := channel.ChannelID(&modParams)
		assert.NotEqual(t, ID, s.State.ID, "Channel ids should differ")
	}
}

func genericSignTest(t *testing.T, s *Setup) {
	_, err := channel.Sign(nil, s.Params, s.State)
	assert.Error(t, err, "Sign should return an error")
	_, err = channel.Sign(s.Account, nil, s.State)
	assert.Error(t, err, "Sign should return an error")
	_, err = channel.Sign(s.Account, s.Params, nil)
	assert.Error(t, err, "Sign should return an error")

	_, err1 := channel.Sign(s.Account, s.Params, s.State)
	assert.NoError(t, err1, "Sign should not return an error")
}

func genericVerifyTest(t *testing.T, s *Setup) {
	addr := s.Account.Address()
	require.Equal(t, channel.ChannelID(s.Params), s.Params.ID(), "Invalid test params")
	sig, err := channel.Sign(s.Account, s.Params, s.State)
	require.NoError(t, err, "Sign should not return an error")

	ok, err := channel.Verify(addr, s.Params, s.State, sig)
	assert.NoError(t, err, "Verify should not return an error")
	assert.True(t, ok, "Verify should return true")

	// Different state and same params
	ok, err = channel.Verify(addr, s.Params, s.State2, sig)
	assert.NoError(t, err, "Verify should not return an error")
	assert.False(t, ok, "Verify should return false")

	// Different params and same state
	// -> The backend does not detect this

	// Different params and different state
	for _, modParams := range buildModifiedParams(s.Params, s.Params2, s) {
		for _, fakeState := range buildModifiedStates(s.State, s.State2) {
			ok, err = channel.Verify(addr, &modParams, &fakeState, sig)
			assert.NoError(t, err, "Verify should not return an error")
			assert.False(t, ok, "Verify should return false")
		}
	}

	// Different address and same state and params
	for i := 0; i < 10; i++ {
		ok, err := channel.Verify(s.RandomAddress(), s.Params, s.State, sig)
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
			ret = append(ret, modParams)
		}
		// Modify ChallengeDuration
		{
			modParams := *p1
			modParams.ChallengeDuration = p2.ChallengeDuration
			ret = append(ret, modParams)
		}
		// Modify Parts
		{
			// Modify complete Parts
			{
				modParams := *p1
				modParams.Parts = p2.Parts
				ret = append(ret, modParams)
			}
			// Modify Parts[0]
			{
				modParams := *p1
				modParams.Parts = make([]wallet.Address, len(p1.Parts))
				copy(modParams.Parts, p1.Parts)
				modParams.Parts[0] = s.RandomAddress()
				ret = append(ret, modParams)
			}
		}
		// Modify App
		{
			modParams := *p1
			modParams.App = p2.App
			ret = append(ret, modParams)
		}
		// Modify Nonce
		{
			modParams := *p1
			modParams.Nonce = p2.Nonce
			ret = append(ret, modParams)
		}
	}

	return
}

// buildModifiedStates returns a slice of States that are different from `s1` assuming that `s2` differs in
// every member from `s1`.
func buildModifiedStates(s1, s2 *channel.State) (ret []channel.State) {
	// Modify state
	{
		// Modify complete state
		{
			modState := *s2
			ret = append(ret, modState)
		}
		// Modify ID
		{
			modState := *s1
			modState.ID = s2.ID
			ret = append(ret, modState)
		}
		// Modify Version
		{
			modState := *s1
			modState.Version = s2.Version
			ret = append(ret, modState)
		}
		// Modify Allocation
		{
			// Modify complete Allocation
			{
				modState := *s1
				modState.Allocation = s2.Allocation
				ret = append(ret, modState)
			}
			// Modify Assets
			{
				// Modify complete Assets
				{
					modState := *s1
					modState.Allocation.Assets = s2.Allocation.Assets
					ret = append(ret, modState)
				}
				// Modify Assets[0]
				{
					modState := *s1
					modState.Assets = make([]channel.Asset, len(s1.Allocation.Assets))
					copy(modState.Allocation.Assets, s1.Allocation.Assets)
					modState.Allocation.Assets[0] = s2.Allocation.Assets[0]
					ret = append(ret, modState)
				}
			}
			// Modify OfParts
			{
				// Modify complete OfParts
				{
					modState := *s1
					modState.Allocation.OfParts = s2.Allocation.OfParts
					ret = append(ret, modState)
				}
				// Modify OfParts[0]
				{
					modState := *s1
					modState.Allocation.OfParts[0] = s2.Allocation.OfParts[0]
					ret = append(ret, modState)
				}
				// Modify OfParts[0][0]
				{
					modState := *s1
					modState.Allocation.OfParts[0][0] = s2.Allocation.OfParts[0][0]
					ret = append(ret, modState)
				}
			}
			// Modify Locked
			{
				// Modify complete Locked
				{
					modState := *s1
					modState.Allocation.Locked = s2.Allocation.Locked
					ret = append(ret, modState)
				}
				// Modify AppID
				{
					modState := *s1
					modState.Allocation.Locked[0].ID = s2.Allocation.Locked[0].ID
					ret = append(ret, modState)
				}
				// Modify Bals
				{
					modState := *s1
					modState.Allocation.Locked[0].Bals = s2.Allocation.Locked[0].Bals
					ret = append(ret, modState)
				}
				// Modify Bals[0]
				{
					modState := *s1
					modState.Allocation.Locked[0].Bals[0] = s2.Allocation.Locked[0].Bals[0]
					ret = append(ret, modState)
				}
			}
		}
		// Modify Data
		{
			modState := *s1
			modState.Data = s2.Data
			ret = append(ret, modState)
		}
		// Modify IsFinal
		{
			modState := *s1
			modState.IsFinal = s2.IsFinal
			ret = append(ret, modState)
		}
	}

	return
}

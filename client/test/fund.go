// Copyright 2022 - See NOTICE file for copyright holders.
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
	"context"
	"errors"
	"math/big"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/client"
	"perun.network/go-perun/wire"
	"polycry.pt/poly-go/test"
)

// FundSetup represents the fund recovery test parameters.
type FundSetup struct {
	ChallengeDuration uint64
	FridaInitBal      channel.Bal
	FredInitBal       channel.Bal
	BalanceDelta      channel.Bal
}

// TestFundRecovery performs a test of the fund recovery functionality for the
// given setup.
func TestFundRecovery( //nolint:revive // test.Test... stutters but this is OK in this special case.
	ctx context.Context,
	t *testing.T,
	params FundSetup,
	setup func(*rand.Rand) ([2]RoleSetup, channel.Asset),
) {
	rng := test.Prng(t)

	t.Run("failing funder proposer", func(t *testing.T) {
		roles, asset := setup(rng)
		roles[0].Funder = FailingFunder{}

		runFredFridaTest(ctx, t, rng, params, roles, asset)
	})

	t.Run("failing funder proposee", func(t *testing.T) {
		roles, asset := setup(rng)
		roles[1].Funder = FailingFunder{}

		runFredFridaTest(ctx, t, rng, params, roles, asset)
	})

	t.Run("failing funder both sides", func(t *testing.T) {
		roles, asset := setup(rng)
		roles[0].Funder = FailingFunder{}
		roles[1].Funder = FailingFunder{}

		runFredFridaTest(ctx, t, rng, params, roles, asset)
	})
}

//nolint:thelper // The linter thinks this is a helper function, but it is not.
func runFredFridaTest(
	ctx context.Context,
	t *testing.T,
	rng *rand.Rand,
	params FundSetup,
	setups [2]RoleSetup,
	asset channel.Asset,
) {
	const (
		fridaIdx = 0
		fredIdx  = 1
		numParts = 2
	)
	var (
		challengeDuration = params.ChallengeDuration
		fridaInitBal      = params.FridaInitBal
		fredInitBal       = params.FredInitBal
	)

	clients := NewClients(t, rng, setups[:])
	frida, fred := clients[fridaIdx], clients[fredIdx]
	fridaWireAddr, fredWireAddr := frida.Identity.Address(), fred.Identity.Address()
	fridaWalletAddr, fredWalletAddr := frida.WalletAddress, fred.WalletAddress

	// Store client balances before running test.
	balancesBefore := channel.Balances{
		{
			frida.BalanceReader.Balance(asset),
			fred.BalanceReader.Balance(asset),
		},
	}

	// Setup proposal handling.
	// New channels and errors are passed via the corresponding Go channels.
	chsFred := make(chan *client.Channel, 1)
	errsFred := make(chan error, 1)
	go fred.Handle(
		AlwaysAcceptChannelHandler(ctx, fredWalletAddr, chsFred, errsFred),
		AlwaysRejectUpdateHandler(ctx, errsFred),
	)

	// Create the proposal.
	initAlloc := channel.NewAllocation(numParts, []int{0}, asset)
	initAlloc.SetAssetBalances(asset, []*big.Int{fridaInitBal, fredInitBal})
	parts := []map[int]wire.Address{fridaWireAddr, fredWireAddr}
	prop, err := client.NewLedgerChannelProposal(
		challengeDuration,
		fridaWalletAddr,
		initAlloc,
		parts,
	)
	require.NoError(t, err, "creating ledger channel proposal")

	// Frida sends the proposal.
	chFrida, err := frida.ProposeChannel(ctx, prop)
	t.Log(err.Error())
	require.IsType(t, &client.ChannelFundingError{}, err)
	require.NotNil(t, chFrida)
	// Frida settles the channel.
	require.NoError(t, chFrida.Settle(ctx, false))

	// Fred gets the channel and settles it afterwards.
	chFred := <-chsFred
	require.NotNil(t, chFred)
	select {
	case err := <-errsFred:
		require.IsType(t, &client.ChannelFundingError{}, err)
	case <-ctx.Done():
		require.NoError(t, ctx.Err())
	}
	// Fred settles the channel.
	require.NoError(t, chFred.Settle(ctx, false))

	// Test the final balances.
	balancesAfter := channel.Balances{
		{
			frida.BalanceReader.Balance(asset),
			fred.BalanceReader.Balance(asset),
		},
	}
	balancesDiff := balancesAfter.Sub(balancesBefore)
	expectedBalancesDiff := channel.Balances{{big.NewInt(0), big.NewInt(0)}}
	balanceDelta := params.BalanceDelta
	eq := EqualBalancesWithDelta(expectedBalancesDiff, balancesDiff, balanceDelta)
	assert.Truef(t, eq, "final ledger balances incorrect: expected balance difference %v +- %v, got %v", expectedBalancesDiff, balanceDelta, balancesDiff)
}

// FailingFunder is a funder that always fails and is used for testing.
type FailingFunder struct{}

// Fund returns an error to simulate failed funding.
func (m FailingFunder) Fund(ctx context.Context, req channel.FundingReq) error {
	return errors.New("funding failed")
}

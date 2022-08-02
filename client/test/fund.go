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
	chtest "perun.network/go-perun/channel/test"
	"perun.network/go-perun/client"
	"perun.network/go-perun/wire"
	"polycry.pt/poly-go/test"
)

// FundSetup represents the fund recovery test parameters.
type FundSetup struct {
	ChallengeDuration uint64
	FridaInitBal      channel.Bal
	FredInitBal       channel.Bal
}

// TestFundRecovery performs a test of the fund recovery functionality for the
// given setup.
func TestFundRecovery(ctx context.Context, t *testing.T, params FundSetup, setupRoles func(*rand.Rand) [2]RoleSetup) { //nolint:revive // test.Test... stutters but this is OK in this special case.
	rng := test.Prng(t)

	t.Run("failing funder proposer", func(t *testing.T) {
		setups := setupRoles(rng)
		setups[0].Funder = FailingFunder{}

		runFredFridaTest(ctx, t, rng, params, setups)
	})

	t.Run("failing funder proposee", func(t *testing.T) {
		setups := setupRoles(rng)
		setups[1].Funder = FailingFunder{}

		runFredFridaTest(ctx, t, rng, params, setups)
	})

	t.Run("failing funder both sides", func(t *testing.T) {
		setups := setupRoles(rng)
		setups[0].Funder = FailingFunder{}
		setups[1].Funder = FailingFunder{}

		runFredFridaTest(ctx, t, rng, params, setups)
	})
}

//nolint:thelper // The linter thinks this is a helper function, but it is not.
func runFredFridaTest(ctx context.Context, t *testing.T, rng *rand.Rand, params FundSetup, setups [2]RoleSetup) {
	const (
		fridaIdx = 0
		fredIdx  = 1
		numParts = 2
	)
	var (
		challengeDuration = params.ChallengeDuration
		fridaInitBal      = params.FredInitBal
		fredInitBal       = params.FredInitBal
	)

	clients := NewClients(t, rng, setups[:])
	frida, fred := clients[fridaIdx], clients[fredIdx]
	fridaWireAddr, fredWireAddr := frida.Identity.Address(), fred.Identity.Address()
	fridaWalletAddr, fredWalletAddr := frida.WalletAddress, fred.WalletAddress

	// The channel into which Fred's created ledger channel is sent into.
	chsFred := make(chan *client.Channel, 1)
	errsFred := make(chan error, 1)
	go fred.Handle(
		AlwaysAcceptChannelHandler(ctx, fredWalletAddr, chsFred, errsFred),
		AlwaysRejectUpdateHandler(ctx, errsFred),
	)

	// Create the proposal.
	asset := chtest.NewRandomAsset(rng)
	initAlloc := channel.NewAllocation(numParts, asset)
	initAlloc.SetAssetBalances(asset, []*big.Int{fridaInitBal, fredInitBal})
	parts := []wire.Address{fridaWireAddr, fredWireAddr}
	prop, err := client.NewLedgerChannelProposal(
		challengeDuration,
		fridaWalletAddr,
		initAlloc,
		parts,
	)
	require.NoError(t, err, "creating ledger channel proposal")

	// Frida sends the proposal.
	chFrida, err := frida.ProposeChannel(ctx, prop)
	require.IsType(t, &client.ChannelFundingError{}, err)
	require.NotNil(t, chFrida)
	// Frida settles the channel.
	require.NoError(t, chFrida.Settle(ctx, false))

	// Fred gets the channel and settles it afterwards.
	chFred := <-chsFred
	err = <-errsFred
	require.IsType(t, &client.ChannelFundingError{}, err)
	require.NotNil(t, chFred)
	// Fred settles the channel.
	require.NoError(t, chFred.Settle(ctx, false))

	// Test the final balances.
	fridaFinalBal := frida.BalanceReader.Balance(fridaWalletAddr, asset)
	assert.Truef(t, fridaFinalBal.Cmp(big.NewInt(0)) == 0, "frida: wrong final balance: got %v, expected %v", fridaFinalBal, fridaInitBal)
	fredFinalBal := fred.BalanceReader.Balance(fredWalletAddr, asset)
	assert.Truef(t, fredFinalBal.Cmp(big.NewInt(0)) == 0, "fred: wrong final balance: got %v, expected %v", fredFinalBal, fredInitBal)
}

// FailingFunder is a funder that always fails and is used for testing.
type FailingFunder struct{}

// Fund returns an error to simulate failed funding.
func (m FailingFunder) Fund(ctx context.Context, req channel.FundingReq) error {
	return errors.New("funding failed")
}

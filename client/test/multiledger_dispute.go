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
	"context"
	"math/big"
	"testing"
	"time"

	"perun.network/go-perun/wallet"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/client"
	"perun.network/go-perun/wire"
)

// TestMultiLedgerDispute runs an end-to-end test of the multi-ledger
// functionality in the dispute case for the implementation specified in the
// test setup.
//
//nolint:revive // test.Test... stutters but this is OK in this special case.
func TestMultiLedgerDispute(
	ctx context.Context,
	t *testing.T,
	mlt MultiLedgerSetup,
	challengeDuration uint64,
) {
	require := require.New(t)
	alice, bob := mlt.Client1, mlt.Client2

	// Store client balances before running test.
	balancesBefore := channel.Balances{
		{
			mlt.Client1.BalanceReader1.Balance(mlt.Asset1),
			mlt.Client2.BalanceReader1.Balance(mlt.Asset1),
		},
		{
			mlt.Client1.BalanceReader2.Balance(mlt.Asset2),
			mlt.Client2.BalanceReader2.Balance(mlt.Asset2),
		},
	}

	// Define initial and update balances.
	initBals := mlt.InitBalances
	updateBals := mlt.UpdateBalances1

	// Establish ledger channel between Alice and Bob.

	bID1 := wallet.BackendID(mlt.Asset1.LedgerBackendID().BackendID())
	bID2 := wallet.BackendID(mlt.Asset2.LedgerBackendID().BackendID())
	// Create channel proposal.
	parts := []map[wallet.BackendID]wire.Address{alice.WireAddress, bob.WireAddress}
	initAlloc := channel.NewAllocation(len(parts), []wallet.BackendID{bID1, bID2}, mlt.Asset1, mlt.Asset2)
	initAlloc.Balances = initBals
	prop, err := client.NewLedgerChannelProposal(
		challengeDuration,
		alice.WalletAddress,
		initAlloc,
		parts,
	)
	require.NoError(err, "creating ledger channel proposal")

	// Setup proposal handler.
	channels := make(chan *client.Channel, 1)
	errs := make(chan error)
	go alice.Handle(
		AlwaysRejectChannelHandler(ctx, errs),
		AlwaysAcceptUpdateHandler(ctx, errs),
	)
	go bob.Handle(
		AlwaysAcceptChannelHandler(ctx, bob.WalletAddress, channels, errs),
		AlwaysAcceptUpdateHandler(ctx, errs),
	)

	// Open channel.
	chAliceBob, err := alice.ProposeChannel(ctx, prop)
	require.NoError(err, "opening channel between Alice and Bob")
	var chBobAlice *client.Channel
	select {
	case chBobAlice = <-channels:
	case err := <-errs:
		t.Fatalf("Error in go-routine: %v", err)
	}

	// Start Bob's watcher.
	go func() {
		errs <- chBobAlice.Watch(bob)
	}()
	// Wait until watcher is active.
	time.Sleep(100 * time.Millisecond) //nolint:gomnd // The 100ms is a guess on how long the watcher needs to setup.

	// Notify Bob when an update is complete.
	done := make(chan struct{}, 1)
	chBobAlice.OnUpdate(func(from, to *channel.State) {
		done <- struct{}{}
	})

	// Update channel.
	err = chAliceBob.Update(ctx, func(s *channel.State) {
		s.Balances = updateBals
	})
	require.NoError(err)

	// Wait until Bob's watcher processed the update.
	<-done
	time.Sleep(100 * time.Millisecond) //nolint:gomnd // The 100ms is a guess on how long the watcher needs to catch up.

	// Alice registers state on L1 adjudicator.
	req1 := client.NewTestChannel(chAliceBob).AdjudicatorReq()
	err = alice.Adjudicator1.Register(ctx, req1, nil)
	require.NoError(err)

	e := <-bob.Events
	require.IsType(e, &channel.RegisteredEvent{})
	err = e.(*channel.RegisteredEvent).TimeoutV.Wait(ctx)
	require.NoError(err)

	// Settle.
	err = chAliceBob.Settle(ctx, false)
	require.NoError(err)
	err = chBobAlice.Settle(ctx, false)
	require.NoError(err)

	// Close the channels.
	require.NoError(chAliceBob.Close())
	require.NoError(chBobAlice.Close())

	// Check final balances.
	balancesAfter := channel.Balances{
		{
			mlt.Client1.BalanceReader1.Balance(mlt.Asset1),
			mlt.Client2.BalanceReader1.Balance(mlt.Asset1),
		},
		{
			mlt.Client1.BalanceReader2.Balance(mlt.Asset2),
			mlt.Client2.BalanceReader2.Balance(mlt.Asset2),
		},
	}

	balancesDiff := balancesAfter.Sub(balancesBefore)
	expectedBalancesDiff := updateBals.Sub(initBals)
	eq := EqualBalancesWithDelta(expectedBalancesDiff, balancesDiff, mlt.BalanceDelta)
	assert.Truef(t, eq, "final ledger balances incorrect: expected balance difference %v +- %v, got %v", expectedBalancesDiff, mlt.BalanceDelta, balancesDiff)
}

// EqualBalancesWithDelta checks whether the given balances are equal up to
// delta.
func EqualBalancesWithDelta(
	bals1 channel.Balances,
	bals2 channel.Balances,
	delta channel.Bal,
) bool {
	if len(bals1) != len(bals2) {
		return false
	}

	for i, assetBals1 := range bals1 {
		assetBals2 := bals2[i]
		if len(assetBals1) != len(assetBals2) {
			return false
		}

		for j, bal1 := range assetBals1 {
			bal2 := assetBals2[j]
			lb := new(big.Int).Sub(bal1, delta)
			ub := new(big.Int).Add(bal1, delta)
			if bal2.Cmp(lb) < 0 || bal2.Cmp(ub) > 0 {
				return false
			}
		}
	}
	return true
}

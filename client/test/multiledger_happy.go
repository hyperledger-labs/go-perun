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
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/client"
	"perun.network/go-perun/wire"
)

// TestMultiLedgerHappy runs an end-to-end test of the multi-ledger
// functionality in the optimistic case for the implementation specified in the
// test setup.
//nolint:revive // test.Test... stutters but this is OK in this special case.
func TestMultiLedgerHappy(t *testing.T, mlt MultiLedgerSetup, challengeDuration uint64) {
	ctx := context.Background()
	require := require.New(t)
	alice, bob := mlt.Client1, mlt.Client2

	// Define initial balances.
	//nolint:gomnd // We allow the balances to be magic numbers.
	initBals := channel.Balances{
		{big.NewInt(10), big.NewInt(0)}, // Asset 1.
		{big.NewInt(0), big.NewInt(10)}, // Asset 2.
	}
	//nolint:gomnd
	updateBals1 := channel.Balances{
		{big.NewInt(5), big.NewInt(5)}, // Asset 1.
		{big.NewInt(3), big.NewInt(7)}, // Asset 2.
	}
	//nolint:gomnd
	updateBals2 := channel.Balances{
		{big.NewInt(1), big.NewInt(9)}, // Asset 1.
		{big.NewInt(5), big.NewInt(5)}, // Asset 2.
	}

	// Establish ledger channel between Alice and Bob.

	// Create channel proposal.
	parts := []wire.Address{alice.WireAddress, bob.WireAddress}
	initAlloc := channel.NewAllocation(len(parts), mlt.Asset1, mlt.Asset2)
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
	require.NoError(err, "opening channel between Alice and Ingrid")
	var chBobAlice *client.Channel
	select {
	case chBobAlice = <-channels:
	case err := <-errs:
		t.Fatalf("Error in go-routine: %v", err)
	}

	// Update channel.
	err = chAliceBob.Update(ctx, func(s *channel.State) {
		s.Balances = updateBals1
	})
	require.NoError(err)

	err = chBobAlice.Update(ctx, func(s *channel.State) {
		s.Balances = updateBals2
	})
	require.NoError(err)

	err = chAliceBob.Update(ctx, func(s *channel.State) {
		s.IsFinal = true
	})
	require.NoError(err)

	// Close channel.
	err = chAliceBob.Settle(ctx, false)
	require.NoError(err)
	err = chBobAlice.Settle(ctx, false)
	require.NoError(err)
}

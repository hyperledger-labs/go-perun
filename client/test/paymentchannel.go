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
	"math/rand"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/client"
	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wire"
)

// PaymentChannelSetup contains the setup for a payment channel test.
type PaymentChannelSetup struct {
	Clients            [2]RoleSetup
	ChallengeDuration  uint64
	Asset              channel.Asset
	Balances           PaymentChannelBalances
	BalanceDelta       channel.Bal
	Rng                *rand.Rand
	WaitWatcherTimeout time.Duration
	IsUTXO             bool
}

// PaymentChannelBalances contains a description of the balances that will be
// used during a virtual channel test.
type PaymentChannelBalances struct {
	InitBalsAliceBob []*big.Int
	BalsUpdated      []*big.Int
	FinalBals        []*big.Int
}

type paymentChannelTest struct {
	alice          *Client
	bob            *Client
	chAliceBob     *client.Channel
	chBobAlice     *client.Channel
	balsUpdated    []*big.Int
	initBalsAlice  *big.Int
	initBalsBob    *big.Int
	finalBalsAlice *big.Int
	finalBalsBob   *big.Int
	errs           chan error
	asset          channel.Asset
	balancesBefore channel.Balances
	isUTXO         bool
}

// TestPaymentChannelOptimistic tests payment channel functionality in the happy case.
func TestPaymentChannelOptimistic( // test.Test... stutters but OK here.
	ctx context.Context,
	t *testing.T,
	setup PaymentChannelSetup,
) {
	pct := setupPaymentChannelTest(t, ctx, setup)
	assert := assert.New(t)

	err := pct.chAliceBob.Update(ctx, func(s *channel.State) {
		s.IsFinal = true
	})
	assert.NoError(err, "update final state")

	isSecondary := false

	// Settle the channels in a random order.
	chs := []*client.Channel{pct.chAliceBob, pct.chBobAlice}
	perm := rand.Perm(len(chs))
	t.Logf("Settle order = %v", perm)
	for _, i := range perm {
		err := chs[i].Settle(ctx, isSecondary)
		isSecondary = true

		assert.NoErrorf(err, "settle channel: %d", i)
	}

	// Check final balances.
	balancesAfter := channel.Balances{
		{
			pct.alice.BalanceReader.Balance(pct.asset),
			pct.bob.BalanceReader.Balance(pct.asset),
		},
	}

	balancesDiff := balancesAfter.Sub(pct.balancesBefore)
	expectedBalancesDiff := channel.Balances{
		{
			new(big.Int).Sub(pct.finalBalsAlice, pct.initBalsAlice),
			new(big.Int).Sub(pct.finalBalsBob, pct.initBalsBob),
		},
	}
	balanceDelta := setup.BalanceDelta
	eq := EqualBalancesWithDelta(expectedBalancesDiff, balancesDiff, balanceDelta)
	assert.Truef(eq, "final ledger balances incorrect: expected balance difference %v +- %v, got %v", expectedBalancesDiff, balanceDelta, balancesDiff)
}

// TestPaymentChannelDispute tests payment channel functionality in the dispute case.
func TestPaymentChannelDispute( //nolint:revive // test.Test... stutters but OK here. {
	ctx context.Context,
	t *testing.T,
	setup PaymentChannelSetup,
) {
	pct := setupPaymentChannelTest(t, ctx, setup)
	assert := assert.New(t)
	waitTimeout := setup.WaitWatcherTimeout
	chs := []*client.Channel{pct.chAliceBob, pct.chBobAlice}
	// Register the channels in a random order.
	perm := rand.Perm(len(chs))
	t.Logf("perm = %v", perm)
	for _, i := range perm {
		err := client.NewTestChannel(chs[i]).Register(ctx)
		assert.NoErrorf(err, "register channel: %d", i)
		time.Sleep(waitTimeout) // Sleep to ensure that events have been processed and local client states have been updated.
	}

	// Settle the channels in a random order.
	isSecondary := false
	perm = rand.Perm(len(chs))
	t.Logf("Settle order = %v", perm)
	for _, i := range perm {
		err := chs[i].Settle(ctx, isSecondary)
		isSecondary = true
		assert.NoErrorf(err, "settle channel: %d", i)
	}

	// Check final balances.
	balancesAfter := channel.Balances{
		{
			pct.alice.BalanceReader.Balance(pct.asset),
			pct.bob.BalanceReader.Balance(pct.asset),
		},
	}

	balancesDiff := balancesAfter.Sub(pct.balancesBefore)
	expectedBalancesDiff := channel.Balances{
		{
			new(big.Int).Sub(pct.finalBalsAlice, pct.initBalsAlice),
			new(big.Int).Sub(pct.finalBalsBob, pct.initBalsBob),
		},
	}
	balanceDelta := setup.BalanceDelta
	eq := EqualBalancesWithDelta(expectedBalancesDiff, balancesDiff, balanceDelta)
	assert.Truef(eq, "final ledger balances incorrect: expected balance difference %v +- %v, got %v", expectedBalancesDiff, balanceDelta, balancesDiff)
}

func setupPaymentChannelTest(
	t *testing.T,
	ctx context.Context,
	setup PaymentChannelSetup,
) (pct paymentChannelTest) {
	t.Helper()

	// Set test values.
	asset := setup.Asset
	pct.asset = asset
	pct.initBalsAlice = setup.Balances.InitBalsAliceBob[0]
	pct.initBalsBob = setup.Balances.InitBalsAliceBob[1]
	pct.balsUpdated = setup.Balances.BalsUpdated
	pct.finalBalsAlice = setup.Balances.FinalBals[0]
	pct.finalBalsBob = setup.Balances.FinalBals[1]
	pct.isUTXO = setup.IsUTXO

	const errBufferLen = 10
	pct.errs = make(chan error, errBufferLen)

	// Setup clients.
	roles := setup.Clients
	clients := NewClients(t, setup.Rng, roles[:])
	alice, bob := clients[0], clients[1]
	pct.alice, pct.bob = alice, bob

	// Store client balances before running test.
	pct.balancesBefore = channel.Balances{
		{
			pct.alice.BalanceReader.Balance(pct.asset),
			pct.bob.BalanceReader.Balance(pct.asset),
		},
	}

	// Setup Bob's proposal and update handler.
	channelsBob := make(chan *client.Channel, 1)
	var openingProposalHandlerBob client.ProposalHandlerFunc = func(cp client.ChannelProposal, pr *client.ProposalResponder) {
		switch cp := cp.(type) {
		case *client.LedgerChannelProposalMsg:
			ch, err := pr.Accept(ctx, cp.Accept(bob.WalletAddress, client.WithRandomNonce()))
			if err != nil {
				pct.errs <- errors.WithMessage(err, "accepting ledger channel proposal")
			}
			channelsBob <- ch
		default:
			pct.errs <- errors.Errorf("invalid channel proposal: %v", cp)
		}
	}
	var updateProposalHandlerBob client.UpdateHandlerFunc = func(
		s *channel.State, cu client.ChannelUpdate, ur *client.UpdateResponder,
	) {
		err := ur.Accept(ctx)
		if err != nil {
			pct.errs <- errors.WithMessage(err, "Bob: accepting channel update")
		}
	}

	go bob.Handle(openingProposalHandlerBob, updateProposalHandlerBob) //nolint:contextcheck

	// Establish ledger channel between Alice and Ingrid.
	peersAlice := []map[wallet.BackendID]wire.Address{wire.AddressMapfromAccountMap(alice.Identity), wire.AddressMapfromAccountMap(bob.Identity)}
	var bID wallet.BackendID
	for i := range peersAlice[0] {
		bID = i
		break
	}
	initAllocAlice := channel.NewAllocation(len(peersAlice), []wallet.BackendID{bID}, asset)
	initAllocAlice.SetAssetBalances(asset, []channel.Bal{pct.initBalsAlice, pct.initBalsBob})
	lcpAlice, err := client.NewLedgerChannelProposal(
		setup.ChallengeDuration,
		alice.WalletAddress,
		initAllocAlice,
		peersAlice,
	)
	require.NoError(t, err, "creating ledger channel proposal")

	pct.chAliceBob, err = alice.ProposeChannel(ctx, lcpAlice)
	require.NoError(t, err, "opening channel between Alice and Ingrid")

	select {
	case pct.chBobAlice = <-channelsBob:
	case err := <-pct.errs:
		t.Fatalf("Error in go-routine: %v", err)
	}

	err = pct.chAliceBob.Update(ctx, func(s *channel.State) {
		s.Balances = channel.Balances{pct.balsUpdated}
	})
	require.NoError(t, err, "updating virtual channel")

	return pct
}

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
	"math/rand"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/client"
	"perun.network/go-perun/wire"
	"polycry.pt/poly-go/sync"
)

// VirtualChannelSetup contains the setup for a virtual channel test.
type VirtualChannelSetup struct {
	Clients            [3]RoleSetup
	ChallengeDuration  uint64
	Asset              channel.Asset
	Balances           VirtualChannelBalances
	BalanceDelta       channel.Bal
	Rng                *rand.Rand
	WaitWatcherTimeout time.Duration
	IsUTXO             bool
}

// TestVirtualChannelOptimistic tests virtual channel functionality in the
// optimistic case.
func TestVirtualChannelOptimistic( //nolint:revive // test.Test... stutters but OK here.
	ctx context.Context,
	t *testing.T,
	setup VirtualChannelSetup,
) {
	vct := setupVirtualChannelTest(t, ctx, setup)

	// Settle.
	var success sync.WaitGroup
	chs := []*client.Channel{vct.chAliceBob, vct.chBobAlice}
	success.Add(len(chs))
	for _, ch := range chs {
		go func(ch *client.Channel) {
			err := ch.Settle(ctx, false)
			if err != nil {
				vct.errs <- err
				return
			}
			success.Done()
		}(ch)
	}

	// Wait for success or error.
	select {
	case <-success.WaitCh():
	case err := <-vct.errs:
		t.Fatalf("Error in go-routine: %v", err)
	}

	// Test final balances.
	err := vct.chAliceIngrid.State().Balances.AssertEqual(channel.Balances{vct.finalBalsAlice})
	assert.NoError(t, err, "Alice: invalid final balances")
	err = vct.chBobIngrid.State().Balances.AssertEqual(channel.Balances{vct.finalBalsBob})
	assert.NoError(t, err, "Bob: invalid final balances")

	// Close the parents.
	chs = []*client.Channel{vct.chAliceIngrid, vct.chIngridAlice, vct.chBobIngrid, vct.chIngridBob}
	isSecondary := [2]bool{false, false}
	perm := rand.Perm(len(chs))
	t.Logf("Settle order = %v", perm)
	for _, i := range perm {
		var err error
		if i < 2 {
			err = chs[i].Settle(ctx, isSecondary[0])
			isSecondary[0] = true
		} else {
			err = chs[i].Settle(ctx, isSecondary[1])
			isSecondary[1] = true
		}
		assert.NoErrorf(t, err, "settle channel: %d", i)
	}

	// Check final balances.
	balancesAfter := channel.Balances{
		{
			vct.alice.BalanceReader.Balance(vct.asset),
			vct.bob.BalanceReader.Balance(vct.asset),
			vct.ingrid.BalanceReader.Balance(vct.asset),
		},
	}

	balancesDiff := balancesAfter.Sub(vct.balancesBefore)
	expectedBalancesDiff := channel.Balances{
		{
			new(big.Int).Sub(vct.finalBalsAlice[0], vct.initBalsAlice[0]),
			new(big.Int).Sub(vct.finalBalsBob[0], vct.initBalsBob[0]),
			big.NewInt(0),
		},
	}
	balanceDelta := setup.BalanceDelta
	eq := EqualBalancesWithDelta(expectedBalancesDiff, balancesDiff, balanceDelta)
	assert.Truef(t, eq, "final ledger balances incorrect: expected balance difference %v +- %v, got %v", expectedBalancesDiff, balanceDelta, balancesDiff)
}

// TestVirtualChannelDispute tests virtual channel functionality in the dispute
// case.
func TestVirtualChannelDispute( //nolint:revive // test.Test... stutters but OK here.
	ctx context.Context,
	t *testing.T,
	setup VirtualChannelSetup,
) {
	vct := setupVirtualChannelTest(t, ctx, setup)
	assert := assert.New(t)
	waitTimeout := setup.WaitWatcherTimeout

	time.Sleep(waitTimeout) // Sleep to ensure that events have been processed and local client states have been updated.

	chs := []*client.Channel{vct.chAliceIngrid, vct.chIngridAlice, vct.chBobIngrid, vct.chIngridBob}
	// Register the channels in a random order.
	perm := rand.Perm(len(chs))
	t.Logf("perm = %v", perm)
	for _, i := range perm {
		err := client.NewTestChannel(chs[i]).Register(ctx)
		assert.NoErrorf(err, "register channel: %d", i)
		time.Sleep(waitTimeout) // Sleep to ensure that events have been processed and local client states have been updated.
	}

	// Settle the channels in a random order.
	isSecondary := [2]bool{false, false}
	perm = rand.Perm(len(chs))
	t.Logf("Settle order = %v", perm)
	for _, i := range perm {
		var err error
		if i < 2 {
			err = chs[i].Settle(ctx, isSecondary[0])
			isSecondary[0] = true
		} else {
			err = chs[i].Settle(ctx, isSecondary[1])
			isSecondary[1] = true
		}
		assert.NoErrorf(err, "settle channel: %d", i)
	}

	// Check final balances.
	balancesAfter := channel.Balances{
		{
			vct.alice.BalanceReader.Balance(vct.asset),
			vct.bob.BalanceReader.Balance(vct.asset),
			vct.ingrid.BalanceReader.Balance(vct.asset),
		},
	}

	balancesDiff := balancesAfter.Sub(vct.balancesBefore)
	expectedBalancesDiff := channel.Balances{
		{
			new(big.Int).Sub(vct.finalBalsAlice[0], vct.initBalsAlice[0]),
			new(big.Int).Sub(vct.finalBalsBob[0], vct.initBalsBob[0]),
			big.NewInt(0),
		},
	}
	balanceDelta := setup.BalanceDelta
	eq := EqualBalancesWithDelta(expectedBalancesDiff, balancesDiff, balanceDelta)
	assert.Truef(eq, "final ledger balances incorrect: expected balance difference %v +- %v, got %v", expectedBalancesDiff, balanceDelta, balancesDiff)
}

type virtualChannelTest struct {
	alice              *Client
	bob                *Client
	ingrid             *Client
	chAliceIngrid      *client.Channel
	chIngridAlice      *client.Channel
	chBobIngrid        *client.Channel
	chIngridBob        *client.Channel
	chAliceBob         *client.Channel
	chBobAlice         *client.Channel
	virtualBalsUpdated []*big.Int
	initBalsAlice      []*big.Int
	initBalsBob        []*big.Int
	finalBalsAlice     []*big.Int
	finalBalsBob       []*big.Int
	finalBalIngrid     *big.Int
	errs               chan error
	asset              channel.Asset
	balancesBefore     channel.Balances
	isUTXO             bool
	parentIDs          [2]channel.ID
}

// VirtualChannelBalances contains a description of the balances that will be
// used during a virtual channel test.
type VirtualChannelBalances struct {
	InitBalsAliceIngrid []*big.Int
	InitBalsBobIngrid   []*big.Int
	InitBalsAliceBob    []*big.Int
	VirtualBalsUpdated  []*big.Int
	FinalBalsAlice      []*big.Int
	FinalBalsBob        []*big.Int
}

func setupVirtualChannelTest(
	t *testing.T,
	ctx context.Context,
	setup VirtualChannelSetup,
) (vct virtualChannelTest) {
	t.Helper()
	require := require.New(t)

	// Set test values.
	asset := setup.Asset
	vct.asset = asset
	vct.initBalsAlice = setup.Balances.InitBalsAliceIngrid
	vct.initBalsBob = setup.Balances.InitBalsBobIngrid
	initBalsVirtual := setup.Balances.InitBalsAliceBob
	vct.virtualBalsUpdated = setup.Balances.VirtualBalsUpdated
	vct.finalBalsAlice = setup.Balances.FinalBalsAlice
	vct.finalBalsBob = setup.Balances.FinalBalsBob
	vct.finalBalIngrid = new(big.Int).Add(vct.finalBalsAlice[1], vct.finalBalsBob[1])
	vct.isUTXO = setup.IsUTXO

	const errBufferLen = 10
	vct.errs = make(chan error, errBufferLen)

	// Setup clients.
	roles := setup.Clients
	clients := NewClients(t, setup.Rng, roles[:])
	alice, bob, ingrid := clients[0], clients[1], clients[2]
	vct.alice, vct.bob, vct.ingrid = alice, bob, ingrid

	// Store client balances before running test.
	vct.balancesBefore = channel.Balances{
		{
			vct.alice.BalanceReader.Balance(vct.asset),
			vct.bob.BalanceReader.Balance(vct.asset),
			vct.ingrid.BalanceReader.Balance(vct.asset),
		},
	}

	channelsIngrid := make(chan *client.Channel, 1)
	var openingProposalHandlerIngrid client.ProposalHandlerFunc = func(cp client.ChannelProposal, pr *client.ProposalResponder) {
		switch cp := cp.(type) {
		case *client.LedgerChannelProposalMsg:
			ch, err := pr.Accept(ctx, cp.Accept(ingrid.WalletAddress, client.WithRandomNonce()))
			if err != nil {
				vct.errs <- errors.WithMessage(err, "accepting ledger channel proposal")
			}
			channelsIngrid <- ch
		default:
			vct.errs <- errors.Errorf("invalid channel proposal: %v", cp)
		}
	}
	var updateProposalHandlerIngrid client.UpdateHandlerFunc = func(
		s *channel.State, cu client.ChannelUpdate, ur *client.UpdateResponder,
	) {
	}
	go ingrid.Client.Handle(openingProposalHandlerIngrid, updateProposalHandlerIngrid)

	// Establish ledger channel between Alice and Ingrid.
	peersAlice := []wire.Address{alice.Identity.Address(), ingrid.Identity.Address()}
	initAllocAlice := channel.NewAllocation(len(peersAlice), asset)
	initAllocAlice.SetAssetBalances(asset, vct.initBalsAlice)
	lcpAlice, err := client.NewLedgerChannelProposal(
		setup.ChallengeDuration,
		alice.WalletAddress,
		initAllocAlice,
		peersAlice,
	)
	require.NoError(err, "creating ledger channel proposal")

	vct.chAliceIngrid, err = alice.ProposeChannel(ctx, lcpAlice)
	require.NoError(err, "opening channel between Alice and Ingrid")
	select {
	case vct.chIngridAlice = <-channelsIngrid:
	case err := <-vct.errs:
		t.Fatalf("Error in go-routine: %v", err)
	}

	// Establish ledger channel between Bob and Ingrid.
	peersBob := []wire.Address{bob.Identity.Address(), ingrid.Identity.Address()}
	initAllocBob := channel.NewAllocation(len(peersBob), asset)
	initAllocBob.SetAssetBalances(asset, vct.initBalsBob)
	lcpBob, err := client.NewLedgerChannelProposal(
		setup.ChallengeDuration,
		bob.WalletAddress,
		initAllocBob,
		peersBob,
	)
	require.NoError(err, "creating ledger channel proposal")

	vct.chBobIngrid, err = bob.ProposeChannel(ctx, lcpBob)
	require.NoError(err, "opening channel between Bob and Ingrid")
	select {
	case vct.chIngridBob = <-channelsIngrid:
	case err := <-vct.errs:
		t.Fatalf("Error in go-routine: %v", err)
	}

	// Setup Bob's proposal and update handler.
	channelsBob := make(chan *client.Channel, 1)
	var openingProposalHandlerBob client.ProposalHandlerFunc = func(
		cp client.ChannelProposal, pr *client.ProposalResponder,
	) {
		switch cp := cp.(type) {
		case *client.VirtualChannelProposalMsg:
			ch, err := pr.Accept(ctx, cp.Accept(bob.WalletAddress))
			if err != nil {
				vct.errs <- errors.WithMessage(err, "accepting virtual channel proposal")
			}
			channelsBob <- ch
		default:
			vct.errs <- errors.Errorf("invalid channel proposal: %v", cp)
		}
	}
	var updateProposalHandlerBob client.UpdateHandlerFunc = func(
		s *channel.State, cu client.ChannelUpdate, ur *client.UpdateResponder,
	) {
		err := ur.Accept(ctx)
		if err != nil {
			vct.errs <- errors.WithMessage(err, "Bob: accepting channel update")
		}
	}
	go bob.Client.Handle(openingProposalHandlerBob, updateProposalHandlerBob)

	// Establish virtual channel between Alice and Bob via Ingrid.
	initAllocVirtual := channel.Allocation{
		Assets:   []channel.Asset{asset},
		Balances: [][]channel.Bal{initBalsVirtual},
	}
	indexMapAlice := []channel.Index{0, 1}
	indexMapBob := []channel.Index{1, 0}

	vct.parentIDs[0] = vct.chAliceIngrid.ID()
	vct.parentIDs[1] = vct.chBobIngrid.ID()

	var vcp *client.VirtualChannelProposalMsg
	if setup.IsUTXO {
		// UTXO Chains need additional auxiliary data to be able to
		// create a virtual channel.
		var aux channel.Aux
		copy(aux[:channel.IDLen], vct.parentIDs[0][:])
		copy(aux[channel.IDLen:], vct.parentIDs[1][:])

		vcp, err = client.NewVirtualChannelProposal(
			setup.ChallengeDuration,
			alice.WalletAddress,
			&initAllocVirtual,
			[]wire.Address{alice.Identity.Address(), bob.Identity.Address()},
			[]channel.ID{vct.chAliceIngrid.ID(), vct.chBobIngrid.ID()},
			[][]channel.Index{indexMapAlice, indexMapBob},
			client.WithAux(aux),
		)
	} else {
		vcp, err = client.NewVirtualChannelProposal(
			setup.ChallengeDuration,
			alice.WalletAddress,
			&initAllocVirtual,
			[]wire.Address{alice.Identity.Address(), bob.Identity.Address()},
			[]channel.ID{vct.chAliceIngrid.ID(), vct.chBobIngrid.ID()},
			[][]channel.Index{indexMapAlice, indexMapBob},
		)
	}
	require.NoError(err, "creating virtual channel proposal")

	vct.chAliceBob, err = alice.ProposeChannel(ctx, vcp)
	require.NoError(err, "opening channel between Alice and Bob")
	select {
	case vct.chBobAlice = <-channelsBob:
	case err := <-vct.errs:
		t.Fatalf("Error in go-routine: %v", err)
	}

	err = vct.chAliceBob.Update(ctx, func(s *channel.State) {
		s.Balances = channel.Balances{vct.virtualBalsUpdated}
	})
	require.NoError(err, "updating virtual channel")

	err = vct.chAliceBob.Update(ctx, func(s *channel.State) {
		s.IsFinal = true
	})
	require.NoError(err, "updating virtual channel")
	return vct
}

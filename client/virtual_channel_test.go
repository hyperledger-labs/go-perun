// Copyright 2021 - See NOTICE file for copyright holders.
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

package client_test

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
	chtest "perun.network/go-perun/channel/test"
	"perun.network/go-perun/client"
	ctest "perun.network/go-perun/client/test"
	"perun.network/go-perun/wire"
	"polycry.pt/poly-go/sync"
	"polycry.pt/poly-go/test"
)

const (
	challengeDuration = 10
	testDuration      = 10 * time.Second
)

func TestVirtualChannelsOptimistic(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testDuration)
	defer cancel()

	vct := setupVirtualChannelTest(t, ctx)

	// Settle.
	var success sync.WaitGroup
	settleCh := func(ch *client.Channel) {
		err := ch.Settle(ctx, false)
		if err != nil {
			vct.errs <- err
			return
		}
		success.Done()
	}
	success.Add(2)
	go settleCh(vct.chAliceBob)
	go settleCh(vct.chBobAlice)

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
}

func TestVirtualChannelsDispute(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testDuration)
	defer cancel()

	vct := setupVirtualChannelTest(t, ctx)
	assert := assert.New(t)
	waitTimeout := 100 * time.Millisecond

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
	for _, i := range rand.Perm(len(chs)) {
		err := chs[i].Settle(ctx, false)
		assert.NoErrorf(err, "settle channel: %d", i)
	}

	// Test final balances.
	vct.testFinalBalancesDispute(t)
}

func (vct *virtualChannelTest) testFinalBalancesDispute(t *testing.T) {
	t.Helper()
	assert := assert.New(t)
	backend, asset := vct.balanceReader, vct.asset
	got, expected := backend.Balance(vct.alice.Identity.Address(), asset), vct.finalBalsAlice[0]
	assert.Truef(got.Cmp(expected) == 0, "alice: wrong final balance: got %v, expected %v", got, expected)
	got, expected = backend.Balance(vct.bob.Identity.Address(), asset), vct.finalBalsBob[0]
	assert.Truef(got.Cmp(expected) == 0, "bob: wrong final balance: got %v, expected %v", got, expected)
	got, expected = backend.Balance(vct.ingrid.Identity.Address(), asset), vct.finalBalIngrid
	assert.Truef(got.Cmp(expected) == 0, "ingrid: wrong final balance: got %v, expected %v", got, expected)
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
	finalBalsAlice     []*big.Int
	finalBalsBob       []*big.Int
	finalBalIngrid     *big.Int
	errs               chan error
	balanceReader      ctest.BalanceReader
	asset              channel.Asset
}

func setupVirtualChannelTest(t *testing.T, ctx context.Context) (vct virtualChannelTest) {
	t.Helper()
	rng := test.Prng(t)
	require := require.New(t)

	// Set test values.
	asset := chtest.NewRandomAsset(rng)
	vct.asset = asset
	initBalsAlice := []*big.Int{big.NewInt(10), big.NewInt(10)}       // with Ingrid
	initBalsBob := []*big.Int{big.NewInt(10), big.NewInt(10)}         // with Ingrid
	initBalsVirtual := []*big.Int{big.NewInt(5), big.NewInt(5)}       // Alice proposes
	vct.virtualBalsUpdated = []*big.Int{big.NewInt(2), big.NewInt(8)} // Send 3.
	vct.finalBalsAlice = []*big.Int{big.NewInt(7), big.NewInt(13)}
	vct.finalBalsBob = []*big.Int{big.NewInt(13), big.NewInt(7)}
	vct.finalBalIngrid = new(big.Int).Add(vct.finalBalsAlice[1], vct.finalBalsBob[1])
	vct.errs = make(chan error, 10)

	// Setup clients.
	clients, _ := NewClients(
		t,
		rng,
		[]string{"Alice", "Bob", "Ingrid"},
	)
	alice, bob, ingrid := clients[0], clients[1], clients[2]
	vct.alice, vct.bob, vct.ingrid = alice, bob, ingrid
	vct.balanceReader = alice.BalanceReader // Assumes all clients have same backend.

	_channelsIngrid := make(chan *client.Channel, 1)
	var openingProposalHandlerIngrid client.ProposalHandlerFunc = func(cp client.ChannelProposal, pr *client.ProposalResponder) {
		switch cp := cp.(type) {
		case *client.LedgerChannelProposalMsg:
			ch, err := pr.Accept(ctx, cp.Accept(ingrid.Identity.Address(), client.WithRandomNonce()))
			if err != nil {
				vct.errs <- errors.WithMessage(err, "accepting ledger channel proposal")
			}
			_channelsIngrid <- ch
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
	initAllocAlice.SetAssetBalances(asset, initBalsAlice)
	lcpAlice, err := client.NewLedgerChannelProposal(
		challengeDuration,
		alice.Identity.Address(),
		initAllocAlice,
		peersAlice,
	)
	require.NoError(err, "creating ledger channel proposal")

	vct.chAliceIngrid, err = alice.ProposeChannel(ctx, lcpAlice)
	require.NoError(err, "opening channel between Alice and Ingrid")
	select {
	case vct.chIngridAlice = <-_channelsIngrid:
	case err := <-vct.errs:
		t.Fatalf("Error in go-routine: %v", err)
	}

	// Establish ledger channel between Bob and Ingrid.
	peersBob := []wire.Address{bob.Identity.Address(), ingrid.Identity.Address()}
	initAllocBob := channel.NewAllocation(len(peersBob), asset)
	initAllocBob.SetAssetBalances(asset, initBalsBob)
	lcpBob, err := client.NewLedgerChannelProposal(
		challengeDuration,
		bob.Identity.Address(),
		initAllocBob,
		peersBob,
	)
	require.NoError(err, "creating ledger channel proposal")

	vct.chBobIngrid, err = bob.ProposeChannel(ctx, lcpBob)
	require.NoError(err, "opening channel between Bob and Ingrid")
	select {
	case vct.chIngridBob = <-_channelsIngrid:
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
			ch, err := pr.Accept(ctx, cp.Accept(bob.Identity.Address()))
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
	vcp, err := client.NewVirtualChannelProposal(
		challengeDuration,
		alice.Identity.Address(),
		&initAllocVirtual,
		[]wire.Address{alice.Identity.Address(), bob.Identity.Address()},
		[]channel.ID{vct.chAliceIngrid.ID(), vct.chBobIngrid.ID()},
		[][]channel.Index{indexMapAlice, indexMapBob},
	)
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

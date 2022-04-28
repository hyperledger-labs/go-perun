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

package client_test

import (
	"bytes"
	"context"
	"math/big"
	"math/rand"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"perun.network/go-perun/backend/multi"
	"perun.network/go-perun/channel"
	chtest "perun.network/go-perun/channel/test"
	"perun.network/go-perun/client"
	ctest "perun.network/go-perun/client/test"
	wtest "perun.network/go-perun/wallet/test"
	"perun.network/go-perun/watcher/local"
	"perun.network/go-perun/wire"
	"perun.network/go-perun/wire/perunio"
	"polycry.pt/poly-go/test"
)

type multiLedgerTest struct {
	c1, c2 testClient
	l1, l2 *ctest.MockBackend
	a1, a2 multi.Asset
}

func TestMultiLedgerHappy(t *testing.T) {
	require := require.New(t)
	ctx, cancel := context.WithTimeout(context.Background(), testDuration)
	defer cancel()

	mlt := setupMultiLedgerTest(t)
	alice, bob := mlt.c1, mlt.c2

	// Define initial balances.
	initBals := channel.Balances{
		{big.NewInt(10), big.NewInt(0)}, // Asset 1.
		{big.NewInt(0), big.NewInt(10)}, // Asset 2.
	}
	updateBals1 := channel.Balances{
		{big.NewInt(5), big.NewInt(5)}, // Asset 1.
		{big.NewInt(3), big.NewInt(7)}, // Asset 2.
	}
	updateBals2 := channel.Balances{
		{big.NewInt(1), big.NewInt(9)}, // Asset 1.
		{big.NewInt(5), big.NewInt(5)}, // Asset 2.
	}

	// Establish ledger channel between Alice and Bob.

	// Create channel proposal.
	parts := []wire.Address{alice.wireAddress, bob.wireAddress}
	initAlloc := channel.NewAllocation(len(parts), mlt.a1, mlt.a2)
	initAlloc.Balances = initBals
	prop, err := client.NewLedgerChannelProposal(
		challengeDuration,
		alice.wireAddress,
		initAlloc,
		parts,
	)
	require.NoError(err, "creating ledger channel proposal")

	// Setup proposal handler.
	channels := make(chan *client.Channel, 1)
	errs := make(chan error)
	var channelHandler client.ProposalHandlerFunc = func(cp client.ChannelProposal, pr *client.ProposalResponder) {
		switch cp := cp.(type) {
		case *client.LedgerChannelProposalMsg:
			ch, err := pr.Accept(ctx, cp.Accept(bob.wireAddress, client.WithRandomNonce()))
			if err != nil {
				errs <- errors.WithMessage(err, "accepting ledger channel proposal")
				return
			}
			channels <- ch
		default:
			errs <- errors.Errorf("invalid channel proposal: %v", cp)
			return
		}
	}
	var updateHandler client.UpdateHandlerFunc = func(
		s *channel.State, cu client.ChannelUpdate, ur *client.UpdateResponder,
	) {
		err := ur.Accept(ctx)
		if err != nil {
			errs <- errors.WithMessage(err, "Bob: accepting channel update")
		}
	}
	go alice.Handle(channelHandler, updateHandler)
	go bob.Handle(channelHandler, updateHandler)

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

func setupMultiLedgerTest(t *testing.T) (mlt multiLedgerTest) {
	t.Helper()
	rng := test.Prng(t)

	// Setup ledgers.
	l1 := ctest.NewMockBackend(rng, "1337")
	l2 := ctest.NewMockBackend(rng, "1338")

	// Setup message bus.
	bus := wire.NewLocalBus()

	// Setup clients.
	c1 := setupClient(t, rng, l1, l2, bus)
	c2 := setupClient(t, rng, l1, l2, bus)

	// Define assets.
	a1 := NewMultiLedgerAsset(l1.ID(), chtest.NewRandomAsset(rng))
	a2 := NewMultiLedgerAsset(l2.ID(), chtest.NewRandomAsset(rng))

	return multiLedgerTest{
		c1: c1,
		c2: c2,
		l1: l1,
		l2: l2,
		a1: a1,
		a2: a2,
	}
}

type MultiLedgerAsset struct {
	id    ctest.LedgerID
	asset channel.Asset
}

func NewMultiLedgerAsset(id ctest.LedgerID, asset channel.Asset) *MultiLedgerAsset {
	return &MultiLedgerAsset{
		id:    id,
		asset: asset,
	}
}

func (a *MultiLedgerAsset) Equal(b channel.Asset) bool {
	bm, ok := b.(*MultiLedgerAsset)
	if !ok {
		return false
	}

	return a.id.MapKey() == bm.id.MapKey() && a.asset.Equal(bm.asset)
}

func (a *MultiLedgerAsset) LedgerID() multi.LedgerID {
	return a.id
}

func (a *MultiLedgerAsset) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	err := perunio.Encode(&buf, string(a.id), a.asset)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (a *MultiLedgerAsset) UnmarshalBinary(data []byte) error {
	buf := bytes.NewBuffer(data)
	return perunio.Decode(buf, string(a.id), a.asset)
}

type testClient struct {
	*client.Client
	wireAddress wire.Address
}

func setupClient(
	t *testing.T, rng *rand.Rand,
	l1, l2 *ctest.MockBackend, bus wire.Bus,
) testClient {
	t.Helper()
	require := require.New(t)

	// Setup wallet and account.
	w := wtest.NewWallet()
	acc := w.NewRandomAccount(rng)

	// Setup funder.
	funder := multi.NewFunder()
	funder.RegisterFunder(l1.ID(), l1)
	funder.RegisterFunder(l2.ID(), l2)

	// Setup adjudicator.
	adj := multi.NewAdjudicator()
	adj.RegisterAdjudicator(l1.ID(), l1)
	adj.RegisterAdjudicator(l2.ID(), l2)

	// Setup watcher.
	watcher, err := local.NewWatcher(adj)
	require.NoError(err)

	c, err := client.New(
		acc.Address(),
		bus,
		funder,
		adj,
		w,
		watcher,
	)
	require.NoError(err)

	return testClient{
		Client:      c,
		wireAddress: acc.Address(),
	}
}

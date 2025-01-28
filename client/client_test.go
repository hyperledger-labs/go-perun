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

package client_test

import (
	"context"
	"testing"
	"time"

	"perun.network/go-perun/wallet"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/channel"
	chtest "perun.network/go-perun/channel/test"
	"perun.network/go-perun/client"
	ctest "perun.network/go-perun/client/test"
	wtest "perun.network/go-perun/wallet/test"
	"perun.network/go-perun/watcher/local"
	"perun.network/go-perun/wire"
	wiretest "perun.network/go-perun/wire/test"
	"polycry.pt/poly-go/test"
)

type DummyBus struct {
	t *testing.T
}

func (d DummyBus) Publish(context.Context, *wire.Envelope) error {
	d.t.Error("DummyBus.Publish called")
	return errors.New("DummyBus.Publish called")
}

func (d DummyBus) SubscribeClient(wire.Consumer, map[wallet.BackendID]wire.Address) error {
	return nil
}

func TestClient_New_NilArgs(t *testing.T) {
	rng := test.Prng(t)
	id := wiretest.NewRandomAddress(rng)
	backend := &ctest.MockBackend{}
	b, f, a, w := &DummyBus{t}, &ctest.MockFunder{}, &ctest.MockAdjudicator{}, map[wallet.BackendID]wallet.Wallet{channel.TestBackendID: wtest.RandomWallet(channel.TestBackendID)}
	watcher, err := local.NewWatcher(backend)
	require.NoError(t, err, "initializing the watcher should not error")
	assert.Panics(t, func() { client.New(nil, b, f, a, w, watcher) })  //nolint:errcheck
	assert.Panics(t, func() { client.New(id, nil, f, a, w, watcher) }) //nolint:errcheck
	assert.Panics(t, func() { client.New(id, b, nil, a, w, watcher) }) //nolint:errcheck
	assert.Panics(t, func() { client.New(id, b, f, nil, w, watcher) }) //nolint:errcheck
	assert.Panics(t, func() { client.New(id, b, f, a, nil, watcher) }) //nolint:errcheck
	assert.Panics(t, func() { client.New(id, b, f, a, w, nil) })       //nolint:errcheck
}

func TestClient_Handle_NilArgs(t *testing.T) {
	rng := test.Prng(t)
	backend := &ctest.MockBackend{}
	watcher, err := local.NewWatcher(backend)
	require.NoError(t, err, "initializing the watcher should not error")
	c, err := client.New(wiretest.NewRandomAddress(rng),
		&DummyBus{t}, &ctest.MockFunder{}, &ctest.MockAdjudicator{}, map[wallet.BackendID]wallet.Wallet{channel.TestBackendID: wtest.RandomWallet(channel.TestBackendID)}, watcher)
	require.NoError(t, err)

	dummyUH := client.UpdateHandlerFunc(func(*channel.State, client.ChannelUpdate, *client.UpdateResponder) {})
	assert.Panics(t, func() { c.Handle(nil, dummyUH) })
	dummyPH := client.ProposalHandlerFunc(func(client.ChannelProposal, *client.ProposalResponder) {})
	assert.Panics(t, func() { c.Handle(dummyPH, nil) })
}

func TestClient_New(t *testing.T) {
	rng := test.Prng(t)
	backend := &ctest.MockBackend{}
	watcher, err := local.NewWatcher(backend)
	require.NoError(t, err, "initializing the watcher should not error")
	c, err := client.New(wiretest.NewRandomAddress(rng),
		&DummyBus{t}, &ctest.MockFunder{}, &ctest.MockAdjudicator{}, map[wallet.BackendID]wallet.Wallet{channel.TestBackendID: wtest.RandomWallet(channel.TestBackendID)}, watcher)
	assert.NoError(t, err)
	require.NotNil(t, c)
}

func TestChannelRejection(t *testing.T) {
	rng := test.Prng(t)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	roles := NewSetups(rng, []string{"Alice", "Bob"}, channel.TestBackendID)
	asset := chtest.NewRandomAsset(rng, channel.TestBackendID)
	clients := ctest.NewClients(t, rng, roles)
	require := require.New(t)
	alice, bob := clients[0], clients[1]

	// Setup proposal handler. Reject all channels.
	errs := make(chan error)
	go alice.Handle(
		ctest.AlwaysRejectChannelHandler(ctx, errs),
		ctest.AlwaysAcceptUpdateHandler(ctx, errs),
	)
	go bob.Handle(
		ctest.AlwaysRejectChannelHandler(ctx, errs),
		ctest.AlwaysAcceptUpdateHandler(ctx, errs),
	)

	// Create channel proposal.
	parts := []map[wallet.BackendID]wire.Address{wire.AddressMapfromAccountMap(alice.Identity), wire.AddressMapfromAccountMap(bob.Identity)}
	initAlloc := channel.NewAllocation(len(parts), []wallet.BackendID{channel.TestBackendID}, asset)
	prop, err := client.NewLedgerChannelProposal(
		challengeDuration,
		alice.WalletAddress,
		initAlloc,
		parts,
	)
	require.NoError(err, "creating ledger channel proposal")

	// Propose channel and await rejection.
	_, err = alice.ProposeChannel(ctx, prop)
	require.Error(err)
	require.NoError(ctx.Err())
}

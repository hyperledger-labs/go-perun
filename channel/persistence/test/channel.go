// Copyright 2024 - See NOTICE file for copyright holders.
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
	"math/rand"

	"github.com/stretchr/testify/require"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/persistence"
	ctest "perun.network/go-perun/channel/test"
	"perun.network/go-perun/wallet"
	wtest "perun.network/go-perun/wallet/test"
	"perun.network/go-perun/wire"
)

// Channel is a wrapper around a persisted channel and its participants, as well
// as the associated persister and restorer.
type Channel struct {
	accounts []map[wallet.BackendID]wallet.Account
	peers    []map[wallet.BackendID]wire.Address
	parent   *map[wallet.BackendID]channel.ID
	*persistence.StateMachine

	pr  persistence.PersistRestorer
	ctx context.Context //nolint:containedctx // This is just done for testing. Could be revised.
}

// NewRandomChannel creates a random channel with the requested persister and
// restorer, as well as the selected peer addresses for the participants (other
// than the owner's). The owner's index in the channel participants can be
// controlled via the 'user' argument. The wallet accounts and addresses used by
// the participants are generated randomly.
// The persister is notified and called to persist the new channel before it is
// returned.
func NewRandomChannel(
	ctx context.Context,
	t require.TestingT,
	pr persistence.PersistRestorer,
	user channel.Index,
	peers []map[wallet.BackendID]wire.Address,
	parent *Channel,
	rng *rand.Rand,
) (c *Channel) {
	bID := wallet.BackendID(channel.TestBackendID)
	if len(peers) > 0 {
		for i := range peers[0] {
			bID = i
		}
	}
	accs, parts := wtest.NewRandomAccounts(rng, len(peers), bID)
	params := ctest.NewRandomParams(rng, ctest.WithParts(parts))
	csm, err := channel.NewStateMachine(accs[0], *params)
	require.NoError(t, err)

	var parentID *map[wallet.BackendID]channel.ID
	if parent != nil {
		parentID = new(map[wallet.BackendID]channel.ID)
		*parentID = parent.ID()
	}

	sm := persistence.FromStateMachine(csm, pr)
	c = &Channel{
		accounts:     accs,
		peers:        peers,
		StateMachine: &sm,
		pr:           pr,
		ctx:          ctx,
		parent:       parentID,
	}

	require.NoError(t, pr.ChannelCreated(ctx, c.StateMachine, c.peers, c.parent))
	c.AssertPersisted(ctx, t)
	return c
}

func requireEqualPeers(t require.TestingT, expected, actual []map[wallet.BackendID]wire.Address) {
	require.Equal(t, len(expected), len(actual))
	for i, p := range expected {
		if !channel.EqualWireMaps(p, actual[i]) {
			t.Errorf("restored peers for channel do not match\nexpected: %v\nactual: %v",
				actual, expected)
			t.FailNow()
		}
	}
}

// AssertPersisted reads the channel state from the restorer and compares it
// to the actual channel state. If an error occurs while restoring the channel
// or if the restored channel does not match the actual channel state, then the
// test fails.
func (c *Channel) AssertPersisted(ctx context.Context, t require.TestingT) {
	ch, err := c.pr.RestoreChannel(ctx, c.ID())
	require.NoError(t, err)
	require.NotNil(t, ch)
	c.RequireEqual(t, ch)
	requireEqualPeers(t, c.peers, ch.PeersV)
	require.Equal(t, c.parent, ch.Parent)
}

// RequireEqual asserts that the channel is equal to the provided channel state.
func (c *Channel) RequireEqual(t require.TestingT, ch channel.Source) {
	require.Equal(t, c.Idx(), ch.Idx(), "Idx")
	require.Equal(t, c.Params(), ch.Params(), "Params")
	requireEqualStagingTX(t, c.StagingTX(), ch.StagingTX())
	require.Equal(t, c.CurrentTX(), ch.CurrentTX(), "CurrentTX")
	require.Equal(t, c.Phase(), ch.Phase(), "Phase")
}

// EqualStagingLoose is a test for loose equality between two staging states,
// where it is allowed for signatures to be a nil slice iff the transaction
// which it is compared to also has a nil slice OR a slice of nil sigs.
func requireEqualStagingTX(t require.TestingT, expected, actual channel.Transaction) {
	require.Equal(t, expected.State, actual.State, "StagingTX.State")
	requireEqualSigs(t, expected.Sigs, actual.Sigs)
}

func requireEqualSigs(t require.TestingT, expected, actual []wallet.Sig) {
	if expected == nil && actual == nil {
		return
	}
	actualNil := isNilSigs(actual)
	expectedNil := isNilSigs(expected)
	if (expected == nil && actualNil) ||
		(expectedNil && actual == nil) {
		return
	}
	if actualNil && expectedNil {
		if len(expected) != len(actual) {
			t.FailNow()
		}
		return
	}
	require.Equal(t, expected, actual, "StagingTX.Sigs")
}

func isNilSigs(s []wallet.Sig) bool {
	for _, el := range s {
		if el != nil {
			return false
		}
	}
	return true
}

// Init calls Init on the state machine and then checks the persistence.
func (c *Channel) Init(ctx context.Context, t require.TestingT, rng *rand.Rand) {
	initAlloc := *ctest.NewRandomAllocation(rng, ctest.WithNumParts(len(c.accounts)))
	initData := channel.NewMockOp(channel.OpValid)
	err := c.StateMachine.Init(ctx, initAlloc, initData)
	require.NoError(t, err)
	c.AssertPersisted(ctx, t)
}

// EnableInit calls EnableInit on the state machine and then checks the persistence.
func (c *Channel) EnableInit(t require.TestingT) {
	err := c.StateMachine.EnableInit(c.ctx)
	require.NoError(t, err)
	c.AssertPersisted(c.ctx, t)
}

// SignAll signs the current staged state by all parties.
func (c *Channel) SignAll(ctx context.Context, t require.TestingT) {
	// trigger local signing
	c.Sig(ctx) //nolint:errcheck
	c.AssertPersisted(ctx, t)
	// remote signers
	for i := range c.accounts {
		sig, err := channel.Sign(c.accounts[i][channel.TestBackendID], c.StagingState(), channel.TestBackendID)
		require.NoError(t, err)
		c.AddSig(ctx, channel.Index(i), sig) //nolint:errcheck
		c.AssertPersisted(ctx, t)
	}
}

// SetFunded calls SetFunded on the state machine and then checks the persistence.
func (c *Channel) SetFunded(t require.TestingT) {
	require.NoError(t, c.StateMachine.SetFunded(c.ctx))
	c.AssertPersisted(c.ctx, t)
}

// Update calls Update on the state machine and then checks the persistence.
func (c *Channel) Update(t require.TestingT, state *channel.State, idx channel.Index) error {
	err := c.StateMachine.Update(c.ctx, state, idx)
	c.AssertPersisted(c.ctx, t)
	return err
}

// EnableUpdate calls EnableUpdate on the state machine and then checks the persistence.
func (c *Channel) EnableUpdate(t require.TestingT) {
	require.NoError(t, c.StateMachine.EnableUpdate(c.ctx))
	c.AssertPersisted(c.ctx, t)
}

// EnableFinal calls EnableFinal on the state machine and then checks the persistence.
func (c *Channel) EnableFinal(t require.TestingT) {
	require.NoError(t, c.StateMachine.EnableFinal(c.ctx))
	c.AssertPersisted(c.ctx, t)
}

// DiscardUpdate calls DiscardUpdate on the state machine and then checks the persistence.
func (c *Channel) DiscardUpdate(t require.TestingT) {
	require.NoError(t, c.StateMachine.DiscardUpdate(c.ctx))
	c.AssertPersisted(c.ctx, t)
}

// SetRegistering calls SetRegistering on the state machine and then checks the persistence.
func (c *Channel) SetRegistering(t require.TestingT) {
	require.NoError(t, c.StateMachine.SetRegistering(c.ctx))
	c.AssertPersisted(c.ctx, t)
}

// SetRegistered calls SetRegistered on the state machine and then checks the persistence.
func (c *Channel) SetRegistered(t require.TestingT) {
	require.NoError(t, c.StateMachine.SetRegistered(c.ctx))
	c.AssertPersisted(c.ctx, t)
}

// SetWithdrawing calls SetWithdrawing on the state machine and then checks the persistence.
func (c *Channel) SetWithdrawing(t require.TestingT) {
	require.NoError(t, c.StateMachine.SetWithdrawing(c.ctx))
	c.AssertPersisted(c.ctx, t)
}

// SetWithdrawn calls SetWithdrawn on the state machine and then checks the persistence.
func (c *Channel) SetWithdrawn(t require.TestingT) {
	require.NoError(t, c.StateMachine.SetWithdrawn(c.ctx))
	rc, err := c.pr.RestoreChannel(c.ctx, c.ID())
	require.Error(t, err, "restoring of a non-existing channel")
	require.Nil(t, rc, "restoring of a non-existing channel")
}

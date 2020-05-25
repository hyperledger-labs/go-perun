// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package persistence_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/persistence"
	"perun.network/go-perun/channel/persistence/test"
	ctest "perun.network/go-perun/channel/test"
	wtest "perun.network/go-perun/wallet/test"
)

// TestStateMachine tests the StateMachine embedding by advancing the
// StateMachine step by step and asserting that the persisted data matches the
// expected.
//
// TODO: After #316 (custom random gens) this test can be greatly improved.
func TestStateMachine(t *testing.T) {
	require := require.New(t)
	rng := rand.New(rand.NewSource(0x3a57))

	const n = 5                                    // number of participants
	accs, parts := wtest.NewRandomAccounts(rng, n) // local participant idx 0
	params := ctest.NewRandomParams(rng, ctest.WithParts(parts...))
	csm, err := channel.NewStateMachine(accs[0], *params)
	require.NoError(err)

	tpr := test.NewPersistRestorer(t)
	sm := persistence.FromStateMachine(csm, tpr)

	// Newly created channel
	tpr.ChannelCreated(nil, &sm, nil) // nil peers since we only test StateMachine
	tpr.AssertEqual(csm)

	// Init state
	initAlloc := *ctest.NewRandomAllocation(rng, ctest.WithNumParts(n))
	initData := channel.NewMockOp(channel.OpValid)
	err = sm.Init(nil, initAlloc, initData)
	require.NoError(err)
	tpr.AssertEqual(csm)

	signAll := func() {
		_, err := sm.Sig(nil) // trigger local signing
		require.NoError(err)
		tpr.AssertEqual(csm)
		// remote signers
		for i := 1; i < n; i++ {
			sig, err := channel.Sign(accs[i], params, csm.StagingState())
			require.NoError(err)
			sm.AddSig(nil, channel.Index(i), sig)
			tpr.AssertEqual(csm)
		}
	}

	// Sign init state
	signAll()

	// Enable init state
	err = sm.EnableInit(nil)
	require.NoError(err)
	tpr.AssertEqual(csm)

	// Set Funded
	err = sm.SetFunded(nil)
	require.NoError(err)
	tpr.AssertEqual(csm)

	// Update state
	state1 := sm.State().Clone()
	state1.Version++
	// Stage update.
	err = sm.Update(nil, state1, sm.Idx())
	require.NoError(err)
	tpr.AssertEqual(csm)
	// Discard update.
	require.NoError(sm.DiscardUpdate(nil))
	tpr.AssertEqual(csm)
	// Re-stage update.
	err = sm.Update(nil, state1, sm.Idx())
	require.NoError(err)
	tpr.AssertEqual(csm)

	// Sign new state
	signAll()

	// Enable new state
	err = sm.EnableUpdate(nil)
	require.NoError(err)
	tpr.AssertEqual(csm)

	// Final state
	statef := sm.State().Clone()
	statef.Version++
	statef.IsFinal = true
	err = sm.Update(nil, statef, n-1)
	require.NoError(err)
	tpr.AssertEqual(csm)

	// Sign final state
	signAll()

	// Enable final state
	err = sm.EnableFinal(nil)
	require.NoError(err)
	tpr.AssertEqual(csm)

	// Set Registering
	err = sm.SetRegistering(nil)
	require.NoError(err)
	tpr.AssertEqual(csm)

	// Set Registered
	reg := &channel.RegisteredEvent{
		ID:      csm.ID(),
		Version: statef.Version,
		Timeout: new(channel.ElapsedTimeout),
	}
	err = sm.SetRegistered(nil, reg)
	require.NoError(err)
	tpr.AssertEqual(csm)

	// Set Withdrawing
	err = sm.SetWithdrawing(nil)
	require.NoError(err)
	tpr.AssertEqual(csm)

	// Set Withdrawn
	err = sm.SetWithdrawn(nil)
	require.NoError(err)
	tpr.AssertEqual(csm)
}

// Copyright 2020 - See NOTICE file for copyright holders.
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

package persistence_test

import (
	"context"
	"perun.network/go-perun/wallet"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/persistence"
	"perun.network/go-perun/channel/persistence/test"
	ctest "perun.network/go-perun/channel/test"
	wtest "perun.network/go-perun/wallet/test"
	pkgtest "polycry.pt/poly-go/test"
)

const defaultTestTimeout = 10 * time.Second

// TestStateMachine tests the StateMachine embedding by advancing the
// StateMachine step by step and asserting that the persisted data matches the
// expected.
func TestStateMachine(t *testing.T) {
	require := require.New(t)
	rng := pkgtest.Prng(t)

	const n = 5                                       // number of participants
	accs, parts := wtest.NewRandomAccounts(rng, n, 0) // local participant idx 0
	params := ctest.NewRandomParams(rng, ctest.WithParts(parts))
	t.Log("Participants:", parts, "accs:", accs)
	t.Log("Params:", params)
	csm, err := channel.NewStateMachine(accs[0], *params)
	require.NoError(err)

	tpr := test.NewPersistRestorer(t)
	sm := persistence.FromStateMachine(csm, tpr)

	ctx, cancel := context.WithTimeout(context.Background(), defaultTestTimeout)
	defer cancel()
	// Newly created channel
	err = tpr.ChannelCreated(ctx, &sm, nil, nil) // nil peers since we only test StateMachine
	require.NoError(err)
	tpr.AssertEqual(csm)

	// Init state
	initAlloc := *ctest.NewRandomAllocation(rng, ctest.WithNumParts(n))
	initData := channel.NewMockOp(channel.OpValid)
	err = sm.Init(ctx, initAlloc, initData)
	require.NoError(err)
	tpr.AssertEqual(csm)

	signAll := func() {
		_, err := sm.Sig(ctx) // trigger local signing
		require.NoError(err)
		tpr.AssertEqual(csm)
		// remote signers
		for i := 1; i < n; i++ {
			var sig wallet.Sig
			for b, acc := range accs[i] {
				sig, err = channel.Sign(acc, csm.StagingState(), b)
				require.NoError(err)
			}
			err = sm.AddSig(ctx, channel.Index(i), sig)
			require.NoError(err)
			tpr.AssertEqual(csm)
		}
	}

	// Sign init state
	signAll()

	// Enable init state
	err = sm.EnableInit(ctx)
	require.NoError(err)
	tpr.AssertEqual(csm)

	// Set Funded
	err = sm.SetFunded(ctx)
	require.NoError(err)
	tpr.AssertEqual(csm)

	// Update state
	state1 := sm.State().Clone()
	state1.Version++
	// Stage update.
	err = sm.Update(ctx, state1, sm.Idx())
	require.NoError(err)
	tpr.AssertEqual(csm)
	// Discard update.
	require.NoError(sm.DiscardUpdate(ctx))
	tpr.AssertEqual(csm)
	// Re-stage update.
	err = sm.Update(ctx, state1, sm.Idx())
	require.NoError(err)
	tpr.AssertEqual(csm)

	// Sign new state
	signAll()

	// Enable new state
	err = sm.EnableUpdate(ctx)
	require.NoError(err)
	tpr.AssertEqual(csm)

	// Final state
	statef := sm.State().Clone()
	statef.Version++
	statef.IsFinal = true
	err = sm.Update(ctx, statef, n-1)
	require.NoError(err)
	tpr.AssertEqual(csm)

	// Sign final state
	signAll()

	// Enable final state
	err = sm.EnableFinal(ctx)
	require.NoError(err)
	tpr.AssertEqual(csm)

	// Set Registering
	err = sm.SetRegistering(ctx)
	require.NoError(err)
	tpr.AssertEqual(csm)

	// Set Registered
	err = sm.SetRegistered(ctx)
	require.NoError(err)
	tpr.AssertEqual(csm)

	// Set Progressing
	s := ctest.NewRandomState(rng)
	err = sm.SetProgressing(ctx, s)
	require.NoError(err)
	tpr.AssertEqual(csm)

	// Set Progressed
	timeout := ctest.NewRandomTimeout(rng)
	idx := channel.Index(rng.Intn(s.NumParts()))
	e := channel.NewProgressedEvent(s.ID[0], timeout, s, idx)
	err = sm.SetProgressed(ctx, e)
	require.NoError(err)
	tpr.AssertEqual(csm)

	// Set Withdrawing
	err = sm.SetWithdrawing(ctx)
	require.NoError(err)
	tpr.AssertEqual(csm)

	// Set Withdrawn
	err = sm.SetWithdrawn(ctx)
	require.NoError(err)
	tpr.AssertNotExists(csm.ID())
}

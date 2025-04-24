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

package local_test

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	testifyMock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	_ "perun.network/go-perun/backend/sim/channel" // Initialize randomizer.
	_ "perun.network/go-perun/backend/sim/wallet"  // Initialize randomizer.
	"perun.network/go-perun/channel"
	channeltest "perun.network/go-perun/channel/test"
	cltest "perun.network/go-perun/client/test"
	"perun.network/go-perun/watcher"
	"perun.network/go-perun/watcher/internal/mocks"
	"perun.network/go-perun/watcher/local"
	"polycry.pt/poly-go/test"
)

func Test_StartWatching(t *testing.T) {
	rng := test.Prng(t)
	rs := &mocks.RegisterSubscriber{}
	rs.On("Subscribe", testifyMock.Anything, testifyMock.Anything).Return(cltest.NewMockSubscription(context.TODO()), nil)

	t.Run("ledger_channel", func(t *testing.T) {
		// Setup
		w := newWatcher(t, rs)
		params, state := channeltest.NewRandomParamsAndState(rng, channeltest.WithVersion(0))
		signedState := makeSignedStateWDummySigs(params, state)

		// Start Watching for a ledger channel for the first time.
		t.Run("happy", func(t *testing.T) {
			statesPub, eventsSub, err := w.StartWatchingLedgerChannel(context.Background(), signedState)
			require.NoError(t, err)
			require.NotNil(t, statesPub)
			require.NotNil(t, eventsSub)
		})

		// Start watching for the same ledger channel for the second time.
		t.Run("error/start_watching_repeatedly", func(t *testing.T) {
			_, _, err := w.StartWatchingLedgerChannel(context.Background(), signedState)
			require.Error(t, err)
			t.Log(err)
		})
	})

	t.Run("sub_channel", func(t *testing.T) {
		// Setup
		w := newWatcher(t, rs)
		parentParams, parentState := channeltest.NewRandomParamsAndState(rng, channeltest.WithVersion(0))
		parentSignedState := makeSignedStateWDummySigs(parentParams, parentState)
		statesPub, eventsSub, err := w.StartWatchingLedgerChannel(context.Background(), parentSignedState)
		require.NoError(t, err)
		require.NotNil(t, statesPub)
		require.NotNil(t, eventsSub)

		childParams, childState := channeltest.NewRandomParamsAndState(rng, channeltest.WithVersion(0))
		childSignedState := makeSignedStateWDummySigs(childParams, childState)

		// Start watching for a sub-channel for the first time.
		// Parent is not registered with the watcher.
		t.Run("error/unknown_parent", func(t *testing.T) {
			randomID := channeltest.NewRandomChannelID(rng)
			_, _, err = w.StartWatchingSubChannel(context.Background(), randomID, childSignedState)
			require.Error(t, err)
			t.Log(err)
		})

		// Start watching for a sub-channel for the first time.
		// Parent is a ledger channel that is registered with the watcher.
		t.Run("happy", func(t *testing.T) {
			_, _, err = w.StartWatchingSubChannel(context.Background(), parentState.ID, childSignedState)
			require.NoError(t, err)
			require.NotNil(t, statesPub)
			require.NotNil(t, eventsSub)
		})

		// Start watching for a sub-channel for the second time.
		// Parent is a ledger channel that is registered with the watcher.
		t.Run("error/start_watching_repeatedly", func(t *testing.T) {
			_, _, err = w.StartWatchingSubChannel(context.Background(), parentState.ID, childSignedState)
			require.Error(t, err)
			t.Log(err)
		})

		// Start watching for a sub-channel for the second time.
		// Parent is a sub-channel that is registered with the watcher.
		t.Run("error/parent_is_sub_channel", func(t *testing.T) {
			// Setup: Generate the params and the initial state for another sub-channel.
			parent := childState.ID
			childParams2, childState2 := channeltest.NewRandomParamsAndState(rng, channeltest.WithVersion(0))
			childSignedState2 := makeSignedStateWDummySigs(childParams2, childState2)

			_, _, err = w.StartWatchingSubChannel(context.Background(), parent, childSignedState2)
			require.Error(t, err)
			t.Log(err)
		})
	})
}

func Test_Watcher_WithoutSubchannel(t *testing.T) {
	rng := test.Prng(t)
	// Send a registered event on the adjudicator subscription, with the latest state.
	// Watcher should relay the event and not refute.
	t.Run("happy/latest_state_registered", func(t *testing.T) {
		// Setup
		params, txs := randomTxsForSingleCh(rng, 3)
		adjSub := &mocks.AdjudicatorSubscription{}
		trigger := setExpectationNextCall(adjSub, makeRegisteredEvents(txs[2])...)

		rs := &mocks.RegisterSubscriber{}
		setExpectationSubscribeCall(rs, adjSub, nil)
		w := newWatcher(t, rs)

		// Publish both the states to the watcher.
		statesPub, eventsForClient := startWatchingForLedgerChannel(t, w, makeSignedStateWDummySigs(params, txs[0].State))
		require.NoError(t, statesPub.Publish(context.Background(), txs[1]))
		require.NoError(t, statesPub.Publish(context.Background(), txs[2]))

		// Trigger events and assert.
		triggerAdjEventAndExpectNotification(t, trigger, eventsForClient)
		rs.AssertExpectations(t)
	})

	// Send a registered event on the adjudicator subscription,
	// with a state newer than the latest state (published to the watcher),
	// Watch should relay the event and not refute.
	t.Run("happy/newer_than_latest_state_registered", func(t *testing.T) {
		// Setup
		params, txs := randomTxsForSingleCh(rng, 3)
		adjSub := &mocks.AdjudicatorSubscription{}
		trigger := setExpectationNextCall(adjSub, makeRegisteredEvents(txs[2])...)

		rs := &mocks.RegisterSubscriber{}
		setExpectationSubscribeCall(rs, adjSub, nil)
		w := newWatcher(t, rs)

		// Publish only one of the two newly created off-chain states to the watcher.
		statesPub, eventsForClient := startWatchingForLedgerChannel(t, w, makeSignedStateWDummySigs(params, txs[0].State))
		require.NoError(t, statesPub.Publish(context.Background(), txs[1]))

		// Trigger adjudicator events with a state newer than the latest state (published to the watcher).
		triggerAdjEventAndExpectNotification(t, trigger, eventsForClient)
		rs.AssertExpectations(t)
	})

	// First, send a registered event on the adjudicator subscription,
	// with a state older than the latest state (published to the watcher),
	// Watch should relay the event and refute by registering the latest state.
	//
	// Next, send a registered event on adjudicator subscription,
	// with the state that was registered.
	// This time, watcher should relay the event and not dispute.
	t.Run("happy/older_state_registered", func(t *testing.T) {
		// Setup
		params, txs := randomTxsForSingleCh(rng, 3)
		adjSub := &mocks.AdjudicatorSubscription{}
		trigger := setExpectationNextCall(adjSub, makeRegisteredEvents(txs[1], txs[2])...)

		rs := &mocks.RegisterSubscriber{}
		setExpectationSubscribeCall(rs, adjSub, nil)
		setExpectationRegisterCalls(t, rs, &channelTree{txs[2], []channel.Transaction{}})
		w := newWatcher(t, rs)

		// Publish both the states to the watcher.
		statesPub, eventsForClient := startWatchingForLedgerChannel(t, w, makeSignedStateWDummySigs(params, txs[0].State))
		require.NoError(t, statesPub.Publish(context.Background(), txs[1]))
		require.NoError(t, statesPub.Publish(context.Background(), txs[2]))

		// Trigger adjudicator events with an older state and assert if Register was called once.
		triggerAdjEventAndExpectNotification(t, trigger, eventsForClient)
		time.Sleep(50 * time.Millisecond) // Wait for the watcher to refute.
		rs.AssertNumberOfCalls(t, "Register", 1)

		// Trigger adjudicator events with the registered state and assert.
		triggerAdjEventAndExpectNotification(t, trigger, eventsForClient)
		rs.AssertExpectations(t)
	})

	testIfEventsAreRelayed := func(
		t *testing.T,
		eventConstructor func(txs ...channel.Transaction,
		) []channel.AdjudicatorEvent,
	) {
		t.Helper()
		// Setup: Generate the params and off-chain states for a ledger channel.
		params, txs := randomTxsForSingleCh(rng, 2)

		// Setup: Adjudicator event subscription for the ledger.
		adjSub := &mocks.AdjudicatorSubscription{}
		trigger := setExpectationNextCall(adjSub, eventConstructor(txs[1])...)
		rs := &mocks.RegisterSubscriber{}
		setExpectationSubscribeCall(rs, adjSub, nil)

		// Setup: Initialize the watcher and start watching for the ledger.
		w := newWatcher(t, rs)
		_, eventsForClient := startWatchingForLedgerChannel(t, w, makeSignedStateWDummySigs(params, txs[0].State))

		// Trigger the event for ledger channel and assert if they are relayed to the adjudicator subscription
		// (eventsForClient).
		triggerAdjEventAndExpectNotification(t, trigger, eventsForClient)
		rs.AssertExpectations(t)
	}
	// Test if progressed events are relayed to the adjudicator subscription.
	t.Run("happy/progressed_event", func(t *testing.T) {
		testIfEventsAreRelayed(t, makeProgressedEvents)
	})
	// Test if concluded events are relayed to the adjudicator subscription.
	t.Run("happy/concluded_event", func(t *testing.T) {
		testIfEventsAreRelayed(t, makeConcludedEvents)
	})
}

func Test_Watcher_WithSubchannel(t *testing.T) {
	rng := test.Prng(t)
	// For both, the parent and the sub-channel,
	// send a registered event on the adjudicator subscription, with the latest state.
	// Watcher should relay the events and not refute.
	t.Run("happy/latest_state_registered", func(t *testing.T) {
		// Setup
		parentParams, parentTxs := randomTxsForSingleCh(rng, 3)
		childParams, childTxs := randomTxsForSingleCh(rng, 3)
		// Add sub-channel to allocation. This transaction represents funding of the sub-channel.
		parentTxs[2].Allocation.Locked = []channel.SubAlloc{{ID: childTxs[0].ID}}

		adjSubParent := &mocks.AdjudicatorSubscription{}
		triggerParent := setExpectationNextCall(adjSubParent, makeRegisteredEvents(parentTxs[2])...)
		adjSubChild := &mocks.AdjudicatorSubscription{}
		triggerChild := setExpectationNextCall(adjSubChild, makeRegisteredEvents(childTxs[2])...)
		rs := &mocks.RegisterSubscriber{}
		setExpectationSubscribeCall(rs, adjSubParent, nil)
		setExpectationSubscribeCall(rs, adjSubChild, nil)

		w := newWatcher(t, rs)

		// Parent: Publish both the states to the watcher.
		parentSignedState := makeSignedStateWDummySigs(parentParams, parentTxs[0].State)
		statesPubParent, eventsForClientParent := startWatchingForLedgerChannel(t, w, parentSignedState)
		require.NoError(t, statesPubParent.Publish(context.Background(), parentTxs[1]))
		require.NoError(t, statesPubParent.Publish(context.Background(), parentTxs[2]))

		// Child: Publish both the states to the watcher.
		childSignedState := makeSignedStateWDummySigs(childParams, childTxs[0].State)
		statesPubChild, eventsForClientChild := startWatchingForSubChannel(t, w, childSignedState, parentTxs[0].State.ID)
		require.NoError(t, statesPubChild.Publish(context.Background(), childTxs[1]))
		require.NoError(t, statesPubChild.Publish(context.Background(), childTxs[2]))

		// Parent and child: Trigger adjudicator events with the latest states and assert.
		triggerAdjEventAndExpectNotification(t, triggerParent, eventsForClientParent)
		triggerAdjEventAndExpectNotification(t, triggerChild, eventsForClientChild)
		rs.AssertExpectations(t)
	})

	// For both, the parent and the sub-channel,
	// send a registered event on the adjudicator subscription,
	// with a state newer than the latest state (published to the watcher),
	// Watch should relay the event and not refute.
	t.Run("happy/newer_than_latest_state_registered", func(t *testing.T) {
		// Setup
		parentParams, parentTxs := randomTxsForSingleCh(rng, 3)
		childParams, childTxs := randomTxsForSingleCh(rng, 3)
		// Add sub-channel to allocation. This transaction represents funding of the sub-channel.
		parentTxs[2].Allocation.Locked = []channel.SubAlloc{{ID: childTxs[0].ID}}

		adjSubParent := &mocks.AdjudicatorSubscription{}
		triggerParent := setExpectationNextCall(adjSubParent, makeRegisteredEvents(parentTxs[2])...)
		adjSubChild := &mocks.AdjudicatorSubscription{}
		triggerChild := setExpectationNextCall(adjSubChild, makeRegisteredEvents(childTxs[2])...)

		rs := &mocks.RegisterSubscriber{}
		setExpectationSubscribeCall(rs, adjSubParent, nil)
		setExpectationSubscribeCall(rs, adjSubChild, nil)

		w := newWatcher(t, rs)

		// Parent: Publish only one of the two newly created off-chain states to the watcher.
		parentSignedState := makeSignedStateWDummySigs(parentParams, parentTxs[0].State)
		statesPubParent, eventsForClientParent := startWatchingForLedgerChannel(t, w, parentSignedState)
		require.NoError(t, statesPubParent.Publish(context.Background(), parentTxs[1]))

		// Child: Publish only one of the two newly created off-chain states to the watcher.
		childSignedState := makeSignedStateWDummySigs(childParams, childTxs[0].State)
		statesPubChild, eventsForClientChild := startWatchingForSubChannel(t, w, childSignedState, parentTxs[0].State.ID)
		require.NoError(t, statesPubChild.Publish(context.Background(), childTxs[1]))

		// Parent, Child: Trigger adjudicator events with a state newer than
		// the latest state (published to the watcher).
		triggerAdjEventAndExpectNotification(t, triggerParent, eventsForClientParent)
		triggerAdjEventAndExpectNotification(t, triggerChild, eventsForClientChild)
		rs.AssertExpectations(t)
	})

	// For both, the parent and the sub-channel,
	//
	// First, send a registered event on the adjudicator subscription,
	// with a state older than the latest state (published to the watcher),
	// Watch should relay the event and refute by registering the latest state.
	//
	// Next, send a registered event on adjudicator subscription,
	// with the state that was registered.
	// This time, watcher should relay the event and not dispute.
	t.Run("happy/older_state_registered", func(t *testing.T) {
		// Setup
		parentParams, parentTxs := randomTxsForSingleCh(rng, 3)
		childParams, childTxs := randomTxsForSingleCh(rng, 3)
		// Add sub-channel to allocation. This transaction represents funding of the sub-channel.
		parentTxs[2].Allocation.Locked = []channel.SubAlloc{{ID: childTxs[0].ID}}

		adjSubParent := &mocks.AdjudicatorSubscription{}
		triggerParent := setExpectationNextCall(adjSubParent, makeRegisteredEvents(parentTxs[1], parentTxs[2])...)
		adjSubChild := &mocks.AdjudicatorSubscription{}
		triggerChild := setExpectationNextCall(adjSubChild, makeRegisteredEvents(childTxs[1], childTxs[2])...)

		rs := &mocks.RegisterSubscriber{}
		setExpectationSubscribeCall(rs, adjSubParent, nil)
		setExpectationSubscribeCall(rs, adjSubChild, nil)
		setExpectationRegisterCalls(t, rs, &channelTree{parentTxs[2], []channel.Transaction{childTxs[2]}})

		w := newWatcher(t, rs)

		// Parent: Publish both the states to the watcher.
		parentSignedState := makeSignedStateWDummySigs(parentParams, parentTxs[0].State)
		parentStatesPub, eventsForClientParent := startWatchingForLedgerChannel(t, w, parentSignedState)
		require.NoError(t, parentStatesPub.Publish(context.Background(), parentTxs[1]))
		require.NoError(t, parentStatesPub.Publish(context.Background(), parentTxs[2]))

		// Child: Publish both the states to the watcher.
		childSignedState := makeSignedStateWDummySigs(childParams, childTxs[0].State)
		childStatesPub, eventsForClientChild := startWatchingForSubChannel(t, w, childSignedState, parentTxs[0].State.ID)
		require.NoError(t, childStatesPub.Publish(context.Background(), childTxs[1]))
		require.NoError(t, childStatesPub.Publish(context.Background(), childTxs[2]))

		// Parent and Child: Trigger adjudicator events with an older state and assert if Register was called once.
		triggerAdjEventAndExpectNotification(t, triggerParent, eventsForClientParent)
		triggerAdjEventAndExpectNotification(t, triggerChild, eventsForClientChild)
		time.Sleep(50 * time.Millisecond) // Wait for the watcher to refute.
		rs.AssertNumberOfCalls(t, "Register", 1)

		// Parent and Child: Trigger adjudicator events with the registered state and assert.
		triggerAdjEventAndExpectNotification(t, triggerParent, eventsForClientParent)
		triggerAdjEventAndExpectNotification(t, triggerChild, eventsForClientChild)
		rs.AssertExpectations(t)
	})

	// First, for both, the parent and the sub-channel,
	// Send a registered event on the adjudicator subscription,
	// with a state older than the latest state (published to the watcher),
	// Watch should relay the event and refute by registering the latest state.
	//
	// Next, for the sub-channel, publish another off-chain state to the watcher.
	//
	// Next, for both, the parent and the sub-channel,
	// Send a registered event on adjudicator subscription,
	// with the state that was registered.
	// This time, because the registered state is older than the latest state for the sub-channel,
	// Watch should relay the event and refute by registering the latest state.
	//
	// Next, for the sub-channel, send a registered event on the adjudicator
	// subscription. This time, the watcher should relay the event and not
	// dispute.
	t.Run("happy/older_state_registered_then_newer_state_received", func(t *testing.T) {
		// Setup
		parentParams, parentTxs := randomTxsForSingleCh(rng, 3)
		childParams, childTxs := randomTxsForSingleCh(rng, 4)
		// Add sub-channel to allocation. This transaction represents funding of the sub-channel.
		parentTxs[2].Allocation.Locked = []channel.SubAlloc{{ID: childTxs[0].ID}}

		adjSubParent := &mocks.AdjudicatorSubscription{}
		triggerParent := setExpectationNextCall(adjSubParent, makeRegisteredEvents(parentTxs[1], parentTxs[2])...)
		adjSubChild := &mocks.AdjudicatorSubscription{}
		triggerChild := setExpectationNextCall(adjSubChild, makeRegisteredEvents(childTxs[1], childTxs[2], childTxs[3])...)

		rs := &mocks.RegisterSubscriber{}
		setExpectationSubscribeCall(rs, adjSubParent, nil)
		setExpectationSubscribeCall(rs, adjSubChild, nil)
		setExpectationRegisterCalls(t, rs,
			&channelTree{parentTxs[2], []channel.Transaction{childTxs[2]}},
			&channelTree{parentTxs[2], []channel.Transaction{childTxs[3]}})

		w := newWatcher(t, rs)

		// Parent: Publish both the states to the watcher.
		parentSignedState := makeSignedStateWDummySigs(parentParams, parentTxs[0].State)
		parentStatesPub, eventsForClientParent := startWatchingForLedgerChannel(t, w, parentSignedState)
		require.NoError(t, parentStatesPub.Publish(context.Background(), parentTxs[1]))
		require.NoError(t, parentStatesPub.Publish(context.Background(), parentTxs[2]))

		// Child: Publish both the states to the watcher.
		childSignedState := makeSignedStateWDummySigs(childParams, childTxs[0].State)
		childStatesPub, eventsForClientChild := startWatchingForSubChannel(t, w, childSignedState, parentTxs[0].State.ID)
		require.NoError(t, childStatesPub.Publish(context.Background(), childTxs[1]))
		require.NoError(t, childStatesPub.Publish(context.Background(), childTxs[2]))

		// Parent and Child: Trigger adjudicator events with an older state and assert if Register was called once.
		triggerAdjEventAndExpectNotification(t, triggerParent, eventsForClientParent)
		triggerAdjEventAndExpectNotification(t, triggerChild, eventsForClientChild)
		time.Sleep(50 * time.Millisecond) // Wait for the watcher to refute.
		rs.AssertNumberOfCalls(t, "Register", 1)

		// Child: After register was called, publish a new state to the watcher.
		require.NoError(t, childStatesPub.Publish(context.Background(), childTxs[3]))

		// Parent and Child: Trigger adjudicator events with the registered state and assert if Register was called once.
		triggerAdjEventAndExpectNotification(t, triggerParent, eventsForClientParent)
		triggerAdjEventAndExpectNotification(t, triggerChild, eventsForClientChild)
		time.Sleep(50 * time.Millisecond) // Wait for the watcher to refute.
		rs.AssertNumberOfCalls(t, "Register", 2)

		// Child: Trigger adjudicator events with the new state and assert.
		triggerAdjEventAndExpectNotification(t, triggerChild, eventsForClientChild)
		rs.AssertExpectations(t)
	})

	testIfEventsAreRelayed := func(
		t *testing.T,
		eventConstructor func(txs ...channel.Transaction,
		) []channel.AdjudicatorEvent,
	) {
		t.Helper()
		// Setup: Generate the params and off-chain states for a ledger channel and a sub-channel.
		parentParams, parentTxs := randomTxsForSingleCh(rng, 2)
		childParams, childTxs := randomTxsForSingleCh(rng, 2)
		parentTxs[1].Allocation.Locked = []channel.SubAlloc{{ID: childTxs[0].ID}} // Add sub-channel to allocation.

		// Setup: Adjudicator event subscription for the ledger and sub-channel.
		adjSubParent := &mocks.AdjudicatorSubscription{}
		triggerParent := setExpectationNextCall(adjSubParent, eventConstructor(parentTxs[1])...)
		adjSubChild := &mocks.AdjudicatorSubscription{}
		triggerChild := setExpectationNextCall(adjSubChild, eventConstructor(childTxs[1])...)
		rs := &mocks.RegisterSubscriber{}
		setExpectationSubscribeCall(rs, adjSubParent, nil)
		setExpectationSubscribeCall(rs, adjSubChild, nil)

		// Setup: Initialize the watcher and start watching for the ledger and sub-channel.
		w := newWatcher(t, rs)
		parentSignedState := makeSignedStateWDummySigs(parentParams, parentTxs[0].State)
		_, eventsForClientParent := startWatchingForLedgerChannel(t, w, parentSignedState)
		childSignedState := makeSignedStateWDummySigs(childParams, childTxs[0].State)
		_, eventsForClientChild := startWatchingForSubChannel(t, w, childSignedState, parentTxs[0].State.ID)

		// Trigger the events for both (the ledger channel and the sub-channel) and, assert if they are relayed to
		// the adjudicator subscription (eventsForClient).
		triggerAdjEventAndExpectNotification(t, triggerParent, eventsForClientParent)
		triggerAdjEventAndExpectNotification(t, triggerChild, eventsForClientChild)

		rs.AssertExpectations(t)
	}
	// Test if progressed events are relayed to the adjudicator subscription.
	t.Run("happy/progressed_event", func(t *testing.T) {
		testIfEventsAreRelayed(t, makeProgressedEvents)
	})
	// Test if concluded events are relayed to the adjudicator subscription.
	t.Run("happy/concluded_event", func(t *testing.T) {
		testIfEventsAreRelayed(t, makeConcludedEvents)
	})
}

func Test_Watcher_StopWatching(t *testing.T) {
	t.Helper()
	rng := test.Prng(t)

	t.Run("ledger_channel_without_sub_channel", func(t *testing.T) {
		f := func(t *testing.T, errOnClose error) {
			t.Helper()
			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

			params, txs := randomTxsForSingleCh(rng, 1)
			adjSub := &mocks.AdjudicatorSubscription{}
			trigger := setExpectationNextCall(adjSub)
			setExpectationCloseCallErrCall(adjSub, trigger, errOnClose)

			rs := &mocks.RegisterSubscriber{}
			setExpectationSubscribeCall(rs, adjSub, nil)
			w := newWatcher(t, rs)
			startWatchingForLedgerChannel(t, w, makeSignedStateWDummySigs(params, txs[0].State))

			require.NoError(t, w.StopWatching(context.Background(), txs[0].State.ID))
			rs.AssertExpectations(t)
		}

		// Errors from adjudicator subscription (events from chain) does not affect stop watching.
		// They are only logged by the watcher and not returned to the user.
		t.Run("happy/adjSub_noError", func(t *testing.T) { f(t, nil) })
		t.Run("happy/adjSub_error", func(t *testing.T) { f(t, assert.AnError) })
		t.Run("happy/concurrency", func(t *testing.T) {
			// defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

			params, txs := randomTxsForSingleCh(rng, 1)
			adjSub := &mocks.AdjudicatorSubscription{}
			trigger := setExpectationNextCall(adjSub)
			setExpectationCloseCallErrCall(adjSub, trigger, nil)

			rs := &mocks.RegisterSubscriber{}
			setExpectationSubscribeCall(rs, adjSub, nil)
			w := newWatcher(t, rs)
			startWatchingForLedgerChannel(t, w, makeSignedStateWDummySigs(params, txs[0].State))

			wg := sync.WaitGroup{}
			for i := 0; i < 2; i++ {
				wg.Add(1)
				go func() {
					w.StopWatching(context.Background(), txs[0].State.ID) //nolint:errcheck
					wg.Done()
				}()
			}
			wg.Wait()
			rs.AssertExpectations(t)
		})
	})

	t.Run("ledger_channel_with_sub_channel", func(t *testing.T) {
		t.Run("happy/settleOnLedger_stopChild_stopParent", func(t *testing.T) {
			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
			parentParams, parentTxs := randomTxsForSingleCh(rng, 3)
			childParams, childTxs := randomTxsForSingleCh(rng, 1)
			parentTxs[1].Allocation.Locked = []channel.SubAlloc{{ID: childTxs[0].ID}} // sub-channel funding.
			parentTxs[2].Allocation.Locked = []channel.SubAlloc{}                     // sub-channel withdrawal.

			adjSubParent := &mocks.AdjudicatorSubscription{}
			triggerParent := setExpectationNextCall(adjSubParent)
			setExpectationCloseCallErrCall(adjSubParent, triggerParent, nil)
			adjSubChild := &mocks.AdjudicatorSubscription{}
			triggerChild := setExpectationNextCall(adjSubChild)
			setExpectationCloseCallErrCall(adjSubChild, triggerChild, nil)

			rs := &mocks.RegisterSubscriber{}
			setExpectationSubscribeCall(rs, adjSubParent, nil)
			setExpectationSubscribeCall(rs, adjSubChild, nil)

			w := newWatcher(t, rs)
			// Parent: Start watching and publish sub-channel funding transaction.
			parentSignedState := makeSignedStateWDummySigs(parentParams, parentTxs[0].State)
			statesPub, _ := startWatchingForLedgerChannel(t, w, parentSignedState)
			require.NoError(t, statesPub.Publish(context.Background(), parentTxs[1]))

			// Child: Start watching. Parent: Publish sub-channel withdrawal transaction
			childSignedState := makeSignedStateWDummySigs(childParams, childTxs[0].State)
			startWatchingForSubChannel(t, w, childSignedState, parentTxs[0].State.ID)
			require.NoError(t, statesPub.Publish(context.Background(), parentTxs[2]))

			// Child, then Parent: Stop Watching.
			require.NoError(t, w.StopWatching(context.Background(), childTxs[0].State.ID))
			require.NoError(t, w.StopWatching(context.Background(), parentTxs[0].State.ID))

			rs.AssertExpectations(t)
		})

		t.Run("happy/stopChild_registeredEventNoDispute_stopParent", func(t *testing.T) {
			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
			parentParams, parentTxs := randomTxsForSingleCh(rng, 3)
			childParams, childTxs := randomTxsForSingleCh(rng, 3)
			parentTxs[2].Allocation.Locked = []channel.SubAlloc{{ID: childTxs[0].ID}} // sub-channel funding.

			adjSubParent := &mocks.AdjudicatorSubscription{}
			triggerParent := setExpectationNextCall(adjSubParent, makeRegisteredEvents(parentTxs[2])[0])
			setExpectationCloseCallErrCall(adjSubParent, triggerParent, nil)
			adjSubChild := &mocks.AdjudicatorSubscription{}
			triggerChild := setExpectationNextCall(adjSubChild)
			setExpectationCloseCallErrCall(adjSubChild, triggerChild, nil)

			rs := &mocks.RegisterSubscriber{}
			setExpectationSubscribeCall(rs, adjSubParent, nil)
			setExpectationSubscribeCall(rs, adjSubChild, nil)

			w := newWatcher(t, rs)
			// Parent: Start watching and publish sub-channel funding transaction.
			parentSignedState := makeSignedStateWDummySigs(parentParams, parentTxs[0].State)
			parentStatesPub, eventsForClientParent := startWatchingForLedgerChannel(t, w, parentSignedState)
			require.NoError(t, parentStatesPub.Publish(context.Background(), parentTxs[1]))
			require.NoError(t, parentStatesPub.Publish(context.Background(), parentTxs[2]))

			// Child: Start watching.
			childSignedState := makeSignedStateWDummySigs(childParams, childTxs[0].State)
			startWatchingForSubChannel(t, w, childSignedState, parentTxs[0].State.ID)

			// Child: Stop watching.
			require.NoError(t, w.StopWatching(context.Background(), childTxs[0].State.ID))
			// Parent: Trigger event with latest state and expect refutation with parent, child transactions.
			triggerAdjEventAndExpectNotification(t, triggerParent, eventsForClientParent)
			// Parent: Stop watching.
			require.NoError(t, w.StopWatching(context.Background(), parentTxs[0].State.ID))

			rs.AssertExpectations(t)
		})

		t.Run("happy/stopChild_registeredEventDispute_stopParent", func(t *testing.T) {
			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
			parentParams, parentTxs := randomTxsForSingleCh(rng, 3)
			childParams, childTxs := randomTxsForSingleCh(rng, 3)
			parentTxs[2].Allocation.Locked = []channel.SubAlloc{{ID: childTxs[0].ID}} // sub-channel funding.

			adjSubParent := &mocks.AdjudicatorSubscription{}
			triggerParent := setExpectationNextCall(adjSubParent, makeRegisteredEvents(parentTxs[1])[0])
			setExpectationCloseCallErrCall(adjSubParent, triggerParent, nil)
			adjSubChild := &mocks.AdjudicatorSubscription{}
			triggerChild := setExpectationNextCall(adjSubChild)
			setExpectationCloseCallErrCall(adjSubChild, triggerChild, nil)

			rs := &mocks.RegisterSubscriber{}
			setExpectationSubscribeCall(rs, adjSubParent, nil)
			setExpectationSubscribeCall(rs, adjSubChild, nil)
			setExpectationRegisterCalls(t, rs, &channelTree{parentTxs[2], []channel.Transaction{childTxs[0]}})

			w := newWatcher(t, rs)
			// Parent: Start watching and publish sub-channel funding transaction.
			parentSignedState := makeSignedStateWDummySigs(parentParams, parentTxs[0].State)
			parentStatesPub, eventsForClientParent := startWatchingForLedgerChannel(t, w, parentSignedState)
			require.NoError(t, parentStatesPub.Publish(context.Background(), parentTxs[1]))
			require.NoError(t, parentStatesPub.Publish(context.Background(), parentTxs[2]))

			// Child: Start watching.
			childSignedState := makeSignedStateWDummySigs(childParams, childTxs[0].State)
			startWatchingForSubChannel(t, w, childSignedState, parentTxs[0].State.ID)

			// Child: Stop watching.
			require.NoError(t, w.StopWatching(context.Background(), childTxs[0].State.ID))
			// Parent: Trigger event with latest state and expect refutation with parent, child transactions.
			triggerAdjEventAndExpectNotification(t, triggerParent, eventsForClientParent)
			// Parent: Stop watching.
			require.NoError(t, w.StopWatching(context.Background(), parentTxs[0].State.ID))

			rs.AssertExpectations(t)
		})

		t.Run("error/stopParent_woSettleOnParent_woStopChild", func(t *testing.T) {
			parentParams, parentTxs := randomTxsForSingleCh(rng, 3)
			childParams, childTxs := randomTxsForSingleCh(rng, 3)
			parentTxs[2].Allocation.Locked = []channel.SubAlloc{{ID: childTxs[0].ID}} // sub-channel funding.

			adjSubParent := &mocks.AdjudicatorSubscription{}
			triggerParent := setExpectationNextCall(adjSubParent)
			setExpectationCloseCallErrCall(adjSubParent, triggerParent, nil)
			adjSubChild := &mocks.AdjudicatorSubscription{}
			triggerChild := setExpectationNextCall(adjSubChild)
			setExpectationCloseCallErrCall(adjSubChild, triggerChild, nil)

			rs := &mocks.RegisterSubscriber{}
			setExpectationSubscribeCall(rs, adjSubParent, nil)
			setExpectationSubscribeCall(rs, adjSubChild, nil)

			w := newWatcher(t, rs)
			// Parent: Start watching and publish sub-channel funding transaction.
			parentSignedState := makeSignedStateWDummySigs(parentParams, parentTxs[0].State)
			statesSub, _ := startWatchingForLedgerChannel(t, w, parentSignedState)
			require.NoError(t, statesSub.Publish(context.Background(), parentTxs[1]))
			require.NoError(t, statesSub.Publish(context.Background(), parentTxs[2]))

			// Child: Start watching.
			childSignedState := makeSignedStateWDummySigs(childParams, childTxs[0].State)
			startWatchingForSubChannel(t, w, childSignedState, parentTxs[0].State.ID)

			// Parent: Stop watching (error).
			require.Error(t, w.StopWatching(context.Background(), parentTxs[0].State.ID))

			rs.AssertExpectations(t)
		})

		t.Run("error/stopParent_SettleOnParent_woStopChild", func(t *testing.T) {
			parentParams, parentTxs := randomTxsForSingleCh(rng, 3)
			childParams, childTxs := randomTxsForSingleCh(rng, 3)
			parentTxs[1].Allocation.Locked = []channel.SubAlloc{{ID: childTxs[0].ID}} // sub-channel funding.
			parentTxs[2].Allocation.Locked = []channel.SubAlloc{}                     // sub-channel withdrawal.

			adjSubParent := &mocks.AdjudicatorSubscription{}
			triggerParent := setExpectationNextCall(adjSubParent)
			setExpectationCloseCallErrCall(adjSubParent, triggerParent, nil)
			adjSubChild := &mocks.AdjudicatorSubscription{}
			triggerChild := setExpectationNextCall(adjSubChild)
			setExpectationCloseCallErrCall(adjSubChild, triggerChild, nil)

			rs := &mocks.RegisterSubscriber{}
			setExpectationSubscribeCall(rs, adjSubParent, nil)
			setExpectationSubscribeCall(rs, adjSubChild, nil)

			w := newWatcher(t, rs)
			// Parent: Start watching and publish sub-channel funding transaction.
			parentSignedState := makeSignedStateWDummySigs(parentParams, parentTxs[0].State)
			statesSub, _ := startWatchingForLedgerChannel(t, w, parentSignedState)
			require.NoError(t, statesSub.Publish(context.Background(), parentTxs[1]))
			require.NoError(t, statesSub.Publish(context.Background(), parentTxs[2]))

			// Child: Start watching.
			childSignedState := makeSignedStateWDummySigs(childParams, childTxs[0].State)
			startWatchingForSubChannel(t, w, childSignedState, parentTxs[0].State.ID)

			// Parent: Stop watching (error).
			require.Error(t, w.StopWatching(context.Background(), parentTxs[0].State.ID))

			rs.AssertExpectations(t)
		})
	})
}

func newWatcher(t *testing.T, rs channel.RegisterSubscriber) *local.Watcher {
	t.Helper()

	w, err := local.NewWatcher(rs)
	require.NoError(t, err)
	require.NotNil(t, w)
	return w
}

func makeSignedStateWDummySigs(params *channel.Params, state *channel.State) channel.SignedState {
	return channel.SignedState{Params: params, State: state}
}

// randomTxsForSingleCh returns "n" transactions for a random channel.
func randomTxsForSingleCh(rng *rand.Rand, n int) (*channel.Params, []channel.Transaction) {
	params, initialState := channeltest.NewRandomParamsAndState(rng, channeltest.WithVersion(0),
		channeltest.WithNumParts(2), channeltest.WithNumAssets(1), channeltest.WithIsFinal(false), channeltest.WithNumLocked(0))

	txs := make([]channel.Transaction, n)
	for i := range txs {
		txs[i] = channel.Transaction{State: initialState.Clone()}
		txs[i].State.Version = uint64(i)
	}
	return params, txs
}

// adjEventSource is used for triggering events on the mock adjudicator subscription.
//
// adjEvents contains the adjudicator events that can be triggered from this source.
// handles contains the handles (signalling channels of type (chan time.Time)) for triggering these events.
type adjEventSource struct {
	adjEvents chan channel.AdjudicatorEvent

	// chan time.Time is the signalling channel type required by testify/mock.WaitUntil.
	handles     chan chan time.Time
	closeHandle chan chan time.Time
}

// trigger closes a signalling channel and returns the corresponding adjudicator event.
func (t *adjEventSource) trigger() channel.AdjudicatorEvent {
	select {
	case handle := <-t.handles:
		close(handle)
		return <-t.adjEvents
	default:
		panic(fmt.Sprintf("Number of triggers exceeded maximum limit of %d", cap(t.handles)))
	}
}

func (t *adjEventSource) close() {
	select {
	case closeHandle := <-t.closeHandle:
		close(closeHandle)
	default:
		panic("Close called more than once")
	}
}

// setExpectationSubscribeCall configures the mock RegisterSubscriber to expect
// the "Subscribe" method to be called once. The adjSub and the err are set as
// the return values for the call and will be returned when the method is
// called.
func setExpectationSubscribeCall(rs *mocks.RegisterSubscriber, adjSub channel.AdjudicatorSubscription, err error) {
	rs.On("Subscribe", testifyMock.Anything, testifyMock.Anything).Return(adjSub, err).Once()
}

// setExpectationNextCall initializes and returns a mock adjudicator subscription
// and an adjEventSource.
//
// Calling "trigger" on the adjEventSource, will send the events in adjEvents
// (in the sequential order) on the mock adjudicator subscription. If number of
// "trigger" calls exceeds the number of transactions, it will result in panic.
//
// After all triggers are used, the subscription blocks.
func setExpectationNextCall(
	adjSub *mocks.AdjudicatorSubscription,
	adjEvents ...channel.AdjudicatorEvent,
) adjEventSource {
	triggers := adjEventSource{
		handles:     make(chan chan time.Time, len(adjEvents)),
		adjEvents:   make(chan channel.AdjudicatorEvent, len(adjEvents)),
		closeHandle: make(chan chan time.Time, 1),
	}

	for i := range adjEvents {
		handle := make(chan time.Time)
		triggers.handles <- handle
		triggers.adjEvents <- adjEvents[i]

		adjSub.On("Next").Return(adjEvents[i]).WaitUntil(handle).Once()
	}
	// Set up a transaction, that cannot be triggered.
	closeHandle := make(chan time.Time)
	triggers.closeHandle <- closeHandle
	adjSub.On("Next").Return(nil).WaitUntil(closeHandle).Once()

	return triggers
}

func setExpectationCloseCallErrCall(adjSub *mocks.AdjudicatorSubscription, trigger adjEventSource, errOnClose error) {
	adjSub.On("Close").Run(func(_ testifyMock.Arguments) {
		time.Sleep(1 * time.Millisecond) // Simulate closure of actual subscription.
		trigger.close()
	}).Return(errOnClose)
	adjSub.On("Err").Return(nil)
}

type channelTree struct {
	rootTx channel.Transaction
	subTxs []channel.Transaction
}

// setExpectationRegisterCalls configures the mock RegisterSubscriber to expect
// the "Register" method to be called, once for each channelTree.
//
// On each call, the mock will check if the parameters of the call are correct
// for the given channel tree and fails the test if incorrect. The return value
// of "Register" calls is always nil.
func setExpectationRegisterCalls(t *testing.T, rs *mocks.RegisterSubscriber, channelTrees ...*channelTree) {
	t.Helper()
	limit := len(channelTrees)
	mtx := sync.Mutex{}
	iChannelTree := 0

	// The functions passed to the "MatchedBy" method check for the correctness
	// of the call parameters. These functions are closures over the variables:
	// mtx, iChannelTree and channelTrees. iChannelTree is incremented on each
	// call.
	//
	// This allows to validate the parameters for different channel trees in
	// each call to the Register method.
	rs.On("Register", testifyMock.Anything,
		testifyMock.MatchedBy(func(req channel.AdjudicatorReq) bool {
			mtx.Lock()
			if iChannelTree >= limit {
				return false
			}
			return assertEqualAdjudicatorReq(t, req, channelTrees[iChannelTree].rootTx.State)
		}),
		testifyMock.MatchedBy(func(subStates []channel.SignedState) bool {
			defer func() {
				iChannelTree++
				mtx.Unlock()
			}()
			if iChannelTree >= limit {
				return false
			}
			return assertEqualSignedStates(t, subStates, channelTrees[iChannelTree].subTxs)
		})).Return(nil).Times(limit)
}

func assertEqualAdjudicatorReq(t *testing.T, got channel.AdjudicatorReq, want *channel.State) bool {
	t.Helper()
	if nil != got.Tx.State.Equal(want) {
		t.Logf("Got %+v, expected %+v", got.Tx.State, want)
		return false
	}
	return true
}

func assertEqualSignedStates(t *testing.T, got []channel.SignedState, want []channel.Transaction) bool {
	t.Helper()
	if len(got) != len(want) {
		t.Logf("Got %d sub states, expected %d sub states", len(got), len(want))
		return false
	}
	for iSubState := range got {
		t.Logf("Got %+v, expected %+v", got[iSubState].State, want[iSubState].State)
		if nil != got[iSubState].State.Equal(want[iSubState].State) {
			return false
		}
	}
	return true
}

func makeRegisteredEvents(txs ...channel.Transaction) []channel.AdjudicatorEvent {
	events := make([]channel.AdjudicatorEvent, len(txs))
	for i, tx := range txs {
		events[i] = &channel.RegisteredEvent{
			State: tx.State,
			Sigs:  tx.Sigs,
			AdjudicatorEventBase: channel.AdjudicatorEventBase{
				IDV:      tx.State.ID,
				TimeoutV: &channel.ElapsedTimeout{},
				VersionV: tx.State.Version,
			},
		}
	}
	return events
}

func makeProgressedEvents(txs ...channel.Transaction) []channel.AdjudicatorEvent {
	events := make([]channel.AdjudicatorEvent, len(txs))
	for i, tx := range txs {
		events[i] = &channel.ProgressedEvent{
			State: tx.State,
			Idx:   channel.Index(0),
			AdjudicatorEventBase: channel.AdjudicatorEventBase{
				IDV:      tx.State.ID,
				TimeoutV: &channel.ElapsedTimeout{},
				VersionV: tx.State.Version,
			},
		}
	}
	return events
}

func makeConcludedEvents(txs ...channel.Transaction) []channel.AdjudicatorEvent {
	events := make([]channel.AdjudicatorEvent, len(txs))
	for i, tx := range txs {
		events[i] = &channel.ConcludedEvent{
			AdjudicatorEventBase: channel.AdjudicatorEventBase{
				IDV:      tx.State.ID,
				TimeoutV: &channel.ElapsedTimeout{},
				VersionV: tx.State.Version,
			},
		}
	}
	return events
}

func startWatchingForLedgerChannel(
	t *testing.T,
	w *local.Watcher,
	signedState channel.SignedState,
) (watcher.StatesPub, watcher.AdjudicatorSub) {
	t.Helper()
	statesPub, eventsSub, err := w.StartWatchingLedgerChannel(context.TODO(), signedState)

	require.NoError(t, err)
	require.NotNil(t, statesPub)
	require.NotNil(t, eventsSub)

	return statesPub, eventsSub
}

func startWatchingForSubChannel(
	t *testing.T,
	w *local.Watcher,
	signedState channel.SignedState,
	parentID channel.ID,
) (watcher.StatesPub, watcher.AdjudicatorSub) {
	t.Helper()
	statesPub, eventsSub, err := w.StartWatchingSubChannel(context.TODO(), parentID, signedState)

	require.NoError(t, err)
	require.NotNil(t, eventsSub)

	return statesPub, eventsSub
}

func triggerAdjEventAndExpectNotification(
	t *testing.T,
	trigger adjEventSource,
	eventsForClient watcher.AdjudicatorSub,
) {
	t.Helper()
	wantEvent := trigger.trigger()
	t.Logf("waiting for adjudicator event for ch %x, version: %v", wantEvent.ID(), wantEvent.Version())
	gotEvent := <-eventsForClient.EventStream()
	require.EqualValues(t, gotEvent, wantEvent)
	t.Logf("received adjudicator event for ch %x, version: %v", wantEvent.ID(), wantEvent.Version())
}

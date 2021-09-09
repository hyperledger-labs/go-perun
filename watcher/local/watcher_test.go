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

package local_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	ethChannel "perun.network/go-perun/backend/ethereum/channel"
	_ "perun.network/go-perun/backend/ethereum/channel/test" // For initilizing channeltest
	"perun.network/go-perun/channel"
	channeltest "perun.network/go-perun/channel/test"
	"perun.network/go-perun/pkg/test"
	"perun.network/go-perun/watcher"
	"perun.network/go-perun/watcher/internal/mocks"
	"perun.network/go-perun/watcher/local"
)

// func init() {
// 	l := logrus.New()
// 	l.SetLevel(logrus.DebugLevel)
// 	log.Set(plogrus.FromLogrus(l))
// }

func Test_StartWatching(t *testing.T) {
	rng := test.Prng(t)
	rs := &mocks.RegisterSubscriber{}
	rs.On("Subscribe", mock.Anything, mock.Anything).Return(&ethChannel.RegisteredSub{}, nil)

	t.Run("happy/ledger_channel", func(t *testing.T) {
		w := newWatcher(t, rs)
		params, state := channeltest.NewRandomParamsAndState(rng, channeltest.WithVersion(0))
		signedState := makeSignedStateWDummySigs(params, state)

		statesPub, eventsSub, err := w.StartWatchingLedgerChannel(context.TODO(), signedState)

		require.NoError(t, err)
		require.NotNil(t, statesPub)
		require.NotNil(t, eventsSub)

		_, _, err = w.StartWatchingLedgerChannel(context.TODO(), signedState)
		require.Error(t, err, "StartWatching twice for the same channel must error")
	})

	t.Run("happy/sub_channel", func(t *testing.T) {
		w := newWatcher(t, rs)

		// Start watching for ledger channel.
		parentParams, parentState := channeltest.NewRandomParamsAndState(rng, channeltest.WithVersion(0))
		parentSignedState := makeSignedStateWDummySigs(parentParams, parentState)
		statesPub, eventsSub, err := w.StartWatchingLedgerChannel(context.TODO(), parentSignedState)
		require.NoError(t, err)
		require.NotNil(t, statesPub)
		require.NotNil(t, eventsSub)

		// Start watching for a sub-channel with unknown parent ch id.
		childParams, childState := channeltest.NewRandomParamsAndState(rng, channeltest.WithVersion(0))
		childSignedState := makeSignedStateWDummySigs(childParams, childState)
		randomID := channeltest.NewRandomChannelID(rng)
		_, _, err = w.StartWatchingSubChannel(context.TODO(), randomID, childSignedState)
		require.Error(t, err, "Start watching for a sub-channel with unknown parent ch id must error")

		// Start watching for a sub-channel with known parent ch id.
		_, _, err = w.StartWatchingSubChannel(context.TODO(), parentState.ID, childSignedState)
		require.NoError(t, err)
		require.NotNil(t, statesPub)
		require.NotNil(t, eventsSub)

		// Repeat Start watching for the sub-channel.
		_, _, err = w.StartWatchingSubChannel(context.TODO(), parentState.ID, childSignedState)
		require.Error(t, err, "StartWatching twice for the same channel must error")
	})
}

func Test_Watcher_Working(t *testing.T) {
	rng := test.Prng(t)

	t.Run("ledger_channel_without_sub_channel", func(t *testing.T) {
		t.Run("happy/latest_state_registered", func(t *testing.T) {
			params, txs := randomTxsForSingleCh(rng, 3)
			adjSub, trigger := setupAdjudicatorSub(makeRegisteredEvent(txs[2])...)

			rs := &mocks.RegisterSubscriber{}
			rs.On("Subscribe", mock.Anything, mock.Anything).Return(adjSub, nil)
			w := newWatcher(t, rs)
			// Start watching and publish states.
			statesPub, eventsForClient := startWatchingForLedgerChannel(t, w, makeSignedStateWDummySigs(params, txs[0].State))
			require.NoError(t, statesPub.Publish(txs[1]))
			require.NoError(t, statesPub.Publish(txs[2]))

			// Trigger events and assert.
			triggerAdjEventAndExpectNotif(t, trigger, eventsForClient)
			rs.AssertExpectations(t)
		})
		t.Run("happy/newer_than_latest_state_registered", func(t *testing.T) {
			params, txs := randomTxsForSingleCh(rng, 3)
			adjSub, trigger := setupAdjudicatorSub(makeRegisteredEvent(txs[2])...)

			rs := &mocks.RegisterSubscriber{}
			rs.On("Subscribe", mock.Anything, mock.Anything).Return(adjSub, nil)
			w := newWatcher(t, rs)
			// Start watching and publish states.
			statesPub, eventsForClient := startWatchingForLedgerChannel(t, w, makeSignedStateWDummySigs(params, txs[0].State))
			require.NoError(t, statesPub.Publish(txs[1]))

			// Trigger events.
			triggerAdjEventAndExpectNotif(t, trigger, eventsForClient)

			rs.AssertExpectations(t)
		})
	})

	t.Run("ledger_channel_with_sub_channel", func(t *testing.T) {
		t.Run("happy/latest_state_registered", func(t *testing.T) {
			parentParams, parentTxs := randomTxsForSingleCh(rng, 3)
			childParams, childTxs := randomTxsForSingleCh(rng, 3)
			parentTxs[2].Allocation.Locked = []channel.SubAlloc{{ID: childTxs[0].ID}} // Add sub-channel to allocation.

			adjSubParent, triggerParent := setupAdjudicatorSub(makeRegisteredEvent(parentTxs[2])...)
			adjSubChild, triggerChild := setupAdjudicatorSub(makeRegisteredEvent(childTxs[2])...)

			rs := &mocks.RegisterSubscriber{}
			rs.On("Subscribe", mock.Anything, mock.Anything).Return(adjSubParent, nil).Once()
			rs.On("Subscribe", mock.Anything, mock.Anything).Return(adjSubChild, nil).Once()

			w := newWatcher(t, rs)
			// Parent: Start watching and publish states.
			parentSignedState := makeSignedStateWDummySigs(parentParams, parentTxs[0].State)
			statesPubParent, eventsForClientParent := startWatchingForLedgerChannel(t, w, parentSignedState)
			require.NoError(t, statesPubParent.Publish(parentTxs[1]))
			require.NoError(t, statesPubParent.Publish(parentTxs[2]))

			// Child: Start watching and publish states.
			childSignedState := makeSignedStateWDummySigs(childParams, childTxs[0].State)
			statesPubChild, eventsForClientChild := startWatchingForSubChannel(t, w, childSignedState, parentTxs[0].State.ID)
			require.NoError(t, statesPubChild.Publish(childTxs[1]))
			require.NoError(t, statesPubChild.Publish(childTxs[2]))

			// Parent, Child: Trigger events.
			triggerAdjEventAndExpectNotif(t, triggerParent, eventsForClientParent)
			triggerAdjEventAndExpectNotif(t, triggerChild, eventsForClientChild)

			rs.AssertExpectations(t)
		})
		t.Run("happy/newer_than_latest_state_registered", func(t *testing.T) {
			parentParams, parentTxs := randomTxsForSingleCh(rng, 3)
			childParams, childTxs := randomTxsForSingleCh(rng, 3)
			parentTxs[2].Allocation.Locked = []channel.SubAlloc{{ID: childTxs[0].ID}} // Add sub-channel to allocation.

			adjSubParent, triggerParent := setupAdjudicatorSub(makeRegisteredEvent(parentTxs[2])...)
			adjSubChild, triggerChild := setupAdjudicatorSub(makeRegisteredEvent(childTxs[2])...)

			rs := &mocks.RegisterSubscriber{}
			rs.On("Subscribe", mock.Anything, mock.Anything).Return(adjSubParent, nil).Once()
			rs.On("Subscribe", mock.Anything, mock.Anything).Return(adjSubChild, nil).Once()

			w := newWatcher(t, rs)
			// Parent: Start watching and publish states.
			parentSignedState := makeSignedStateWDummySigs(parentParams, parentTxs[0].State)
			statesPubParent, eventsForClientParent := startWatchingForLedgerChannel(t, w, parentSignedState)
			require.NoError(t, statesPubParent.Publish(parentTxs[1]))

			// Child: Start watching and publish states.
			childSignedState := makeSignedStateWDummySigs(childParams, childTxs[0].State)
			statesPubChild, eventsForClientChild := startWatchingForSubChannel(t, w, childSignedState, parentTxs[0].State.ID)
			require.NoError(t, statesPubChild.Publish(childTxs[1]))

			// Parent, Child: Trigger events.
			triggerAdjEventAndExpectNotif(t, triggerParent, eventsForClientParent)
			triggerAdjEventAndExpectNotif(t, triggerChild, eventsForClientChild)
			rs.AssertExpectations(t)
		})
	})
}

func newWatcher(t *testing.T, rs channel.RegisterSubscriber) *local.Watcher {
	t.Helper()

	w, err := local.NewWatcher(local.Config{RegisterSubscriber: rs})
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

// trigger is used for triggering events on the mock adjudicator subscription.
//
// Adjudicator subscription blocks until it is closed.
type trigger struct {
	handle    chan chan time.Time
	adjEvents chan channel.AdjudicatorEvent
}

func (t *trigger) trigger() channel.AdjudicatorEvent {
	select {
	case handles := <-t.handle:
		close(handles)
		return <-t.adjEvents
	default:
		panic(fmt.Sprintf("Number of triggers exceeded maximum limit of %d", cap(t.handle)))
	}
}

// setupAdjudicatorSub returns an adjudicator subscription and a trigger.
//
// On each trigger, an event is sent on the mock adjudicator subscription and
// the corresponding transaction is returned.
//
// If trigger is triggered more times than the number of transactions, it panics.
//
// After all triggers are used, the subscription blocks.
func setupAdjudicatorSub(adjEvents ...channel.AdjudicatorEvent) (*mocks.AdjudicatorSubscription, trigger) {
	adjSub := &mocks.AdjudicatorSubscription{}
	triggers := trigger{
		handle:    make(chan chan time.Time, len(adjEvents)),
		adjEvents: make(chan channel.AdjudicatorEvent, len(adjEvents)),
	}

	for i := range adjEvents {
		handle := make(chan time.Time)
		triggers.handle <- handle
		triggers.adjEvents <- adjEvents[i]

		adjSub.On("Next").Return(adjEvents[i]).WaitUntil(handle).Once()
	}
	// Set up a transaction, that cannot be triggered.
	handle := make(chan time.Time)
	adjSub.On("Next").Return(channel.RegisteredEvent{}).WaitUntil(handle).Once()

	return adjSub, triggers
}

func makeRegisteredEvent(txs ...channel.Transaction) []channel.AdjudicatorEvent {
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

func startWatchingForLedgerChannel(t *testing.T, w *local.Watcher, signedState channel.SignedState) (
	watcher.StatesPub, watcher.AdjudicatorSub) {
	statesPub, eventsSub, err := w.StartWatchingLedgerChannel(context.TODO(), signedState)

	require.NoError(t, err)
	require.NotNil(t, statesPub)
	require.NotNil(t, eventsSub)

	return statesPub, eventsSub
}

func startWatchingForSubChannel(t *testing.T, w *local.Watcher, signedState channel.SignedState, parentID channel.ID) (
	watcher.StatesPub, watcher.AdjudicatorSub) {
	statesPub, eventsSub, err := w.StartWatchingSubChannel(context.TODO(), parentID, signedState)

	require.NoError(t, err)
	require.NotNil(t, eventsSub)

	return statesPub, eventsSub
}

func triggerAdjEventAndExpectNotif(t *testing.T, trigger trigger,
	eventsForClient watcher.AdjudicatorSub) {
	wantEvent := trigger.trigger()
	t.Logf("waiting for adjudicator event for ch %x, version: %v", wantEvent.ID(), wantEvent.Version())
	gotEvent := <-eventsForClient.EventStream()
	require.EqualValues(t, gotEvent, wantEvent)
	t.Logf("received adjudicator event for ch %x, version: %v", wantEvent.ID(), wantEvent.Version())
}

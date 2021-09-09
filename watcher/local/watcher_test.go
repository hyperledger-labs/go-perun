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
	"testing"

	testifyMock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	ethChannel "perun.network/go-perun/backend/ethereum/channel"
	_ "perun.network/go-perun/backend/ethereum/channel/test" // For initilizing channeltest
	"perun.network/go-perun/channel"
	channeltest "perun.network/go-perun/channel/test"
	"perun.network/go-perun/pkg/test"
	"perun.network/go-perun/watcher/internal/mock"
	"perun.network/go-perun/watcher/local"
)

func Test_StartWatching(t *testing.T) {
	rng := test.Prng(t)
	rs := &mock.RegisterSubscriber{}
	rs.On("Subscribe", testifyMock.Anything, testifyMock.Anything).Return(&ethChannel.RegisteredSub{}, nil)

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

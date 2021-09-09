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

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	ethChannel "perun.network/go-perun/backend/ethereum/channel"
	_ "perun.network/go-perun/backend/ethereum/channel/test" // For initilizing channeltest
	"perun.network/go-perun/channel"
	channeltest "perun.network/go-perun/channel/test"
	"perun.network/go-perun/pkg/test"
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

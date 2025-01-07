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

package watcher

import (
	"context"

	"perun.network/go-perun/channel"
)

type (
	// Watcher is the interface used by a client to interact with the Watcher.
	//
	// When a new channel is established, the client should register it with the
	// watcher by calling the StartWatching method.
	//
	// After that, it could publish each state on the StatesPub. If any state is
	// registered or progressed on the blockchain, the corresponding adjudicator
	// event will be relayed by the watcher on the AdjudicatorSub. If the
	// registered state is not the latest state (published to the watcher) then,
	// the watcher will automatically refute by registering the latest state on
	// the blockchain. In case of a multi-ledger channel, the watcher also
	// ensures that the contracts on different ledgers are kept in-sync.
	//
	// Once a channel is closed, the client can de-register the channel from the
	// watcher by calling the StopWatching function. This will close the
	// AdjudicatorSub and StatesSub. The client must not publish states after
	// this.
	Watcher interface {
		StartWatchingLedgerChannel(context.Context, channel.SignedState) (StatesPub, AdjudicatorSub, error)
		StartWatchingSubChannel(_ context.Context, parent channel.ID, _ channel.SignedState) (
			StatesPub, AdjudicatorSub, error)
		StopWatching(context.Context, channel.ID) error
	}

	// StatesPub is the interface used to send newer off-chain states from the
	// client to the watcher.
	//
	// This is initialized when a client starts watching for a given channel.
	// The client can send a state by calling the publish method. Each time a
	// state is published, the watcher will treat it as the latest state
	// without any validation. It is the responsibility of the client to
	// publish states in the correct order and to ensure that they are valid.
	//
	// The publisher will be closed when the client requests the watcher to
	// stop watching or when there is an error. After the client requests the
	// watcher to stop watching, method `Publish` must not be called anymore.
	StatesPub interface {
		Publish(context.Context, channel.Transaction) error
	}

	// AdjudicatorSub is the interface used to relay the adjudicator events
	// from the watcher to the client.
	//
	// This is initialized when a client starts watching for a given channel.
	// The client receives events via the channel returned by the method
	// EventStream. It always returns the same channel and does not support
	// multiplexing.
	//
	// This channel will be closed when client requests the watcher to stop
	// watching or when there is an error. The method Err should tell the
	// possible error.
	AdjudicatorSub interface {
		EventStream() <-chan channel.AdjudicatorEvent
		Err() error
	}
)

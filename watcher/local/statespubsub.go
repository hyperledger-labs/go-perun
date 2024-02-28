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

package local

import (
	"context"
	"sync"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/watcher"
)

var _ watcher.StatesPub = &statesPubSub{}

const statesPubSubBufferSize = 10

type (
	statesPubSub struct {
		once sync.Once
		pipe chan channel.Transaction
	}

	// statesSub is used by the watcher to receive newer off-chain states from
	// the client.
	statesSub interface {
		statesStream() <-chan channel.Transaction
		close()
	}
)

func newStatesPubSub() *statesPubSub {
	return &statesPubSub{
		pipe: make(chan channel.Transaction, statesPubSubBufferSize),
	}
}

// Publish publishes the given transaction (state and signatures on it) to the
// subscriber. It always returns nil. The error result is for implementing
// watcher.StatesPub.
//
// This function panics if the pub-sub instance is already closed. This design
// choice is made because:
//  1. The Watcher requires that the Publish method must not be called after
//     stop watching for a channel. See the documentation of watcher.StatesPub
//     for more details.
//  2. By properly integrating the watcher into the client, it can be guaranteed
//     that this method will never be called after the pub-sub instance is
//     closed and that this method will never panic.
func (s *statesPubSub) Publish(_ context.Context, tx channel.Transaction) error {
	s.pipe <- tx
	return nil
}

// close closes the publisher instance and the associated subscription. Any
// further call to Publish, after a pub-sub is closed will panic.
func (s *statesPubSub) close() {
	s.once.Do(func() { close(s.pipe) })
}

// statesStream returns a channel for consuming the published states. It always
// returns the same channel and does not support multiplexing.
func (s *statesPubSub) statesStream() <-chan channel.Transaction {
	return s.pipe
}

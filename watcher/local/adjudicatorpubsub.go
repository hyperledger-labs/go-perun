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
	"sync"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/watcher"
)

var _ watcher.AdjudicatorSub = &adjudicatorPubSub{}

const adjPubSubBufferSize = 10

type (
	adjudicatorPubSub struct {
		once sync.Once
		pipe chan channel.AdjudicatorEvent
	}

	// adjudicatorPub is used by the watcher to publish the adjudicator events
	// received from the blockchain to the client.
	adjudicatorPub interface {
		publish(adjEvent channel.AdjudicatorEvent)
		close()
	}
)

func newAdjudicatorEventsPubSub() *adjudicatorPubSub {
	return &adjudicatorPubSub{
		pipe: make(chan channel.AdjudicatorEvent, adjPubSubBufferSize),
	}
}

// publish publishes the given adjudicator event to the subscriber.
//
// Panics if the pub-sub instance is already closed. It is implemented this
// way, because:
//  1. The watcher will publish on this pub-sub only when it receives an
//     adjudicator event from the blockchain.
//  2. When de-registering a channel from the watcher, watcher will close the
//     subscription for adjudicator events from blockchain, before closing this
//     pub-sub.
//  3. This way, it can be guaranteed that, this method will never be called
//     after the pub-sub instance is closed.
func (a *adjudicatorPubSub) publish(e channel.AdjudicatorEvent) {
	a.pipe <- e
}

// close closes the publisher instance and the associated subscription. Any
// further call to publish, after a pub-sub is closed will panic.
func (a *adjudicatorPubSub) close() {
	a.once.Do(func() { close(a.pipe) })
}

// EventStream returns a channel for consuming the published adjudicator
// events. It always returns the same channel and does not support
// multiplexing.
//
// The channel will be closed when the pub-sub instance is closed and Err
// should tell the possible error.
func (a *adjudicatorPubSub) EventStream() <-chan channel.AdjudicatorEvent {
	return a.pipe
}

// Err always returns nil. Because, there will be no errors when closing a
// local subscription.
func (a *adjudicatorPubSub) Err() error {
	return nil
}

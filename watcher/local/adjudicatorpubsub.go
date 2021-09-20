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
	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/pkg/sync/atomic"
	"perun.network/go-perun/watcher"
)

var _ watcher.AdjudicatorSub = &adjudicatorPubSub{}

type (
	adjudicatorPubSub struct {
		closed atomic.Bool
		err    error
		pipe   chan channel.AdjudicatorEvent
	}

	adjudicatorPub interface {
		publish(channel.AdjudicatorEvent)
		close() error
	}
)

func newAdjudicatorEventsPubSub() *adjudicatorPubSub {
	return &adjudicatorPubSub{
		pipe: make(chan channel.AdjudicatorEvent, 10),
	}
}

func (a *adjudicatorPubSub) publish(e channel.AdjudicatorEvent) {
	a.pipe <- e
}

// Close closes the publisher instance and all the subscriptions associated
// with it. Any further call to Publish should immediately return error.
func (a *adjudicatorPubSub) close() error {
	if a.closed.IsSet() {
		return errors.New("publisher is closed")
	}
	a.err = errors.WithStack(ErrAlreadyClosed)
	close(a.pipe)
	a.closed.Set()
	return nil
}

func (a *adjudicatorPubSub) EventStream() <-chan channel.AdjudicatorEvent {
	return a.pipe
}

// Err returns the error after the the subscription is closed.
func (a *adjudicatorPubSub) Err() error {
	return a.err
}

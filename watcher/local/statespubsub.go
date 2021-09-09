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
)

// ErrAlreadyClosed indicates the pub-sub has already been closed.
var ErrAlreadyClosed error = errors.New("already closed")

type (
	statesPubSub struct {
		err    error
		closed atomic.Bool
		pipe   chan channel.Transaction
	}

	// statesSub is used by the watcher to receive transactions.
	statesSub interface {
		statesStream() <-chan channel.Transaction
		error() error
		close() error
	}
)

func newStatesPubSub() *statesPubSub {
	return &statesPubSub{
		pipe: make(chan channel.Transaction, 10),
	}
}

// Publish publishes the given state to all the subscribers.
//
// Each time when a transaction is published, watcher will treat it as the
// latest transaction without any validation. It is the responsibility of the
// client to publish transactions in correct order and ensure they are the
// valid.
//
// Returns nil if the state is published. Panics if the subscriptions is
// already closed.
func (s *statesPubSub) Publish(tx channel.Transaction) error {
	s.pipe <- tx
	return nil
}

// Close closes the publisher instance and all the subscriptions associated
// with it. Any further call to Publish should panic.
func (s *statesPubSub) close() error {
	if s.closed.IsSet() {
		return s.err
	}
	s.err = errors.WithStack(ErrAlreadyClosed)
	close(s.pipe)
	s.closed.Set()
	return nil
}

func (s *statesPubSub) statesStream() <-chan channel.Transaction {
	return s.pipe
}

func (s *statesPubSub) error() error {
	return s.err
}

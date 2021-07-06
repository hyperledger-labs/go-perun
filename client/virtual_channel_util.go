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

package client

import (
	"context"
	"math/big"
	"sync"

	"perun.network/go-perun/channel"
)

func (c *Channel) translateBalances(indexMap []channel.Index) channel.Balances {
	state := c.state()
	return transformBalances(state.Balances, state.NumParts(), indexMap)
}

func transformBalances(b channel.Balances, numParts int, indexMap []channel.Index) (_b channel.Balances) {
	_b = make(channel.Balances, len(b))
	for a := range _b {
		_b[a] = make([]*big.Int, numParts)
		// Init with zero.
		for p := range _b[a] {
			_b[a][p] = big.NewInt(0)
		}
		// Fill at specified indices.
		for p, _p := range indexMap {
			_b[a][_p] = b[a][p]
		}
	}
	return
}

func (c *Client) rejectProposal(responder *UpdateResponder, reason string) {
	ctx, cancel := context.WithTimeout(c.Ctx(), responseTimeout)
	defer cancel()
	err := responder.Reject(ctx, reason)
	if err != nil {
		c.log.Warn("Rejecting proposal with reason '%s': %+v", reason, err)
	}
}

func (c *Client) acceptProposal(responder *UpdateResponder) {
	ctx, cancel := context.WithTimeout(c.Ctx(), responseTimeout)
	defer cancel()
	err := responder.Accept(ctx)
	if err != nil {
		c.log.Warn("Accepting proposal error: %+v", err)
	}
}

type watcherEntry struct {
	state interface{}
	done  chan struct{}
}

type stateWatcher struct {
	sync.Mutex
	entries   map[interface{}]watcherEntry
	condition func(ctx context.Context, a, b interface{}) bool
}

func newStateWatcher(c func(ctx context.Context, a, b interface{}) bool) *stateWatcher {
	return &stateWatcher{
		entries:   make(map[interface{}]watcherEntry),
		condition: c,
	}
}

// Await blocks until the condition is met. For any set of matching states, the
// condition function is evaluated at most once.
func (w *stateWatcher) Await(
	ctx context.Context,
	state interface{},
) (err error) {
	match := make(chan struct{}, 1)
	w.register(ctx, state, match)
	defer w.deregister(state)
	select {
	case <-match:
	case <-ctx.Done():
		err = ctx.Err()
	}
	return
}

func (w *stateWatcher) register(
	ctx context.Context,
	state interface{},
	done chan struct{},
) {
	w.Lock()
	defer w.Unlock()

	for k, e := range w.entries {
		if w.condition(ctx, state, e.state) {
			done <- struct{}{}
			e.done <- struct{}{}
			delete(w.entries, k)
			return
		}
	}

	w.entries[state] = watcherEntry{state: state, done: done}
}

func (w *stateWatcher) deregister(
	state interface{},
) {
	w.Lock()
	defer w.Unlock()

	delete(w.entries, state)
}

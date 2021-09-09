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

	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/log"
	"perun.network/go-perun/watcher"
)

type (
	// Watcher implements a local watcher.
	Watcher struct {
		rs channel.RegisterSubscriber
		*registry
	}

	// Config represents the configuration for initializing a local watcher.
	Config struct {
		RegisterSubscriber channel.RegisterSubscriber
	}

	ch struct {
		id     channel.ID
		params *channel.Params
		parent channel.ID

		subChsAccess sync.Mutex
	}
)

// NewWatcher initializes a local watcher.
//
// Local watcher implements pub-sub interfaces over "go channels".
func NewWatcher(cfg Config) (*Watcher, error) {
	w := &Watcher{
		rs:       cfg.RegisterSubscriber,
		registry: newRegistry(),
	}
	return w, nil
}

// StartWatchingLedgerChannel starts watching for a ledger channel.
func (w *Watcher) StartWatchingLedgerChannel(ctx context.Context, signedState channel.SignedState) (
	watcher.StatesPub, watcher.AdjudicatorSub, error) {
	return w.startWatching(ctx, channel.Zero, signedState)
}

// StartWatchingSubChannel starts watching for a sub-channel or virtual channel.
func (w *Watcher) StartWatchingSubChannel(ctx context.Context, parent channel.ID, signedState channel.SignedState) (
	watcher.StatesPub, watcher.AdjudicatorSub, error) {
	parentCh, ok := w.registry.retrieve(parent)
	if !ok {
		return nil, nil, errors.New("parent channel not registered with the watcher")
	}
	parentCh.subChsAccess.Lock()
	defer parentCh.subChsAccess.Unlock()
	return w.startWatching(ctx, parent, signedState)
}

func (w *Watcher) startWatching(ctx context.Context, parent channel.ID, signedState channel.SignedState) (
	watcher.StatesPub, watcher.AdjudicatorSub, error) {
	id := signedState.State.ID
	w.registry.lock()
	defer w.registry.unlock()

	if _, ok := w.registry.retrieveUnsafe(id); ok {
		return nil, nil, errors.New("already watching for this channel")
	}

	eventsFromChainSub, err := w.rs.Subscribe(ctx, id)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "subscribing to adjudicator events from blockchain")
	}
	statesPubSub := newStatesPubSub()
	eventsToClientPubSub := newAdjudicatorEventsPubSub()
	ch := newCh(id, parent, signedState.Params)

	w.registry.addUnsafe(ch)

	go w.handleEventsFromChain(eventsFromChainSub, eventsToClientPubSub, ch)

	return statesPubSub, eventsToClientPubSub, nil
}

func newCh(id, parent channel.ID, params *channel.Params) *ch {
	return &ch{
		id:     id,
		params: params,
		parent: parent,
	}
}

func (w *Watcher) handleEventsFromChain(eventsFromChainSub channel.AdjudicatorSubscription, eventsToClientPubSub adjudicatorPub,
	thisCh *ch) {
	for e := eventsFromChainSub.Next(); e != nil; e = eventsFromChainSub.Next() {
		switch e.(type) {
		case *channel.RegisteredEvent:
			log := log.WithFields(log.Fields{"ID": e.ID(), "Version": e.Version()})
			log.Debug("Received registered event from chain")
			eventsToClientPubSub.Publish(e)
		default:
		}
	}
}

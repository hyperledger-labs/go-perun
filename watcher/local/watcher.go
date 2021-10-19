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

	ch struct {
		id     channel.ID
		params *channel.Params
		parent *ch

		eventsFromChainSub channel.AdjudicatorSubscription
		eventsToClientPub  adjudicatorPub
		statesSub          statesSub

		// subChsAccess mutex is used for thread-safe access of a parent
		// channel and all of its sub-channels. For example, while adding new
		// sub-channels to a ledger channel or while registering dispute for
		// the ledger channel and all its children.
		subChsAccess sync.Mutex
	}

	chInitializer func() (*ch, error)
)

// NewWatcher initializes a local watcher.
//
// It implements the pub-sub interfaces using go channels.
func NewWatcher(rs channel.RegisterSubscriber) (*Watcher, error) {
	w := &Watcher{
		rs:       rs,
		registry: newRegistry(),
	}
	return w, nil
}

// StartWatchingLedgerChannel starts watching for a ledger channel.
func (w *Watcher) StartWatchingLedgerChannel(
	ctx context.Context,
	signedState channel.SignedState,
) (watcher.StatesPub, watcher.AdjudicatorSub, error) {
	return w.startWatching(ctx, nil, signedState)
}

// StartWatchingSubChannel starts watching for a sub-channel or virtual channel.
//
// Parent must be a ledger channel. Because, currently only one level of
// sub-channels is supported.
func (w *Watcher) StartWatchingSubChannel(
	ctx context.Context,
	parent channel.ID,
	signedState channel.SignedState,
) (watcher.StatesPub, watcher.AdjudicatorSub, error) {
	parentCh, ok := w.registry.retrieve(parent)
	if !ok {
		return nil, nil, errors.New("parent channel not registered with the watcher")
	}
	if parentCh.parent != nil {
		return nil, nil, errors.New("parent must be a ledger channel")
	}
	parentCh.subChsAccess.Lock()
	defer parentCh.subChsAccess.Unlock()
	return w.startWatching(ctx, parentCh, signedState)
}

func (w *Watcher) startWatching(
	ctx context.Context,
	parent *ch,
	signedState channel.SignedState,
) (watcher.StatesPub, watcher.AdjudicatorSub, error) {
	id := signedState.State.ID

	var statesPubSub *statesPubSub
	var eventsToClientPubSub *adjudicatorPubSub
	chInitializer := func() (*ch, error) {
		eventsFromChainSub, err := w.rs.Subscribe(ctx, id)
		if err != nil {
			return nil, errors.WithMessage(err, "subscribing to adjudicator events from blockchain")
		}
		statesPubSub = newStatesPubSub()
		eventsToClientPubSub = newAdjudicatorEventsPubSub()
		return newCh(id, parent, signedState.Params, eventsFromChainSub, eventsToClientPubSub, statesPubSub), nil
	}

	ch, err := w.registry.addIfSucceeds(id, chInitializer)
	if err != nil {
		return nil, nil, err
	}
	go ch.handleEventsFromChain(w.rs, w.registry)

	return statesPubSub, eventsToClientPubSub, nil
}

func newCh(
	id channel.ID,
	parent *ch,
	params *channel.Params,
	eventsFromChainSub channel.AdjudicatorSubscription,
	eventsToClientPub adjudicatorPub,
	statesSub statesSub,
) *ch {
	return &ch{
		id:     id,
		params: params,
		parent: parent,

		eventsFromChainSub: eventsFromChainSub,
		eventsToClientPub:  eventsToClientPub,
		statesSub:          statesSub,
	}
}

// handleEventsFromChain receives adjudicator events from the blockchain and
// relays it to the client.
//
// It should be started as a go-routine and returns when the subscription for
// adjudicator events from blockchain is closed.
func (ch *ch) handleEventsFromChain(registerer channel.Registerer, chRegistry *registry) {
	for e := ch.eventsFromChainSub.Next(); e != nil; e = ch.eventsFromChainSub.Next() {
		switch e.(type) {
		case *channel.RegisteredEvent:
			log := log.WithFields(log.Fields{"ID": e.ID(), "Version": e.Version()})
			log.Debug("Received registered event from chain")
			ch.eventsToClientPub.publish(e)

		default:
		}
	}
	err := ch.eventsFromChainSub.Err()
	if err != nil {
		log.Error("Subscription to adjudicator events from chain was closed with error: %v", err)
	}
}

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
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/log"
	"perun.network/go-perun/watcher"
)

const (
	// Duration for which the watcher will wait for latest transactions
	// when an adjudicator event is received.
	statesFromClientWaitTime = 1 * time.Millisecond
)

type (
	// closer is the interface that wraps the close method.
	closer interface {
		close() error
	}

	// Closer is the interface that wraps the Close method.
	Closer interface {
		Close() error
	}

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

		parent              channel.ID
		subChs              map[channel.ID]struct{}
		archivedSubChStates map[channel.ID]channel.SignedState

		registeredVersion uint64
		requestLatestTx   chan struct{}
		latestTx          chan channel.Transaction

		eventsFromChainSub Closer
		eventsToClientSub  closer
		statesPubSub       closer

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
	statesPub, eventsSub, err := w.startWatching(ctx, parent, signedState)
	if err != nil {
		return nil, nil, err
	}
	parentCh.subChs[signedState.State.ID] = struct{}{}
	return statesPub, eventsSub, nil
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
	ch := newCh(id, parent, signedState.Params, eventsFromChainSub, eventsToClientPubSub, statesPubSub)

	w.registry.addUnsafe(ch)

	tx := channel.Transaction{
		State: signedState.State,
		Sigs:  signedState.Sigs,
	}
	go handleStatesFromClient(tx, statesPubSub, ch.requestLatestTx, ch.latestTx)
	go w.handleEventsFromChain(eventsFromChainSub, eventsToClientPubSub, ch)

	return statesPubSub, eventsToClientPubSub, nil
}

func newCh(id, parent channel.ID, params *channel.Params,
	eventsFromChainSub Closer, eventsToClientSub, statesPubSub closer) *ch {
	return &ch{
		id:     id,
		params: params,

		parent:              parent,
		subChs:              make(map[channel.ID]struct{}),
		archivedSubChStates: make(map[channel.ID]channel.SignedState),

		registeredVersion: 0,
		requestLatestTx:   make(chan struct{}),
		latestTx:          make(chan channel.Transaction),

		eventsFromChainSub: eventsFromChainSub,
		eventsToClientSub:  eventsToClientSub,
		statesPubSub:       statesPubSub,
	}
}

func (ch *ch) retreiveLatestTx() channel.Transaction {
	ch.requestLatestTx <- struct{}{}
	return <-ch.latestTx
}

func handleStatesFromClient(currentTx channel.Transaction, statesSub statesSub, requestLatestTx chan struct{},
	latestTx chan channel.Transaction) {
	var _tx channel.Transaction
	var ok bool
	for {
		select {
		case _tx, ok = <-statesSub.statesStream():
			if !ok {
				// TODO: Read error.
				log.WithField("ID", currentTx.State.ID).Info("States sub closed by client. Shutting down handler")
				return
			}
			currentTx = _tx
			log.WithField("ID", currentTx.ID).Debugf("Received state from client %v", currentTx.State)
		case <-requestLatestTx:
			currentTx = receiveTxUntil(statesSub, time.NewTimer(statesFromClientWaitTime).C, currentTx)
			latestTx <- currentTx
		}
	}
}

// receiveTxUntil wait for the transactions on statesPub until the timeout channel is closed
// or statesPub is closed and returns the last received transaction.
// If no transaction was received, then returns currentTx itself.
func receiveTxUntil(statesSub statesSub, timeout <-chan time.Time, currentTx channel.Transaction) channel.Transaction {
	var _tx channel.Transaction
	var ok bool
	for {
		select {
		case _tx, ok = <-statesSub.statesStream():
			if !ok {
				return currentTx // states sub was closed, send the latest event.
			}
			currentTx = _tx
			log.WithField("ID", currentTx.ID).Debugf("Received state from client %v", currentTx.State)
		case <-timeout:
			return currentTx // timer expired, send the latest the event.
		}
	}
}

func (w *Watcher) handleEventsFromChain(eventsFromChainSub channel.AdjudicatorSubscription,
	eventsToClientPubSub adjudicatorPub, thisCh *ch) {
	parent := thisCh
	if thisCh.parent != channel.Zero {
		parent, _ = w.registry.retrieve(thisCh.parent)
	}

	for e := eventsFromChainSub.Next(); e != nil; e = eventsFromChainSub.Next() {
		switch e.(type) {
		case *channel.RegisteredEvent:
			parent.subChsAccess.Lock()
			func() {
				defer parent.subChsAccess.Unlock()

				log := log.WithFields(log.Fields{"ID": e.ID(), "Version": e.Version()})
				log.Debugf("Received registered event from chain: %v", e)

				eventsToClientPubSub.publish(e)

				latestTx := thisCh.retreiveLatestTx()
				log.Debugf("Latest version is (%d)", latestTx.Version)

				if e.Version() < latestTx.Version {
					if e.Version() < thisCh.registeredVersion {
						log.Debugf("Latest version (%d) already registered ", thisCh.registeredVersion)
						return
					}
					log.Debugf("Registering latest version (%d)", latestTx.Version)
					err := registerDispute(w.registry, w.rs, parent)
					if err != nil {
						log.Error("Error registering dispute")
						// TODO: Should the subscription be closed ?
						return
					}
					log.Debug("Registered successfully")
				}
			}()
		case *channel.ProgressedEvent:
			log.Debugf("Received progressed event from chain: %v", e)
			eventsToClientPubSub.publish(e)
		default:
		}
	}

	err := eventsFromChainSub.Err()
	log := log.WithField("ID", thisCh.id)
	log.Errorf("Events from chain sub was closed with error: %v", err)
}

// registerDispute collects the latest transaction for the parent channel and
// each of its children. It then registers dispute for the channel tree.
//
// This function assumes the callers has locked the parent channel.
func registerDispute(r *registry, registerer channel.Registerer, parentCh *ch) error {
	parentTx, subStates := retreiveLatestSubStates(r, parentCh)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := registerer.Register(ctx, makeAdjudicatorReq(parentCh.params, parentTx), subStates)
	if err != nil {
		return err
	}

	parentCh.registeredVersion = parentTx.Version
	for i := range subStates {
		subCh, ok := r.retrieve(parentTx.Allocation.Locked[i].ID)
		if ok {
			subCh.registeredVersion = subStates[i].State.Version
		}
	}
	return nil
}

func retreiveLatestSubStates(r *registry, parent *ch) (channel.Transaction, []channel.SignedState) {
	parentTx := parent.retreiveLatestTx()
	subStates := make([]channel.SignedState, len(parentTx.Allocation.Locked))
	for i := range parentTx.Allocation.Locked {
		// Can be done concurrently.
		subCh, ok := r.retrieve(parentTx.Allocation.Locked[i].ID)
		if ok {
			subChTx := subCh.retreiveLatestTx()
			subStates[i] = makeSignedState(subCh.params, subChTx)
		} else {
			subStates[i] = parent.archivedSubChStates[parentTx.Allocation.Locked[i].ID]
		}
	}
	return parentTx, subStates
}

func makeAdjudicatorReq(params *channel.Params, tx channel.Transaction) channel.AdjudicatorReq {
	return channel.AdjudicatorReq{
		Params:    params,
		Tx:        tx,
		Secondary: false,
	}
}

// StopWatching stops watching for adjudicator events, closes the pub-sub
// instances and removes the channel from the registry.
//
// Client should invoke stop watching for all the sub-channels before invoking
// for the parent ledger channel.
//
// In case of stop watching for sub-channels, watcher ensures that, when it
// receives a registered event for its parent channel or any other sub-channels
// of the parent channel, it is able to successfully refute with the latest
// states for the ledger channel and all its sub-channels (even if the watcher
// has stopped watching for some of the sub-channel).
func (w *Watcher) StopWatching(id channel.ID) error {
	ch, ok := w.retrieve(id)
	if !ok {
		return errors.New("channel not registered with the watcher")
	}

	parent := ch.parent
	if parent != channel.Zero { // Sub channel.
		parentCh, ok := w.retrieve(parent)
		if !ok {
			// Code MUST NOT reach this point
			return errors.New("Fatal error: parent channel not registered with watcher")
		}
		parentCh.subChsAccess.Lock()
		defer parentCh.subChsAccess.Unlock()
		if _, ok := parentCh.retreiveLatestTx().SubAlloc(id); ok {
			parentCh.archivedSubChStates[id] = makeSignedState(ch.params, ch.retreiveLatestTx())
		}
		delete(parentCh.subChs, id)
	} else { // Ledger channel.
		ch.subChsAccess.Lock()
		defer ch.subChsAccess.Unlock()

		if len(ch.subChs) > 0 {
			return fmt.Errorf("cannot de-register when sub-channels are present: %d %v", len(ch.subChs), ch.id)
		}
	}

	errMsg := closePubSubs(ch)
	w.remove(ch.id)

	if errMsg != "" {
		err := errors.New("Stop Watching errors: " + errMsg)
		log.WithField("id", id).Error(err.Error())
		return err
	}
	return nil
}

func closePubSubs(ch *ch) string {
	errMsg := ""
	if err := ch.eventsFromChainSub.Close(); err != nil {
		errMsg += fmt.Sprintf("closing events from chain sub: %v:", err)
	}
	if err := ch.eventsToClientSub.close(); err != nil {
		errMsg += fmt.Sprintf("closing events to client pub-sub: %v:", err)
	}
	if err := ch.statesPubSub.close(); err != nil {
		errMsg += fmt.Sprintf("closing states from client pub-sub: %v:", err)
	}
	return errMsg
}

func makeSignedState(params *channel.Params, tx channel.Transaction) channel.SignedState {
	return channel.SignedState{
		Params: params,
		State:  tx.State,
		Sigs:   tx.Sigs,
	}
}

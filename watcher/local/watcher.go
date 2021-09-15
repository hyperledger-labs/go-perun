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

		registeredVersion uint64
		requestLatestTx   chan struct{}
		latestTx          chan channel.Transaction

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

	tx := channel.Transaction{
		State: signedState.State,
		Sigs:  signedState.Sigs,
	}
	go handleStatesFromClient(tx, statesPubSub, ch.requestLatestTx, ch.latestTx)
	go w.handleEventsFromChain(eventsFromChainSub, eventsToClientPubSub, ch)

	return statesPubSub, eventsToClientPubSub, nil
}

func newCh(id, parent channel.ID, params *channel.Params) *ch {
	return &ch{
		id:                id,
		params:            params,
		parent:            parent,
		registeredVersion: 0,
		requestLatestTx:   make(chan struct{}),
		latestTx:          make(chan channel.Transaction),
	}
}

func (ch *ch) retreiveLatestTx() channel.Transaction {
	ch.requestLatestTx <- struct{}{}
	return <-ch.latestTx
}

func handleStatesFromClient(currentTx channel.Transaction, statesSub statesSub, requestLatestTxn chan struct{},
	latestTx chan channel.Transaction) {
	var _tx channel.Transaction
	var ok bool
	for {
		select {
		case _tx, ok = <-statesSub.statesStream():
			if !ok {
				log.WithField("ID", currentTx.State.ID).Info("States sub closed by client. Shutting down handler")
				return
			}
			currentTx = _tx
			log.WithField("ID", currentTx.ID).Debugf("Received state from client", currentTx.Version, currentTx.ID)
		case <-requestLatestTxn:
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
			log.WithField("ID", currentTx.ID).Debugf("Received state from client", currentTx.Version, currentTx.ID)
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
				log.Debug("Received registered event from chain")

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
}

// registerDispute collects the latest transaction for the parent channel and
// each of its children. It then registers dispute for the channel tree.
//
// This function assumes the callers has locked the parent channel.
func registerDispute(r *registry, registerer channel.Registerer, parentCh *ch) error {
	parentTx := parentCh.retreiveLatestTx()
	subStates := retreiveLatestSubStates(r, parentTx)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := registerer.Register(ctx, makeAdjudicatorReq(parentCh.params, parentTx), subStates)
	if err != nil {
		return err
	}

	parentCh.registeredVersion = parentTx.Version
	for i := range subStates {
		subCh, _ := r.retrieve(parentTx.Allocation.Locked[i].ID)
		subCh.registeredVersion = subStates[i].State.Version
	}
	return nil
}

func retreiveLatestSubStates(r *registry, parentTx channel.Transaction) []channel.SignedState {
	subStates := make([]channel.SignedState, len(parentTx.Allocation.Locked))
	for i := range parentTx.Allocation.Locked {
		// Can be done concurrently.
		subCh, _ := r.retrieve(parentTx.Allocation.Locked[i].ID)
		subChTx := subCh.retreiveLatestTx()
		subStates[i] = channel.SignedState{
			Params: subCh.params,
			State:  subChTx.State,
			Sigs:   subChTx.Sigs,
		}
	}
	return subStates
}

func makeAdjudicatorReq(params *channel.Params, tx channel.Transaction) channel.AdjudicatorReq {
	return channel.AdjudicatorReq{
		Params:    params,
		Tx:        tx,
		Secondary: false,
	}
}

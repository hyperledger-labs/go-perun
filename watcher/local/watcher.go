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

var _ watcher.Watcher = &Watcher{}

const (
	// Duration for which the watcher will wait for latest transactions
	// when an adjudicator event is received.
	//
	// Currently, this is set to an insignificant value of 1ms.
	// Later, when the concept of "sync interval" will be introduced to avoid
	// two registrations for the same channel (as described in test case
	// "happy/older_state_registered_then_newer_state_received", this can be
	// set to the sync interval.
	statesFromClientWaitTime = 1 * time.Millisecond
)

type (
	// Watcher implements a local watcher.
	Watcher struct {
		rs channel.RegisterSubscriber
		*registry
	}

	txRetriever struct {
		request  chan struct{}
		response chan channel.Transaction
	}

	ch struct {
		id       channel.ID
		params   *channel.Params
		isClosed bool
		parent   *ch

		// For keeping track of the sub-channels of this ledger channel
		// registered with the watcher.
		// Sub-channels are added when they are registered with the watcher and
		// removed when they are de-registered from the watcher.
		subChs map[channel.ID]struct{}

		// For keeping track of the last received signed states for a
		// sub-channel, after it is de-registered from the watcher.  These
		// states will be used, when the state of a particular sub-channel is
		// required while disputing any other channel in the channel tree and
		// the particular sub-channel has already been de-registered from the
		// watcher.
		archivedSubChStates map[channel.ID]channel.SignedState

		// For keeping track of the version registered on the blockchain for
		// this channel. This is used to prevent registering the same state
		// more than once.
		registeredVersion uint64

		// For retrieving the latest state (from the handler for receiving
		// off-chain states) when processing events from the blockchain.
		txRetriever txRetriever

		eventsFromChainSub channel.AdjudicatorSubscription
		eventsToClientPub  adjudicatorPub
		statesSub          statesSub

		// subChsAccess mutex is used for thread-safe access of a parent
		// channel and all of its sub-channels. For example, while adding new
		// sub-channels to a ledger channel or while registering dispute for
		// the ledger channel and all its children.
		subChsAccess sync.Mutex

		// wg is used to synchronize the starting and stopping of handler go
		// routines for this channel.
		wg sync.WaitGroup
	}

	chInitializer func() (*ch, error)
)

// Go executes a function in a goroutine, updating the channel's wait group
// before and after execution.
func (ch *ch) Go(fn func()) {
	ch.wg.Add(1)
	go func() {
		defer ch.wg.Done()
		fn()
	}()
}

func (ch *ch) isSubChannel() bool { return ch.parent != nil }

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
	if parentCh.isSubChannel() {
		return nil, nil, errors.New("parent must be a ledger channel")
	}
	parentCh.subChsAccess.Lock()
	defer parentCh.subChsAccess.Unlock()
	statesPub, eventsSub, err := w.startWatching(ctx, parentCh, signedState)
	if err != nil {
		return nil, nil, err
	}
	parentCh.subChs[signedState.State.ID] = struct{}{}
	return statesPub, eventsSub, nil
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
	initialTx := channel.Transaction{
		State: signedState.State,
		Sigs:  signedState.Sigs,
	}
	ch.Go(func() { ch.handleStatesFromClient(initialTx) })
	ch.Go(func() { ch.handleEventsFromChain(w.rs, w.registry) })

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

		subChs:              make(map[channel.ID]struct{}),
		archivedSubChStates: make(map[channel.ID]channel.SignedState),

		registeredVersion: 0,
		txRetriever: txRetriever{
			request:  make(chan struct{}),
			response: make(chan channel.Transaction),
		},

		eventsFromChainSub: eventsFromChainSub,
		eventsToClientPub:  eventsToClientPub,
		statesSub:          statesSub,
	}
}

func (lt txRetriever) retrieve() channel.Transaction {
	lt.request <- struct{}{}
	return <-lt.response
}

// handleStatesFromClient keeps receiving off-chain states published on the
// states subscription.
//
// It also listens for requests on the channel's transaction retriever. When a
// request is received, it sends the latest state on the transaction
// retriever's response channel.
//
// It should be started as a go-routine and returns when the subscription for
// states is closed.
func (ch *ch) handleStatesFromClient(initialTx channel.Transaction) {
	currentTx := initialTx
	var _tx channel.Transaction
	var ok bool
	for {
		select {
		case _tx, ok = <-ch.statesSub.statesStream():
			if !ok {
				log.WithField("ID", currentTx.State.ID).Info("States sub closed by client. Shutting down handler")
				return
			}
			currentTx = _tx
			log.WithField("ID", currentTx.ID).Debugf("Received state from client", currentTx.Version, currentTx.ID)

		case <-ch.txRetriever.request:
			pendingTx, found := readPendingTxs(ch.statesSub, statesFromClientWaitTime)
			if found {
				currentTx = pendingTx
			}
			ch.txRetriever.response <- currentTx
		}
	}
}

// readPendingTxs reads all pending transactions on the states subscription and
// returns the last read transaction.
func readPendingTxs(statesSub statesSub, timeout time.Duration) (channel.Transaction, bool) {
	var currentTx, temp channel.Transaction
	var ok bool
	found := false

	for {
		select {
		case temp, ok = <-statesSub.statesStream():
			if !ok {
				return currentTx, found
			}
			found = true
			currentTx = temp
			log.WithField("ID", currentTx.ID).Debugf("Received state from client", currentTx.Version, currentTx.ID)
		case <-time.NewTimer(timeout).C:
			return currentTx, found
		}
	}
}

// handleEventsFromChain receives adjudicator events from the blockchain and
// relays it to the client. If received state is not the latest, it disputes by
// registering the latest state.
//
// It should be started as a go-routine and returns when the subscription for
// adjudicator events from blockchain is closed.
func (ch *ch) handleEventsFromChain(registerer channel.Registerer, chRegistry *registry) {
	parent := ch
	if ch.isSubChannel() {
		parent = ch.parent
	}
	for e := ch.eventsFromChainSub.Next(); e != nil; e = ch.eventsFromChainSub.Next() {
		switch e.(type) {
		case *channel.RegisteredEvent:
			// This lock ensures, when there are one or more sub-channels and
			// adjudicator event is received for each channel, the events are
			// processed one after the other.
			//
			// Even if older states have been registered for more than one
			// channel in the same family, the channel tree will be registered
			// only once.
			parent.subChsAccess.Lock()

			func() {
				defer parent.subChsAccess.Unlock()

				log := log.WithFields(log.Fields{"ID": e.ID(), "Version": e.Version()})
				log.Debug("Received registered event from chain")

				ch.eventsToClientPub.publish(e)

				latestTx := ch.txRetriever.retrieve()
				log.Debugf("Latest version is (%d)", latestTx.Version)

				if e.Version() < latestTx.Version {
					if e.Version() < ch.registeredVersion {
						log.Debugf("Latest version (%d) already registered ", ch.registeredVersion)
						return
					}

					log.Debugf("Registering latest version (%d)", latestTx.Version)
					err := registerDispute(chRegistry, registerer, parent)
					if err != nil {
						log.Error("Error registering dispute")
						// TODO: Should the subscription be closed with an error ?
						return
					}
					log.Debug("Registered successfully")
				}
			}()
		case *channel.ProgressedEvent:
			log.Debugf("Received progressed event from chain: %v", e)
			ch.eventsToClientPub.publish(e)
		case *channel.ConcludedEvent:
			log.Debugf("Received concluded event from chain: %v", e)
			ch.eventsToClientPub.publish(e)
		default:
			// This should never happen.
			log.Error("Received adjudicator event of unknown type (%T) from chain: %v", e)
		}
	}
	err := ch.eventsFromChainSub.Err()
	if err != nil {
		log.Errorf("Subscription to adjudicator events from chain was closed with error: %v", err)
	}
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
	parentTx := parent.txRetriever.retrieve()
	subStates := make([]channel.SignedState, len(parentTx.Allocation.Locked))
	for i := range parentTx.Allocation.Locked {
		// Can be done concurrently.
		subCh, ok := r.retrieve(parentTx.Allocation.Locked[i].ID)
		if ok {
			subChTx := subCh.txRetriever.retrieve()
			subStates[i] = makeSignedState(subCh.params, subChTx)
		} else {
			subStates[i] = parent.archivedSubChStates[parentTx.Allocation.Locked[i].ID]
		}
	}
	return parentTx, subStates
}

func makeSignedState(params *channel.Params, tx channel.Transaction) channel.SignedState {
	return channel.SignedState{
		Params: params,
		State:  tx.State,
		Sigs:   tx.Sigs,
	}
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
// The client should invoke stop watching for all the sub-channels before
// invoking for the parent ledger channel.
//
// In case of stop watching for sub-channels, watcher ensures that, when it
// receives a registered event for its parent channel or any other sub-channels
// of the parent channel, it is able to successfully refute with the latest
// states for the ledger channel and all its sub-channels (even if the watcher
// has stopped watching for some of the sub-channel).
//
// Context is not used, it is for implementing watcher.Watcher interface.
func (w *Watcher) StopWatching(_ context.Context, id channel.ID) error {
	ch, ok := w.retrieve(id)
	if !ok {
		return errors.New("channel not registered with the watcher")
	}

	parent := ch
	if ch.isSubChannel() {
		parent = ch.parent
	}
	parent.subChsAccess.Lock()
	defer parent.subChsAccess.Unlock()
	if ch.isClosed {
		// Channel could have been closed while were waiting for the mutex locked.
		return errors.New("channel not registered with the watcher")
	}

	if ch.isSubChannel() {
		latestParentTx := ch.parent.txRetriever.retrieve()
		if _, ok := latestParentTx.SubAlloc(id); ok {
			parent.archivedSubChStates[id] = makeSignedState(ch.params, ch.txRetriever.retrieve())
		}
		delete(parent.subChs, id)
	} else if len(ch.subChs) > 0 {
		return fmt.Errorf("cannot de-register when sub-channels are present: %d %v", len(ch.subChs), ch.id)
	}

	closePubSubs(ch)
	w.remove(ch.id)
	ch.isClosed = true
	return nil
}

func closePubSubs(ch *ch) {
	if err := ch.eventsFromChainSub.Close(); err != nil {
		err := errors.WithMessage(err, "closing events from chain sub")
		log.WithField("id", ch.id).Error(err.Error())
	}
	ch.statesSub.close()
	ch.wg.Wait()

	ch.eventsToClientPub.close()
}

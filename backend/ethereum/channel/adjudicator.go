// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package channel // import "perun.network/go-perun/backend/ethereum/channel"

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/pkg/errors"

	"perun.network/go-perun/backend/ethereum/bindings/adjudicator"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/log"
	psync "perun.network/go-perun/pkg/sync"
)

// compile time check that we implement the perun adjudicator interface
var _ channel.Adjudicator = (*Adjudicator)(nil)

// We only loop maxRegisteredEvents times in register, prevents endless loop
var maxRegisteredEvents = 10

// The Adjudicator struct implements the channel.Adjudicator interface
// It provides all functionality to close a channel.
type Adjudicator struct {
	ContractBackend
	contract *adjudicator.Adjudicator
	// The address to which we send all funds.
	Receiver common.Address
	// Structured logger
	log log.Logger
	// Transaction mutex
	mu psync.Mutex
}

// NewAdjudicator creates a new ethereum adjudicator.
func NewAdjudicator(backend ContractBackend, contract common.Address, onchainAddress common.Address) *Adjudicator {
	contr, err := adjudicator.NewAdjudicator(contract, backend)
	if err != nil {
		panic("Could not create a new instance of adjudicator")
	}
	return &Adjudicator{
		ContractBackend: backend,
		contract:        contr,
		Receiver:        onchainAddress,
		log:             log.WithField("account", backend.account.Address),
	}
}

// Withdraw calls conclude and withdraw on a channel.
// Withdraw implements the following logic:
//  - Search for concluded event
//  - If not found -> Call conclude
//  - Wait for concluded event
//  - Search for withdrawn event on my partID
//  - If not found -> call withdraw
//  - Wait for withdrawn event
func (a *Adjudicator) Withdraw(ctx context.Context, request channel.AdjudicatorReq) error {
	// Filter old concluded events.
	err := a.filterConcludedConfirmations(ctx, request.Params.ID())
	if err != nil {
		if err == errConcludedNotFound {
			if err := a.conclude(ctx, request.Params, request.Tx); err != nil {
				return errors.WithMessage(err, "calling conclude")
			}
		} else {
			return errors.WithMessage(err, "filter concluded confirmations")
		}
	}
	return a.withdraw(ctx, request)
}

// SubscribeRegistered returns a new subscription to registered events.
func (a *Adjudicator) SubscribeRegistered(ctx context.Context, params *channel.Params) (channel.RegisteredSubscription, error) {
	stored := make(chan *adjudicator.AdjudicatorStored)
	sub, iter, err := a.waitForStoredEvent(ctx, stored, params)
	if err != nil {
		return nil, errors.WithMessage(err, "waiting for stored event")
	}

	go func() {
		var ev *adjudicator.AdjudicatorStored
		for iter.Next() {
			ev = iter.Event
		}
		if ev != nil {
			stored <- ev
		}
		iter.Close()
	}()

	return &RegisteredSub{
		sub:       sub,
		eventChan: stored,
	}, nil
}

func (a *Adjudicator) waitForStoredEvent(ctx context.Context, stored chan *adjudicator.AdjudicatorStored, params *channel.Params) (event.Subscription, *adjudicator.AdjudicatorStoredIterator, error) {
	// Watch new events
	watchOpts, err := a.newWatchOpts(ctx)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "creating watchopts")
	}
	sub, err := a.contract.WatchStored(watchOpts, stored, []channel.ID{params.ID()})
	if err != nil {
		return nil, nil, errors.Wrapf(err, "watching stored events")
	}

	// Filter old Events
	filterOpts, err := a.newFilterOpts(ctx)
	if err != nil {
		sub.Unsubscribe()
		return nil, nil, errors.WithMessage(err, "creating filter opts")
	}
	iter, err := a.contract.FilterStored(filterOpts, []channel.ID{params.ID()})
	if err != nil {
		sub.Unsubscribe()
		return nil, nil, errors.Wrap(err, "filtering stored events")
	}

	return sub, iter, nil
}

// RegisteredSub implements the channel.RegisteredSub interface.
type RegisteredSub struct {
	sub       event.Subscription
	eventChan chan *adjudicator.AdjudicatorStored
}

// Next returns the next blockchain event.
// Next blocks until a event is returned from the blockchain.
func (r *RegisteredSub) Next() *channel.Registered {
	event := <-r.eventChan
	if event == nil {
		return nil
	}
	return storedToRegisteredEvent(event)
}

// Err returns an error if the subscription to a blockchain failed.
func (r *RegisteredSub) Err() error {
	return <-r.sub.Err()
}

// Close closes this subscription.
func (r *RegisteredSub) Close() error {
	r.sub.Unsubscribe()
	close(r.eventChan)
	return <-r.sub.Err()
}

func storedToRegisteredEvent(event *adjudicator.AdjudicatorStored) *channel.Registered {
	return &channel.Registered{
		ID: event.ChannelID,
		// Idx:     event.Idx,
		Version: event.Version,
		Timeout: time.Unix(int64(event.Timeout), 0),
	}
}

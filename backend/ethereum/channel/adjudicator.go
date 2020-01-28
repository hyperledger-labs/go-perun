// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package channel // import "perun.network/go-perun/backend/ethereum/channel"

import (
	"context"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/pkg/errors"
	"perun.network/go-perun/backend/ethereum/bindings/adjudicator"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/log"
)

// compile time check that we implement the perun adjudicator interface
var _ channel.Adjudicator = (*Adjudicator)(nil)

// The Adjudicator struct implements the channel.Adjudicator interface
// It provides all functionality to close a channel.
type Adjudicator struct {
	ContractBackend
	contract *adjudicator.Adjudicator
	// The address to which we send all funds.
	OnChainAddress common.Address
	// Structured logger
	log log.Logger
	// This mutex prevents us from executing parallel transactions
	mu sync.Mutex
}

// NewETHAdjudicator creates a new ethereum funder.
func NewETHAdjudicator(backend ContractBackend, contract common.Address, onchainAddress common.Address) *Adjudicator {
	contr, err := adjudicator.NewAdjudicator(contract, backend)
	if err != nil {
		panic("Could not create a new instance of adjudicator")
	}
	return &Adjudicator{
		ContractBackend: backend,
		contract:        contr,
		OnChainAddress:  onchainAddress,
		log:             log.WithField("account", backend.account.Address),
	}
}

// Register registers a state on-chain.
// If the state is a final state, register becomes a no-op.
func (a *Adjudicator) Register(ctx context.Context, request channel.AdjudicatorReq) error {
	stored := make(chan *adjudicator.AdjudicatorStored)
	sub, iter, err := a.waitForStoredEvent(ctx, stored, request.Params)
	if err != nil {
		return errors.WithMessage(err, "waiting for stored event")
	}
	defer sub.Unsubscribe()
	if iter.Next() {
		ev := iter.Event
		if request.Tx.Version > ev.Version.Uint64() {
			if err := a.refute(ctx, request); err != nil {
				return errors.WithMessage(err, "refuting with higher version")
			}
		} else {
			//return ev.Timeout, nil
		}
	}
	go func() {
		for iter.Next() {
			stored <- iter.Event
		}
	}()
	if err := a.register(ctx, request); err != nil {
		return errors.WithMessage(err, "registering state")
	}
	select {
	case ev := <-stored:
		for request.Tx.Version > ev.Version.Uint64() {
			if err := a.refute(ctx, request); err != nil {
				return errors.WithMessage(err, "refuting with higher version")
			}
		}
		//return ev.Timeout, nil
	case <-ctx.Done():
		return errors.New("did not receive stored event in time")
	}
	panic("should never happen")
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
	// Filter old concluded events
	filterOpts, err := a.newFilterOpts(ctx)
	if err != nil {
		return errors.WithMessage(err, "creating filteropts")
	}
	iter, err := a.contract.FilterConcluded(filterOpts, []channel.ID{request.Params.ID()})
	if err != nil {
		return errors.Wrap(err, "filtering concluded events")
	}
	if !iter.Next() {
		// Call concluded
		if err := a.conclude(ctx, request.Params, request.Tx); err != nil {
			return errors.WithMessage(err, "calling conclude")
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
		for iter.Next() {
			stored <- iter.Event
		}
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
	return &channel.Registered{
		ID: event.ChannelID,
		//Idx: event.Idx,
		//Version: event.Version.Uint64(),
		Timeout: time.Unix(event.Timeout.Int64(), 0),
	}
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

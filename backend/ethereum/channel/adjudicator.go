// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package channel // import "perun.network/go-perun/backend/ethereum/channel"

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/event"
	"github.com/pkg/errors"
	"perun.network/go-perun/backend/ethereum/bindings/adjudicator"
	"perun.network/go-perun/channel"
)

type Adjudicator struct {
	ContractBackend
	contract adjudicator.Adjudicator
}

func (a *Adjudicator) Register(ctx context.Context, params *channel.Params, tx channel.Transaction) error {
	return nil
}

// Search for concluded event
// If not found -> Call conclude
// search for withdraw event on my partID
// call -> withdraw
// wait for withdraw event
func (a *Adjudicator) Withdraw(ctx context.Context, params *channel.Params, tx channel.Transaction) error {
	// Filter old concluded events
	filterOpts := bind.FilterOpts{
		Start:   uint64(1),
		End:     nil,
		Context: ctx}
	iter, err := a.contract.FilterConcluded(&filterOpts, []channel.ID{params.ID()})
	if err != nil {
		return errors.Wrap(err, "filtering concluded events")
	}
	if !iter.Next() {
		// Call concluded
	}

	return nil
}

// SubRegistered returns a new subscription to registered events.
func (a *Adjudicator) SubRegistered(ctx context.Context, params *channel.Params) (*RegisteredSub, error) {
	registered := make(chan *adjudicator.AdjudicatorRegistered)
	// Watch new events
	watchOpts, err := a.newWatchOpts(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "error creating watchopts")
	}
	sub, err := a.contract.WatchRegistered(watchOpts, registered, []channel.ID{params.ID()})
	if err != nil {
		return nil, errors.Wrapf(err, "watching registered events")
	}

	// Filter old Events
	filterOpts := bind.FilterOpts{
		Start:   uint64(1),
		End:     nil,
		Context: ctx}
	iter, err := a.contract.FilterRegistered(&filterOpts, []channel.ID{params.ID()})
	if err != nil {
		return nil, errors.Wrap(err, "filtering registered events")
	}
	go func() {
		for iter.Next() {
			registered <- iter.Event
		}
	}()

	return &RegisteredSub{
		sub:       sub,
		eventChan: registered,
	}, nil
}

// RegisteredSub implements the channel.RegisteredSub interface.
type RegisteredSub struct {
	sub       event.Subscription
	eventChan chan *adjudicator.AdjudicatorRegistered
}

// Next returns the next blockchain event.
// Next blocks until a event is returned from the blockchain.
func (r *RegisteredSub) Next() *Registered {
	event := <-r.eventChan
	if event == nil {
		return nil
	}
	return &Registered{
		ID: event.ChannelID,
		//Idx: event.Idx,
		Version: event.Version.Uint64(),
		//Timeout: event.Timeout,
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

type Registered struct {
	ID      channel.ID
	Idx     channel.Index // index of the participant who registered the event
	Version uint64
	Timeout time.Time // Timeout when the event can be concluded or progressed
}

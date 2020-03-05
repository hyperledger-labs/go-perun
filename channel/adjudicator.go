// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package channel

import (
	"context"
	"time"

	"perun.network/go-perun/wallet"
)

type (
	// An Adjudicator represents an adjudicator contract on the blockchain. It
	// has methods for state registration and withdrawal of channel funds.
	// A channel state needs to be registered before the concluded state can be
	// withdrawn after a possible timeout.
	//
	// Furthermore, it has a method for subscribing to RegisteredEvents. Those
	// events might be triggered by a Register call on the adjudicator from any
	// channel participant.
	Adjudicator interface {
		// Register should register the given channel state on-chain. It must be
		// taken into account that a peer might already have registered the same or
		// even an old state for the same channel. If registration was successful,
		// it should return the timeout when withdrawal can be initiated with
		// Withdraw.
		Register(context.Context, AdjudicatorReq) (*RegisteredEvent, error)

		// Withdraw should conclude and withdraw the registered state, so that the
		// final outcome is set on the asset holders and funds are withdrawn
		// (dependent on the architecture of the contracts). It must be taken into
		// account that a peer might already have concluded the same channel.
		Withdraw(context.Context, AdjudicatorReq) error

		// SubscribeRegistered returns a RegisteredEvent subscription. The
		// subscription should be a subscription of the newest past as well as
		// future events. The subscription should only be valid within the given
		// context: If the context is canceled, its Next method should return nil
		// and Err should return the context's error.
		SubscribeRegistered(context.Context, *Params) (RegisteredSubscription, error)
	}

	// An AdjudicatorReq collects all necessary information to make calls to the
	// adjudicator.
	AdjudicatorReq struct {
		Params *Params
		Acc    wallet.Account
		Tx     Transaction
		Idx    Index
	}

	// RegisteredEvent is the abstract event that signals a successful state
	// registration on the blockchain.
	RegisteredEvent struct {
		ID      ID        // Channel ID
		Idx     Index     // Index of the participant who registered the event.
		Version uint64    // Registered version.
		Timeout time.Time // Timeout when the event can be concluded or progressed
	}

	// A RegisteredSubscription is a subscription to RegisteredEvents for a
	// specific channel. The subscription should also return the newest past
	// RegisteredEvent, if there is any.
	//
	// The usage of the subscription should be similar to that of an iterator.
	// Next calls should block until a new event is generated (or the first past
	// event has been found). If the channel is closed or an error is produced,
	// Next should return nil and Err should tell the possible error.
	RegisteredSubscription interface {
		// Next returns the newest past or next future event. If the subscription is
		// closed or any other error occurs, it should return nil.
		Next() *RegisteredEvent

		// Err returns the error status of the subscription. After Next returns nil,
		// Err should be checked for an error.
		Err() error

		// Close closes the subscription. Any call to Next should immediately return
		// nil.
		Close() error
	}
)

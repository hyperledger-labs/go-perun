// Copyright 2020 - See NOTICE file for copyright holders.
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

package channel

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"perun.network/go-perun/wallet"
)

type (
	// An Adjudicator represents an adjudicator contract on the blockchain. It has
	// methods for state registration, progression and withdrawal of channel
	// funds. A channel state needs to be registered before it can be progressed
	// or withdrawn. No-App channels skip the force-execution phase and can be
	// withdrawn directly after the refutation phase has finished. Channels in a
	// final state can directly be withdrawn after registration.
	//
	// Furthermore, an Adjudicator has a method for subscribing to
	// AdjudicatorEvents. Those events might be triggered by a Register or
	// Progress call on the adjudicator from any channel participant.
	Adjudicator interface {
		// Register should register the given ledger channel state on-chain.
		// If the channel has locked funds into sub-channels, the corresponding
		// signed sub-channel states must be provided.
		Register(context.Context, AdjudicatorReq, []SignedState) error

		// Withdraw should conclude and withdraw the registered state, so that the
		// final outcome is set on the asset holders and funds are withdrawn
		// (dependent on the architecture of the contracts). It must be taken into
		// account that a peer might already have concluded the same channel.
		// If the channel has locked funds in sub-channels, the states of the
		// corresponding sub-channels need to be supplied additionally.
		Withdraw(context.Context, AdjudicatorReq, StateMap) error

		// Progress should try to progress an on-chain registered state to the new
		// state given in ProgressReq. The Transaction field only needs to
		// contain the state, the signatures can be nil, since the old state is
		// already registered on the adjudicator.
		Progress(context.Context, ProgressReq) error

		// Subscribe returns an AdjudicatorEvent subscription.
		//
		// The context should only be used to establish the subscription. The
		// framework will call Close on the subscription once the respective channel
		// controller shuts down.
		Subscribe(context.Context, *Params) (AdjudicatorSubscription, error)
	}

	// An AdjudicatorReq collects all necessary information to make calls to the
	// adjudicator.
	//
	// If the Secondary flag is set to true, it is assumed that this is an
	// on-chain request that is executed by the other channel participants as well
	// and the Adjudicator backend may run an optimized on-chain transaction
	// protocol, possibly saving unnecessary double sending of transactions.
	AdjudicatorReq struct {
		Params    *Params
		Acc       wallet.Account
		Tx        Transaction
		Idx       Index // Always the own index
		Secondary bool  // Optimized secondary call protocol
	}

	// SignedState represents a signed channel state including parameters.
	SignedState struct {
		Params *Params
		State  *State
		Sigs   []wallet.Sig
	}

	// A ProgressReq collects all necessary information to do a progress call to
	// the adjudicator.
	ProgressReq struct {
		AdjudicatorReq            // Tx should refer to the currently registered state
		NewState       *State     // New state to progress into
		Sig            wallet.Sig // Own signature on the new state
	}

	// An AdjudicatorSubscription is a subscription to AdjudicatorEvents for a
	// specific channel. The subscription should also return the most recent past
	// event, if there is any. It should skip all Registered events and only
	// return the most recent Progressed event if the channel already got progressed.
	//
	// The usage of the subscription should be similar to that of an iterator.
	// Next calls should block until a new event is generated (or the first past
	// event has been found). If the subscription is closed or an error is
	// produced, Next should return nil and Err should tell the possible error.
	AdjudicatorSubscription interface {
		// Next returns the most recent past or next future event. If the subscription is
		// closed or any other error occurs, it should return nil.
		Next() AdjudicatorEvent

		// Err returns the error status of the subscription. After Next returns nil,
		// Err should be checked for an error.
		Err() error

		// Close closes the subscription. Any call to Next should immediately return
		// nil.
		Close() error
	}

	// An AdjudicatorEvent is any event that an on-chain adjudicator call might
	// cause, currently either a Registered or Progressed event.
	// The type of the event should be checked with a type switch.
	AdjudicatorEvent interface {
		ID() ID
		Timeout() Timeout
		Version() uint64
	}

	// An AdjudicatorEventBase implements the AdjudicatorEvent interface. It can
	// be embedded to implement an AdjudicatorEvent.
	AdjudicatorEventBase struct {
		IDV      ID      // Channel ID
		TimeoutV Timeout // Current phase timeout
		VersionV uint64  // Registered version
	}

	// ProgressedEvent is the abstract event that signals an on-chain progression.
	ProgressedEvent struct {
		AdjudicatorEventBase        // Channel ID and ForceExec phase timeout
		State                *State // State that was progressed into
		Idx                  Index  // Index of the participant who progressed
	}

	// RegisteredEvent is the abstract event that signals a successful state
	// registration on the blockchain.
	RegisteredEvent struct {
		AdjudicatorEventBase // Channel ID and Refutation phase timeout
		State                *State
		Sigs                 []wallet.Sig
	}

	// ConcludedEvent signals channel conclusion.
	ConcludedEvent struct {
		AdjudicatorEventBase
	}

	// A Timeout is an abstract timeout of a channel dispute. A timeout can be
	// elapsed and it can be waited on it to elapse.
	Timeout interface {
		// IsElapsed should return whether the timeout has elapsed at the time of
		// the call of this method.
		IsElapsed(context.Context) bool

		// Wait waits for the timeout to elapse. If the context is canceled, Wait
		// should return immediately with the context's error.
		Wait(context.Context) error
	}

	// StateMap represents a channel state tree.
	StateMap map[ID]*State
)

// NewProgressReq creates a new ProgressReq object.
func NewProgressReq(ar AdjudicatorReq, newState *State, sig wallet.Sig) *ProgressReq {
	return &ProgressReq{ar, newState, sig}
}

// NewAdjudicatorEventBase creates a new AdjudicatorEventBase object.
func NewAdjudicatorEventBase(c ID, t Timeout, v uint64) *AdjudicatorEventBase {
	return &AdjudicatorEventBase{
		IDV:      c,
		TimeoutV: t,
		VersionV: v,
	}
}

// ID returns the channel ID.
func (b AdjudicatorEventBase) ID() ID { return b.IDV }

// Timeout returns the phase timeout.
func (b AdjudicatorEventBase) Timeout() Timeout { return b.TimeoutV }

// Version returns the channel version.
func (b AdjudicatorEventBase) Version() uint64 { return b.VersionV }

// NewRegisteredEvent creates a new RegisteredEvent.
func NewRegisteredEvent(id ID, timeout Timeout, version uint64, state *State, sigs []wallet.Sig) *RegisteredEvent {
	return &RegisteredEvent{
		AdjudicatorEventBase: AdjudicatorEventBase{
			IDV:      id,
			TimeoutV: timeout,
			VersionV: version,
		},
		State: state,
		Sigs:  sigs,
	}
}

// NewProgressedEvent creates a new ProgressedEvent.
func NewProgressedEvent(id ID, timeout Timeout, state *State, idx Index) *ProgressedEvent {
	return &ProgressedEvent{
		AdjudicatorEventBase: AdjudicatorEventBase{
			IDV:      id,
			TimeoutV: timeout,
			VersionV: state.Version,
		},
		State: state,
		Idx:   idx,
	}
}

// ElapsedTimeout is a Timeout that is always elapsed.
type ElapsedTimeout struct{}

// IsElapsed returns true.
func (t *ElapsedTimeout) IsElapsed(context.Context) bool { return true }

// Wait immediately return nil.
func (t *ElapsedTimeout) Wait(context.Context) error { return nil }

// String says that this is an always elapsed timeout.
func (t *ElapsedTimeout) String() string { return "<Always elapsed timeout>" }

// TimeTimeout is a Timeout that elapses after a fixed time.Time.
type TimeTimeout struct{ time.Time }

// IsElapsed returns whether the current time is after the fixed timeout.
func (t *TimeTimeout) IsElapsed(context.Context) bool { return t.After(time.Now()) }

// Wait waits until the timeout has elapsed or the context is cancelled.
func (t *TimeTimeout) Wait(ctx context.Context) error {
	select {
	case <-time.After(time.Until(t.Time)):
		return nil
	case <-ctx.Done():
		return errors.Wrap(ctx.Err(), "ctx done")
	}
}

// String returns the timeout's date and time string.
func (t *TimeTimeout) String() string {
	return fmt.Sprintf("<Timeout: %v>", t.Time)
}

// MakeStateMap creates a new StateMap object.
func MakeStateMap() StateMap {
	return make(map[ID]*State)
}

// Add adds the given states to the state map.
func (m StateMap) Add(states ...*State) {
	for _, s := range states {
		m[s.ID] = s
	}
}

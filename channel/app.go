// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel

import (
	"perun.network/go-perun/pkg/io"
	"perun.network/go-perun/wallet"
)

type (
	// An App is an abstract interface for an app definition. Either a StateApp or
	// ActionApp should be implemented.
	App interface {
		// Def is an identifier of the channel application. It is usually the
		// (counterfactual) on-chain address of the stateless contract that defines
		// what is a validTransition
		Def() wallet.Address
	}

	// A StateApp is advanced by full state udpates. The validity of state
	// transitions is checked with method ValidTransition.
	StateApp interface {
		App
		// ValidTransition checks if the application specific rules of the given
		// transition from `from` to `to` are fulfilled.
		// The implementation should return a TransitionError describing the
		// invalidity of the transition, if it is not valid. It should return a normal
		// error (with attached stacktrace from pkg/errors) if there was any other
		// runtime error, not related to the invalidity of the transition itself.
		ValidTransition(parameters *Params, from, to *State) error

		// ValidInit checks whether the given State is a valid initial state.
		// Note that version == 0 and correct channel IDs are checked by the
		// framework. This method should only perform app-specific checks.
		ValidInit(*Params, *State) error
	}

	// An ActionApp is advanced by first collecting actions from the participants
	// and then applying those actions to the state. In a sense it is a more
	// fine-grained version of a StateApp and allows for more optimized
	// state channel applications.
	ActionApp interface {
		App
		// ValidAction checks if the provided Action by the participant at the given
		// index is valid, applied to the provided Params and State.
		// The implementation should return an ActionError describing the invalidity
		// of the action. It should return a normal error (with attached stacktrace
		// from pkg/errors) if there was any other runtime error, not related to the
		// invalidity of the action itself.
		ValidAction(*Params, *State, uint, Action) error

		// ApplyAction applies the given actions to the provided channel state and
		// returns the resulting new state.
		// the version counter should be increased by one.
		// The implementation should return an ActionError describing the invalidity
		// of the action. It should return a normal error (with attached stacktrace
		// from pkg/errors) if there was any other runtime error, not related to the
		// invalidity of the action itself.
		ApplyActions(*Params, *State, []Action) (*State, error)

		// InitState creates the initial state from the given actions. The actual
		// State will be created by the machine and only the initial allocation of
		// funds and app data can be set, as the channel id is specified by the
		// parameters and version must be 0.
		InitState(*Params, []Action) (Allocation, Data, error)
	}

	// Actions are applied to channel states to result in new states
	Action = io.Serializable
)

func IsStateApp(app App) bool {
	_, ok := app.(StateApp)
	return ok
}

func IsActionApp(app App) bool {
	_, ok := app.(ActionApp)
	return ok
}

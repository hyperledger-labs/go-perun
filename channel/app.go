// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel

import (
	"io"

	perunio "perun.network/go-perun/pkg/io"
	"perun.network/go-perun/wallet"
)

type (
	// An App is an abstract interface for an app definition. Either a StateApp or
	// ActionApp should be implemented.
	App interface {
		// Def is an identifier of the channel application. It is usually the
		// (counterfactual) on-chain address of the stateless contract that defines
		// what valid actions or transitions are.
		Def() wallet.Address

		// DecodeData decodes data specific to this application. This has to be
		// defined on an application-level because every app can have completely
		// different data; during decoding the application needs to be known to
		// know how to decode the data.
		DecodeData(io.Reader) (Data, error)
	}

	// A StateApp is advanced by full state updates. The validity of state
	// transitions is checked with method ValidTransition.
	StateApp interface {
		App

		// ValidTransition should check that the app-specific rules of the given
		// transition from `from` to `to` are fulfilled.
		// `actor` is the index of the acting party whose action resulted in the new state.
		// The implementation should return a StateTransitionError describing the
		// invalidity of the transition, if it is not valid. It should return a normal
		// error (with attached stacktrace from pkg/errors) if there was any other
		// runtime error, not related to the invalidity of the transition itself.
		ValidTransition(parameters *Params, from, to *State, actor Index) error

		// ValidInit should perform app-specific checks for a valid initial state.
		// The framework guarantees to only pass initial states with version == 0,
		// correct channel ID and valid initial allocation.
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
		ValidAction(*Params, *State, Index, Action) error

		// ApplyAction applies the given actions to the provided channel state and
		// returns the resulting new state.
		// The version counter should be increased by one.
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

		// DecodeAction decodes actions specific to this application. This has to be
		// defined on an application-level because every app can have completely
		// different actions; during decoding the application needs to be known to
		// know how to decode an action.
		DecodeAction(io.Reader) (Action, error)
	}

	// Actions are applied to channel states to result in new states.
	// Actions need to be Encoders so they can be sent over the wire.
	// Decoding happens with ActionApp.DecodeAction() since the app context needs
	// to be known at the point of decoding.
	Action = perunio.Encoder

	AppBackend interface {
		// AppFromDefinition creates an app from its defining address. It is
		// possible that multiple apps are in use, which is why creation happens
		// over a central AppFromDefinition function.  One possible implementation
		// is that the app is just read from an app registry, mapping addresses to
		// apps.
		AppFromDefinition(wallet.Address) (App, error)

		// DecodeAsset decodes an asset from a stream.
		DecodeAsset(io.Reader) (Asset, error)
	}
)

func IsStateApp(app App) bool {
	_, ok := app.(StateApp)
	return ok
}

func IsActionApp(app App) bool {
	_, ok := app.(ActionApp)
	return ok
}

// appBackend stores the AppBackend globally for the channel package.
var appBackend AppBackend

// SetAppBackend sets the channel package's app backend. This is more specific
// than the blockchain backend, so it has to be set separately.
// It should be set in the init() function of the package that implements the
// application.
func SetAppBackend(b AppBackend) {
	if appBackend != nil {
		panic("App backend already set")
	}
	appBackend = b
}

// AppFromDefinition is a global wrapper call to the app backend function.
func AppFromDefinition(def wallet.Address) (App, error) {
	return appBackend.AppFromDefinition(def)
}

// DecodeAsset is a global wrapper call to the app backend function.
func DecodeAsset(r io.Reader) (Asset, error) {
	return appBackend.DecodeAsset(r)
}

// Copyright 2019 - See NOTICE file for copyright holders.
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
	"io"

	"github.com/pkg/errors"

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

	// An Action is applied to a channel state to result in new state.
	// Actions need to be Encoders so they can be sent over the wire.
	// Decoding happens with ActionApp.DecodeAction() since the app context needs
	// to be known at the point of decoding.
	Action = perunio.Encoder

	// AppBackend provides functionality to create an App from an Address.
	// The AppBackend needs to be implemented for every state channel application.
	AppBackend interface {
		// AppFromDefinition creates an app from its defining address. It is
		// possible that multiple apps are in use, which is why creation happens
		// over a central AppFromDefinition function.  One possible implementation
		// is that the app is just read from an app registry, mapping addresses to
		// apps.
		AppFromDefinition(wallet.Address) (App, error)
	}
)

// IsStateApp returns true if the app is a StateApp.
func IsStateApp(app App) bool {
	_, ok := app.(StateApp)
	return ok
}

// IsActionApp returns true if the app is an ActionApp.
func IsActionApp(app App) bool {
	_, ok := app.(ActionApp)
	return ok
}

// appBackend stores the AppBackend globally for the channel package.
var appBackend AppBackend = &MockAppBackend{}

// isAppBackendSet whether the appBackend was already set with `SetAppBackend`.
var isAppBackendSet bool

// SetAppBackend sets the channel package's app backend. This is more specific
// than the blockchain backend, so it has to be set separately.
// The app backend is set to the MockAppBackend by default. Because the MockApp is in
// package channel, we cannot set it through the usual init.go idiom.
// The app backend can be changed once by another app (by a SetAppBackend call
// of the app package's init() function).
func SetAppBackend(b AppBackend) {
	if isAppBackendSet {
		panic("app backend already set")
	}
	isAppBackendSet = true
	appBackend = b
}

// AppFromDefinition is a global wrapper call to the app backend function.
func AppFromDefinition(def wallet.Address) (App, error) {
	return appBackend.AppFromDefinition(def)
}

// OptAppEnc makes an optional App value encodable.
type OptAppEnc struct {
	App
}

// Encode encodes an optional App value.
func (e OptAppEnc) Encode(w io.Writer) error {
	if e.App != nil {
		return perunio.Encode(w, true, e.App.Def())
	}
	return perunio.Encode(w, false)
}

// OptAppDec makes an optional App value decodable.
type OptAppDec struct {
	App *App
}

// Decode decodes an optional App value.
func (d OptAppDec) Decode(r io.Reader) (err error) {
	var hasApp bool
	if err = perunio.Decode(r, &hasApp); err != nil {
		return err
	}
	if !hasApp {
		*d.App = nil
		return nil
	}
	appDef, err := wallet.DecodeAddress(r)
	if err != nil {
		return errors.WithMessage(err, "decode app address")
	}
	*d.App, err = AppFromDefinition(appDef)
	return errors.WithMessage(err, "resolve app")
}

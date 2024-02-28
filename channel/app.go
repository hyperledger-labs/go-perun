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
	"encoding"
	"io"

	"github.com/pkg/errors"

	"perun.network/go-perun/wire/perunio"
)

type (
	// AppID represents an app identifier.
	AppID interface {
		encoding.BinaryMarshaler
		encoding.BinaryUnmarshaler
		Equal(appID AppID) bool

		// Key returns the object key which can be used as a map key.
		Key() AppIDKey
	}

	// An App is an abstract interface for an app definition. Either a StateApp or
	// ActionApp should be implemented.
	App interface {
		// Def is an identifier of the channel application. It is usually the
		// (counterfactual) on-chain address of the stateless contract that defines
		// what valid actions or transitions are.
		// Calling this function on a NoApp panics, so ensure that IsNoApp
		// returns false.
		Def() AppID

		// NewData returns a new instance of data specific to NoApp, intialized
		// to its zero value.
		//
		// This should be used for unmarshalling the data from its binary
		// representation.
		//
		// This has to be defined on an application-level because every app can
		// have completely different data.
		NewData() Data
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
		ValidInit(params *Params, state *State) error
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
		ValidAction(params *Params, state *State, idx Index, action Action) error

		// ApplyAction applies the given actions to the provided channel state and
		// returns the resulting new state.
		// The version counter should be increased by one.
		// The implementation should return an ActionError describing the invalidity
		// of the action. It should return a normal error (with attached stacktrace
		// from pkg/errors) if there was any other runtime error, not related to the
		// invalidity of the action itself.
		ApplyActions(params *Params, state *State, actions []Action) (*State, error)

		// InitState creates the initial state from the given actions. The actual
		// State will be created by the machine and only the initial allocation of
		// funds and app data can be set, as the channel id is specified by the
		// parameters and version must be 0.
		InitState(Params *Params, actions []Action) (Allocation, Data, error)

		// NewAction returns an instance of action specific to this
		// application. This has to be defined on an application-level because
		// every app can have completely different action. This instance should
		// be used for unmarshalling the binary representation of the action.
		//
		// It is zero-value is safe to use.
		NewAction() Action
	}

	// An Action is applied to a channel state to result in new state.
	//
	// It is sent as binary data over the wire. For unmarshalling an action from
	// its binary representation, use ActionApp.NewAction() to create an Action
	// instance specific to the application and then unmarshal using it.
	Action interface {
		encoding.BinaryMarshaler
		encoding.BinaryUnmarshaler
	}

	// AppResolver provides functionality to create an App from an Address.
	// The AppResolver needs to be implemented for every state channel application.
	AppResolver interface {
		// Resolve creates an app from its defining identifier. It is possible that
		// multiple apps are in use, which is why creation happens over a central
		// Resolve function. This function is intended to resolve app definitions
		// coming in on the wire.
		Resolve(appID AppID) (App, error)
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

// OptAppEnc makes an optional App value encodable.
type OptAppEnc struct {
	App
}

// Encode encodes an optional App value.
func (e OptAppEnc) Encode(w io.Writer) error {
	if IsNoApp(e.App) {
		return perunio.Encode(w, false)
	}
	return perunio.Encode(w, true, e.App.Def())
}

// OptAppDec makes an optional App value decodable.
type OptAppDec struct {
	App *App
}

// Decode decodes an optional App value.
func (d OptAppDec) Decode(r io.Reader) error {
	var err error
	var hasApp bool
	if err = perunio.Decode(r, &hasApp); err != nil {
		return err
	}
	if !hasApp {
		*d.App = NoApp()
		return nil
	}
	appDef := backend.NewAppID()
	err = perunio.Decode(r, appDef)
	if err != nil {
		return errors.WithMessage(err, "decode app address")
	}
	*d.App, err = Resolve(appDef)
	return errors.WithMessage(err, "resolve app")
}

// OptAppAndDataEnc makes an optional pair of App definition and Data encodable.
type OptAppAndDataEnc struct {
	App  App
	Data Data
}

// Encode encodes an optional pair of App definition and Data.
func (o OptAppAndDataEnc) Encode(w io.Writer) error {
	return perunio.Encode(w, OptAppEnc{App: o.App}, o.Data)
}

// OptAppAndDataDec makes an optional pair of App definition and Data decodable.
type OptAppAndDataDec struct {
	App  *App
	Data *Data
}

// Decode decodes an optional pair of App definition and Data.
func (o OptAppAndDataDec) Decode(r io.Reader) error {
	if err := perunio.Decode(r, OptAppDec{App: o.App}); err != nil {
		return err
	}

	*o.Data = (*o.App).NewData()
	return perunio.Decode(r, *o.Data)
}

// AppShouldEqual compares two Apps for equality.
func AppShouldEqual(expected, actual App) error {
	if IsNoApp(expected) && IsNoApp(actual) {
		return nil
	}

	if !IsNoApp(expected) && IsNoApp(actual) ||
		IsNoApp(expected) && !IsNoApp(actual) {
		return errors.New("(non-)nil App definitions")
	}

	if !expected.Def().Equal(actual.Def()) {
		return errors.New("different App definitions")
	}

	return nil
}

// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package test // import "perun.network/go-perun/channel/test"

import (
	"io"

	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
	perun "perun.network/go-perun/wallet"
	"perun.network/go-perun/wire"
)

// MockApp a mocked App whose behaviour is determined by the MockOp passed to it either as State.Data or Action.
// It is a StateApp and ActionApp at the same time.
type MockApp struct {
	definition perun.Address
}

var _ channel.ActionApp = new(MockApp)
var _ channel.StateApp = new(MockApp)

// MockOp serves as Action and State.Data for MockApp.
type MockOp uint8

var _ channel.Action = new(MockOp)
var _ channel.Data = new(MockOp)

const (
	// OpValid function call should succeed.
	OpValid MockOp = iota
	// OpErr function call should return an error.
	OpErr
	// OpTransitionErr function call should return a TransisionError.
	OpTransitionErr
	// OpActionErr function call should return an ActionError.
	OpActionErr
	// OpPanic function call should panic.
	OpPanic
)

// NewMockOp returns a pointer to a MockOp with the given value
// this is needed, since &MockOp{OpValid} does not work.
func NewMockOp(op MockOp) *MockOp {
	return &op
}

// Encode encodes a MockOp into an io.Writer.
func (o MockOp) Encode(w io.Writer) error {
	return wire.Encode(w, uint8(o))
}

// Decode decodes a MockOp from an io.Reader.
func (o *MockOp) Decode(r io.Reader) error {
	return wire.Decode(r, (*uint8)(o))
}

// Clone returns a deep copy of a
func (o MockOp) Clone() channel.Data {
	return &o
}

// NewMockApp create an App with the given definition.
func NewMockApp(definition perun.Address) *MockApp {
	return &MockApp{definition: definition}
}

// Def returns the definition on the MockApp.
func (a MockApp) Def() perun.Address {
	return a.definition
}

// DecodeAction returns a decoded MockOp or an error.
func (a MockApp) DecodeAction(r io.Reader) (channel.Action, error) {
	var act MockOp
	return &act, act.Decode(r)
}

// DecodeData returns a decoded MockOp or an error.
func (a MockApp) DecodeData(r io.Reader) (channel.Data, error) {
	var data MockOp
	return &data, data.Decode(r)
}

// ValidTransition checks the transition for validity.
func (a MockApp) ValidTransition(params *channel.Params, from, to *channel.State) error {
	return a.execMockOp(from.Data.(*MockOp))
}

// ValidInit checks the initial state for validity.
func (a MockApp) ValidInit(params *channel.Params, state *channel.State) error {
	return a.execMockOp(state.Data.(*MockOp))
}

// ValidAction checks the action for validity.
func (a MockApp) ValidAction(params *channel.Params, state *channel.State, part uint, act channel.Action) error {
	return a.execMockOp(act.(*MockOp))
}

// ApplyActions applies the actions unto a copy of state and returns the result or an error.
func (a MockApp) ApplyActions(params *channel.Params, state *channel.State, acts []channel.Action) (*channel.State, error) {
	ret := state.Clone()
	ret.Version++

	return ret, a.execMockOp(acts[0].(*MockOp))
}

// InitState Checks for the validity of the passed arguments as initial state.
func (a MockApp) InitState(params *channel.Params, rawActs []channel.Action) (channel.Allocation, channel.Data, error) {
	return channel.Allocation{}, nil, a.execMockOp(rawActs[0].(*MockOp))
}

// execMockOp executes the operation indicated by the MockOp from the MockOp.
func (a MockApp) execMockOp(op *MockOp) error {
	switch *op {
	case OpErr:
		return errors.New("MockOp: runtime error")
	case OpTransitionErr:
		return channel.NewStateTransitionError(channel.ID{}, "MockOp: state transition error")
	case OpActionErr:
		return channel.NewActionError(channel.ID{}, "MockOp: action error")
	case OpPanic:
		panic("MockOp: panic")
	case OpValid:
		return nil
	default:
		panic("MockOp: unhandled switch case")
	}
}

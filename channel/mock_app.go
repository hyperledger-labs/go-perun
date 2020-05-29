// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package channel

import (
	"io"

	"github.com/pkg/errors"

	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wire"
)

// MockApp a mocked App whose behaviour is determined by the MockOp passed to it either as State.Data or Action.
// It is a StateApp and ActionApp at the same time.
type MockApp struct {
	definition wallet.Address
}

var _ ActionApp = (*MockApp)(nil)
var _ StateApp = (*MockApp)(nil)

// MockOp serves as Action and State.Data for MockApp.
type MockOp uint64

var _ Action = (*MockOp)(nil)
var _ Data = (*MockOp)(nil)

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
	return wire.Encode(w, uint64(o))
}

// Decode decodes a MockOp from an io.Reader.
func (o *MockOp) Decode(r io.Reader) error {
	return wire.Decode(r, (*uint64)(o))
}

// Clone returns a deep copy of a MockOp.
func (o MockOp) Clone() Data {
	return &o
}

// NewMockApp create an App with the given definition.
func NewMockApp(definition wallet.Address) *MockApp {
	return &MockApp{definition: definition}
}

// Def returns the definition on the MockApp.
func (a MockApp) Def() wallet.Address {
	return a.definition
}

// DecodeAction returns a decoded MockOp or an error.
func (a MockApp) DecodeAction(r io.Reader) (Action, error) {
	var act MockOp
	return &act, act.Decode(r)
}

// DecodeData returns a decoded MockOp or an error.
func (a MockApp) DecodeData(r io.Reader) (Data, error) {
	var data MockOp
	return &data, data.Decode(r)
}

// ValidTransition checks the transition for validity.
func (a MockApp) ValidTransition(params *Params, from, to *State, actor Index) error {
	return a.execMockOp(from.Data.(*MockOp))
}

// ValidInit checks the initial state for validity.
func (a MockApp) ValidInit(params *Params, state *State) error {
	return a.execMockOp(state.Data.(*MockOp))
}

// ValidAction checks the action for validity.
func (a MockApp) ValidAction(params *Params, state *State, part Index, act Action) error {
	return a.execMockOp(act.(*MockOp))
}

// ApplyActions applies the actions unto a copy of state and returns the result or an error.
func (a MockApp) ApplyActions(params *Params, state *State, acts []Action) (*State, error) {
	ret := state.Clone()
	ret.Version++

	return ret, a.execMockOp(acts[0].(*MockOp))
}

// InitState Checks for the validity of the passed arguments as initial state.
func (a MockApp) InitState(params *Params, rawActs []Action) (Allocation, Data, error) {
	return Allocation{}, nil, a.execMockOp(rawActs[0].(*MockOp))
}

// execMockOp executes the operation indicated by the MockOp from the MockOp.
func (a MockApp) execMockOp(op *MockOp) error {
	switch *op {
	case OpErr:
		return errors.New("MockOp: runtime error")
	case OpTransitionErr:
		return NewStateTransitionError(ID{}, "MockOp: state transition error")
	case OpActionErr:
		return NewActionError(ID{}, "MockOp: action error")
	case OpPanic:
		panic("MockOp: panic")
	case OpValid:
		return nil
	default:
		panic("MockOp: unhandled switch case")
	}
}

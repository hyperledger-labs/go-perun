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
	"encoding/binary"
	"fmt"

	"github.com/pkg/errors"

	"perun.network/go-perun/wallet"
)

// MockApp a mocked App whose behaviour is determined by the MockOp passed to it either as State.Data or Action.
// It is a StateApp and ActionApp at the same time.
type MockApp struct {
	definition wallet.Address
}

var (
	_ ActionApp = (*MockApp)(nil)
	_ StateApp  = (*MockApp)(nil)
)

// MockOp serves as Action and State.Data for MockApp.
type MockOp uint64

// mockOpLen is the length of binary representation of mockOp, in bytes.
const mockOpLen = 8

var (
	_ Action = (*MockOp)(nil)
	_ Data   = (*MockOp)(nil)

	// byteOrder used for marshalling/unmarshalling mockOp to/from its binary
	// representation.
	byteOrder = binary.BigEndian
)

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

// MarshalBinary marshals MockOp to its binary representation.
func (o MockOp) MarshalBinary() ([]byte, error) {
	data := make([]byte, mockOpLen)
	byteOrder.PutUint64(data, uint64(o))
	return data, nil
}

// UnmarshalBinary unmarshals MockOp from its binary representation.
func (o *MockOp) UnmarshalBinary(data []byte) error {
	if len(data) != mockOpLen {
		//nolint: goerr113  // We do not want to define this as constant error.
		return fmt.Errorf("unexpected length %d, want %d", len(data), mockOpLen)
	}
	*o = MockOp(byteOrder.Uint64(data))
	return nil
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

// NewAction returns an instance of Action specific to MockApp.
func (a MockApp) NewAction() Action {
	return new(MockOp)
}

// NewData returns a new instance of data specific to MockApp, intialized to
// its zero value.
//
// This should be used for unmarshalling the data from its binary
// representation.
func (a MockApp) NewData() Data {
	return new(MockOp)
}

// ValidTransition checks the transition for validity.
func (a MockApp) ValidTransition(params *Params, from, to *State, actor Index) error {
	op, ok := from.Data.(*MockOp)
	if !ok {
		return fmt.Errorf("wrong data type: expected *MockOp, got %T", from.Data)
	}
	return a.execMockOp(op)
}

// ValidInit checks the initial state for validity.
func (a MockApp) ValidInit(params *Params, state *State) error {
	op, ok := state.Data.(*MockOp)
	if !ok {
		return fmt.Errorf("wrong data type: expected *MockOp, got %T", state.Data)
	}
	return a.execMockOp(op)
}

// ValidAction checks the action for validity.
func (a MockApp) ValidAction(params *Params, state *State, part Index, act Action) error {
	op, ok := act.(*MockOp)
	if !ok {
		return fmt.Errorf("wrong data type: expected *MockOp, got %T", act)
	}
	return a.execMockOp(op)
}

// ApplyActions applies the actions unto a copy of state and returns the result or an error.
func (a MockApp) ApplyActions(params *Params, state *State, acts []Action) (*State, error) {
	ret := state.Clone()
	ret.Version++

	op, ok := acts[0].(*MockOp)
	if !ok {
		return nil, fmt.Errorf("wrong data type: expected *MockOp, got %T", acts[0])
	}

	return ret, a.execMockOp(op)
}

// InitState Checks for the validity of the passed arguments as initial state.
func (a MockApp) InitState(params *Params, rawActs []Action) (Allocation, Data, error) {
	op, ok := rawActs[0].(*MockOp)
	if !ok {
		return Allocation{}, nil, fmt.Errorf("wrong data type: expected *MockOp, got %T", rawActs[0])
	}

	return Allocation{}, nil, a.execMockOp(op)
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

// MockAppResolver resolves every given `wallet.Address` to a `mockApp`.
type MockAppResolver struct{}

var _ AppResolver = &MockAppResolver{}

// Resolve creates an app from its defining address.
func (m *MockAppResolver) Resolve(addr wallet.Address) (App, error) {
	return NewMockApp(addr), nil
}

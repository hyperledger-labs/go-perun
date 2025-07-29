// Copyright 2025 - See NOTICE file for copyright holders.
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
	"github.com/pkg/errors"

	"perun.network/go-perun/wallet"
)

// A StateMachine is the channel pushdown automaton around a StateApp.
// It implements the state transitions specific for StateApps: Init and Update.
type StateMachine struct {
	*machine

	app StateApp `cloneable:"shallow"`
}

// NewStateMachine creates a new StateMachine.
func NewStateMachine(acc map[wallet.BackendID]wallet.Account, params Params) (*StateMachine, error) {
	app, ok := params.App.(StateApp)
	if !ok {
		return nil, errors.New("app must be StateApp")
	}

	m, err := newMachine(acc, params)
	if err != nil {
		return nil, err
	}

	return &StateMachine{
		machine: m,
		app:     app,
	}, nil
}

// RestoreStateMachine restores a state machine to the data given by Source.
func RestoreStateMachine(acc map[wallet.BackendID]wallet.Account, source Source) (*StateMachine, error) {
	app, ok := source.Params().App.(StateApp)
	if !ok {
		return nil, errors.New("app must be StateApp")
	}

	m, err := restoreMachine(acc, source)
	if err != nil {
		return nil, err
	}

	return &StateMachine{
		machine: m,
		app:     app,
	}, nil
}

// Init sets the initial staging state to the given balance and data.
// It returns the initial state and own signature on it.
func (m *StateMachine) Init(initBals Allocation, initData Data) error {
	if err := m.expect(PhaseTransition{InitActing, InitSigning}); err != nil {
		return err
	}

	// we start with the initial state being the staging state
	initState, err := newState(&m.params, initBals, initData)
	if err != nil {
		return err
	}
	if err := m.app.ValidInit(&m.params, initState); err != nil {
		return err
	}

	m.setStaging(InitSigning, initState)
	return nil
}

// Update makes the provided state the staging state.
// It is checked whether this is a valid state transition.
func (m *StateMachine) Update(stagingState *State, actor Index) error {
	if err := m.expect(PhaseTransition{Acting, Signing}); err != nil {
		return err
	}

	if err := m.validTransition(stagingState, actor); err != nil {
		return err
	}

	m.setStaging(Signing, stagingState)
	return nil
}

// ForceUpdate makes the provided state the staging state.
func (m *StateMachine) ForceUpdate(stagingState *State, actor Index) error {
	m.setStaging(Signing, stagingState)
	return nil
}

// CheckUpdate checks if the given state is a valid transition from the current
// state and if the given signature is valid. It is a read-only operation that
// does not advance the state machine.
func (m *StateMachine) CheckUpdate(
	state *State, actor Index,
	sig wallet.Sig, sigIdx Index,
) error {
	if err := m.validTransition(state, actor); err != nil {
		return err
	}
	for _, add := range m.params.Parts[sigIdx] {
		if ok, err := Verify(add, state, sig); err != nil {
			return errors.WithMessagef(err, "verifying signature[%d]", sigIdx)
		} else if !ok {
			return errors.Errorf("invalid signature[%d]", sigIdx)
		}
	}
	return nil
}

// Clone returns a deep copy of StateMachine.
func (m *StateMachine) Clone() *StateMachine {
	return &StateMachine{
		machine: m.machine.Clone(),
		app:     m.app,
	}
}

// validTransition makes all the default transition checks and additionally
// checks for a valid application specific transition.
// This is where a StateMachine and ActionMachine differ. In an ActionMachine,
// every action is checked as being a valid action by the application definition
// and the resulting state by applying all actions to the old state is by
// definition a valid new state.
func (m *StateMachine) validTransition(to *State, actor Index) (err error) {
	if actor >= m.N() {
		return errors.New("actor index is out of range")
	}
	if err := m.ValidTransition(to); err != nil {
		return err
	}

	if err = m.app.ValidTransition(&m.params, m.currentTX.State, to, actor); IsStateTransitionError(err) {
		return err
	}
	return errors.WithMessagef(err, "runtime error in application's ValidTransition()")
}

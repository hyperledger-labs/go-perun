// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel

import (
	"github.com/pkg/errors"

	"perun.network/go-perun/wallet"
)

type StateMachine struct {
	*machine

	app StateApp
}

func NewStateMachine(acc wallet.Account, params Params) (*StateMachine, error) {
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
// It returns the initial state and own signature on it.
// It is checked whether this is a valid state transition.
func (m *StateMachine) Update(stagingState *State) error {
	if err := m.expect(PhaseTransition{Acting, Signing}); err != nil {
		return err
	}

	if err := m.validTransition(stagingState); err != nil {
		return err
	}

	m.setStaging(Signing, stagingState)
	return nil
}

// validTransition makes all the default transition checks and additionally
// checks for a valid application specific transition.
// This is where a StateMachine and ActionMachine differ. In an ActionMachine,
// every action is checked as being a valid action by the application definition
// and the resulting state by applying all actions to the old state is by
// definition a valid new state.
func (m *StateMachine) validTransition(to *State) error {
	if err := m.machine.validTransition(to); err != nil {
		return err
	}

	if err := m.app.ValidTransition(&m.params, m.currentTX.State, to); IsStateTransitionError(err) {
		return err
	} else if err != nil {
		return errors.WithMessagef(err, "runtime error in application's ValidTransition() (ID: %x)", m.params.id)
	}

	return nil
}

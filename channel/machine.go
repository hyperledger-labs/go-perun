// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel

import (
	"fmt"

	"github.com/pkg/errors"

	"perun.network/go-perun/log"
	"perun.network/go-perun/pkg/io"
	"perun.network/go-perun/wallet"
)

type (
	// Phase is a phase of the channel pushdown automaton
	Phase uint8

	// PhaseTransition represents a transition between two phases
	PhaseTransition struct {
		From, To Phase
	}
)

const (
	Nil Phase = iota
	Initializing
	Funding
	Open
	Updating
	Finalizing
	Final
	Settled
)

func (p Phase) String() string {
	return [...]string{"Nil", "Initializing", "Funding", "Open", "Updating", "Finalizing", "Final", "Settled"}[p]
}

func (t PhaseTransition) String() string {
	return fmt.Sprintf("%v->%v", t.From, t.To)
}

var stagingPhases = []Phase{Initializing, Updating, Finalizing}

// Machine is the channel pushdown automaton that handles phase transitions.
// It checks for correct signatures and valid state transitions.
type Machine struct {
	phase     Phase
	acc       wallet.Account
	idx       uint
	params    Params
	stagingTX Transaction
	currentTX Transaction
	prevTXs   []Transaction

	// subs contains subscribers to each phase transition
	subs map[Phase]map[string]chan<- PhaseTransition
}

// NewMachine retruns a new uninitialized machine for the given parameters.
func NewMachine(acc wallet.Account, params Params) (*Machine, error) {
	idx := wallet.AddrIdx(params.Parts, acc.Address())
	if idx == -1 {
		return nil, errors.New("account not part of participant set.")
	}

	return &Machine{
		phase:   Initializing,
		acc:     acc,
		idx:     uint(idx),
		params:  params,
		prevTXs: make([]Transaction, 0),
		subs:    make(map[Phase]map[string]chan<- PhaseTransition),
	}, nil
}

// N returns the number of participants of the channel parameters of this machine
func (m *Machine) N() uint {
	return uint(len(m.params.Parts))
}

// Phase returns the current phase
func (m *Machine) Phase() Phase {
	return m.phase
}

// setPhase is internally used to set the phase and notify all subscribers of
// the phase transition
func (m *Machine) setPhase(p Phase) {
	oldPhase := m.phase
	m.phase = p
	m.notifySubs(oldPhase)
}

// Sig returns the signature on the currently staged state.
// The signature is caclulated and saved to the staging TX's signature slice
// if it was not calculated before.
func (m *Machine) Sig() (sig Sig, err error) {
	if m.stagingTX.Sigs[m.idx] == nil {
		sig, err = Sign(m.acc, &m.params, m.stagingTX.State)
		if err != nil {
			return
		}
		m.stagingTX.Sigs[m.idx] = sig
	} else {
		sig = m.stagingTX.Sigs[m.idx]
	}
	return
}

// AddSig verifies the provided signature of another participant on the staging
// state and if successful adds it to the staged transaction. It also checks
// whether the signature has already been set and in that case errors.
// It should not happen that a signature of the same participant is set twice.
// It is also checked that the current phase is any staging phase.
func (m *Machine) AddSig(idx uint, sig Sig) error {
	if !inPhase(m.phase, stagingPhases) {
		return m.error(m.selfTransition(), "can only add signature in a staging phase")
	}

	if m.stagingTX.Sigs[idx] != nil {
		return m.errorf(m.selfTransition(), "signature for idx %d already present", idx)
	}

	if ok, err := Verify(m.params.Parts[idx], &m.params, m.stagingTX.State, sig); err != nil {
		return err
	} else if !ok {
		return m.error(m.selfTransition(), "invalid signature")
	}

	m.stagingTX.Sigs[idx] = sig
	return nil
}

// Init sets the initial staging state to the given balance and data.
// It returns the own signature on the initial state.
func (m *Machine) Init(initBals Allocation, initData io.Serializable) (Sig, error) {
	if m.phase != Nil {
		return nil, m.error(PhaseTransition{Nil, Initializing}, "can only initialize from Nil state")
	}

	// we start with the initial state being the staging state
	initState, err := newState(&m.params, initBals, initData)
	if err != nil {
		return nil, err
	}

	return m.setStaging(Initializing, initState)
}

// Update initiates the update routine of the state machine and makes the
// provided state the staging state.
// It returns the own signature on the staging state.
func (m *Machine) Update(newState *State) (Sig, error) {
	if newState.IsFinal {
		return nil, m.error(PhaseTransition{Open, Updating}, "can only update to non-final state (use Finalize() instead)")
	}
	return m.update(Updating, newState)
}

// Finalize initiates the finalization routine of the state machine and makes
// the provided state the staging state.
// It returns the own signature on the staging state.
func (m *Machine) Finalize(newState *State) (Sig, error) {
	if !newState.IsFinal {
		return nil, m.error(PhaseTransition{Open, Finalizing}, "can only finalize final state (use Update() instead)")
	}
	return m.update(Finalizing, newState)
}

// update makes the provided state the staging state.
// It returns the own signature on the staging state.
// It is checked whether this is a valid phase transition from Open and valid
// state transition.
func (m *Machine) update(to Phase, toState *State) (Sig, error) {
	if m.phase != Open {
		return nil, m.error(PhaseTransition{Open, to}, "can only update from Open state")
	}

	// TODO(Seb): change ValidTransition to only return error?
	if _, err := m.params.ValidTransition(m.currentTX.State, toState); err != nil {
		return nil, err
	}

	return m.setStaging(to, toState)
}

// setStaging sets the given phase and state as staging state.
// It returns the signature on the staging state and an error if signature
// generation fails.
func (m *Machine) setStaging(phase Phase, state *State) (Sig, error) {
	m.stagingTX = Transaction{
		State: state,
		Sigs:  make([]Sig, m.N()),
	}

	sig, err := m.Sig()
	if err != nil {
		return nil, errors.WithMessage(err, "signing state update")
	}

	m.setPhase(phase)
	return sig, err
}

// EnableFunding promotes the initial staging state to the current funding state.
// A valid phase transition and the existence of all signatures is checked.
func (m *Machine) EnableFunding() error {
	return m.enableStaged(Initializing, Funding)
}

// EnableUpdate promotes the current staging state to the current state.
// A valid phase transition and the existence of all signatures is checked.
func (m *Machine) EnableUpdate() error {
	return m.enableStaged(Updating, Open)
}

// EnableFinal promotes the final staging state to the final current state.
// A valid phase transition and the existence of all signatures is checked.
func (m *Machine) EnableFinal() error {
	return m.enableStaged(Finalizing, Final)
}

// enableStaged checks that
//   1. the current phase is `from` and
//   2. all signatures of the staging transactions have been set.
// If successful, the staging transaction is promoted to be the current
// transaction. If not, an error is returned.
func (m *Machine) enableStaged(from, to Phase) error {
	if m.phase != from {
		return m.error(PhaseTransition{from, to}, "no staging phase")
	}

	for i, sig := range m.stagingTX.Sigs {
		if sig == nil {
			return m.errorf(PhaseTransition{from, to}, "signature %d missing from staging TX", i)
		}
	}

	m.prevTXs = append(m.prevTXs, m.currentTX) // push current to previous
	m.currentTX = m.stagingTX                  // promote staging to current
	m.stagingTX = Transaction{}                // clear staging

	m.phase = to
	return nil
}

// SetFunded tells the state machine that the channel got funded and progresses
// to the Open phase
func (m *Machine) SetFunded() error {
	if m.phase != Funding {
		return m.error(PhaseTransition{Funding, Open}, "can only set as funded in phase Funding")
	}

	m.setPhase(Open)
	return nil
}

// SetSettled tells the state machine that the final state was settled on the
// blockchain or funding channel and progresses to the Settled state
func (m *Machine) SetSettled() error {
	if m.phase != Final {
		return m.error(PhaseTransition{Final, Settled}, "can only settle from Final phase")
	}

	m.setPhase(Settled)
	return nil
}

// Subscribe subscribes go-channel `sub` to phase `phase` under the name `who`.
// If the machine changes into phase `phase`, the phase transition is sent on
// channel `sub`.
func (m *Machine) Subscribe(phase Phase, who string, sub chan<- PhaseTransition) {
	if m.subs[phase] == nil {
		m.subs[phase] = make(map[string]chan<- PhaseTransition)
	}

	m.subs[phase][who] = sub
}

// notifySubs notifies all subscribers to the current phase that a phase
// transition from the provided phase `from` has happened.
func (m *Machine) notifySubs(from Phase) {
	if m.subs[m.phase] == nil {
		// no subscribers
		return
	}

	transition := PhaseTransition{from, m.phase}
	for who, sub := range m.subs[m.phase] {
		log.Tracef("PhaseTransition: %v, notifying subscriber %s", transition, who)
		sub <- transition
	}
}

// error constructs a new PhaseTransitionError
func (m *Machine) error(expected PhaseTransition, msg string) *PhaseTransitionError {
	return newPhaseTransitionError(m.params.ID(), m.phase, expected, msg)
}

// error constructs a new PhaseTransitionError
func (m *Machine) errorf(expected PhaseTransition, format string, args ...interface{}) *PhaseTransitionError {
	return newPhaseTransitionErrorf(m.params.ID(), m.phase, expected, format, args...)
}

// inPhase returns whether phase is in phases
func inPhase(phase Phase, phases []Phase) bool {
	for _, p := range phases {
		if p == phase {
			return true
		}
	}
	return false
}

// selfTransition returns a PhaseTransition from current to current phase
func (m *Machine) selfTransition() PhaseTransition {
	return PhaseTransition{m.phase, m.phase}
}

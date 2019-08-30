// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel

import (
	"fmt"

	"github.com/pkg/errors"

	"perun.network/go-perun/log"
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
	InitActing Phase = iota
	InitSigning
	Funding
	Acting
	Signing
	Final
	Settled
)

func (p Phase) String() string {
	return [...]string{"InitActing", "InitSigning", "Funding", "Acting", "Signing", "Final", "Settled"}[p]
}

func (t PhaseTransition) String() string {
	return fmt.Sprintf("%v->%v", t.From, t.To)
}

var signingPhases = []Phase{InitSigning, Signing}

// A machine is the channel pushdown automaton that handles phase transitions.
// It checks for correct signatures and valid state transitions.
type machine struct {
	phase     Phase
	acc       wallet.Account
	idx       uint
	params    Params
	stagingTX Transaction
	currentTX Transaction
	prevTXs   []Transaction

	// subs contains subscribers to each phase transition
	subs map[Phase]map[string]chan<- PhaseTransition
	// log is a fields logger for this machine
	log log.Logger
}

// newMachine retruns a new uninitialized machine for the given parameters.
func newMachine(acc wallet.Account, params Params) (*machine, error) {
	idx := wallet.IndexOfAddr(params.Parts, acc.Address())
	if idx == -1 {
		return nil, errors.New("account not part of participant set.")
	}

	return &machine{
		phase:  InitActing,
		acc:    acc,
		idx:    uint(idx),
		params: params,
		subs:   make(map[Phase]map[string]chan<- PhaseTransition),
		log:    log.WithField("ID", params.id),
	}, nil
}

// N returns the number of participants of the channel parameters of this machine
func (m *machine) N() uint {
	return uint(len(m.params.Parts))
}

// Phase returns the current phase
func (m *machine) Phase() Phase {
	return m.phase
}

// setPhase is internally used to set the phase and notify all subscribers of
// the phase transition
func (m *machine) setPhase(p Phase) {
	m.log.Tracef("phase transition: %v", PhaseTransition{m.phase, p})
	oldPhase := m.phase
	m.phase = p
	m.notifySubs(oldPhase)
}

// inPhase returns whether phase is in phases.
func inPhase(phase Phase, phases []Phase) bool {
	for _, p := range phases {
		if p == phase {
			return true
		}
	}
	return false
}

// Sig returns the own signature on the currently staged state.
// The signature is caclulated and saved to the staging TX's signature slice
// if it was not calculated before.
// A call to Sig only makes sense in a signing phase.
func (m *machine) Sig() (sig Sig, err error) {
	if !inPhase(m.phase, signingPhases) {
		return nil, m.error(m.selfTransition(), "can only create own signature in a signing phase")
	}

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

// StagingState returns the staging state. It should usually be called after
// entering a signing phase to get the new staging state, which might have been
// created during Init() or Update() (for ActionApps).
func (m *machine) StagingState() *State {
	return m.stagingTX.State
}

// AddSig verifies the provided signature of another participant on the staging
// state and if successful adds it to the staged transaction. It also checks
// whether the signature has already been set and in that case errors.
// It should not happen that a signature of the same participant is set twice.
// It is also checked that the current phase is a signing phase.
// If the index is out of bounds, a panic occurs as this is an invalid usage of
// the machine.
func (m *machine) AddSig(idx uint, sig Sig) error {
	if !inPhase(m.phase, signingPhases) {
		return m.error(m.selfTransition(), "can only add signature in a signing phase")
	}

	if m.stagingTX.Sigs[idx] != nil {
		return errors.Errorf("signature for idx %d already present (ID: %x)", idx, m.params.id)
	}

	if ok, err := Verify(m.params.Parts[idx], &m.params, m.stagingTX.State, sig); err != nil {
		return err
	} else if !ok {
		return errors.Errorf("invalid signature for idx %d (ID: %x)", idx, m.params.id)
	}

	m.stagingTX.Sigs[idx] = sig
	return nil
}

// setStaging sets the given phase and state as staging state.
// It returns the signature on the staging state and an error if signature
// generation fails.
func (m *machine) setStaging(phase Phase, state *State) {
	m.stagingTX = Transaction{
		State: state,
		Sigs:  make([]Sig, m.N()),
	}

	m.setPhase(phase)
}

// EnableInit promotes the initial staging state to the current funding state.
// A valid phase transition and the existence of all signatures is checked.
func (m *machine) EnableInit() error {
	return m.enableStaged(PhaseTransition{InitSigning, Funding})
}

// EnableUpdate promotes the current staging state to the current state.
// A valid phase transition and the existence of all signatures is checked.
func (m *machine) EnableUpdate() error {
	return m.enableStaged(PhaseTransition{Signing, Acting})
}

// EnableFinal promotes the final staging state to the final current state.
// A valid phase transition and the existence of all signatures is checked.
func (m *machine) EnableFinal() error {
	return m.enableStaged(PhaseTransition{Signing, Final})
}

// enableStaged checks that
//   1. the current phase is `expected.From` and
//   2. all signatures of the staging transactions have been set.
// If successful, the staging transaction is promoted to be the current
// transaction. If not, an error is returned.
func (m *machine) enableStaged(expected PhaseTransition) error {
	if err := m.expect(expected); err != nil {
		return errors.WithMessage(err, "no staging phase")
	}
	if (expected.To == Final) != m.stagingTX.State.IsFinal {
		return m.error(expected, "State.IsFinal and target phase don't match")
	}

	for i, sig := range m.stagingTX.Sigs {
		if sig == nil {
			return m.errorf(expected, "signature %d missing from staging TX", i)
		}
	}

	m.prevTXs = append(m.prevTXs, m.currentTX) // push current to previous
	m.currentTX = m.stagingTX                  // promote staging to current
	m.stagingTX = Transaction{}                // clear staging

	m.setPhase(expected.To)
	return nil
}

// SetFunded tells the state machine that the channel got funded and progresses
// to the Acting phase
func (m *machine) SetFunded() error {
	if err := m.expect(PhaseTransition{Funding, Acting}); err != nil {
		return err
	}

	m.setPhase(Acting)
	return nil
}

// SetSettled tells the state machine that the final state was settled on the
// blockchain or funding channel and progresses to the Settled state
func (m *machine) SetSettled() error {
	if err := m.expect(PhaseTransition{Final, Settled}); err != nil {
		return err
	}

	m.setPhase(Settled)
	return nil
}

var validPhaseTransitions = map[PhaseTransition]bool{
	PhaseTransition{InitActing, InitSigning}: true,
	PhaseTransition{InitSigning, Funding}:    true,
	PhaseTransition{Funding, Acting}:         true,
	PhaseTransition{Acting, Signing}:         true,
	PhaseTransition{Signing, Acting}:         true,
	PhaseTransition{Signing, Final}:          true,
	PhaseTransition{Final, Settled}:          true,
}

func (m *machine) expect(tr PhaseTransition) error {
	if m.phase != tr.From {
		return m.error(tr, "not in correct phase")
	}
	if ok := validPhaseTransitions[PhaseTransition{m.phase, tr.To}]; !ok {
		return m.error(tr, "forbidden phase transition")
	}
	return nil
}

// validTransition checks that the transition from the current to the provided
// state is valid. The following checks are run:
// * matching channel ids
// * no transition from final state
// * version increase by 1
// * preservation of balances
// A StateMachine will additionally check the validity of the app-specific
// transition whereas an ActionMachine checks each Action as being valid.
func (m *machine) validTransition(to *State) error {
	if to.ID != m.params.id {
		return errors.New("new state's ID doesn't match")
	}

	if m.currentTX.IsFinal == true {
		return NewStateTransitionError(m.params.id, "cannot advance final state")
	}

	if m.currentTX.Version+1 != to.Version {
		return NewStateTransitionError(m.params.id, "version must increase by one")
	}

	if eq, err := equalSum(m.currentTX.Allocation, to.Allocation); err != nil {
		return err
	} else if !eq {
		return NewStateTransitionError(m.params.id, "allocations must be preserved")
	}

	return nil
}

// Subscribe subscribes go-channel `sub` to phase `phase` under the name `who`.
// If the machine changes into phase `phase`, the phase transition is sent on
// channel `sub`.
// If a subscription for `who` to this phase already exists, it is overwritten.
func (m *machine) Subscribe(phase Phase, who string, sub chan<- PhaseTransition) {
	if m.subs[phase] == nil {
		m.subs[phase] = make(map[string]chan<- PhaseTransition)
	}

	m.subs[phase][who] = sub
}

// notifySubs notifies all subscribers to the current phase that a phase
// transition from the provided phase `from` has happened.
func (m *machine) notifySubs(from Phase) {
	if m.subs[m.phase] == nil {
		// no subscribers
		return
	}

	transition := PhaseTransition{from, m.phase}
	for who, sub := range m.subs[m.phase] {
		m.log.Tracef("phase transition: %v, notifying subscriber %s", transition, who)
		sub <- transition
	}
}

// error constructs a new PhaseTransitionError
func (m *machine) error(expected PhaseTransition, msg string) error {
	return newPhaseTransitionError(m.params.ID(), m.phase, expected, msg)
}

// error constructs a new PhaseTransitionError
func (m *machine) errorf(expected PhaseTransition, format string, args ...interface{}) error {
	return newPhaseTransitionErrorf(m.params.ID(), m.phase, expected, format, args...)
}

// selfTransition returns a PhaseTransition from current to current phase
func (m *machine) selfTransition() PhaseTransition {
	return PhaseTransition{m.phase, m.phase}
}

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
	"fmt"
	stdio "io"

	"github.com/pkg/errors"

	"perun.network/go-perun/log"
	"perun.network/go-perun/pkg/io"
	"perun.network/go-perun/pkg/math/big"
	"perun.network/go-perun/wallet"
)

// Index is the type for the number of participants, assets, sub-allocations, actions and alike.
type Index = uint16

type (
	// Phase is a phase of the channel pushdown automaton.
	Phase uint8

	// PhaseTransition represents a transition between two phases.
	PhaseTransition struct {
		From, To Phase
	}

	// Source is a source of channel data. It allows access to all information
	// needed for persistence. The ID, Idx and Params only need to be persisted
	// once per channel as they stay constant during a channel's lifetime.
	Source interface {
		ID() ID                 // ID is the channel ID of this source. It is the same as Params().ID().
		Idx() Index             // Idx is the own index in the channel.
		Params() *Params        // Params are the channel parameters.
		StagingTX() Transaction // StagingTX is the staged transaction (State+incomplete list of sigs).
		CurrentTX() Transaction // CurrentTX is the current transaction (State+complete list of sigs).
		Phase() Phase           // Phase is the phase in which the channel is currently in.
	}
)

var _ Source = (*machine)(nil)
var _ io.Serializer = (*Phase)(nil)

// Phases known to the channel machine.
const (
	InitActing Phase = iota
	InitSigning
	Funding
	Acting
	Signing
	Final
	Registering
	Registered
	Progressing
	Progressed
	Withdrawing
	Withdrawn
	// LastPhase contains the value of the last phase. This is useful for testing.
	LastPhase = int(Withdrawn)
)

func (p Phase) String() string {
	return [...]string{
		"InitActing",
		"InitSigning",
		"Funding",
		"Acting",
		"Signing",
		"Final",
		"Registering",
		"Registered",
		"Progressing",
		"Progressed",
		"Withdrawing",
		"Withdrawn",
	}[p]
}

// Encode serializes a Phase.
func (p Phase) Encode(w stdio.Writer) error {
	return io.Encode(w, uint8(p))
}

// Decode deserializes a Phase.
func (p *Phase) Decode(r stdio.Reader) error {
	return io.Decode(r, (*uint8)(p))
}

func (t PhaseTransition) String() string {
	return fmt.Sprintf("%v->%v", t.From, t.To)
}

var signingPhases = []Phase{InitSigning, Signing, Progressing}

// A machine is the channel pushdown automaton that handles phase transitions.
// It checks for correct signatures and valid phase transitions.
// It only contains implementations for the phase transitions common to
// both, ActionMachine and StateMachine, that is, AddSig, EnableInit, SetFunded,
// EnableUpdate, EnableFinal and the external phase changes
// Set(Funded|Register(ing|ed)|Withdraw(ing|n)).
// The other transitions are specific to the type of machine and are implemented
// individually.
type machine struct {
	phase     Phase
	acc       wallet.Account `cloneable:"shallow"`
	idx       Index
	params    Params
	stagingTX Transaction
	currentTX Transaction
	prevTXs   []Transaction

	// logger embedding
	log.Embedding
}

// newMachine returns a new uninitialized machine for the given parameters.
func newMachine(acc wallet.Account, params Params) (*machine, error) {
	idx := wallet.IndexOfAddr(params.Parts, acc.Address())
	if idx < 0 {
		return nil, errors.New("account not part of participant set")
	}

	return &machine{
		phase:     InitActing,
		acc:       acc,
		idx:       Index(idx),
		params:    params,
		Embedding: log.MakeEmbedding(log.WithField("ID", params.id)),
	}, nil
}

func restoreMachine(acc wallet.Account, source Source) (*machine, error) {
	m, err := newMachine(acc, *source.Params())
	if err != nil {
		return nil, err
	}
	m.phase = source.Phase()
	m.stagingTX = source.StagingTX()
	m.currentTX = source.CurrentTX()
	return m, nil
}

// ID returns the channel id.
func (m *machine) ID() ID {
	return m.params.ID()
}

// Account returns the account this channel is using for signing state updates.
func (m *machine) Account() wallet.Account {
	return m.acc
}

// Idx returns our index in the channel participants list.
func (m *machine) Idx() Index {
	return m.idx
}

// Params returns the channel parameters.
func (m *machine) Params() *Params {
	return &m.params
}

// N returns the number of participants of the channel parameters of this machine.
func (m *machine) N() Index {
	return Index(len(m.params.Parts))
}

// Phase returns the current phase.
func (m *machine) Phase() Phase {
	return m.phase
}

// setPhase is internally used to set the phase.
func (m *machine) setPhase(p Phase) {
	m.Log().Tracef("phase transition: %v", PhaseTransition{m.phase, p})
	m.phase = p
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
// The signature is calculated and saved to the staging TX's signature slice
// if it was not calculated before.
// A call to Sig only makes sense in a signing phase.
func (m *machine) Sig() (sig wallet.Sig, err error) {
	if !inPhase(m.phase, signingPhases) {
		return nil, m.phaseErrorf(m.selfTransition(), "can only create own signature in a signing phase")
	}

	if m.stagingTX.Sigs[m.idx] == nil {
		sig, err = Sign(m.acc, m.stagingTX.State)
		if err != nil {
			return
		}
		m.stagingTX.Sigs[m.idx] = sig
	} else {
		sig = m.stagingTX.Sigs[m.idx]
	}
	return
}

// State returns the current state.
// Clone the state first if you need to modify it.
func (m *machine) State() *State {
	return m.currentTX.State
}

// CurrentTX returns the current current transaction.
func (m *machine) CurrentTX() Transaction {
	return m.currentTX
}

// AdjudicatorReq returns the adjudicator request for the current channel
// transaction (the current state together with all participants' signatures on
// it).
//
// The Secondary flag is left as false. Set it manually after creating the
// request if you want to use optimized sencondary adjudication logic.
func (m *machine) AdjudicatorReq() AdjudicatorReq {
	return AdjudicatorReq{
		Params: &m.params,
		Acc:    m.acc,
		Idx:    m.idx,
		Tx:     m.currentTX,
	}
}

// StagingState returns the staging state. It should usually be called after
// entering a signing phase to get the new staging state, which might have been
// created during Init() or Update() (for ActionApps).
// Clone the state first if you need to modify it.
func (m *machine) StagingState() *State {
	return m.stagingTX.State
}

// StagingTX returns the current staging transaction.
func (m *machine) StagingTX() Transaction {
	return m.stagingTX
}

// AddSig verifies the provided signature of another participant on the staging
// state and if successful adds it to the staged transaction. It also checks
// whether the signature has already been set and in that case errors.
// It should not happen that a signature of the same participant is set twice.
// It is also checked that the current phase is a signing phase.
// If the index is out of bounds, a panic occurs as this is an invalid usage of
// the machine.
func (m *machine) AddSig(idx Index, sig wallet.Sig) error {
	if !inPhase(m.phase, signingPhases) {
		return m.phaseErrorf(m.selfTransition(), "can only add signature in a signing phase")
	}

	if m.stagingTX.Sigs[idx] != nil {
		return errors.Errorf("signature for idx %d already present (ID: %x)", idx, m.params.id)
	}

	if ok, err := Verify(m.params.Parts[idx], m.stagingTX.State, sig); err != nil {
		return err
	} else if !ok {
		return errors.Errorf("invalid signature for idx %d (ID: %x)", idx, m.params.id)
	}

	m.stagingTX.Sigs[idx] = sig
	return nil
}

// setStaging sets the given phase and state as staging state.
func (m *machine) setStaging(phase Phase, state *State) {
	m.stagingTX = *m.newTransaction(state)
	m.setPhase(phase)
}

// DiscardUpdate discards the current staging transaction and sets the machine's
// phase back to Acting. This method is useful in the case where a valid update
// request is rejected.
func (m *machine) DiscardUpdate() error {
	if err := m.expect(PhaseTransition{Signing, Acting}); err != nil {
		return err
	}

	m.stagingTX = Transaction{} // clear staging tx
	m.setPhase(Acting)
	return nil
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

	// Assert that we transition to phase Final iff state.IsFinal.
	if (expected.To == Final) != m.stagingTX.State.IsFinal {
		return m.phaseErrorf(expected, "State.IsFinal and target phase don't match")
	}

	// Assert that all signatures are present.
	for i, sig := range m.stagingTX.Sigs {
		if sig == nil {
			return m.phaseErrorf(expected, "signature %d missing from staging TX", i)
		}
	}

	m.setPhase(expected.To)
	m.addTx(&m.stagingTX)

	return nil
}

// SetFunded tells the state machine that the channel got funded and progresses
// to the Acting phase.
func (m *machine) SetFunded() error {
	return m.simplePhaseTransition(Funding, Acting)
}

// SetRegistering tells the state machine that the current channel state is
// being registered on the adjudicator. This phase can be reached after the
// initial phases are done, i.e., when there's at least one state with
// signatures.
func (m *machine) SetRegistering() error {
	if m.phase < Funding {
		return m.phaseErrorf(m.selfTransition(), "can only register after init phases")
	}

	m.setPhase(Registering)
	return nil
}

// SetRegistered moves the machine into the Registered phase. The passed event
// gets stored in the machine to record the timeout and registered version.
// This phase can be reached after the initial phases are done, i.e., when
// there's at least one state with signatures.
func (m *machine) SetRegistered() error {
	if m.phase < Funding {
		return m.phaseErrorf(m.selfTransition(), "can only register after init phases")
	}

	m.setPhase(Registered)
	return nil
}

// SetProgressing sets the machine phase to Progressing and the staging state to
// the given state.
func (m *machine) SetProgressing(state *State) error {
	if !inPhase(m.phase, []Phase{Registered, Progressed}) {
		return m.phaseErrorf(m.selfTransition(), "can only progress when registered or progressed")
	}
	m.setStaging(Progressing, state)
	return nil
}

// SetProgressed sets the machine phase to Progressed and the current state to
// the state specified in the given ProgressedEvent.
func (m *machine) SetProgressed(e *ProgressedEvent) error {
	m.forceState(Progressed, e.State)
	return nil
}

// SetWithdrawing sets the state machine to the Withdrawing phase. The current
// state was registered on-chain and funds withdrawal is in progress.
// This phase can only be reached from phase Final, Registered, Progressed, or
// Withdrawing.
func (m *machine) SetWithdrawing() error {
	if !inPhase(m.phase, []Phase{Final, Registered, Progressed, Withdrawing}) {
		return m.phaseErrorf(m.selfTransition(), "can only withdraw after registering")
	}
	m.setPhase(Withdrawing)
	return nil
}

// SetWithdrawn sets the state machine to the final phase Withdrawn. The current
// state was registered on-chain and funds withdrawal was successful.
// This phase can only be reached from the Withdrawing phase.
func (m *machine) SetWithdrawn() error {
	return m.simplePhaseTransition(Withdrawing, Withdrawn)
}

func (m *machine) simplePhaseTransition(from, to Phase) error {
	if err := m.expect(PhaseTransition{from, to}); err != nil {
		return err
	}

	m.setPhase(to)
	return nil
}

var validPhaseTransitions = map[PhaseTransition]struct{}{
	{InitActing, InitSigning}: {},
	{InitSigning, Funding}:    {},
	{Funding, Acting}:         {},
	{Acting, Signing}:         {},
	{Signing, Acting}:         {},
	{Signing, Final}:          {},
	{Funding, Registering}:    {},
	{Acting, Registering}:     {},
	{Signing, Registering}:    {},
	{Final, Registering}:      {},
	{Funding, Registered}:     {},
	{Acting, Registered}:      {},
	{Signing, Registered}:     {},
	{Final, Registered}:       {},
	{Registering, Registered}: {},
	{Registered, Withdrawing}: {},
	{Registered, Progressed}:  {},
	{Progressing, Progressed}: {},
	{Progressed, Withdrawing}: {},
	{Withdrawing, Withdrawn}:  {},
}

func (m *machine) expect(tr PhaseTransition) error {
	if m.phase != tr.From {
		return m.phaseErrorf(tr, "not in correct phase")
	}
	if _, ok := validPhaseTransitions[PhaseTransition{m.phase, tr.To}]; !ok {
		return m.phaseErrorf(tr, "forbidden phase transition")
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

	newError := func(s string) error { return NewStateTransitionError(m.params.id, s) }

	if err := AppShouldEqual(m.params.App, to.App); err != nil {
		return newError(fmt.Sprintf("new state's App doesn't match: %v", err))
	}

	if m.currentTX.IsFinal {
		return newError("cannot advance final state")
	}

	if m.currentTX.Version+1 != to.Version {
		return newError(fmt.Sprintf("expected version %d, got version %d", m.currentTX.Version+1, to.Version))
	}

	if err := to.Allocation.Valid(); err != nil {
		return newError(fmt.Sprintf("invalid allocation: %v", err))
	}

	if eq, err := big.EqualSum(m.currentTX.Allocation, to.Allocation); err != nil {
		return newError(fmt.Sprintf("allocation: %v", err))
	} else if !eq {
		return newError("allocations must be preserved")
	}

	return nil
}

// phaseErrorf constructs a new PhaseTransitionError.
func (m *machine) phaseErrorf(expected PhaseTransition, format string, args ...interface{}) error {
	return newPhaseTransitionErrorf(m.params.ID(), m.phase, expected, format, args...)
}

// selfTransition returns a PhaseTransition from current to current phase.
func (m *machine) selfTransition() PhaseTransition {
	return PhaseTransition{m.phase, m.phase}
}

func (m *machine) Clone() *machine {
	var prevTXs []Transaction
	if m.prevTXs != nil {
		prevTXs = make([]Transaction, len(m.prevTXs))
		for i, tx := range m.prevTXs {
			prevTXs[i] = tx.Clone()
		}
	}

	return &machine{
		phase:     m.phase,
		acc:       m.acc,
		idx:       m.idx,
		params:    *m.params.Clone(),
		stagingTX: m.stagingTX.Clone(),
		currentTX: m.currentTX.Clone(),
		prevTXs:   prevTXs,
		Embedding: m.Embedding,
	}
}

func (m *machine) IsRegistered() bool {
	return m.phase >= Registered
}

func (m *machine) newTransaction(s *State) *Transaction {
	return &Transaction{
		State: s,
		Sigs:  make([]wallet.Sig, m.N()),
	}
}

func (m *machine) addTx(tx *Transaction) {
	m.prevTXs = append(m.prevTXs, m.currentTX) // push current to previous
	m.currentTX = *tx                          // promote staging to current
	m.stagingTX = Transaction{}                // clear staging
}

func (m *machine) forceState(p Phase, s *State) {
	m.setPhase(p)
	m.addTx(m.newTransaction(s))
}

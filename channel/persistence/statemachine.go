// Copyright 2020 - See NOTICE file for copyright holders.
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

package persistence

import (
	"context"

	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"
)

// A StateMachine is a wrapper around a channel.StateMachine that forwards calls
// to it and, if successful, persists changed data using a Persister.
type StateMachine struct {
	*channel.StateMachine
	pr Persister
}

// FromStateMachine creates a persisting StateMachine wrapper around the passed
// StateMachine using the Persister pr.
func FromStateMachine(m *channel.StateMachine, pr Persister) StateMachine {
	return StateMachine{
		StateMachine: m,
		pr:           pr,
	}
}

// SetFunded calls SetFunded on the channel.StateMachine and then persists the
// changed phase.
func (m StateMachine) SetFunded(ctx context.Context) error {
	if err := m.StateMachine.SetFunded(); err != nil {
		return err
	}
	return errors.WithMessage(m.pr.PhaseChanged(ctx, m.StateMachine), "Persister.PhaseChanged")
}

// SetRegistering calls SetRegistering on the channel.StateMachine and then
// persists the changed phase.
func (m StateMachine) SetRegistering(ctx context.Context) error {
	if err := m.StateMachine.SetRegistering(); err != nil {
		return err
	}
	return errors.WithMessage(m.pr.PhaseChanged(ctx, m.StateMachine), "Persister.PhaseChanged")
}

// SetRegistered calls SetRegistered on the channel.StateMachine and then
// persists the changed phase.
func (m StateMachine) SetRegistered(ctx context.Context) error {
	if err := m.StateMachine.SetRegistered(); err != nil {
		return err
	}
	return errors.WithMessage(m.pr.PhaseChanged(ctx, m.StateMachine), "Persister.PhaseChanged")
}

// SetProgressing calls SetProgressing on the channel.StateMachine and then
// persists the changed state.
func (m StateMachine) SetProgressing(ctx context.Context, s *channel.State) error {
	if err := m.StateMachine.SetProgressing(s); err != nil {
		return err
	}
	return errors.WithMessage(m.pr.Staged(ctx, m.StateMachine), "Persister.Staged")
}

// SetProgressed calls SetProgressed on the channel.StateMachine and then
// persists the changed state.
func (m StateMachine) SetProgressed(ctx context.Context, e *channel.ProgressedEvent) error {
	if err := m.StateMachine.SetProgressed(e); err != nil {
		return err
	}
	return errors.WithMessage(m.pr.Enabled(ctx, m.StateMachine), "Persister.Enabled")
}

// SetWithdrawing calls SetWithdrawing on the channel.StateMachine and then
// persists the changed phase.
func (m StateMachine) SetWithdrawing(ctx context.Context) error {
	if err := m.StateMachine.SetWithdrawing(); err != nil {
		return err
	}
	return errors.WithMessage(m.pr.PhaseChanged(ctx, m.StateMachine), "Persister.PhaseChanged")
}

// SetWithdrawn calls SetWithdrawn on the channel.StateMachine and then persists
// the changed phase.
func (m StateMachine) SetWithdrawn(ctx context.Context) error {
	if err := m.StateMachine.SetWithdrawn(); err != nil {
		return err
	}
	return errors.WithMessage(m.pr.ChannelRemoved(ctx, m.ID()), "Persister.ChannelRemoved")
}

// Init calls Init on the channel.StateMachine and then persists the changed
// staging state.
func (m *StateMachine) Init(ctx context.Context, initBals channel.Allocation, initData channel.Data) error {
	if err := m.StateMachine.Init(initBals, initData); err != nil {
		return err
	}
	return errors.WithMessage(m.pr.Staged(ctx, m.StateMachine), "Persister.Staged")
}

// Update calls Update on the channel.StateMachine and then persists the changed
// staging state.
func (m StateMachine) Update(
	ctx context.Context,
	stagingState *channel.State,
	actor channel.Index,
) error {
	if err := m.StateMachine.Update(stagingState, actor); err != nil {
		return err
	}
	return errors.WithMessage(m.pr.Staged(ctx, m.StateMachine), "Persister.Staged")
}

// Sig calls Sig on the channel.StateMachine and then persists the added
// signature.
func (m StateMachine) Sig(ctx context.Context) (sig wallet.Sig, err error) {
	sig, err = m.StateMachine.Sig()
	if err != nil {
		return sig, err
	}
	return sig, errors.WithMessage(m.pr.SigAdded(ctx, m.StateMachine, m.Idx()), "Persister.SigAdded")
}

// AddSig calls AddSig on the channel.StateMachine and then persists the added
// signature.
func (m StateMachine) AddSig(ctx context.Context, idx channel.Index, sig wallet.Sig) error {
	if err := m.StateMachine.AddSig(idx, sig); err != nil {
		return err
	}
	return errors.WithMessage(m.pr.SigAdded(ctx, m.StateMachine, idx), "Persister.SigAdded")
}

// EnableInit calls EnableInit on the channel.StateMachine and then persists the
// enabled transaction.
func (m StateMachine) EnableInit(ctx context.Context) error {
	if err := m.StateMachine.EnableInit(); err != nil {
		return err
	}
	return errors.WithMessage(m.pr.Enabled(ctx, m.StateMachine), "Persister.Enabled")
}

// EnableUpdate calls EnableUpdate on the channel.StateMachine and then persists
// the enabled transaction.
func (m StateMachine) EnableUpdate(ctx context.Context) error {
	if err := m.StateMachine.EnableUpdate(); err != nil {
		return err
	}
	return errors.WithMessage(m.pr.Enabled(ctx, m.StateMachine), "Persister.Enabled")
}

// EnableFinal calls EnableFinal on the channel.StateMachine and then persists
// the enabled transaction.
func (m StateMachine) EnableFinal(ctx context.Context) error {
	if err := m.StateMachine.EnableFinal(); err != nil {
		return err
	}
	return errors.WithMessage(m.pr.Enabled(ctx, m.StateMachine), "Persister.Enabled")
}

// DiscardUpdate calls DiscardUpdate on the channel.StateMachine and then
// removes the state machine's staged state from persistence.
func (m StateMachine) DiscardUpdate(ctx context.Context) error {
	if err := m.StateMachine.DiscardUpdate(); err != nil {
		return err
	}
	return errors.WithMessage(m.pr.Staged(ctx, m.StateMachine), "Persister.Staged")
}

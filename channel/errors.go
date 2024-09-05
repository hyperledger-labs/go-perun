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
	"perun.network/go-perun/wallet"

	"github.com/pkg/errors"
)

type (
	// StateTransitionError happens in case of an invalid channel state transition.
	StateTransitionError struct {
		ID map[wallet.BackendID]ID
	}

	// ActionError happens if an invalid action is applied to a channel state.
	ActionError struct {
		ID map[wallet.BackendID]ID
	}

	// PhaseTransitionError happens in case of an invalid channel machine phase
	// transition.
	PhaseTransitionError struct {
		ID      map[wallet.BackendID]ID
		current Phase
		PhaseTransition
	}
)

func (e *StateTransitionError) Error() string {
	return fmt.Sprintf("invalid state transition (ID: %x)", e.ID)
}

func (e *ActionError) Error() string {
	return fmt.Sprintf("invalid action (ID: %x)", e.ID)
}

func (e *PhaseTransitionError) Error() string {
	return fmt.Sprintf(
		"invalid channel phase transition (ID: %x, current: %v, expected: %v)",
		e.ID, e.current, e.PhaseTransition,
	)
}

// NewStateTransitionError creates a new StateTransitionError.
func NewStateTransitionError(id map[wallet.BackendID]ID, msg string) error {
	return errors.Wrap(&StateTransitionError{
		ID: id,
	}, msg)
}

// NewActionError creates a new ActionError.
func NewActionError(id map[wallet.BackendID]ID, msg string) error {
	return errors.Wrap(&ActionError{
		ID: id,
	}, msg)
}

func newPhaseTransitionError(id map[wallet.BackendID]ID, current Phase, expected PhaseTransition, msg string) error {
	return errors.Wrap(&PhaseTransitionError{
		ID:              id,
		current:         current,
		PhaseTransition: expected,
	}, msg)
}

func newPhaseTransitionErrorf(
	id map[wallet.BackendID]ID,
	current Phase,
	expected PhaseTransition,
	format string,
	args ...interface{},
) error {
	return errors.Wrapf(&PhaseTransitionError{
		ID:              id,
		current:         current,
		PhaseTransition: expected,
	}, format, args...)
}

// IsStateTransitionError returns true if the error was a StateTransitionError.
func IsStateTransitionError(err error) bool {
	cause := errors.Cause(err)
	_, ok := cause.(*StateTransitionError)
	return ok
}

// IsActionError returns true if the error was an ActionError.
func IsActionError(err error) bool {
	cause := errors.Cause(err)
	_, ok := cause.(*ActionError)
	return ok
}

// IsPhaseTransitionError returns true if the error was a PhaseTransitionError.
func IsPhaseTransitionError(err error) bool {
	cause := errors.Cause(err)
	_, ok := cause.(*PhaseTransitionError)
	return ok
}

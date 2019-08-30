// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel

import (
	"fmt"

	"github.com/pkg/errors"
)

type (
	// StateTransitionError happens in case of an invalid channel state transition
	StateTransitionError struct {
		ID ID
	}

	// ActionError happens if an invalid action is applied to a channel state
	ActionError struct {
		ID ID
	}

	// PhaseTransitionError happens in case of an invalid channel machine phase
	// transition
	PhaseTransitionError struct {
		ID      ID
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

func NewStateTransitionError(id ID, msg string) error {
	return errors.Wrap(&StateTransitionError{
		ID: id,
	}, msg)
}

func NewActionError(id ID, msg string) error {
	return errors.Wrap(&ActionError{
		ID: id,
	}, msg)
}

func newPhaseTransitionError(id ID, current Phase, expected PhaseTransition, msg string) error {
	return errors.Wrap(&PhaseTransitionError{
		ID:              id,
		current:         current,
		PhaseTransition: expected,
	}, msg)
}

func newPhaseTransitionErrorf(
	id ID,
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

func IsStateTransitionError(err error) bool {
	cause := errors.Cause(err)
	_, ok := cause.(*StateTransitionError)
	return ok
}

func IsActionError(err error) bool {
	cause := errors.Cause(err)
	_, ok := cause.(*ActionError)
	return ok
}

func IsPhaseTransitionError(err error) bool {
	cause := errors.Cause(err)
	_, ok := cause.(*PhaseTransitionError)
	return ok
}

// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel

import (
	"github.com/pkg/errors"
)

type (
	// StateTransitionError happens in case of an invalid channel state transition
	StateTransitionError struct {
		error
		ID ID
	}

	// PhaseTransitionError happens in case of an invalid channel machine phase
	// transition
	PhaseTransitionError struct {
		error
		ID      ID
		current Phase
		PhaseTransition
	}
)

func newStateTransitionError(id ID, msg string) *StateTransitionError {
	return &StateTransitionError{
		error: errors.New(msg),
		ID:    id,
	}
}

func newPhaseTransitionError(id ID, current Phase, t PhaseTransition, msg string) *PhaseTransitionError {
	return &PhaseTransitionError{
		error:           errors.New(msg),
		ID:              id,
		current:         current,
		PhaseTransition: t,
	}
}

func newPhaseTransitionErrorf(id ID, current Phase, t PhaseTransition, format string, args ...interface{}) *PhaseTransitionError {
	return &PhaseTransitionError{
		error:           errors.Errorf(format, args...),
		ID:              id,
		current:         current,
		PhaseTransition: t,
	}
}

func IsStateTransitionError(err error) bool {
	_, ok := err.(*StateTransitionError)
	return ok
}

func IsPhaseTransitionError(err error) bool {
	_, ok := err.(*PhaseTransitionError)
	return ok
}

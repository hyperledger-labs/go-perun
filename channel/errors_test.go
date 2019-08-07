// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransitionError(t *testing.T) {
	assert.False(t, IsStateTransitionError(errors.New("No StateTransitionError")))
	assert.True(t, IsStateTransitionError(newStateTransitionError(Zero, "A StateTransitionError")))

	assert.False(t, IsPhaseTransitionError(errors.New("No PhaseTransitionError")))
	assert.True(t, IsPhaseTransitionError(newPhaseTransitionError(
		Zero, Nil, PhaseTransition{Nil, Nil}, "A PhaseTransitionError")))
	assert.True(t, IsPhaseTransitionError(newPhaseTransitionErrorf(
		Zero, Nil, PhaseTransition{Nil, Nil}, "A %s", "PhaseTransitionError")))
}

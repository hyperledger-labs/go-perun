// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransitionErrors(t *testing.T) {
	assert.False(t, IsStateTransitionError(errors.New("No StateTransitionError")))
	assert.True(t, IsStateTransitionError(NewStateTransitionError(Zero, "A StateTransitionError")))

	assert.False(t, IsActionError(errors.New("No ActionError")))
	assert.True(t, IsActionError(NewActionError(Zero, "An ActionError")))

	assert.False(t, IsPhaseTransitionError(errors.New("No PhaseTransitionError")))
	assert.True(t, IsPhaseTransitionError(newPhaseTransitionError(
		Zero, InitActing, PhaseTransition{InitActing, InitActing}, "A PhaseTransitionError")))
	assert.True(t, IsPhaseTransitionError(newPhaseTransitionErrorf(
		Zero, InitActing, PhaseTransition{InitActing, InitActing}, "A %s", "PhaseTransitionError")))
}

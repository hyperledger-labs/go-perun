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

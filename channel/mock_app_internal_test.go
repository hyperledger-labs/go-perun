// Copyright 2025 - See NOTICE file for copyright holders.
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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	wiretest "perun.network/go-perun/wire/test"
	pkgtest "polycry.pt/poly-go/test"
)

func TestMockApp(t *testing.T) {
	rng := pkgtest.Prng(t)

	appID := newRandomAppID(rng)
	app := NewMockApp(appID)

	t.Run("App", func(t *testing.T) {
		assert.Equal(t, appID, app.Def())
	})

	t.Run("StateApp", func(t *testing.T) {
		MockStateAppTest(t, *app)
	})

	t.Run("ActionApp", func(t *testing.T) {
		MockActionAppTest(t, *app)
	})

	t.Run("GenericSerializeable", func(t *testing.T) {
		wiretest.GenericMarshalerTest(t, NewMockOp(OpValid))
		wiretest.GenericMarshalerTest(t, NewMockOp(OpErr))
		wiretest.GenericMarshalerTest(t, NewMockOp(OpTransitionErr))
		wiretest.GenericMarshalerTest(t, NewMockOp(OpActionErr))
		wiretest.GenericMarshalerTest(t, NewMockOp(OpPanic))
	})

	// We cant use VerifyClone here since it requires that the same type is returned by
	// Clone() but in this case it returns Data instead of *MockOp
	t.Run("CloneTest", func(t *testing.T) {
		op := NewMockOp(OpValid)
		op2 := op.Clone()
		// Dont use Equal here since it compares the values and not addresses
		assert.False(t, op == op2, "Clone should return a different address")
		assert.Equal(t, op, op2, "Clone should return the same value")
	})
}

func MockStateAppTest(t *testing.T, app MockApp) {
	t.Helper()
	stateValid := createState(OpValid)
	stateErr := createState(OpErr)
	stateTransErr := createState(OpTransitionErr)
	stateActErr := createState(OpActionErr)
	statePanic := createState(OpPanic)

	t.Run("ValidTransition", func(t *testing.T) {
		// ValidTransition only checks the first state.
		require.NoError(t, app.ValidTransition(nil, stateValid, nil, 0))
		assert.Error(t, app.ValidTransition(nil, stateErr, nil, 0))
		assert.True(t, IsStateTransitionError(app.ValidTransition(nil, stateTransErr, nil, 0)))
		assert.True(t, IsActionError(app.ValidTransition(nil, stateActErr, nil, 0)))
		assert.Panics(t, func() { require.NoError(t, app.ValidTransition(nil, statePanic, nil, 0)) })
	})

	t.Run("ValidInit", func(t *testing.T) {
		require.NoError(t, app.ValidInit(nil, stateValid))
		assert.Error(t, app.ValidInit(nil, stateErr))
		assert.True(t, IsStateTransitionError(app.ValidInit(nil, stateTransErr)))
		assert.True(t, IsActionError(app.ValidInit(nil, stateActErr)))
		assert.Panics(t, func() { require.NoError(t, app.ValidInit(nil, statePanic)) })
	})
}

func MockActionAppTest(t *testing.T, app MockApp) {
	t.Helper()
	actValid := NewMockOp(OpValid)
	actErr := NewMockOp(OpErr)
	actTransErr := NewMockOp(OpTransitionErr)
	actActErr := NewMockOp(OpActionErr)
	actPanic := NewMockOp(OpPanic)

	state := createState(OpValid)

	t.Run("InitState", func(t *testing.T) {
		_, _, err := app.InitState(nil, []Action{actValid})
		// Sadly we can not check Allocation.valid() here, since it is private.
		require.NoError(t, err)

		_, _, err = app.InitState(nil, []Action{actErr})
		assert.Error(t, err)

		_, _, err = app.InitState(nil, []Action{actTransErr})
		assert.True(t, IsStateTransitionError(err))

		_, _, err = app.InitState(nil, []Action{actActErr})
		assert.True(t, IsActionError(err))

		assert.Panics(t, func() { app.InitState(nil, []Action{actPanic}) }) //nolint:errcheck
	})

	t.Run("ValidAction", func(t *testing.T) {
		require.NoError(t, app.ValidAction(nil, nil, 0, actValid))
		assert.Error(t, app.ValidAction(nil, nil, 0, actErr))
		assert.True(t, IsStateTransitionError(app.ValidAction(nil, nil, 0, actTransErr)))
		assert.True(t, IsActionError(app.ValidAction(nil, nil, 0, actActErr)))
		assert.Panics(t, func() { app.ValidAction(nil, nil, 0, actPanic) }) //nolint:errcheck
	})

	t.Run("ApplyActions", func(t *testing.T) {
		// ApplyActions increments the Version counter, so we cant pass nil as state.
		retState, err := app.ApplyActions(nil, state, []Action{actValid})
		assert.Equal(t, retState.Version, state.Version+1)
		require.NoError(t, err)

		_, err = app.ApplyActions(nil, state, []Action{actErr})
		assert.Error(t, err)

		_, err = app.ApplyActions(nil, state, []Action{actTransErr})
		assert.True(t, IsStateTransitionError(err))

		_, err = app.ApplyActions(nil, state, []Action{actActErr})
		assert.True(t, IsActionError(err))

		assert.Panics(t, func() { app.ApplyActions(nil, state, []Action{actPanic}) }) //nolint:errcheck
	})
}

func createState(op MockOp) *State {
	return &State{ID: ID{}, Version: 0, Allocation: Allocation{}, Data: NewMockOp(op), IsFinal: false}
}

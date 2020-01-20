// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/pkg/test"
)

func TestAppRandomizerSet(t *testing.T) {
	test.OnlyOnce(t)

	assert.NotNil(t, appRandomizer, "appRandomizer should be default initialized")
	assert.False(t, isAppRandomizerSet, "isAppRandomizerSet should be defaulted to false")

	assert.Panics(t, func() { SetAppRandomizer(nil) }, "nil Randomizer set should panic")
	assert.False(t, isAppRandomizerSet, "isAppRandomizerSet should be false")
	assert.NotNil(t, appRandomizer, "appRandomizer should not be nil")

	old := appRandomizer
	assert.NotPanics(t, func() { SetAppRandomizer(&MockAppRandomizer{}) }, "first SetAppRandomizer() should work")
	assert.True(t, isAppRandomizerSet, "isAppRandomizerSet should be true")
	assert.NotNil(t, appRandomizer, "appRandomizer should not be nil")
	assert.False(t, old == appRandomizer, "appRandomizer should have changed")

	old = appRandomizer
	assert.Panics(t, func() { SetAppRandomizer(&MockAppRandomizer{}) }, "second SetAppRandomizer() should panic")
	assert.True(t, isAppRandomizerSet, "isAppRandomizerSet should be true")
	assert.NotNil(t, appRandomizer, "appRandomizer should not be nil")
	assert.True(t, old == appRandomizer, "appRandomizer should not have changed")
}

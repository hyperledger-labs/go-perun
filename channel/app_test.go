// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package channel

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/pkg/test"
)

func TestAppBackendSet(t *testing.T) {
	test.OnlyOnce(t)

	assert.NotNil(t, appBackend, "appBackend should be default initialized")
	assert.False(t, isAppBackendSet, "isAppBackendSet should be defaulted to false")

	assert.Panics(t, func() { SetAppBackend(nil) }, "nil backend set should panic")
	assert.False(t, isAppBackendSet, "isAppBackendSet should be false")
	assert.NotNil(t, appBackend, "appBackend should not be nil")

	old := appBackend
	assert.NotPanics(t, func() { SetAppBackend(&MockAppBackend{}) }, "first SetAppBackend() should work")
	assert.True(t, isAppBackendSet, "isAppBackendSet should be true")
	assert.NotNil(t, appBackend, "appBackend should not be nil")
	assert.False(t, old == appBackend, "appBackend should have changed")

	old = appBackend
	assert.Panics(t, func() { SetAppBackend(&MockAppBackend{}) }, "second SetAppBackend() should panic")
	assert.True(t, isAppBackendSet, "isAppBackendSet should be true")
	assert.NotNil(t, appBackend, "appBackend should not be nil")
	assert.True(t, old == appBackend, "appBackend should not have changed")
}

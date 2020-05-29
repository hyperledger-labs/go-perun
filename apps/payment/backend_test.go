// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package payment

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "perun.network/go-perun/backend/sim" // backend init
	pkgtest "perun.network/go-perun/pkg/test"
	"perun.network/go-perun/wallet/test"
)

func TestBackend(t *testing.T) {
	pkgtest.OnlyOnce(t)

	rng := rand.New(rand.NewSource(0))
	assert, require := assert.New(t), require.New(t)

	require.NotNil(backend, "init() should have initialized the backend")

	def := test.NewRandomAddress(rng)

	assert.Panics(func() { AppFromDefinition(def) })
	assert.Panics(func() { AppDef() })

	require.NotPanics(func() { SetAppDef(def) })
	assert.Equal(def, AppDef())
	assert.Equal(def, backend.def)
	assert.Panics(func() { AppFromDefinition(nil) })

	app, err := AppFromDefinition(test.NewRandomAddress(rng))
	assert.Error(err)
	assert.Nil(app)

	app, err = AppFromDefinition(def)
	assert.NoError(err)
	require.NotNil(app)
	assert.Equal(&App{def}, app)
}

func TestNoData(t *testing.T) {
	assert := assert.New(t)

	assert.NotPanics(func() {
		data := new(NoData)
		assert.Nil(data.Encode(nil))
	})

	assert.NotPanics(func() {
		app := new(App)
		data, err := app.DecodeData(nil)
		assert.NoError(err)
		assert.NotNil(data)
		assert.IsType(&NoData{}, data)
	})

	data := new(NoData)
	clone := data.Clone()
	assert.IsType(data, clone)
}

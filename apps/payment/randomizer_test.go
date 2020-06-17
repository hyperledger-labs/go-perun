// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package payment

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	_ "perun.network/go-perun/backend/sim" // backend init
	"perun.network/go-perun/wallet/test"
)

func TestRandomizer(t *testing.T) {
	rng := rand.New(rand.NewSource(0))
	if backend.def == nil {
		SetAppDef(test.NewRandomAddress(rng))
		// Reset app def during cleanup in case this test runs before TestBackend,
		// which assumes the app def to not be set yet.
		t.Cleanup(func() { backend.def = nil })
	}

	r := new(Randomizer)
	app := r.NewRandomApp(rng)
	assert.True(t, app.Def().Equals(AppDef()))
	assert.IsType(t, &NoData{}, r.NewRandomData(rng))
}

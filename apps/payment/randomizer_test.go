// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package payment

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	_ "perun.network/go-perun/backend/sim/wallet" // backend init
	"perun.network/go-perun/wallet/test"
)

func TestRandomizer(t *testing.T) {
	rng := rand.New(rand.NewSource(0))
	SetAppDef(test.NewRandomAddress(rng))

	r := new(Randomizer)
	app := r.NewRandomApp(rng)
	assert.True(t, app.Def().Equals(AppDef()))
	assert.IsType(t, &NoData{}, r.NewRandomData(rng))
}

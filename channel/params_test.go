// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package channel_test

import (
	"math/rand"
	"testing"

	"perun.network/go-perun/channel/test"
	pkgtest "perun.network/go-perun/pkg/test"
)

func TestParamsClone(t *testing.T) {
	rng := rand.New(rand.NewSource(0xDDDDD))
	app := test.NewRandomApp(rng)
	params := test.NewRandomParams(rng, app.Def())
	pkgtest.VerifyClone(t, params)
}

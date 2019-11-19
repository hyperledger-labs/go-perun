// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel_test

import (
	"math/rand"
	"testing"

	"perun.network/go-perun/channel/test"
	iotest "perun.network/go-perun/pkg/io/test"

	_ "perun.network/go-perun/backend/sim/channel" // backend init
	_ "perun.network/go-perun/backend/sim/wallet"  // backend init
)

func TestStateSerialization(t *testing.T) {
	rng := rand.New(rand.NewSource(1337))

	app := test.NewRandomApp(rng)
	params := test.NewRandomParams(rng, app)
	state := test.NewRandomState(rng, params)

	iotest.GenericSerializableTest(t, state)
}

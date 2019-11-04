// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package peer

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	sim "perun.network/go-perun/backend/sim/wallet"
	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wire/msg"
)

func init() {
	wallet.SetBackend(new(sim.Backend))
}

func TestAuthResponseMsg(t *testing.T) {
	rng := rand.New(rand.NewSource(1337))
	msg.TestMsg(t, NewAuthResponseMsg(sim.NewRandomAccount(rng)))
}

func TestAuthenticate_NilParams(t *testing.T) {
	rnd := rand.New(rand.NewSource(0xb0ba))
	assert.Panics(t, func() { Authenticate(nil, nil) })
	assert.Panics(t, func() { Authenticate(nil, newMockConn(nil)) })
	assert.Panics(t, func() {
		Authenticate(sim.NewRandomAccount(rnd), nil)
	})
}

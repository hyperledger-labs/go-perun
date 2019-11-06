// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package peer

import (
	"math/rand"
	"sync"
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

func TestExchangeAddrs_NilParams(t *testing.T) {
	rnd := rand.New(rand.NewSource(0xb0ba))
	assert.Panics(t, func() { ExchangeAddrs(nil, nil) })
	assert.Panics(t, func() { ExchangeAddrs(nil, newMockConn(nil)) })
	assert.Panics(t, func() {
		ExchangeAddrs(sim.NewRandomAccount(rnd), nil)
	})
}

func TestExchangeAddrs_Success(t *testing.T) {
	rng := rand.New(rand.NewSource(0xfedd))
	conn0, conn1 := newPipeConnPair()
	defer conn0.Close()
	account0, account1 := sim.NewRandomAccount(rng), sim.NewRandomAccount(rng)
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		defer conn1.Close()

		recvAddr0, err := ExchangeAddrs(account1, conn1)
		assert.NoError(t, err)
		assert.True(t, recvAddr0.Equals(account0.Address()))
	}()

	recvAddr1, err := ExchangeAddrs(account0, conn0)
	assert.NoError(t, err)
	assert.True(t, recvAddr1.Equals(account1.Address()))

	wg.Wait()
}

func TestExchangeAddrs_BogusMsg(t *testing.T) {
	rng := rand.New(rand.NewSource(0xcafe))
	acc := sim.NewRandomAccount(rng)
	conn := newMockConn(nil)
	conn.recvQueue <- msg.NewPingMsg()
	addr, err := ExchangeAddrs(acc, conn)

	assert.Error(t, err, "ExchangeAddrs should error when peer sends a non-AuthResponseMsg")
	assert.Nil(t, addr)
}

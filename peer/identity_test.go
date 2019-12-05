// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package peer

import (
	"context"
	"math/rand"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/pkg/test"
	wallettest "perun.network/go-perun/wallet/test"
	"perun.network/go-perun/wire/msg"
)

func TestAuthResponseMsg(t *testing.T) {
	rng := rand.New(rand.NewSource(1337))
	msg.TestMsg(t, NewAuthResponseMsg(wallettest.NewRandomAccount(rng)))
}

func TestExchangeAddrs_NilParams(t *testing.T) {
	rnd := rand.New(rand.NewSource(0xb0ba))
	assert.Panics(t, func() { ExchangeAddrs(context.Background(), nil, nil) })
	assert.Panics(t, func() { ExchangeAddrs(context.Background(), nil, newMockConn(nil)) })
	assert.Panics(t, func() {
		ExchangeAddrs(context.Background(), wallettest.NewRandomAccount(rnd), nil)
	})
	assert.Panics(t, func() { ExchangeAddrs(nil, wallettest.NewRandomAccount(rnd), newMockConn(nil)) })
}

func TestExchangeAddrs_ConnFail(t *testing.T) {
	rng := rand.New(rand.NewSource(0xDDDDDEDE))
	a, _ := newPipeConnPair()
	a.Close()
	addr, err := ExchangeAddrs(context.Background(), wallettest.NewRandomAccount(rng), a)
	assert.Nil(t, addr)
	assert.Error(t, err)
}

func TestExchangeAddrs_Success(t *testing.T) {
	rng := rand.New(rand.NewSource(0xfedd))
	conn0, conn1 := newPipeConnPair()
	defer conn0.Close()
	account0, account1 := wallettest.NewRandomAccount(rng), wallettest.NewRandomAccount(rng)
	var wg sync.WaitGroup
	wg.Add(1)

	t.Run("remote part", func(t *testing.T) {
		go func() {
			defer wg.Done()
			defer conn1.Close()

			recvAddr0, err := ExchangeAddrs(context.Background(), account1, conn1)
			assert.NoError(t, err)
			assert.True(t, recvAddr0.Equals(account0.Address()))
		}()
	})

	recvAddr1, err := ExchangeAddrs(context.Background(), account0, conn0)
	assert.NoError(t, err)
	assert.True(t, recvAddr1.Equals(account1.Address()))

	wg.Wait()
}

func TestExchangeAddrs_Timeout(t *testing.T) {
	rng := rand.New(rand.NewSource(0xDDDDDeDe))
	a, _ := newPipeConnPair()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	test.AssertTerminates(t, 2*timeout, func() {
		addr, err := ExchangeAddrs(ctx, wallettest.NewRandomAccount(rng), a)
		assert.Nil(t, addr)
		assert.Error(t, err)
	})
}

func TestExchangeAddrs_BogusMsg(t *testing.T) {
	rng := rand.New(rand.NewSource(0xcafe))
	acc := wallettest.NewRandomAccount(rng)
	conn := newMockConn(nil)
	conn.recvQueue <- msg.NewPingMsg()
	addr, err := ExchangeAddrs(context.Background(), acc, conn)

	assert.Error(t, err, "ExchangeAddrs should error when peer sends a non-AuthResponseMsg")
	assert.Nil(t, addr)
}

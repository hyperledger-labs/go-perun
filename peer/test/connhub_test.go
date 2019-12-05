// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package test

import (
	"context"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "perun.network/go-perun/backend/sim/wallet"
	"perun.network/go-perun/pkg/sync"
	"perun.network/go-perun/pkg/test"
	wallettest "perun.network/go-perun/wallet/test"
	"perun.network/go-perun/wire/msg"
)

func TestConnHub_Create(t *testing.T) {
	rng := rand.New(rand.NewSource(0xDDDDDEDE))
	t.Run("create and dial existing", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)

		var c ConnHub
		addr := wallettest.NewRandomAddress(rng)
		d, l := c.NewDialer(), c.NewListener(addr)
		assert.NotNil(d)
		assert.NotNil(l)

		t.Run("accept", func(t *testing.T) {
			go test.AssertTerminates(t, timeout, func() {
				conn, err := l.Accept()
				assert.NoError(err)
				require.NotNil(t, conn)
				assert.NoError(conn.Send(msg.NewPingMsg()))
			})
		})

		test.AssertTerminates(t, timeout, func() {
			conn, err := d.Dial(context.Background(), addr)
			assert.NoError(err)
			require.NotNil(conn)
			m, err := conn.Recv()
			assert.NoError(err)
			assert.IsType(msg.NewPingMsg(), m)
		})
	})

	t.Run("double create", func(t *testing.T) {
		assert := assert.New(t)

		var c ConnHub
		addr := wallettest.NewRandomAddress(rng)

		l := c.NewListener(addr)
		assert.NotNil(l)

		assert.Panics(func() { c.NewListener(addr) })
	})

	t.Run("dial nonexisting", func(t *testing.T) {
		assert := assert.New(t)

		var c ConnHub

		d := c.NewDialer()
		test.AssertTerminates(t, timeout, func() {
			conn, err := d.Dial(context.Background(), wallettest.NewRandomAddress(rng))
			assert.Nil(conn)
			assert.Error(err)
		})
	})

	t.Run("closed create", func(t *testing.T) {
		assert := assert.New(t)

		var c ConnHub
		c.Close()
		addr := wallettest.NewRandomAddress(rng)

		assert.Panics(func() { c.NewDialer() })
		assert.Panics(func() { c.NewListener(addr) })
	})
}

func TestConnHub_Close(t *testing.T) {
	rng := rand.New(rand.NewSource(0xDDDDDEDE))
	t.Run("nonempty close", func(t *testing.T) {
		assert := assert.New(t)

		var c ConnHub
		l := c.NewListener(wallettest.NewRandomAddress(rng))
		assert.NoError(c.Close())
		assert.True(l.(*Listener).IsClosed())
	})

	t.Run("nonempty close with error (listener)", func(t *testing.T) {
		assert := assert.New(t)

		var c ConnHub
		l := c.NewListener(wallettest.NewRandomAddress(rng))
		l2 := NewListener()
		l2.Close()
		c.insert(wallettest.NewRandomAccount(rng).Address(), l2)
		assert.Error(c.Close())
		assert.True(l.(*Listener).IsClosed())
	})

	t.Run("nonempty close with error (dialer)", func(t *testing.T) {
		assert := assert.New(t)

		var c ConnHub
		d := c.NewDialer()
		d2 := &Dialer{}
		d2.Close()
		c.dialers.insert(d2)
		assert.Error(c.Close())
		assert.True(d.(*Dialer).IsClosed())
	})

	t.Run("double close", func(t *testing.T) {
		assert := assert.New(t)

		var c ConnHub
		assert.NoError(c.Close())
		err := c.Close()
		assert.Error(err)
		assert.True(sync.IsAlreadyClosedError(err))
	})
}

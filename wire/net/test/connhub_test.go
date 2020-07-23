// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package test

import (
	"context"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "perun.network/go-perun/backend/sim" // backend init
	"perun.network/go-perun/pkg/sync"
	"perun.network/go-perun/pkg/test"
	wallettest "perun.network/go-perun/wallet/test"
	"perun.network/go-perun/wire"
	wiretest "perun.network/go-perun/wire/test"
)

func TestConnHub_Create(t *testing.T) {
	rng := rand.New(rand.NewSource(0xDDDDDEDE))
	t.Run("create and dial existing", func(t *testing.T) {
		assert := assert.New(t)

		var c ConnHub
		addr := wallettest.NewRandomAddress(rng)
		d, l := c.NewNetDialer(), c.NewNetListener(addr)
		assert.NotNil(d)
		assert.NotNil(l)

		ct := test.NewConcurrent(t)
		go test.AssertTerminates(t, timeout, func() {
			ct.Stage("accept", func(rt require.TestingT) {
				conn, err := l.Accept()
				assert.NoError(err)
				require.NotNil(rt, conn)
				assert.NoError(conn.Send(wiretest.NewRandomEnvelope(rng, wire.NewPingMsg())))
			})
		})

		test.AssertTerminates(t, timeout, func() {
			ct.Stage("dial", func(rt require.TestingT) {
				conn, err := d.Dial(context.Background(), addr)
				assert.NoError(err)
				require.NotNil(rt, conn)
				m, err := conn.Recv()
				assert.NoError(err)
				assert.IsType(wire.NewPingMsg(), m.Msg)
			})
		})

		ct.Wait("accept", "dial")
	})

	t.Run("double create", func(t *testing.T) {
		assert := assert.New(t)

		var c ConnHub
		addr := wallettest.NewRandomAddress(rng)

		l := c.NewNetListener(addr)
		assert.NotNil(l)

		assert.Panics(func() { c.NewNetListener(addr) })
	})

	t.Run("dial nonexisting", func(t *testing.T) {
		assert := assert.New(t)

		var c ConnHub

		d := c.NewNetDialer()
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

		assert.Panics(func() { c.NewNetDialer() })
		assert.Panics(func() { c.NewNetListener(addr) })
	})
}

func TestConnHub_Close(t *testing.T) {
	rng := rand.New(rand.NewSource(0xDDDDDEDE))
	t.Run("nonempty close", func(t *testing.T) {
		assert := assert.New(t)

		var c ConnHub
		l := c.NewNetListener(wallettest.NewRandomAddress(rng))
		assert.NoError(c.Close())
		assert.True(l.IsClosed())
	})

	t.Run("nonempty close with error (listener)", func(t *testing.T) {
		assert := assert.New(t)

		var c ConnHub
		l := c.NewNetListener(wallettest.NewRandomAddress(rng))
		l2 := NewNetListener()
		l2.Close()
		c.insert(wallettest.NewRandomAccount(rng).Address(), l2)
		assert.Error(c.Close())
		assert.True(l.IsClosed())
	})

	t.Run("nonempty close with error (dialer)", func(t *testing.T) {
		assert := assert.New(t)

		var c ConnHub
		d := c.NewNetDialer()
		d2 := &Dialer{}
		d2.Close()
		c.dialers.insert(d2)
		assert.Error(c.Close())
		assert.True(d.IsClosed())
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

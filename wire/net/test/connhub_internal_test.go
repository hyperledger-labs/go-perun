// Copyright 2019 - See NOTICE file for copyright holders.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "perun.network/go-perun/backend/sim" // backend init
	"perun.network/go-perun/wire"
	perunio "perun.network/go-perun/wire/perunio/serializer"
	wiretest "perun.network/go-perun/wire/test"
	ctxtest "polycry.pt/poly-go/context/test"
	"polycry.pt/poly-go/sync"
	pkgtest "polycry.pt/poly-go/test"
)

func TestConnHub_Create(t *testing.T) {
	rng := pkgtest.Prng(t)
	ser := perunio.Serializer()
	t.Run("create and dial existing", func(t *testing.T) {
		assert := assert.New(t)

		var c ConnHub
		addr := wiretest.NewRandomAddress(rng)
		d, l := c.NewNetDialer(), c.NewNetListener(addr)
		assert.NotNil(d)
		assert.NotNil(l)

		ct := pkgtest.NewConcurrent(t)
		go ctxtest.AssertTerminates(t, timeout, func() {
			ct.Stage("accept", func(rt pkgtest.ConcT) {
				conn, err := l.Accept(ser)
				require.NoError(t, err)                                                           //nolint: testifylint
				require.NotNil(rt, conn)                                                          //nolint: testifylint
				require.NoError(t, conn.Send(wiretest.NewRandomEnvelope(rng, wire.NewPingMsg()))) //nolint: testifylint
			})
		})

		ctxtest.AssertTerminates(t, timeout, func() {
			ct.Stage("dial", func(rt pkgtest.ConcT) {
				conn, err := d.Dial(context.Background(), addr, ser)
				require.NoError(t, err)
				require.NotNil(rt, conn)
				m, err := conn.Recv()
				require.NoError(t, err)
				assert.IsType(wire.NewPingMsg(), m.Msg)
			})
		})

		ct.Wait("accept", "dial")
	})

	t.Run("double create", func(t *testing.T) {
		assert := assert.New(t)

		var c ConnHub
		addr := wiretest.NewRandomAddress(rng)

		l := c.NewNetListener(addr)
		assert.NotNil(l)

		assert.Panics(func() { c.NewNetListener(addr) })
	})

	t.Run("dial nonexisting", func(t *testing.T) {
		assert := assert.New(t)

		var c ConnHub

		d := c.NewNetDialer()
		ctxtest.AssertTerminates(t, timeout, func() {
			conn, err := d.Dial(context.Background(), wiretest.NewRandomAddress(rng), ser)
			assert.Nil(conn)
			require.Error(t, err)
		})
	})

	t.Run("closed create", func(t *testing.T) {
		assert := assert.New(t)

		var c ConnHub
		c.Close()
		addr := wiretest.NewRandomAddress(rng)

		assert.Panics(func() { c.NewNetDialer() })
		assert.Panics(func() { c.NewNetListener(addr) })
	})
}

func TestConnHub_Close(t *testing.T) {
	rng := pkgtest.Prng(t)
	t.Run("nonempty close", func(t *testing.T) {
		assert := assert.New(t)

		var c ConnHub
		l := c.NewNetListener(wiretest.NewRandomAddress(rng))
		require.NoError(t, c.Close())
		assert.True(l.IsClosed())
	})

	t.Run("nonempty close with error (listener)", func(t *testing.T) {
		assert := assert.New(t)

		var c ConnHub
		l := c.NewNetListener(wiretest.NewRandomAddress(rng))
		l2 := NewNetListener()
		l2.Close()
		err := c.insert(wiretest.NewRandomAccount(rng).Address(), l2)
		require.NoError(t, err)
		require.Error(t, c.Close())
		assert.True(l.IsClosed())
	})

	t.Run("nonempty close with error (dialer)", func(t *testing.T) {
		assert := assert.New(t)

		var c ConnHub
		d := c.NewNetDialer()
		d2 := &Dialer{}
		d2.Close()
		c.dialers.insert(d2)
		require.Error(t, c.Close())
		assert.True(d.IsClosed())
	})

	t.Run("double close", func(t *testing.T) {
		assert := assert.New(t)

		var c ConnHub
		require.NoError(t, c.Close())
		err := c.Close()
		require.Error(t, err)
		assert.True(sync.IsAlreadyClosedError(err))
	})
}

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

	"perun.network/go-perun/backend/sim/wallet"
	"perun.network/go-perun/peer"
	"perun.network/go-perun/pkg/test"
	"perun.network/go-perun/wire/msg"
)

func TestConnHub_Create(t *testing.T) {
	rng := rand.New(rand.NewSource(0xDDDDDEDE))
	t.Run("create and dial existing", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)

		var c ConnHub
		id := wallet.NewRandomAccount(rng)
		d, l, err := c.Create(id)
		assert.NotNil(d)
		assert.NotNil(l)
		assert.NoError(err)

		t.Run("accept", func(t *testing.T) {
			go test.AssertTerminates(t, timeout, func() {
				conn, err := l.Accept()
				assert.NoError(err)
				require.NotNil(t, conn)
				_, err = peer.ExchangeAddrs(id, conn)
				assert.NoError(err)
				assert.NoError(conn.Send(msg.NewPingMsg()))
			})
		})

		test.AssertTerminates(t, timeout, func() {
			conn, err := d.Dial(context.Background(), id.Address())
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
		id := wallet.NewRandomAccount(rng)

		d, l, err := c.Create(id)
		assert.NotNil(d)
		assert.NotNil(l)
		assert.NoError(err)

		d, l, err = c.Create(id)
		assert.Nil(d)
		assert.Nil(l)
		assert.Error(err)
	})

	t.Run("dial nonexisting", func(t *testing.T) {
		assert := assert.New(t)

		var c ConnHub
		id := wallet.NewRandomAccount(rng)

		d, _, _ := c.Create(id)
		test.AssertTerminates(t, timeout, func() {
			conn, err := d.Dial(context.Background(), wallet.NewRandomAddress(rng))
			assert.Nil(conn)
			assert.Error(err)
		})
	})
}

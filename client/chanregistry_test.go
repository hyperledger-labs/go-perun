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

package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/channel/test"
	"perun.network/go-perun/wire"
	pkgtest "polycry.pt/poly-go/test"
)

func testCh() *Channel {
	r := wire.NewRelay()
	conn := &channelConn{OnCloser: r, r: r}
	return &Channel{OnCloser: conn, conn: conn}
}

func TestChanRegistry_Put(t *testing.T) {
	rng := pkgtest.Prng(t)
	ch := testCh()
	id := test.NewRandomChannelID(rng)

	t.Run("callback", func(t *testing.T) {
		called := false
		r := makeChanRegistry()
		r.OnNewChannel(func(c *Channel) {
			called = true
			require.Same(t, c, ch)
		})
		assert.True(t, r.Put(id, ch))
		require.True(t, called)
	})

	t.Run("single insert", func(t *testing.T) {
		r := makeChanRegistry()
		assert.True(t, r.Put(id, ch))
		c, ok := r.Get(id)
		require.True(t, ok)
		require.Same(t, c, ch)
	})

	t.Run("double insert", func(t *testing.T) {
		r := makeChanRegistry()
		require.True(t, r.Put(id, ch))
		assert.False(t, r.Put(id, ch))
	})
}

func TestChanRegistry_Has(t *testing.T) {
	rng := pkgtest.Prng(t)
	ch := testCh()
	id := test.NewRandomChannelID(rng)

	t.Run("nonexistent has", func(t *testing.T) {
		r := makeChanRegistry()
		assert.False(t, r.Has(id))
	})

	t.Run("existing has", func(t *testing.T) {
		r := makeChanRegistry()
		require.True(t, r.Put(id, ch))
		assert.True(t, r.Has(id))
	})
}

func TestChanRegistry_Get(t *testing.T) {
	rng := pkgtest.Prng(t)
	ch := testCh()
	id := test.NewRandomChannelID(rng)

	t.Run("nonexistent get", func(t *testing.T) {
		r := makeChanRegistry()
		c, ok := r.Get(id)
		assert.False(t, ok)
		assert.Nil(t, c)
	})

	t.Run("existing get", func(t *testing.T) {
		r := makeChanRegistry()
		require.True(t, r.Put(id, ch))
		c, ok := r.Get(id)
		assert.True(t, ok)
		assert.Same(t, c, ch)
	})
}

func TestChanRegistry_Delete(t *testing.T) {
	rng := pkgtest.Prng(t)
	ch := testCh()
	id := test.NewRandomChannelID(rng)

	t.Run("nonexistent delete", func(t *testing.T) {
		r := makeChanRegistry()
		assert.False(t, r.Delete(id))
		require.False(t, r.Has(id))
	})

	t.Run("existing delete", func(t *testing.T) {
		r := makeChanRegistry()
		require.True(t, r.Put(id, ch))
		assert.True(t, r.Delete(id))
		assert.False(t, r.Has(id))
	})

	t.Run("double delete", func(t *testing.T) {
		r := makeChanRegistry()
		require.True(t, r.Put(id, ch))
		require.True(t, r.Delete(id))
		assert.False(t, r.Delete(id))
		assert.False(t, r.Has(id))
	})
}

func TestChanRegistry_CloseAll(t *testing.T) {
	rng := pkgtest.Prng(t)
	id := test.NewRandomChannelID(rng)

	ch := testCh()
	reg := makeChanRegistry()
	reg.Put(id, ch)
	reg.CloseAll()
	assert.True(t, ch.IsClosed())
}

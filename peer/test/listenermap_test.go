// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/wallet/test"
)

func TestListenerMap_find(t *testing.T) {
	rng := rand.New(rand.NewSource(0xb0bafedd))

	t.Run("empty map", func(t *testing.T) {
		m := listenerMap{}
		l, ok := m.find(test.NewRandomAddress(rng))
		assert.Nil(t, l)
		assert.False(t, ok)
	})

	t.Run("map with entry", func(t *testing.T) {
		key := test.NewRandomAddress(rng)
		listener := NewListener()
		entry := listenerMapEntry{key, listener}
		m := listenerMap{entries: []listenerMapEntry{entry}}

		l, ok := m.find(key)
		assert.Same(t, l, listener)
		assert.True(t, ok)

		l, ok = m.find(test.NewRandomAddress(rng))
		assert.Nil(t, l)
		assert.False(t, ok)
	})
}

func TestListenerMap_insert(t *testing.T) {
	rng := rand.New(rand.NewSource(0xb0bafedd))

	t.Run("insert new", func(t *testing.T) {
		m := listenerMap{}
		for i := 0; i < 10; i++ {
			key := test.NewRandomAddress(rng)
			assert.NoError(t, m.insert(key, NewListener()))
			_, ok := m.find(key)
			assert.True(t, ok)
		}
	})

	t.Run("double insert", func(t *testing.T) {
		m := listenerMap{}
		key := test.NewRandomAddress(rng)
		assert.NoError(t, m.insert(key, NewListener()))
		assert.Error(t, m.insert(key, NewListener()))
	})
}

func TestListenerMap_erase(t *testing.T) {
	rng := rand.New(rand.NewSource(0xb0bafedd))

	t.Run("erase existing", func(t *testing.T) {
		m := listenerMap{}
		for i := 0; i < 10; i++ {
			key := test.NewRandomAddress(rng)
			assert.NoError(t, m.insert(key, NewListener()))
			assert.NoError(t, m.erase(key))
			_, ok := m.find(key)
			assert.False(t, ok)
		}
	})

	t.Run("erase nonexistent", func(t *testing.T) {
		m := listenerMap{}
		assert.Error(t, m.erase(test.NewRandomAddress(rng)))
	})
}

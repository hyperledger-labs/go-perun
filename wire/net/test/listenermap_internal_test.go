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
	"testing"

	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/wire/test"
	pkgtest "polycry.pt/poly-go/test"
)

func TestListenerMap_find(t *testing.T) {
	rng := pkgtest.Prng(t)

	t.Run("empty map", func(t *testing.T) {
		m := listenerMap{}
		l, ok := m.find(test.NewRandomAddress(rng))
		assert.Nil(t, l)
		assert.False(t, ok)
	})

	t.Run("map with entry", func(t *testing.T) {
		key := test.NewRandomAddress(rng)
		listener := NewNetListener()
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
	rng := pkgtest.Prng(t)

	t.Run("insert new", func(t *testing.T) {
		m := listenerMap{}
		for i := 0; i < 10; i++ {
			key := test.NewRandomAddress(rng)
			assert.NoError(t, m.insert(key, NewNetListener()))
			_, ok := m.find(key)
			assert.True(t, ok)
		}
	})

	t.Run("double insert", func(t *testing.T) {
		m := listenerMap{}
		key := test.NewRandomAddress(rng)
		assert.NoError(t, m.insert(key, NewNetListener()))
		assert.Error(t, m.insert(key, NewNetListener()))
	})
}

func TestListenerMap_erase(t *testing.T) {
	rng := pkgtest.Prng(t)

	t.Run("erase existing", func(t *testing.T) {
		m := listenerMap{}
		for i := 0; i < 10; i++ {
			key := test.NewRandomAddress(rng)
			assert.NoError(t, m.insert(key, NewNetListener()))
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

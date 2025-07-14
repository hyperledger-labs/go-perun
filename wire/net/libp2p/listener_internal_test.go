// Copyright 2025 - See NOTICE file for copyright holders.
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

package libp2p

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	perunio "perun.network/go-perun/wire/perunio/serializer"

	ctxtest "polycry.pt/poly-go/context/test"
	pkgtest "polycry.pt/poly-go/test"
)

func TestNewListener(t *testing.T) {
	rng := pkgtest.Prng(t)
	acc := NewRandomAccount(rng)
	defer func() {
		assert.NoError(t, acc.Close())
	}()

	l := NewP2PListener(acc)
	defer l.Close()
	assert.NotNil(t, l)
}

func TestListener_Close(t *testing.T) {
	t.Run("double close", func(t *testing.T) {
		rng := pkgtest.Prng(t)
		acc := NewRandomAccount(rng)

		defer func() {
			assert.NoError(t, acc.Close())
		}()

		l := NewP2PListener(acc)
		assert.NoError(t, l.Close(), "first close must not return error")
		assert.Error(t, l.Close(), "second close must result in error")
	})
}

func TestListener_Accept(t *testing.T) {
	// Happy case already tested in TestDialer_Dial.
	rng := pkgtest.Prng(t)
	acc := NewRandomAccount(rng)

	defer func() {
		assert.NoError(t, acc.Close())
	}()

	timeout := 100 * time.Millisecond

	t.Run("timeout", func(t *testing.T) {
		l := NewP2PListener(acc)
		defer l.Close()

		ctxtest.AssertNotTerminates(t, timeout, func() {
			_, err := l.Accept(perunio.Serializer())
			assert.Error(t, err)
		})
	})

	t.Run("closed", func(t *testing.T) {
		l := NewP2PListener(acc)
		l.Close()

		ctxtest.AssertTerminates(t, timeout, func() {
			conn, err := l.Accept(perunio.Serializer())
			assert.Nil(t, conn)
			assert.Error(t, err)
		})
	})
}

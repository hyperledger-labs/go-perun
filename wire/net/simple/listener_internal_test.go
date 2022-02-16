// Copyright 2020 - See NOTICE file for copyright holders.
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

package simple

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	perunio "perun.network/go-perun/wire/perunio/serializer"

	"polycry.pt/poly-go/context/test"
)

const addr = "0.0.0.0:1337"

func TestNewTCPListener(t *testing.T) {
	l, err := NewTCPListener(addr)
	require.NoError(t, err)
	defer l.Close()
}

func TestNewUnixListener(t *testing.T) {
	l, err := NewUnixListener(addr)
	require.NoError(t, err)
	defer l.Close()
}

func TestListener_Close(t *testing.T) {
	t.Run("double close", func(t *testing.T) {
		l, err := NewTCPListener(addr)
		require.NoError(t, err)
		assert.NoError(t, l.Close(), "first close must not return error")
		assert.Error(t, l.Close(), "second close must result in error")
	})
}

func TestNewListener(t *testing.T) {
	t.Run("happy", func(t *testing.T) {
		l, err := NewTCPListener(addr)
		assert.NoError(t, err)
		require.NotNil(t, l)
		l.Close()
	})

	t.Run("sad", func(t *testing.T) {
		l, err := NewTCPListener("not an address")
		assert.Error(t, err)
		assert.Nil(t, l)
	})

	t.Run("address in use", func(t *testing.T) {
		l, err := NewTCPListener(addr)
		require.NoError(t, err)
		_, err = NewTCPListener(addr)
		require.Error(t, err)
		l.Close()
	})
}

func TestListener_Accept(t *testing.T) {
	// Happy case already tested in TestDialer_Dial.

	ser := perunio.Serializer()
	timeout := 100 * time.Millisecond
	t.Run("timeout", func(t *testing.T) {
		l, err := NewTCPListener(addr)
		require.NoError(t, err)
		defer l.Close()

		test.AssertNotTerminates(t, timeout, func() {
			l.Accept(ser) //nolint:errcheck
		})
	})

	t.Run("closed", func(t *testing.T) {
		l, err := NewTCPListener(addr)
		require.NoError(t, err)
		l.Close()

		test.AssertTerminates(t, timeout, func() {
			conn, err := l.Accept(ser)
			assert.Nil(t, conn)
			assert.Error(t, err)
		})
	})
}

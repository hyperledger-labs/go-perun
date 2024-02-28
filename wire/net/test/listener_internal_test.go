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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/wire"
	wirenet "perun.network/go-perun/wire/net"
	perunio "perun.network/go-perun/wire/perunio/serializer"
	ctxtest "polycry.pt/poly-go/context/test"
)

type fakeConn struct{}

func (fakeConn) Send(*wire.Envelope) error     { panic("") }
func (fakeConn) Recv() (*wire.Envelope, error) { panic("") }
func (fakeConn) Close() error                  { panic("") }

// A valid connection needed to check that accept will pass along values
// properly.
var connection wirenet.Conn = new(fakeConn)

const timeout = 100 * time.Millisecond

func TestListener_Accept_Put(t *testing.T) {
	t.Parallel()

	l := NewNetListener()
	done := make(chan struct{})
	go func() {
		defer close(done)

		ctxtest.AssertTerminates(t, timeout, func() {
			conn, err := l.Accept(perunio.Serializer())
			require.NoError(t, err, "Accept must not fail") //nolint:testifylint
			assert.Same(t, connection, conn,
				"Accept must receive connection from Put")
			assert.Equal(t, 1, l.NumAccepted(),
				"Accept must track accepted connections")
		})
	}()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	assert.True(t, l.Put(ctx, connection))
	// there is no select with `time.After()` branch here because the goroutine
	// calls `ctxtest.AssertTerminates`
	<-done
}

func TestListener_Accept_Close(t *testing.T) {
	t.Parallel()
	ser := perunio.Serializer()

	t.Run("close before accept", func(t *testing.T) {
		l := NewNetListener()
		l.Close()
		ctxtest.AssertTerminates(t, timeout, func() {
			conn, err := l.Accept(ser)
			require.Error(t, err, "Accept must fail")
			assert.Nil(t, conn)
			assert.Zero(t, l.NumAccepted())
		})
	})
	t.Run("close during accept", func(t *testing.T) {
		l := NewNetListener()

		go func() {
			<-time.After(timeout)
			l.Close()
		}()

		ctxtest.AssertTerminates(t, 2*timeout, func() {
			conn, err := l.Accept(ser)
			require.Error(t, err, "Accept must fail")
			assert.Nil(t, conn)
			assert.Zero(t, l.NumAccepted())
		})
	})
}

func TestListener_Put(t *testing.T) {
	t.Parallel()

	t.Run("blocking", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		ctxtest.AssertTerminates(t, timeout, func() {
			assert.False(t, NewNetListener().Put(ctx, connection))
		})
	})

	t.Run("close", func(t *testing.T) {
		t.Parallel()

		l := NewNetListener()
		l.Close()
		ctxtest.AssertTerminates(t, timeout, func() {
			// Closed listener must abort Put() calls.
			assert.False(t, l.Put(context.Background(), connection))
			// Accept() must always fail when closed.
			conn, err := l.Accept(perunio.Serializer())
			assert.Nil(t, conn)
			require.Error(t, err)
			assert.Zero(t, l.NumAccepted())
		})
	})
}

func TestListener_Close(t *testing.T) {
	l := NewNetListener()
	assert.False(t, l.IsClosed())
	require.NoError(t, l.Close())
	assert.True(t, l.IsClosed())
	require.Error(t, l.Close())
}

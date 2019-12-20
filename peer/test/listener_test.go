// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/peer"
	"perun.network/go-perun/pkg/test"
	"perun.network/go-perun/wire/msg"
)

type fakeConn struct{}

func (fakeConn) Send(msg.Msg) error     { panic("") }
func (fakeConn) Recv() (msg.Msg, error) { panic("") }
func (fakeConn) Close() error           { panic("") }

// A valid connection needed to check that accept will pass along values
// properly.
var connection peer.Conn = new(fakeConn)

const timeout = 100 * time.Millisecond

func TestListener_Accept_Put(t *testing.T) {
	t.Parallel()

	l := NewListener()
	done := make(chan struct{})
	go func() {
		defer close(done)

		test.AssertTerminates(t, timeout, func() {
			conn, err := l.Accept()
			assert.NoError(t, err, "Accept must not fail")
			assert.Same(t, connection, conn,
				"Accept must receive connection from Put")
			assert.Equal(t, 1, l.NumAccepted(),
				"Accept must track accepted connections")
		})
	}()

	test.AssertTerminates(t, timeout, func() {
		assert.True(t, l.Put(context.Background(), connection))
	})
	// there is no select with `time.After()` branch here because the goroutine
	// calls `test.AssertTerminates`
	<-done
}

func TestListener_Accept_Close(t *testing.T) {
	t.Parallel()

	t.Run("close before accept", func(t *testing.T) {
		l := NewListener()
		l.Close()
		test.AssertTerminates(t, timeout, func() {
			conn, err := l.Accept()
			assert.Error(t, err, "Accept must fail")
			assert.Nil(t, conn)
			assert.Zero(t, l.NumAccepted())
		})
	})
	t.Run("close during accept", func(t *testing.T) {
		l := NewListener()

		go func() {
			<-time.After(timeout)
			l.Close()
		}()

		test.AssertTerminates(t, 2*timeout, func() {
			conn, err := l.Accept()
			assert.Error(t, err, "Accept must fail")
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
		test.AssertTerminates(t, timeout, func() {
			assert.False(t, NewListener().Put(ctx, connection))
		})
	})

	t.Run("close", func(t *testing.T) {
		t.Parallel()

		l := NewListener()
		l.Close()
		test.AssertTerminates(t, timeout, func() {
			// Closed listener must abort Put() calls.
			assert.False(t, l.Put(context.Background(), connection))
			// Accept() must always fail when closed.
			conn, err := l.Accept()
			assert.Nil(t, conn)
			assert.Error(t, err)
			assert.Zero(t, l.NumAccepted())
		})
	})
}

func TestListener_Close(t *testing.T) {
	l := NewListener()
	assert.False(t, l.IsClosed())
	assert.NoError(t, l.Close())
	assert.True(t, l.IsClosed())
	assert.Error(t, l.Close())
}

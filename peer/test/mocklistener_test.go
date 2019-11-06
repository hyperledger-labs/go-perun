// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package test

import (
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

func TestMockListener_Accept_Put(t *testing.T) {
	t.Parallel()

	l := NewMockListener()
	t.Run("accept", func(t *testing.T) {
		t.Parallel()
		test.AssertTerminates(t, timeout, func() {
			conn, err := l.Accept()
			assert.NoError(t, err, "Accept must not fail")
			assert.Same(t, connection, conn,
				"Accept must receive connection from Put")
			assert.Equal(t, 1, l.NumAccepted(),
				"Accept must track accepted connections")
		})
	})
	t.Run("put", func(t *testing.T) {
		t.Parallel()
		test.AssertTerminates(t, timeout, func() {
			l.Put(connection)
		})
	})
}

func TestMockListener_Accept_Close(t *testing.T) {
	t.Parallel()

	t.Run("close before accept", func(t *testing.T) {
		l := NewMockListener()
		l.Close()
		test.AssertTerminates(t, timeout, func() {
			conn, err := l.Accept()
			assert.Error(t, err, "Accept must fail")
			assert.Nil(t, conn)
			assert.Zero(t, l.NumAccepted())
		})
	})
	t.Run("close during accept", func(t *testing.T) {
		l := NewMockListener()

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

func TestMockListener_Put(t *testing.T) {
	t.Parallel()

	t.Run("blocking", func(t *testing.T) {
		t.Parallel()

		test.AssertNotTerminates(t, timeout, func() {
			NewMockListener().Put(connection)
		})
	})

	t.Run("close", func(t *testing.T) {
		t.Parallel()

		l := NewMockListener()
		l.Close()
		test.AssertTerminates(t, timeout, func() {
			// Closed listener must abort Put() calls.
			l.Put(connection)
			// Accept() must always fail when closed.
			conn, err := l.Accept()
			assert.Nil(t, conn)
			assert.Error(t, err)
			assert.Zero(t, l.NumAccepted())
		})
	})
}

func TestMockListener_Close(t *testing.T) {
	l := NewMockListener()
	assert.False(t, l.IsClosed())
	assert.NoError(t, l.Close())
	assert.True(t, l.IsClosed())
	assert.Error(t, l.Close())
}

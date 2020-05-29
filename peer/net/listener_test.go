// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package net

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/pkg/test"
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

	timeout := 100 * time.Millisecond
	t.Run("timeout", func(t *testing.T) {
		l, err := NewTCPListener(addr)
		require.NoError(t, err)
		defer l.Close()

		test.AssertNotTerminates(t, timeout, func() {
			l.Accept()
		})
	})

	t.Run("closed", func(t *testing.T) {
		l, err := NewTCPListener(addr)
		require.NoError(t, err)
		l.Close()

		test.AssertTerminates(t, timeout, func() {
			conn, err := l.Accept()
			assert.Nil(t, conn)
			assert.Error(t, err)
		})
	})
}

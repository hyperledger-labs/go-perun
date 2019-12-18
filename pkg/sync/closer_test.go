// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package sync

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/pkg/test"
)

func TestCloser_IsClosed(t *testing.T) {
	t.Parallel()
	var c Closer

	assert.False(t, c.IsClosed())
	assert.False(t, c.IsClosed())

	require.NoError(t, c.Close())

	assert.True(t, c.IsClosed())
	assert.True(t, c.IsClosed())
}

func TestCloser_Closed(t *testing.T) {
	t.Parallel()
	var c Closer

	assert.NotNil(t, c.Closed())
	select {
	case _, ok := <-c.Closed():
		t.Fatalf("Closed() should not yield a value, ok = %t", ok)
	default:
	}

	require.NoError(t, c.Close())

	test.AssertTerminates(t, timeout, func() {
		_, ok := <-c.Closed()
		assert.False(t, ok)
	})
}

func TestCloser_Close(t *testing.T) {
	t.Parallel()

	t.Run("error return check", func(t *testing.T) {
		var c Closer

		assert.NoError(t, c.Close())
		err := c.Close()
		require.Error(t, err)
		assert.Equal(t, err.Error(), alreadyClosedMsg)
		assert.True(t, IsAlreadyClosedError(err))
	})

	t.Run("handler execute check", func(t *testing.T) {
		var c Closer
		const N = 100
		var called int32 = 0
		for i := 0; i < N; i++ {
			c.OnClose(func() { atomic.AddInt32(&called, 1) })
		}

		<-time.After(timeout)
		assert.Zero(t, atomic.LoadInt32(&called))

		require.NoError(t, c.Close())
		<-time.After(timeout)
		assert.Equal(t, atomic.LoadInt32(&called), int32(N))

		// Check that there is no double execution when closed twice.
		require.Error(t, c.Close())
		<-time.After(timeout)
		assert.Equal(t, atomic.LoadInt32(&called), int32(N))
	})
}

// TestCloser_OnClose_Hammer hammers the Closer to expose any data races.
func TestCloser_OnClose_Hammer(t *testing.T) {
	t.Parallel()
	const N = 128
	const M = 32

	for i := 0; i < N; i++ {
		var wg sync.WaitGroup
		var c Closer
		if i&1 != 0 { // Half the tests operate on closed Closer.
			c.Close()
		}

		wg.Add(M)
		for j := 0; j < M; j++ {
			go func() {
				c.OnClose(func() {})
				wg.Done()
			}()
		}

		wg.Wait()
	}
}

// TestCloser_Close_Hammer hammers teh Closer to expose any data races.
func TestCloser_Close_Hammer(t *testing.T) {
	t.Parallel()
	const N = 128
	const M = 32

	for i := 0; i < N; i++ {
		var wg sync.WaitGroup
		var c Closer
		var errs int32 = 0
		wg.Add(M)
		for j := 0; j < M; j++ {
			go func() {
				if c.Close() != nil {
					atomic.AddInt32(&errs, 1)
				}
				wg.Done()
			}()
		}

		wg.Wait()
		require.Equal(t, atomic.LoadInt32(&errs), int32(M-1))
	}
}

func TestIsAlreadyClosedError(t *testing.T) {
	assert.True(t, IsAlreadyClosedError(newAlreadyClosedError()))
	assert.False(t, IsAlreadyClosedError(errors.New("No alreadyClosedError")))
	assert.False(t, IsAlreadyClosedError(nil))
}

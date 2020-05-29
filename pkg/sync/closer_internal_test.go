// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package sync

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	perunatomic "perun.network/go-perun/pkg/sync/atomic"
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
		var called int32
		for i := 0; i < N; i++ {
			require.True(t, c.OnCloseAlways(func() { atomic.AddInt32(&called, 1) }))
		}

		assert.Zero(t, atomic.LoadInt32(&called))

		require.NoError(t, c.Close())
		assert.Equal(t, atomic.LoadInt32(&called), int32(N))

		// Check that there is no double execution when closed twice.
		require.Error(t, c.Close())
		assert.Equal(t, atomic.LoadInt32(&called), int32(N))
	})
}

func TestCloser_OnCloseAlways(t *testing.T) {
	t.Parallel()

	t.Run("before closing", func(t *testing.T) {
		var c Closer
		var executed perunatomic.Bool

		// OnCloseAlways must return true if called before closing.
		assert.True(t, c.OnCloseAlways(executed.Set))
		c.Close()
		// OnCloseAlways must execute the handler if called before closing.
		assert.True(t, executed.IsSet())

	})

	t.Run("after closing", func(t *testing.T) {
		var c Closer
		var executed perunatomic.Bool

		c.Close()
		// OnCloseAlways must return false if called after closing.
		assert.False(t, c.OnCloseAlways(executed.Set))
		// OnCloseAlways must execute the handler if called before closing.
		assert.True(t, executed.IsSet())
	})
}

func TestCloser_OnClose(t *testing.T) {
	t.Parallel()

	t.Run("before closing", func(t *testing.T) {
		var c Closer
		var executed perunatomic.Bool

		// OnClose must return true if called before closing.
		assert.True(t, c.OnClose(executed.Set))
		c.Close()
		// OnClose must execute the handler if called before closing.
		assert.True(t, executed.IsSet())

	})

	t.Run("after closing", func(t *testing.T) {
		var c Closer
		var executed perunatomic.Bool

		c.Close()
		// OnClose must return false if called after closing.
		assert.False(t, c.OnClose(executed.Set))
		// OnClose must not execute the handler if called after closing.
		assert.False(t, executed.IsSet())
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

		// Half the tests operate on closed Closer.
		isClosed := i&1 != 0

		if isClosed {
			c.Close()
		}

		wg.Add(M)
		for j := 0; j < M; j++ {
			go func() {
				// Half the tests use OnClose, half use OnCloseAlways.
				if i&2 != 0 {
					assert.Equal(t, !isClosed, c.OnClose(func() {}))
				} else {
					assert.Equal(t, !isClosed, c.OnCloseAlways(func() {}))
				}
				wg.Done()
			}()
		}

		wg.Wait()
	}
}

// TestCloser_Close_Hammer hammers the Closer to expose any data races.
func TestCloser_Close_Hammer(t *testing.T) {
	t.Parallel()
	const N = 128
	const M = 32

	for i := 0; i < N; i++ {
		var wg sync.WaitGroup
		var c Closer
		var errs int32
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

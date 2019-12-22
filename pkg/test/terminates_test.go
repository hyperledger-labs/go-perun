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
)

const timeout = 200 * time.Millisecond

func TestTerminatesCtx(t *testing.T) {
	// Test often, to detect if there are rare execution branches (due to
	// 'select' statements).
	for i := 0; i < 256; i++ {
		t.Run("immediate deadline", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			assert.False(t, TerminatesCtx(ctx, func() {}))
		})
	}

	t.Run("delayed deadline", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		cancel()
		assert.False(t, TerminatesCtx(ctx, func() {
			<-time.After(2 * timeout)
		}))
	})

	t.Run("no deadline", func(t *testing.T) {
		assert.True(t, TerminatesCtx(context.Background(), func() {}))
	})
}

func TestTerminates(t *testing.T) {
	// Test often, to detect if there are rare execution branches (due to
	// 'select' statements).
	for i := 0; i < 256; i++ {
		t.Run("immediate deadline", func(t *testing.T) {
			assert.False(t, Terminates(-1, func() { <-time.After(time.Second) }))
		})
	}

	t.Run("delayed deadline", func(t *testing.T) {
		assert.False(t, Terminates(timeout, func() {
			<-time.After(2 * timeout)
		}))

		assert.True(t, Terminates(2*timeout, func() {
			<-time.After(timeout)
		}))
	})
}

func TestAssertTerminatesCtx(t *testing.T) {
	t.Run("error case", func(t *testing.T) {
		AssertError(t, func(t T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			AssertTerminatesCtx(ctx, t, func() {})
		})
	})

	t.Run("success case", func(t *testing.T) {
		AssertTerminatesCtx(context.Background(), t, func() {})
	})
}

func TestAssertNotTerminatesCtx(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		AssertNotTerminatesCtx(ctx, t, func() {})
	})

	t.Run("error case", func(t *testing.T) {
		AssertError(t, func(t T) {
			AssertNotTerminatesCtx(context.Background(), t, func() {})
		})
	})
}

func TestAssertTerminates(t *testing.T) {
	t.Run("error case", func(t *testing.T) {
		AssertError(t, func(t T) {
			AssertTerminates(t, -1, func() {})
		})
	})

	t.Run("success case", func(t *testing.T) {
		AssertTerminates(t, timeout, func() {})
	})
}

func TestAssertNotTerminates(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		AssertNotTerminates(t, -1, func() {})
	})

	t.Run("error case", func(t *testing.T) {
		AssertError(t, func(t T) {
			AssertNotTerminates(t, timeout, func() {})
		})
	})
}

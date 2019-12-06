// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package peer

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/pkg/test"
	wire "perun.network/go-perun/wire/msg"
)

// timeout controls how long to wait until we decide that something will never
// happen.
const timeout = 200 * time.Millisecond

func TestReceiver_Close(t *testing.T) {
	t.Parallel()

	r := NewReceiver()
	assert.NoError(t, r.Close())
	assert.Error(t, r.Close())
}

func TestReceiver_Next(t *testing.T) {
	t.Parallel()
	peer := newPeer(nil, nil, nil)
	msg := wire.NewPingMsg()

	t.Run("Happy case", func(t *testing.T) {
		t.Parallel()
		test.AssertTerminates(t, timeout, func() {
			r := NewReceiver()
			go r.Put(peer, msg)
			p, m := r.Next(context.Background())
			assert.Same(t, p, peer)
			assert.Same(t, m, msg)
		})
	})

	t.Run("Closed before", func(t *testing.T) {
		t.Parallel()
		test.AssertTerminates(t, timeout, func() {
			r := NewReceiver()
			r.Close()
			p, m := r.Next(context.Background())
			assert.Nil(t, p)
			assert.Nil(t, m)
		})
	})

	t.Run("Delayed close", func(t *testing.T) {
		t.Parallel()
		test.AssertTerminates(t, timeout*2, func() {
			r := NewReceiver()
			go func() {
				time.Sleep(timeout)
				r.Close()
			}()
			p, m := r.Next(context.Background())
			assert.Nil(t, p)
			assert.Nil(t, m)
		})
	})

	t.Run("Context instant timeout", func(t *testing.T) {
		t.Parallel()
		test.AssertTerminates(t, timeout, func() {
			r := NewReceiver()
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			p, m := r.Next(ctx)
			assert.Nil(t, p)
			assert.Nil(t, m)
		})
	})

	t.Run("Context delayed timeout", func(t *testing.T) {
		t.Parallel()
		test.AssertTerminates(t, timeout*2, func() {
			r := NewReceiver()
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			p, m := r.Next(ctx)
			assert.Nil(t, p)
			assert.Nil(t, m)
		})
	})

}

// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package peer

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	wire "perun.network/go-perun/wire/msg"
)

// timeout controls how long to wait until we decide that something will never
// happen.
const timeout = 200 * time.Millisecond

func TestNewReceiver(t *testing.T) {
	t.Parallel()

	assert.Zero(t, len(NewReceiver().subs), "fresh receivers must be empty")
}

func TestReceiver_Close(t *testing.T) {
	t.Parallel()

	r := NewReceiver()
	assert.NoError(t, r.Close())
	assert.Error(t, r.Close())
}

func pred(wire.Msg) bool { return true }

func TestReceiver_Subscribe(t *testing.T) {
	t.Parallel()

	r := NewReceiver()
	p := newPeer(nil, nil, nil)

	assert.Zero(t, len(r.subs), "receiver must be empty")
	assert.NoError(t, r.Subscribe(p, pred), "first subscribe must not fail")
	assert.Equal(t, len(r.subs), 1, "receiver must not be empty")
	assert.Panics(t, func() { r.Subscribe(p, pred) }, "double subscription must panic")
	assert.NotPanics(t, func() { r.Unsubscribe(p) }, "first unsubscribe must not panic")
	assert.Panics(t, func() { r.Unsubscribe(p) }, "double unsubscribe must panic")
	assert.Zero(t, len(r.subs), "receiver must be empty")
	assert.NoError(t, r.Subscribe(p, pred), "subscribe on empty must not fail")
	assert.Equal(t, len(r.subs), 1, "receiver must not be empty")
	r.UnsubscribeAll()
	assert.Zero(t, len(r.subs), "receiver must be empty")

	p.Close()
	assert.Error(t, r.Subscribe(p, pred), "subscription on closed peer must fail")
	r.Close()
	assert.Error(t, r.Subscribe(p, pred), "subscription on closed receiver must fail")
}

func TestReceiver_Next(t *testing.T) {
	t.Parallel()

	in, out := newPipeConnPair()
	p := newPeer(nil, in, nil)
	go p.recvLoop()
	r := NewReceiver()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	peer, msg := r.Next(ctx)
	assert.Nil(t, peer, "Next() must return nil when canceled")
	assert.Nil(t, msg, "Next() must return nil when canceled")

	out.Send(wire.NewPingMsg())
	// Ensure that the peer received the message.
	<-time.NewTimer(timeout).C

	r.Subscribe(p, pred)
	ctx, cancel = context.WithTimeout(context.Background(), timeout)
	peer, msg = r.Next(ctx)
	assert.Nil(t, msg, "messages received before subscribing must not appear.")
	assert.Nil(t, peer, "messages received before subscribing must not appear.")
	cancel()

	out.Send(wire.NewPongMsg())
	// The new message must appear.
	ctx, cancel = context.WithTimeout(context.Background(), timeout)
	peer, msg = r.Next(ctx)
	assert.Equal(t, peer, p, "message must come from the subscribed peer")
	assert.IsType(t, &wire.PongMsg{}, msg, "received message must be PongMsg")
	cancel()

	// This will trigger in the middle of the next test.
	go func() {
		<-time.NewTimer(timeout).C
		r.Close()
	}()

	doneTest := make(chan struct{}, 1)
	go func() {
		peer, msg = r.Next(context.Background())
		assert.Nil(t, peer, "Next() must fail")
		assert.Nil(t, msg, "Next() must fail")
		doneTest <- struct{}{}
	}()

	select {
	case <-doneTest:
	case <-time.NewTimer(timeout * 2).C:
		t.Error("Next() was not aborted by Receiver.Close()")
	}
}

// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package peer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/pkg/test"
	wire "perun.network/go-perun/wire/msg"
)

func TestRelay_Put(t *testing.T) {
	t.Parallel()

	relay := NewRelay()
	r := NewReceiver()
	relay.Subscribe(r, func(wire.Msg) bool { return true })

	p := newPeer(nil, nil, nil)
	msg := wire.NewPingMsg()
	go relay.Put(p, msg)

	test.AssertTerminates(t, timeout, func() {
		peer, m := r.Next(context.Background())
		assert.Same(t, m, msg)
		assert.Same(t, peer, p)
	})
}

func TestRelay_WithPeerAndReceiver(t *testing.T) {
	t.Parallel()

	acceptAll := func(wire.Msg) bool { return true }

	send, recv := newPipeConnPair()
	p := newPeer(nil, recv, nil)
	relay := NewRelay()
	receiver := NewReceiver()

	relay.Subscribe(receiver, acceptAll)
	p.Subscribe(relay, acceptAll)

	go p.recvLoop()
	msg := wire.NewPingMsg()
	send.Send(msg)

	test.AssertTerminates(t, timeout, func() {
		origin, receivedMsg := receiver.Next(context.Background())
		assert.Equal(t, msg, receivedMsg)
		assert.Same(t, origin, p)
	})
}

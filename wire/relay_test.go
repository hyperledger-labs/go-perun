// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package wire

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/pkg/test"
)

func TestRelay_Put(t *testing.T) {
	t.Parallel()

	relay := NewRelay()
	r := NewReceiver()
	relay.Subscribe(r, func(*Envelope) bool { return true })

	e := NewRandomEnvelope(test.Prng(t), NewPingMsg())
	go relay.Put(e)

	test.AssertTerminates(t, timeout, func() {
		re, err := r.Next(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, re, e)
	})
}

func TestRelay_WithPeerAndReceiver(t *testing.T) {
	t.Parallel()

	acceptAll := func(*Envelope) bool { return true }

	send, recv := newPipeConnPair()
	p := newEndpoint(nil, recv, nil)
	relay := NewRelay()
	receiver := NewReceiver()

	relay.Subscribe(receiver, acceptAll)
	p.Subscribe(relay, acceptAll)

	go p.recvLoop()
	e := NewRandomEnvelope(test.Prng(t), NewPingMsg())
	send.Send(e)

	test.AssertTerminates(t, timeout, func() {
		re, err := receiver.Next(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, re, e)
	})
}

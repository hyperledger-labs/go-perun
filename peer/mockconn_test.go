// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package peer

import (
	"sync"

	"github.com/pkg/errors"

	"perun.network/go-perun/pkg/sync/atomic"
	"perun.network/go-perun/wire"
)

var _ Conn = (*mockConn)(nil)

type mockConn struct {
	mutex     sync.Mutex
	closed    atomic.Bool
	recvQueue chan wire.Msg

	sent func(wire.Msg) // observes sent messages.
}

func newMockConn(sent func(wire.Msg)) *mockConn {
	if sent == nil {
		sent = func(wire.Msg) {}
	}

	return &mockConn{
		sent:      sent,
		recvQueue: make(chan wire.Msg, 1),
	}
}

func (c *mockConn) Send(m wire.Msg) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.closed.IsSet() {
		return errors.New("closed")
	}
	c.sent(m)
	return nil
}

func (c *mockConn) Recv() (wire.Msg, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.closed.IsSet() {
		return nil, errors.New("closed")
	}
	return <-c.recvQueue, nil
}

func (c *mockConn) Close() error {
	if !c.closed.TrySet() {
		return errors.New("double close")
	}
	return nil
}

// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package net

import (
	"sync"

	"github.com/pkg/errors"

	"perun.network/go-perun/pkg/sync/atomic"
	"perun.network/go-perun/wire"
)

var _ Conn = (*MockConn)(nil)

type MockConn struct {
	mutex     sync.Mutex
	closed    atomic.Bool
	recvQueue chan *wire.Envelope

	sent func(*wire.Envelope) // observes sent messages.
}

func newMockConn(sent func(*wire.Envelope)) *MockConn {
	if sent == nil {
		sent = func(*wire.Envelope) {}
	}

	return &MockConn{
		sent:      sent,
		recvQueue: make(chan *wire.Envelope, 1),
	}
}

func (c *MockConn) Send(e *wire.Envelope) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.closed.IsSet() {
		return errors.New("closed")
	}
	c.sent(e)
	return nil
}

func (c *MockConn) Recv() (*wire.Envelope, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.closed.IsSet() {
		return nil, errors.New("closed")
	}
	return <-c.recvQueue, nil
}

func (c *MockConn) Close() error {
	if !c.closed.TrySet() {
		return errors.New("double close")
	}
	return nil
}

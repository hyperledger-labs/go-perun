// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package wire

import (
	"sync"

	"github.com/pkg/errors"

	"perun.network/go-perun/pkg/sync/atomic"
)

var _ Conn = (*MockConn)(nil)

type MockConn struct {
	mutex     sync.Mutex
	closed    atomic.Bool
	recvQueue chan *Envelope

	sent func(*Envelope) // observes sent messages.
}

func newMockConn(sent func(*Envelope)) *MockConn {
	if sent == nil {
		sent = func(*Envelope) {}
	}

	return &MockConn{
		sent:      sent,
		recvQueue: make(chan *Envelope, 1),
	}
}

func (c *MockConn) Send(e *Envelope) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.closed.IsSet() {
		return errors.New("closed")
	}
	c.sent(e)
	return nil
}

func (c *MockConn) Recv() (*Envelope, error) {
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

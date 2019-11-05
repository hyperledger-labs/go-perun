// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package peer

import (
	"sync"

	"github.com/pkg/errors"

	wire "perun.network/go-perun/wire/msg"
)

var _ Conn = (*mockConn)(nil)

type mockConn struct {
	mutex     sync.Mutex
	closed    bool
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
	if c.closed {
		return errors.New("closed")
	} else {
		c.sent(m)
		return nil
	}
}

func (c *mockConn) Recv() (wire.Msg, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.closed {
		return nil, errors.New("closed")
	}
	return <-c.recvQueue, nil
}

func (c *mockConn) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.closed {
		return errors.New("double close")
	} else {
		c.closed = true
		return nil
	}
}

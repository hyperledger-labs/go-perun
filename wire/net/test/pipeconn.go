// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package test

import (
	"net"

	"github.com/pkg/errors"

	"perun.network/go-perun/pkg/sync/atomic"
	"perun.network/go-perun/wire"
	wirenet "perun.network/go-perun/wire/net"
)

// TestConn is a testing connection.
type TestConn struct {
	closed *atomic.Bool
	conn   wirenet.Conn
}

// Send sends an envelope.
func (c *TestConn) Send(e *wire.Envelope) (err error) {
	if err = c.conn.Send(e); err != nil {
		c.Close()
	}
	return
}

// Recv receives an envelope.
func (c *TestConn) Recv() (e *wire.Envelope, err error) {
	if e, err = c.conn.Recv(); err != nil {
		c.Close()
	}
	return
}

// Close closes the TestConn.
func (c *TestConn) Close() error {
	if !c.closed.TrySet() {
		return errors.New("already closed")
	}
	return c.conn.Close()
}

// IsClosed returns whether the TestConn is already closed.
func (c *TestConn) IsClosed() bool {
	return c.closed.IsSet()
}

// NewTestConnPair creates endpoints that are connected via pipes.
func NewTestConnPair() (a wirenet.Conn, b wirenet.Conn) {
	closed := new(atomic.Bool)
	c0, c1 := net.Pipe()
	return &TestConn{closed, wirenet.NewIoConn(c0)}, &TestConn{closed, wirenet.NewIoConn(c1)}
}

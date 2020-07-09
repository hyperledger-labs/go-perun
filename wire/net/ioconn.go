// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package net

import (
	"io"

	"perun.network/go-perun/wire"
)

var _ Conn = (*ioConn)(nil)

// IoConn is a connection that communicates its messages over an io stream.
type ioConn struct {
	conn io.ReadWriteCloser
}

// NewIoConn creates a peer message connection from an io stream.
func NewIoConn(conn io.ReadWriteCloser) Conn {
	return &ioConn{
		conn: conn,
	}
}

func (c *ioConn) Send(e *wire.Envelope) error {
	if err := e.Encode(c.conn); err != nil {
		c.conn.Close()
		return err
	}
	return nil
}

func (c *ioConn) Recv() (*wire.Envelope, error) {
	var e wire.Envelope
	if err := e.Decode(c.conn); err != nil {
		c.conn.Close()
		return nil, err
	}
	return &e, nil
}

func (c *ioConn) Close() error {
	return c.conn.Close()
}

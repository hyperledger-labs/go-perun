// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package wire

import (
	"io"
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

func (c *ioConn) Send(m Msg) error {
	if err := Encode(m, c.conn); err != nil {
		c.conn.Close()
		return err
	}
	return nil
}

func (c *ioConn) Recv() (Msg, error) {
	m, err := Decode(c.conn)
	if err != nil {
		c.conn.Close()
		return nil, err
	}
	return m, nil
}

func (c *ioConn) Close() error {
	return c.conn.Close()
}

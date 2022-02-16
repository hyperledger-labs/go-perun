// Copyright 2019 - See NOTICE file for copyright holders.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package net

import (
	"io"

	"github.com/pkg/errors"

	"perun.network/go-perun/wire"
	"polycry.pt/poly-go/sync/atomic"
)

var _ Conn = (*ioConn)(nil)

// ioConn is a connection that communicates its messages over an io stream.
type ioConn struct {
	closed     atomic.Bool
	conn       io.ReadWriteCloser
	serializer wire.EnvelopeSerializer
}

// NewIoConn creates a peer message connection from an io stream.
func NewIoConn(conn io.ReadWriteCloser, serializer wire.EnvelopeSerializer) Conn {
	return &ioConn{
		conn:       conn,
		serializer: serializer,
	}
}

func (c *ioConn) Send(e *wire.Envelope) error {
	if err := c.serializer.Encode(c.conn, e); err != nil {
		c.conn.Close()
		return err
	}
	return nil
}

func (c *ioConn) Recv() (*wire.Envelope, error) {
	e, err := c.serializer.Decode(c.conn)
	if err != nil {
		c.conn.Close()
		return nil, err
	}
	return e, nil
}

func (c *ioConn) Close() error {
	if !c.closed.TrySet() {
		return errors.New("already closed")
	}
	return c.conn.Close()
}

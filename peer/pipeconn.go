// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package peer

import (
	"io"

	"github.com/pkg/errors"
)

var _ io.ReadWriteCloser = (*pipeConn)(nil)

// pipeConn is a connection that sends over a local pipe.
// It is probably only useful for simpler testing.
type pipeConn struct {
	io.ReadCloser
	io.WriteCloser
	closed chan struct{}
}

func (c *pipeConn) Close() (err error) {
	c.ReadCloser.Close()
	c.WriteCloser.Close()

	err = errors.New("already closed")
	defer func() { recover() }()
	close(c.closed)
	err = nil
	return
}

// newPipeConnPair creates endpoints that are connected via pipes.
func newPipeConnPair() (a Conn, b Conn) {
	ra, wa := io.Pipe()
	rb, wb := io.Pipe()
	return NewConn(&pipeConn{ra, wb, make(chan struct{})}), NewConn(&pipeConn{rb, wa, make(chan struct{})})
}

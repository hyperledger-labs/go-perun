// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

// Package tcp implements a direct tcp implementation and fulfills
// the io.ReadWriteCloser interface.
package tcp

import (
	"net"
)

// Connection represents a direct tcp connection to a single peer.
// It implements the io.ReadWriteCloser interface.
type Connection struct {
	net.Conn
	server *Server
}

// Connect connects to another server.
func Connect(host, port string) (Connection, error) {
	log.Info("Connecting to a server at " + host + ":" + port)
	conn, err := net.Dial("tcp", host+":"+port)
	return Connection{
		Conn:   conn,
		server: nil,
	}, err
}

// Read reads the next message from a connection.
func (c *Connection) Read(p []byte) (n int, err error) {
	reqLen, err := c.Conn.Read(p)
	return reqLen, err
}

// Write sends the message to a peer.
func (c *Connection) Write(p []byte) (n int, err error) {
	return c.Conn.Write(p)
}

// Close closes this connection.
func (c *Connection) Close() error {
	log.Info("Closed connection with peer" + c.RemoteAddr().String())
	if c.server != nil {
		c.server.removeConnection(c)
	}
	return c.Conn.Close()
}

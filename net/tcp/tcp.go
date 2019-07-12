// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

// Package tcp implements a direct tcp implementation and fulfills
// the io.ReadWriteCloser interface.
package tcp

import (
	"net"
	"sync"

	"github.com/pkg/errors"
)

// Connection represents a connection to a single peer.
type Connection struct {
	net.Conn
	server *Server
}

// Server represents a server to a peer.
type Server struct {
	listener net.Listener
	conns    map[string]*Connection
	mu       sync.RWMutex
}

// Connect connects to another server.
func Connect(host, port string) (Connection, error) {
	conn, err := net.Dial("tcp", host+":"+port)
	return Connection{
		Conn:   conn,
		server: nil,
	}, err
}

// NewTCPServer initializes a new tcp server and listens to incomming connections.
func NewTCPServer(host, port string) (*Server, error) {

	// TODO log the listening here
	listener, err := net.Listen("tcp", host+":"+port)
	if err != nil {
		return &Server{}, errors.Wrap(err, "error trying to open connection on "+host+":"+port)
	}
	server := Server{
		listener: listener,
		conns:    make(map[string]*Connection),
	}

	go server.acceptIncomingConnections()
	return &server, nil
}

func (s *Server) acceptIncomingConnections() {
	for {
		c, err := s.listener.Accept()
		// TODO log accept here
		if err != nil {
			if err.Error() != "accept tcp "+s.listener.Addr().String()+": use of closed network connection" {
				panic("Unknown error")
			}
			continue
		}
		s.mu.Lock()
		conn := Connection{
			Conn:   c,
			server: s,
		}
		s.conns[conn.RemoteAddr().String()] = &conn
		s.mu.Unlock()
	}
}

// Connections returns all open connections of this server.
func (s *Server) Connections() []*Connection {
	s.mu.RLock()
	defer s.mu.RUnlock()

	conns := []*Connection{}
	for _, conn := range s.conns {
		conns = append(conns, conn)
	}
	return conns
}

func (s *Server) removeConnection(conn *Connection) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.conns, conn.RemoteAddr().String())
}

// Close closes all connections of the server.
func (s *Server) Close() error {
	if s.listener == nil {
		return errors.New("Server has no valid listener")
	}
	var wrapErr error
	for _, conn := range s.conns {
		if err := conn.Close(); err != nil {
			wrapErr = errors.Wrap(wrapErr, err.Error())
		}
	}
	if err := s.listener.Close(); err != nil {
		wrapErr = errors.Wrap(wrapErr, err.Error())
	}
	return wrapErr
}

// Read reads the next message from a connection.
func (c *Connection) Read(p []byte) (n int, err error) {
	reqLen, err := c.Conn.Read(p)
	return reqLen, nil
}

// Write sends the message to a peer.
func (c *Connection) Write(p []byte) (n int, err error) {
	return c.Conn.Write(p)
}

// Close closes this connection.
func (c *Connection) Close() error {
	if c.server != nil {
		c.server.removeConnection(c)
	}
	return c.Conn.Close()
}

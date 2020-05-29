// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

// Package net provides the networking backends for go-perun.
package net // import "perun.network/go-perun/net"

import (
	"net"

	"perun.network/go-perun/log"
)

// LoggedConn wraps a network connection with logging abilities.
type LoggedConn struct {
	net.Conn
}

// Close closes a connection to a peer.
func (c *LoggedConn) Close() error {
	log.Info("Closed connection with peer " + c.RemoteAddr().String())
	return c.Conn.Close()
}

// LoggedListener wraps a network listener with logging abilities.
type LoggedListener struct {
	net.Listener
}

// Accept accepts a new connection from a peer.
func (l *LoggedListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err == nil {
		log.Debugf("Accepted connection from peer ", conn.RemoteAddr().String())
	}
	return &LoggedConn{conn}, err
}

// Close closes a listener.
func (l *LoggedListener) Close() error {
	log.Debugf("Closing Listener")
	return l.Listener.Close()
}

// Wrapped helper functions for basic networking types (tcp, http, unix, ipv4).
// For more infos on which networks can be used see: https://golang.org/src/net/dial.go?s=9568:9616#L299

// Dial wraps the net.Dial functionality.
func Dial(network, address string) (net.Conn, error) {
	log.Info("Connecting to a server at " + network + "://" + address)
	conn, err := net.Dial(network, address)
	return &LoggedConn{conn}, err
}

// Listen wraps the net.Listen functionality.
func Listen(network, address string) (net.Listener, error) {
	log.Info("Start listening on " + network + "://" + address)
	conn, err := net.Listen(network, address)
	return &LoggedListener{conn}, err
}

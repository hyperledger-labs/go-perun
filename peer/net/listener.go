// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package net

import (
	"net"

	"github.com/pkg/errors"

	"perun.network/go-perun/peer"
)

// Listener is a TCP implementation of the peer.Listener interface.
type Listener struct {
	net.Listener
}

var _ peer.Listener = (*Listener)(nil)

// NewListener creates a listener reachable under the requested address.
func NewListener(network string, address string) (*Listener, error) {
	l, err := net.Listen(network, address)
	if err != nil {
		return nil, errors.Wrapf(err,
			"failed to create listener for '%s'", address)
	}

	return &Listener{Listener: l}, nil
}

// NewTCPListener is a short-hand version of NewListener for TCP listeners.
func NewTCPListener(address string) (*Listener, error) {
	return NewListener("tcp", address)
}

// NewUnixListener is a short-hand version of NewListener for Unix listeners.
func NewUnixListener(address string) (*Listener, error) {
	return NewListener("unix", address)
}

// Accept implements peer.Dialer.Accept().
func (l *Listener) Accept() (peer.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, errors.Wrap(err, "accept failed")
	}

	return peer.NewIoConn(conn), nil
}

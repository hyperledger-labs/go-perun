// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package wire

import (
	"net"

	"github.com/pkg/errors"
)

// NetListener is a TCP Listener.
type NetListener struct {
	net.Listener
}

var _ Listener = (*NetListener)(nil)

// NewNetListener creates a listener reachable under the requested address.
func NewNetListener(network string, address string) (*NetListener, error) {
	l, err := net.Listen(network, address)
	if err != nil {
		return nil, errors.Wrapf(err,
			"failed to create listener for '%s'", address)
	}

	return &NetListener{Listener: l}, nil
}

// NewTCPListener is a short-hand version of NewNetListener for TCP listeners.
func NewTCPListener(address string) (*NetListener, error) {
	return NewNetListener("tcp", address)
}

// NewUnixListener is a short-hand version of NewNetListener for Unix listeners.
func NewUnixListener(address string) (*NetListener, error) {
	return NewNetListener("unix", address)
}

// Accept implements peer.Dialer.Accept().
func (l *NetListener) Accept() (Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, errors.Wrap(err, "accept failed")
	}

	return NewIoConn(conn), nil
}

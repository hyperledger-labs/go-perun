// Copyright 2020 - See NOTICE file for copyright holders.
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

package simple

import (
	"net"

	"github.com/pkg/errors"
	wirenet "perun.network/go-perun/wire/net"
)

// Listener is a TCP Listener.
type Listener struct {
	net.Listener
}

var _ wirenet.Listener = (*Listener)(nil)

// NewNetListener creates a listener reachable under the requested address.
func NewNetListener(network string, address string) (*Listener, error) {
	l, err := net.Listen(network, address)
	if err != nil {
		return nil, errors.Wrapf(err,
			"failed to create listener for '%s'", address)
	}

	return &Listener{Listener: l}, nil
}

// NewTCPListener is a short-hand version of NewNetListener for TCP listeners.
func NewTCPListener(address string) (*Listener, error) {
	return NewNetListener("tcp", address)
}

// NewUnixListener is a short-hand version of NewNetListener for Unix listeners.
func NewUnixListener(address string) (*Listener, error) {
	return NewNetListener("unix", address)
}

// Accept implements peer.Dialer.Accept().
func (l *Listener) Accept() (wirenet.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, errors.Wrap(err, "accept failed")
	}

	return wirenet.NewIoConn(conn), nil
}

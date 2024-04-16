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
	"crypto/tls"

	"fmt"
	"net"

	"github.com/pkg/errors"
	"perun.network/go-perun/wire"
	wirenet "perun.network/go-perun/wire/net"
	"polycry.pt/poly-go/sync"
)

// Listener is a TCP Listener.
type Listener struct {
	host string
	net.Listener
	sync.Closer
}

var _ wirenet.Listener = (*Listener)(nil)

// NewNetListener creates a listener reachable under the requested address.
func NewNetListener(network string, address string, config *tls.Config) (*Listener, error) {
	l, err := tls.Listen(network, address, config)
	if err != nil {
		return nil, errors.Wrapf(err,
			"failed to create listener for '%s'", address)
	}

	return &Listener{host: address, Listener: l}, nil
}

// NewTCPListener is a short-hand version of NewNetListener for TCP listeners.
func NewTCPListener(address string, config *tls.Config) (*Listener, error) {
	return NewNetListener("tcp", address, config)
}

// NewUnixListener is a short-hand version of NewNetListener for Unix listeners.
func NewUnixListener(address string, config *tls.Config) (*Listener, error) {
	return NewNetListener("unix", address, config)
}

// Accept implements peer.Dialer.Accept().
func (l *Listener) Accept(ser wire.EnvelopeSerializer) (wirenet.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, errors.Wrap(err, "accept failed")
	}

	return wirenet.NewIoConn(conn, ser), nil
}

// Close implements peer.Listener.Close().
func (l *Listener) Close() error {
	if l.IsClosed() {
		return fmt.Errorf("listener already closed")
	}
	err := l.Listener.Close()
	if err != nil {
		return err
	}

	return l.Closer.Close()
}

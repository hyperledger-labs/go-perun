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
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/pkg/errors"
	"perun.network/go-perun/wire"
	wirenet "perun.network/go-perun/wire/net"
	pkgsync "polycry.pt/poly-go/sync"
)

// Dialer is a simple lookup-table based dialer that can dial known peers.
// New peer addresses can be added via Register().
type Dialer struct {
	hub     *ConnHub
	dialer  tls.Dialer // Used to dial connections.
	network string     // The socket type.

	pkgsync.Closer
}

var _ wirenet.Dialer = (*Dialer)(nil)

// NewNetDialer creates a new dialer with a preset default timeout for dial
// attempts. Leaving the timeout as 0 will result in no timeouts. Standard OS
// timeouts may still apply even when no timeout is selected. The network string
// controls the type of connection that the dialer can dial.
// `serializer` defines the message encoding.
func NewNetDialer(network string, defaultTimeout time.Duration, tlsConfig *tls.Config) *Dialer {
	netDialer := &net.Dialer{Timeout: defaultTimeout}

	return &Dialer{
		dialer: tls.Dialer{
			NetDialer: netDialer,
			Config:    tlsConfig,
		},
		network: network,
	}
}

// NewTCPDialer is a short-hand version of NewNetDialer for creating TCP dialers.
func NewTCPDialer(defaultTimeout time.Duration, tlsConfig *tls.Config) *Dialer {
	return NewNetDialer("tcp", defaultTimeout, tlsConfig)
}

// NewUnixDialer is a short-hand version of NewNetDialer for creating Unix dialers.
func NewUnixDialer(defaultTimeout time.Duration, tlsConfig *tls.Config) *Dialer {
	return NewNetDialer("unix", defaultTimeout, tlsConfig)
}

// Dial implements Dialer.Dial().
func (d *Dialer) Dial(ctx context.Context, addr wire.Address, ser wire.EnvelopeSerializer) (wirenet.Conn, error) {
	done := make(chan struct{})
	defer close(done)

	listener, ok := d.hub.findListener(addr)
	if !ok {
		return nil, errors.New("peer not found")
	}

	// To combine the provided context with the Dialer's Closer as specified by
	// the Dialer interface, we have to use some goroutine trickery.
	wrappedCtx, cancel := context.WithCancel(ctx)
	go func() {
		defer cancel()

		select {
		case <-d.Closed():
		case <-done:
		}
	}()

	conn, err := d.dialer.DialContext(wrappedCtx, d.network, listener.host)
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial peer")
	}

	return wirenet.NewIoConn(conn, ser), nil
}

// Close closes the Dialer and cleans up any associated resources.
func (d *Dialer) Close() error {
	if !d.IsClosed() {
		// Mark the Dialer as closed.
		d.Closer.Close()

		if d.hub != nil {
			if err := d.hub.eraseDialer(d); err != nil {
				return errors.Wrap(err, "failed to remove Dialer from ConnHub")
			}
		}
		return nil
	}

	return fmt.Errorf("dialer already closed")
}

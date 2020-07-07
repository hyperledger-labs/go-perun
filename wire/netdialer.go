// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package wire

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/pkg/errors"
	pkgsync "perun.network/go-perun/pkg/sync"
)

// NetDialer is a simple lookup-table based dialer that can dial known peers.
// New peer addresses can be added via Register().
type NetDialer struct {
	mutex   sync.RWMutex       // Protects peers.
	peers   map[Address]string // Known peer addresses.
	dialer  net.Dialer         // Used to dial connections.
	network string             // The socket type.

	pkgsync.Closer
}

var _ Dialer = (*NetDialer)(nil)

// NewNetDialer creates a new dialer with a preset default timeout for dial
// attempts. Leaving the timeout as 0 will result in no timeouts. Standard OS
// timeouts may still apply even when no timeout is selected. The network string
// controls the type of connection that the dialer can dial.
func NewNetDialer(network string, defaultTimeout time.Duration) *NetDialer {
	return &NetDialer{
		peers:   make(map[Address]string),
		dialer:  net.Dialer{Timeout: defaultTimeout},
		network: network,
	}
}

// NewTCPDialer is a short-hand version of NewNetDialer for creating TCP dialers.
func NewTCPDialer(defaultTimeout time.Duration) *NetDialer {
	return NewNetDialer("tcp", defaultTimeout)
}

// NewUnixDialer is a short-hand version of NewNetDialer for creating Unix dialers.
func NewUnixDialer(defaultTimeout time.Duration) *NetDialer {
	return NewNetDialer("unix", defaultTimeout)
}

func (d *NetDialer) get(addr Address) (string, bool) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	host, ok := d.peers[addr]
	return host, ok
}

// Dial implements Dialer.Dial().
func (d *NetDialer) Dial(ctx context.Context, addr Address) (Conn, error) {
	done := make(chan struct{})
	defer close(done)

	host, ok := d.get(addr)
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

	conn, err := d.dialer.DialContext(wrappedCtx, d.network, host)
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial peer")
	}

	return NewIoConn(conn), nil
}

// Register registers a network address for a peer address.
func (d *NetDialer) Register(addr Address, address string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.peers[addr] = address
}

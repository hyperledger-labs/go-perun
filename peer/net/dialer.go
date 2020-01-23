// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package net

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/pkg/errors"

	"perun.network/go-perun/peer"
	pkgsync "perun.network/go-perun/pkg/sync"
)

// Dialer is a simple lookup-table based dialer that can dial known peers.
// New peer addresses can be added via Register().
type Dialer struct {
	mutex   sync.RWMutex            // Protects peers.
	peers   map[peer.Address]string // Known peer addresses.
	dialer  net.Dialer              // Used to dial connections.
	network string                  // The socket type.

	pkgsync.Closer
}

var _ peer.Dialer = (*Dialer)(nil)

// NewDialer creates a new dialer with a preset default timeout for dial
// attempts. Leaving the timeout as 0 will result in no timeouts. Standard OS
// timeouts may still apply even when no timeout is selected. The network string
// controls the type of connection that the dialer can dial.
func NewDialer(network string, defaultTimeout time.Duration) *Dialer {
	return &Dialer{
		peers:   make(map[peer.Address]string),
		dialer:  net.Dialer{Timeout: defaultTimeout},
		network: network,
	}
}

// NewTCPDialer is a short-hand version of NewDialer for creating TCP dialers.
func NewTCPDialer(defaultTimeout time.Duration) *Dialer {
	return NewDialer("tcp", defaultTimeout)
}

// NewUnixDialer is a short-hand version of NewDialer for creating Unix dialers.
func NewUnixDialer(defaultTimeout time.Duration) *Dialer {
	return NewDialer("unix", defaultTimeout)
}

func (d *Dialer) get(addr peer.Address) (string, bool) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	host, ok := d.peers[addr]
	return host, ok
}

// Dial implements peer.Dialer.Dial().
func (d *Dialer) Dial(ctx context.Context, addr peer.Address) (peer.Conn, error) {
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

	return peer.NewIoConn(conn), nil
}

// Register registers a network address for a peer address.
func (d *Dialer) Register(addr peer.Address, address string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.peers[addr] = address
}

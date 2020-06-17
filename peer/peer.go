// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

// Package peer contains the peer connection related code.
package peer // import "perun.network/go-perun/peer"

import (
	"context"

	"github.com/pkg/errors"

	"perun.network/go-perun/log"
	"perun.network/go-perun/pkg/sync"
	"perun.network/go-perun/wire"
)

// Peer is an authenticated connection to a Perun peer.
// It contains the peer's identity. Peers are thread-safe.
// Peers must not be created manually. The creation of peers is handled by the
// Registry, which tracks all existing peers. The registry, in turn, is used by
// the Client.
//
// Sending messages to a peer is done via the Send() method, or via the
// Broadcaster helper type. To receive messages from a Peer, use the Receiver
// helper type (by subscribing).
//
// If a peer is entered into the registry, but still being dialed, then it
// exists in an unfinished state, and all its operations will block until it is
// dialed or closed.
type Peer struct {
	PerunAddress Address // The peer's perun address.

	conn Conn // The peer's connection.

	creating sync.Mutex // Prevent races when concurrently creating the peer.
	sending  sync.Mutex // Blocks multiple Send calls.

	created sync.Closer

	producer
}

// recvLoop continuously receives messages from a peer until it is closed.
// Received messages are relayed via the peer's subscription system. This is
// called by the registry when the peer is registered.
func (p *Peer) recvLoop() {
	// Wait until the peer exists or is closed.
	if !p.waitExists(nil) {
		return // closed before connection set
	}

	for {
		m, err := p.conn.Recv()
		if err != nil {
			p.Close() // Ignore double close.
			log.WithError(err).Errorf("Ending recvLoop on closed connection of peer %v", p.PerunAddress)
			return
		}
		// Broadcast the received message to all interested subscribers.
		p.produce(m, p)
	}
}

// create finishes a peer that does not yet have a connection.
// This is needed in the registry when a peer is still being dialed, but
// already registered. This wakes up all operations that were started on the
// unfinished peer object.
func (p *Peer) create(conn Conn) {
	p.creating.Lock()
	defer p.creating.Unlock()

	if p.conn == nil {
		p.conn = conn
		p.created.Close()
	} else {
		conn.Close()
	}
}

// waitExists waits until the peer is either fully created, or closed.
// The optional context can be used to add a third condition to wait for.
// The functions returns whether the peer connection was set (true) or whether
// the peer was prematurely closed or the context finshed (false).
func (p *Peer) waitExists(ctx context.Context) bool {
	var done <-chan struct{}
	if ctx != nil {
		done = ctx.Done()
	}

	if p.IsClosed() {
		return false
	}

	select {
	case <-p.created.Closed():
		return true
	case <-p.Closed():
	case <-done:
	}
	return false
}

// exists returns whether the peer has been fully created.
func (p *Peer) exists() bool {
	return p.created.IsClosed()
}

// OnCreate calls fn after create is called, but only if it has not yet been
// called. See pkg/sync/Closer.OnClose.
func (p *Peer) OnCreate(fn func()) bool {
	return p.created.OnClose(fn)
}

// OnCreateAlways calls fn after create is called, even if it has already been
// called. See pkg/sync/Closer.OnCloseAlways.
func (p *Peer) OnCreateAlways(fn func()) bool {
	return p.created.OnCloseAlways(fn)
}

// Send sends a single message to a peer.
// Fails if the peer is closed via Close() or the transmission fails.
//
// The passed context is used to timeout the send operation. If the context
// times out, the peer is closed.
func (p *Peer) Send(ctx context.Context, m wire.Msg) error {
	// Wait until peer exists, is closed, or context timeout.
	if !p.waitExists(ctx) {
		p.Close()                        // replace with p.conn.Close() when reintroducing repair.
		return errors.New("peer closed") // closed before connection set
	}

	if !p.sending.TryLockCtx(ctx) {
		p.Close() // replace with p.conn.Close() when reintroducing repair.
		return errors.New("aborted manually")
	}

	sent := make(chan error, 1)
	// Asynchronously send, because we cannot abort Conn.Send().
	go func() {
		defer p.sending.Unlock()
		sent <- p.conn.Send(m)
	}()

	// Return as soon as the sending finishes, times out, or peer is closed.
	select {
	case err := <-sent:
		return err
	case <-p.Closed():
		return errors.New("peer closed")
	case <-ctx.Done():
		p.Close() // replace with p.conn.Close() when reintroducing repair.
		return errors.New("aborted manually")
	}
}

// Close closes the peer's connection. A closed peer is no longer usable.
func (p *Peer) Close() (err error) {
	if err = p.producer.Close(); sync.IsAlreadyClosedError(err) {
		return
	}

	// Close the peer's connection.
	if p.conn != nil {
		if cerr := p.conn.Close(); cerr != nil && err == nil {
			err = errors.WithMessage(cerr, "closing connection")
		}
	}

	return
}

// newPeer creates a new peer from a peer address and connection.
func newPeer(addr Address, conn Conn, _ Dialer) *Peer {
	p := &Peer{
		PerunAddress: addr,

		conn:     conn,
		producer: makeProducer(),
	}

	if p.conn != nil {
		p.created.Close()
	}

	return p
}

// String returns the peer's address string
func (p *Peer) String() string {
	return p.PerunAddress.String()
}

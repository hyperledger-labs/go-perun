// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

// Package peer contains the peer connection related code.
package peer // import "perun.network/go-perun/peer"

import (
	"context"

	"github.com/pkg/errors"

	"perun.network/go-perun/log"
	"perun.network/go-perun/pkg/sync"
	wire "perun.network/go-perun/wire/msg"
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

	conn Conn          // The peer's connection.
	subs subscriptions // The receivers that are subscribed to the peer.

	creating sync.Mutex // Prevent races when concurrently creating the peer.
	sending  sync.Mutex // Blocks multiple Send calls.
	closing  sync.Mutex

	exists chan struct{} // Indicates whether a peer has been created yet.
	closed chan struct{} // Indicates whether the peer is closed.

	closeWork func(*Peer) // Work to be done when the peer is closed.
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
		if m, err := p.conn.Recv(); err != nil {
			log.Debugf("ending recvLoop of closed peer %v", p.PerunAddress)
			return
		} else {
			// Broadcast the received message to all interested subscribers.
			p.subs.put(m, p)
		}
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
		close(p.exists)
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

	select {
	case <-p.exists:
		return true
	case <-p.closed:
	case <-done:
	}
	return false
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
	case <-p.closed:
		return errors.New("peer closed")
	case <-ctx.Done():
		p.Close() // replace with p.conn.Close() when reintroducing repair.
		return errors.New("aborted manually")
	}
}

// isClosed checks whether the peer is marked as closed.
// This is different from the peer's connection being closed, in that a closed
// peer cannot have its connection restored and is marked for deletion.
func (p *Peer) isClosed() bool {
	select {
	case <-p.closed:
		return true
	default:
		return false
	}
}

// Close closes a peer's connection and deletes it from the registry. If the
// peer was already closed, results in an error. A closed peer is no longer
// usable.
func (p *Peer) Close() error {
	p.closing.Lock()
	defer p.closing.Unlock()

	if p.isClosed() {
		return errors.New("already closed")
	}

	// Unregister the peer.
	p.closeWork(p)
	// Mark the peer as closed.
	close(p.closed)
	// Close the peer's connection.
	if p.conn != nil {
		p.conn.Close()
	}
	// Delete this peer from all receivers.
	p.subs.mutex.Lock()
	defer p.subs.mutex.Unlock()
	for _, recv := range p.subs.subs {
		recv.receiver.unsubscribe(p, false)
	}

	return nil
}

// newPeer creates a new peer from a peer address and connection.
func newPeer(addr Address, conn Conn, closeWork func(*Peer), _ Dialer) *Peer {

	// In tests, it is useful to omit the function.
	if closeWork == nil {
		closeWork = func(*Peer) {}
	}

	p := new(Peer)
	*p = Peer{
		PerunAddress: addr,

		conn: conn,
		subs: makeSubscriptions(p),

		exists: make(chan struct{}),
		closed: make(chan struct{}),

		closeWork: closeWork,
	}

	if p.conn != nil {
		close(p.exists)
	}

	return p
}

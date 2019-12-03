// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package peer

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"perun.network/go-perun/log"
	perunsync "perun.network/go-perun/pkg/sync"
)

// Registry is a peer Registry.
// It should not be used manually, but only internally by the client.
type Registry struct {
	mutex sync.RWMutex
	peers []*Peer  // The list of all of the registry's peers.
	id    Identity // The identity of the node.

	exchangeAddrsTimeout int64

	dialer    Dialer      // Used for dialing peers (and later: repairing).
	subscribe func(*Peer) // Sets up peer subscriptions.

	perunsync.Closer
}

const defaultExchangeAddrsTimeout = 10 * time.Second

// NewRegistry creates a new registry.
// The provided callback is used to set up new peer's subscriptions and it is
// called before the peer starts receiving messages.
func NewRegistry(id Identity, subscribe func(*Peer), dialer Dialer) *Registry {
	return &Registry{
		id:        id,
		subscribe: subscribe,
		dialer:    dialer,

		exchangeAddrsTimeout: int64(defaultExchangeAddrsTimeout),
	}
}

func (r *Registry) SetExchangeAddrsTimeout(d time.Duration) {
	atomic.StoreInt64(&r.exchangeAddrsTimeout, int64(d))
}

// Close closes the registry's dialer and all its peers.
func (r *Registry) Close() (err error) {
	if err = r.Closer.Close(); err != nil {
		return
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	for _, p := range r.peers {
		// When peers are closed, they delete themselves from the registry.
		if cerr := p.Close(); cerr != nil && err == nil {
			err = errors.WithMessage(cerr, "closing peer")
		}
	}

	if r.dialer != nil {
		if cerr := r.dialer.Close(); cerr != nil && err == nil {
			err = errors.WithMessage(cerr, "closing dialer")
		}
	}
	return
}

// Listen starts listening for incoming connections on the provided listener and
// currently just automatically accepts them after successful authentication.
// This function does not start go routines but instead should be started by the
// user as `go registry.Listen()`.
func (r *Registry) Listen(listener Listener) {
	// Start listener and accept all incoming peer connections, writing them to
	// the registry.
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Debugf("peer listener closed: %v", err)
			return
		}

		// setup connection in a separate routine so that new incoming
		// connections can immediately be handled.
		go r.setupConn(conn)
	}
}

// setupConn authenticates a fresh connection, and if successful, adds it to the
// registry.
func (r *Registry) setupConn(conn Conn) error {
	timeout := time.Duration(atomic.LoadInt64(&r.exchangeAddrsTimeout))
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var peerAddr Address
	var err error
	if peerAddr, err = ExchangeAddrs(ctx, r.id, conn); err != nil {
		conn.Close()
		return errors.WithMessage(err, "could not authenticate peer")
	}
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if peer, _ := r.find(peerAddr); peer == nil {
		r.addPeer(peerAddr, conn)
	} else {
		peer.create(conn)
	}
	return nil
}

// find looks up a peer via its Perun address.
// If found, returns the peer and its index, otherwise returns a nil peer.
// find is not thread safe and is assumed to be called from a method which has
// the r.mutex lock.
func (r *Registry) find(addr Address) (*Peer, int) {
	for i, peer := range r.peers {
		if peer.PerunAddress.Equals(addr) {
			if peer.IsClosed() {
				// remove from slice
				r.peers[i] = r.peers[len(r.peers)-1]
				r.peers = r.peers[:len(r.peers)-1]
				return nil, -1
			}
			return peer, i
		}
	}

	return nil, -1
}

// prune removes all closed peers from the Registry.
// prune is not thread safe and is assumed to be called from a method which has
// the r.mutex lock.
func (r *Registry) prune() {
	peers := r.peers[:0]
	for _, peer := range r.peers {
		if !peer.IsClosed() {
			peers = append(peers, peer)
		}
	}
	r.peers = peers
}

// Get looks up the peer via its perun address.
// If the peer does not exist yet, creates a placeholder peer and dials the
// requested address. When the dialling finishes, completes the peer or closes
// it, depending on the success of the dialing operation. The unfinished peer
// object can be used already, but it will block until the peer is finished or
// closed. If the registry is already closed, returns a closed peer.
func (r *Registry) Get(ctx context.Context, addr Address) (*Peer, error) {
	r.mutex.Lock()
	if p, i := r.find(addr); i != -1 {
		r.mutex.Unlock()
		if p.waitExists(ctx) {
			return p, nil
		} else {
			return nil, errors.New("peer was not created in time")
		}
	}

	// Create "nonexistent" peer (nil connection).
	peer := r.addPeer(addr, nil)
	r.mutex.Unlock()

	if err := r.authenticatedDial(ctx, peer, addr); err != nil {
		peer.Close()
		return nil, errors.WithMessage(err, "failed to dial peer")
	}

	return peer, nil
}

func (r *Registry) authenticatedDial(ctx context.Context, peer *Peer, addr Address) error {
	conn, err := r.dialer.Dial(ctx, addr)

	if peer.exists() {
		if conn != nil {
			conn.Close()
		}
		return nil
	} else if err != nil {
		return errors.WithMessage(err, "failed to dial")
	}

	a, err := ExchangeAddrs(ctx, r.id, conn)
	if err != nil || !a.Equals(addr) {
		conn.Close()
		if !peer.exists() {
			peer.Close()
			if err != nil {
				return errors.WithMessage(err, "ExchangeAddrs failed")
			} else {
				return errors.New("Dialed impersonator")
			}
		}
		return nil
	}

	peer.create(conn)
	return nil
}

// NumPeers returns the current number of peers in the registry including
// placeholder peers (cf. Registry.Get).
func (r *Registry) NumPeers() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	r.prune()
	return len(r.peers)
}

// Has return true if and only if there is a peer with the given address in the
// registry. The function does not differentiate between regular and
// placeholder peers.
func (r *Registry) Has(addr Address) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	p, _ := r.find(addr)

	return p != nil
}

// addPeer adds a new peer to the registry.
// addPeer is not thread safe and is assumed to be called from a method which has
// the r.mutex lock.
func (r *Registry) addPeer(addr Address, conn Conn) *Peer {
	// Create and register a new peer.
	peer := newPeer(addr, conn, r.dialer)
	r.peers = append(r.peers, peer)
	// Setup the peer's subscriptions.
	r.subscribe(peer)
	// Start receiving messages.
	go peer.recvLoop()

	return peer
}

// delete deletes a peer from the registry.
// If the peer does not exist in the registry, panics. Does not close the peer.
func (r *Registry) delete(peer *Peer) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, i := r.find(peer.PerunAddress); i != -1 {
		// Delete the i-th entry.
		r.peers[i] = r.peers[len(r.peers)-1]
		r.peers = r.peers[:len(r.peers)-1]
	} else {
		log.Panic("tried to delete non-existent peer!")
	}
}

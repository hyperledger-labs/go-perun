// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package peer

import (
	"context"
	"sync"
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

	exchangeAddrsTimeout time.Duration

	dialer    Dialer      // Used for dialing peers (and later: repairing).
	subscribe func(*Peer) // Sets up peer subscriptions.

	perunsync.Closer
}

// NewRegistry creates a new registry.
// The provided callback is used to set up new peer's subscriptions and it is
// called before the peer starts receiving messages.
func NewRegistry(id Identity, subscribe func(*Peer), dialer Dialer) *Registry {
	return &Registry{
		id:        id,
		subscribe: subscribe,
		dialer:    dialer,

		exchangeAddrsTimeout: 10 * time.Second,
	}
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

		// setup connection in a serparate routine so that new incoming
		// connections can immediately be handled.
		go r.setupConn(conn)
	}
}

// setupConn authenticates a fresh connection, and if successful, adds it to the
// registry.
func (r *Registry) setupConn(conn Conn) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.exchangeAddrsTimeout)
	defer cancel()

	if peerAddr, err := ExchangeAddrs(ctx, r.id, conn); err != nil {
		conn.Close()
		return errors.WithMessage(err, "could not authenticate peer")
	} else {
		// the peer registry is thread safe
		r.Register(peerAddr, conn)
		return nil
	}
}

// find looks up a peer via its Perun address.
// If found, returns the peer and its index, otherwise returns a nil peer.
func (r *Registry) find(addr Address) (*Peer, int) {
	for i, peer := range r.peers {
		if peer.PerunAddress.Equals(addr) {
			return peer, i
		}
	}

	return nil, -1
}

func (r *Registry) authenticatedDial(peer *Peer, addr Address) error {
	conn, err := r.dialer.Dial(context.Background(), addr)
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if err != nil {
		if peer.exists() {
			return nil // Failed to dial, but peer was created anyway.
		} else {
			peer.Close()
			return errors.WithMessage(err, "failed to dial")
		}
	}

	a, err := ExchangeAddrs(context.Background(), r.id, conn)
	if err != nil {
		conn.Close()
		peer.Close()
		return errors.WithMessage(err, "ExchangeAddrs failed")
	}
	if !a.Equals(addr) {
		conn.Close()
		peer.Close()
		return errors.New("Dialed impersonator")
	}

	peer.create(conn)
	return nil
}

// Get looks up the peer via its perun address.
// If the peer does not exist yet, creates a placeholder peer and dials the
// requested address. When the dialling finishes, completes the peer or closes
// it, depending on the success of the dialing operation. The unfinished peer
// object can be used already, but it will block until the peer is finished or
// closed. If the registry is already closed, returns a closed peer.
func (r *Registry) Get(addr Address) *Peer {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if p, i := r.find(addr); i != -1 {
		return p
	}

	// Create "nonexistent" peer (nil connection).
	peer := r.addPeer(addr, nil)

	// Dial the peer in the background.
	go r.authenticatedDial(peer, addr)

	return peer
}

// Register registers a peer in the registry.
// If a peer with the same perun address already exists, it is returned,
// initialized with the given connection, if it did not already have a
// connection. Otherwise, enters a new peer into the registry and returns it.
func (r *Registry) Register(addr Address, conn Conn) (peer *Peer) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if peer, _ = r.find(addr); peer == nil {
		return r.addPeer(addr, conn)
	}
	peer.create(conn)
	return peer
}

// NumPeers returns the current number of peers in the registry including
// placeholder peers (cf. Registry.Get).
func (r *Registry) NumPeers() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

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
func (r *Registry) addPeer(addr Address, conn Conn) *Peer {
	// Create and register a new peer.
	peer := newPeer(addr, conn, r.dialer)
	peer.OnClose(func() { r.delete(peer) })
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

// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package peer

import (
	"context"
	"sync"

	"perun.network/go-perun/log"
)

// Registry is a peer Registry.
// It should not be used manually, but only internally by the client.
type Registry struct {
	mutex sync.RWMutex
	peers []*Peer // The list of all of the registry's peers.

	dialer    Dialer      // Used for dialing peers (and later: repairing).
	subscribe func(*Peer) // Sets up peer subscriptions.
}

// NewRegistry creates a new registry.
// The provided callback is used to set up new peer's subscriptions and it is
// called before the peer starts receiving messages.
func NewRegistry(subscribe func(*Peer), dialer Dialer) *Registry {
	return &Registry{
		subscribe: subscribe,
		dialer:    dialer,
	}
}

func (r *Registry) Close() (err error) {
	if r.dialer != nil {
		r.mutex.Lock()
		err = r.dialer.Close()
		r.mutex.Unlock()
	}
	for len(r.peers) > 0 {
		// when peers are closed, they delete themselves from the registry via their
		// close hook, so we always just close the first peer.
		if cerr := r.peers[0].Close(); cerr != nil && err == nil {
			err = cerr // record first non-nil error
		}
	}
	return err
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

// Get looks up the peer via its perun address.
// If the peer does not exist yet, creates a placeholder peer and dials the
// requested address. When the dialling finishes, completes the peer or closes
// it, depending on the success of the dialing operation. The unfinished peer
// object can be used already, but it will block until the peer is finished or
// closed.
func (r *Registry) Get(addr Address) *Peer {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if p, i := r.find(addr); i != -1 {
		return p
	}

	// Create "nonexistent" peer (nil connection).
	peer := r.addPeer(addr, nil)

	// Dial the peer in the background.
	go func() {
		conn, err := r.dialer.Dial(context.Background(), addr)
		if err != nil {
			peer.Close()
		} else {
			peer.create(conn)
		}
	}()
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
	peer := newPeer(addr, conn, func(p *Peer) { r.delete(p) }, r.dialer)
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

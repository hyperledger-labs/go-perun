// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package net

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"

	"perun.network/go-perun/log"
	perunsync "perun.network/go-perun/pkg/sync"
	"perun.network/go-perun/wire"
)

// EndpointRegistry is a peer EndpointRegistry.
// It should not be used manually, but only internally by the client.
type EndpointRegistry struct {
	mutex sync.RWMutex
	peers []*Endpoint  // The list of all of the registry's peers.
	id    wire.Account // The identity of the node.

	dialer    Dialer          // Used for dialing peers (and later: repairing).
	subscribe func(*Endpoint) // Sets up peer subscriptions.

	log log.Logger
	perunsync.Closer
}

const exchangeAddrsTimeout = 10 * time.Second

// NewEndpointRegistry creates a new registry.
// The provided callback is used to set up new peer's subscriptions and it is
// called before the peer starts receiving messages.
func NewEndpointRegistry(id wire.Account, subscribe func(*Endpoint), dialer Dialer) *EndpointRegistry {
	return &EndpointRegistry{
		id:        id,
		subscribe: subscribe,
		dialer:    dialer,

		log: log.WithField("id", id.Address()),
	}
}

// Close closes the registry's dialer and all its peers.
func (r *EndpointRegistry) Close() (err error) {
	if err = r.Closer.Close(); err != nil {
		return
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	for _, p := range r.peers {
		// When peers are closed, they are lazily deleted from the registry during
		// Get, Has or NumPeers calls.
		if cerr := p.Close(); !perunsync.IsAlreadyClosedError(cerr) && cerr != nil && err == nil {
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
func (r *EndpointRegistry) Listen(listener Listener) {
	if !r.OnCloseAlways(func() {
		if err := listener.Close(); err != nil {
			r.log.Debugf("Registry.Listen: closing listener OnClose: %v", err)
		}
	}) {
		return
	}

	// Start listener and accept all incoming peer connections, writing them to
	// the registry.
	for {
		conn, err := listener.Accept()
		if err != nil {
			r.log.Debugf("Registry.Listen: Accept() loop: %v", err)
			return
		}

		r.log.Debug("Registry.Listen: setting up incoming connection")
		// setup connection in a separate routine so that new incoming
		// connections can immediately be handled.
		go r.setupConn(conn)
	}
}

// setupConn authenticates a fresh connection, and if successful, adds it to the
// registry.
func (r *EndpointRegistry) setupConn(conn Conn) error {
	timeout := time.Duration(exchangeAddrsTimeout)
	ctx, cancel := context.WithTimeout(r.Ctx(), timeout)
	defer cancel()

	r.mutex.Lock()
	unfinishedPeer := r.addPeer(nil, nil)
	r.mutex.Unlock()

	var peerAddr wire.Address
	var err error
	if peerAddr, err = ExchangeAddrsPassive(ctx, r.id, conn); err != nil {
		conn.Close()
		return errors.WithMessage(err, "could not authenticate peer")
	}
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if peer, _ := r.find(peerAddr); peer == nil {
		unfinishedPeer.PerunAddress = peerAddr
		unfinishedPeer.create(conn)
	} else {
		peer.create(conn)
		r.delete(unfinishedPeer)
	}
	return nil
}

// find looks up a peer via its Perun address.
// If found, returns the peer and its index, otherwise returns a nil peer.
// find is not thread safe and is assumed to be called from a method which has
// the r.mutex lock.
// While iterating over the peers, find removes closed ones.
func (r *EndpointRegistry) find(addr wire.Address) (*Endpoint, int) {
	for i, peer := range r.peers {
		if peer.PerunAddress != nil && peer.PerunAddress.Equals(addr) {
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
func (r *EndpointRegistry) prune() {
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
func (r *EndpointRegistry) Get(ctx context.Context, addr wire.Address) (*Endpoint, error) {
	log := r.log.WithField("peer", addr)
	log.Trace("Registry.Get")
	r.mutex.Lock()
	if p, i := r.find(addr); i != -1 {
		r.mutex.Unlock()
		log.Trace("Registry.Get: peer found, waiting for conn...")
		if !p.waitExists(ctx) {
			return nil, errors.New("peer was not created in time")
		}
		log.Trace("Registry.Get: peer connection established")
		return p, nil
	}

	log.Trace("Registry.Get: peer not found, dialing...")
	// Create "nonexistent" peer (nil connection).
	peer := r.addPeer(addr, nil)
	r.mutex.Unlock()

	if err := r.authenticatedDial(ctx, peer, addr); err != nil {
		peer.Close()
		return nil, errors.WithMessage(err, "failed to dial peer")
	}

	return peer, nil
}

func (r *EndpointRegistry) authenticatedDial(ctx context.Context, peer *Endpoint, addr wire.Address) error {
	conn, err := r.dialer.Dial(ctx, addr)

	if peer.exists() {
		if conn != nil {
			conn.Close()
		}
		return nil
	} else if err != nil {
		return errors.WithMessage(err, "failed to dial")
	}

	if err := ExchangeAddrsActive(ctx, r.id, addr, conn); err != nil {
		conn.Close()
		return errors.WithMessage(err, "ExchangeAddrs failed")
	}

	peer.create(conn)
	return nil
}

// NumPeers returns the current number of peers in the registry including
// placeholder peers (cf. Registry.Get).
func (r *EndpointRegistry) NumPeers() int {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.prune()
	l := 0
	for _, p := range r.peers {
		if p.PerunAddress != nil {
			l++
		}
	}
	return l
}

// Has return true if and only if there is a peer with the given address in the
// registry. The function does not differentiate between regular and
// placeholder peers.
func (r *EndpointRegistry) Has(addr wire.Address) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	p, _ := r.find(addr)

	return p != nil
}

// addPeer adds a new peer to the registry.
// addPeer is not thread safe and is assumed to be called from a method which has
// the r.mutex lock.
func (r *EndpointRegistry) addPeer(addr wire.Address, conn Conn) *Endpoint {
	r.log.WithField("peer", addr).Trace("Registry.addPeer")
	// Create and register a new peer.
	peer := newEndpoint(addr, conn, r.dialer)
	r.peers = append(r.peers, peer)
	// Setup the peer's subscriptions.
	r.subscribe(peer)
	// Start receiving messages.
	go peer.recvLoop()

	return peer
}

// delete deletes a peer from the registry.
// If the peer does not exist in the registry, panics. Does not close the peer.
// This function is not thread-safe and is assumed to only be called when the
// mutex lock is held.
func (r *EndpointRegistry) delete(peer *Endpoint) {
	for i := range r.peers {
		if r.peers[i] == peer {
			// Delete the i-th entry.
			r.peers[i] = r.peers[len(r.peers)-1]
			r.peers = r.peers[:len(r.peers)-1]
			return
		}
	}
	log.Panic("tried to delete non-existent peer!")
}

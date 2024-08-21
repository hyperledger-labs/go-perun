// Copyright 2019 - See NOTICE file for copyright holders.
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

package net

import (
	"context"
	"perun.network/go-perun/channel/persistence/test"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/pkg/errors"

	"perun.network/go-perun/log"
	"perun.network/go-perun/wire"
	perunsync "polycry.pt/poly-go/sync"
)

// dialingEndpoint is an endpoint that is being dialed, but has no connection
// associated with it yet.
type dialingEndpoint struct {
	Address   map[int]wire.Address // The Endpoint's address.
	created   chan struct{}        // Triggered when the Endpoint is created.
	createdAt *Endpoint            // Contains the finished Endpoint when it exists.
}

// fullEndpoint describes an endpoint that is held within the registry.
type fullEndpoint struct {
	endpoint unsafe.Pointer // *Endpoint
}

func (p *fullEndpoint) Endpoint() *Endpoint {
	return (*Endpoint)(atomic.LoadPointer(&p.endpoint))
}

func newFullEndpoint(e *Endpoint) *fullEndpoint {
	return &fullEndpoint{
		endpoint: unsafe.Pointer(e),
	}
}

func newDialingEndpoint(addr map[int]wire.Address) *dialingEndpoint {
	return &dialingEndpoint{
		Address: addr,
		created: make(chan struct{}),
	}
}

// EndpointRegistry is a peer Endpoint registry and manages the establishment of
// new connections and acts as a dictionary for looking up established
// connections. It should not be used manually, but only internally by a
// wire.Bus.
type EndpointRegistry struct {
	id            wire.Account                             // The identity of the node.
	dialer        Dialer                                   // Used for dialing peers.
	onNewEndpoint func(map[int]wire.Address) wire.Consumer // Selects Consumer for new Endpoints' receive loop.
	ser           wire.EnvelopeSerializer

	endpoints map[wire.AddrKey]*fullEndpoint // The list of all of all established Endpoints.
	dialing   map[wire.AddrKey]*dialingEndpoint
	mutex     sync.RWMutex // protects peers and dialing.

	log.Embedding
	perunsync.Closer
}

const exchangeAddrsTimeout = 10 * time.Second

// NewEndpointRegistry creates a new registry.
// The provided callback is used to set up new peer's subscriptions and it is
// called before the peer starts receiving messages.
func NewEndpointRegistry(
	id wire.Account,
	onNewEndpoint func(map[int]wire.Address) wire.Consumer,
	dialer Dialer,
	ser wire.EnvelopeSerializer,
) *EndpointRegistry {
	return &EndpointRegistry{
		id:            id,
		onNewEndpoint: onNewEndpoint,
		dialer:        dialer,
		ser:           ser,

		endpoints: make(map[wire.AddrKey]*fullEndpoint),
		dialing:   make(map[wire.AddrKey]*dialingEndpoint),

		Embedding: log.MakeEmbedding(log.WithField("id", id.Address())),
	}
}

// Close closes the registry's dialer and all its peers.
func (r *EndpointRegistry) Close() (err error) {
	if err = r.Closer.Close(); err != nil {
		return
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.dialer != nil {
		if cerr := r.dialer.Close(); cerr != nil {
			err = errors.WithMessage(cerr, "closing dialer")
		}
	}

	for _, p := range r.endpoints {
		e := p.Endpoint()
		if e == nil {
			continue
		}
		if cerr := e.Close(); cerr != nil && err == nil {
			err = errors.WithMessage(cerr, "closing peer")
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
			r.Log().Debugf("EndpointRegistry.Listen: closing listener OnClose: %v", err)
		}
	}) {
		return
	}

	// Start listener and accept all incoming peer connections, writing them to
	// the registry.
	for {
		conn, err := listener.Accept(r.ser)
		if err != nil {
			r.Log().Debugf("EndpointRegistry.Listen: Accept() loop: %v", err)
			return
		}

		r.Log().Debug("EndpointRegistry.Listen: setting up incoming connection")
		// setup connection in a separate routine so that new incoming
		// connections can immediately be handled.
		go func() {
			if err := r.setupConn(conn); err != nil {
				log.WithError(err).Error("EndpointRegistry could not setup wire/net.Conn")
			}
		}()
	}
}

// setupConn authenticates a fresh connection, and if successful, adds it to the
// registry.
func (r *EndpointRegistry) setupConn(conn Conn) error {
	ctx, cancel := context.WithTimeout(r.Ctx(), exchangeAddrsTimeout)
	defer cancel()

	var peerAddr map[int]wire.Address
	var err error
	if peerAddr, err = ExchangeAddrsPassive(ctx, r.id, conn); err != nil {
		conn.Close()
		r.Log().WithField("peer", peerAddr).Error("could not authenticate peer:", err)
		return err
	}

	if test.EqualWireMaps(peerAddr, r.id.Address()) {
		r.Log().Error("dialed by self")
		return errors.New("dialed by self")
	}

	r.addEndpoint(peerAddr, conn, false)
	return nil
}

// Endpoint looks up an Endpoint via its perun address. If the Endpoint does not
// exist yet, it is dialed. Does not return until the peer is dialed or the
// context is closed.
func (r *EndpointRegistry) Endpoint(ctx context.Context, addr map[int]wire.Address) (*Endpoint, error) {
	log := r.Log().WithField("peer", addr)
	key := wire.Keys(addr)

	if test.EqualWireMaps(addr, r.id.Address()) {
		log.Panic("tried to dial self")
	}

	log.Trace("EndpointRegistry.Get")

	r.mutex.Lock()
	fe, ok := r.endpoints[key]
	if ok {
		if e := fe.Endpoint(); e != nil {
			r.mutex.Unlock()
			log.Trace("EndpointRegistry.Get: peer connection established")
			return e, nil
		}
	}
	de, created := r.dialingEndpoint(addr)
	r.mutex.Unlock()

	log.Trace("EndpointRegistry.Get: peer not found, dialing...")

	e, err := r.authenticatedDial(ctx, addr, de, created)
	return e, errors.WithMessage(err, "failed to dial peer")
}

func (r *EndpointRegistry) authenticatedDial(
	ctx context.Context,
	addr map[int]wire.Address,
	de *dialingEndpoint,
	created bool,
) (ret *Endpoint, _ error) {
	key := wire.Keys(addr)

	// Short cut: another dial for that peer is already in progress.
	if !created {
		select {
		case <-r.Ctx().Done():
			return nil, errors.New("waiting for dial, registry closed")
		case <-ctx.Done():
			return nil, errors.Wrap(ctx.Err(), "waiting for dial, context")
		case <-de.created:
			if de.createdAt == nil {
				return nil, errors.New("waiting for dial, dial failed")
			}
			return de.createdAt, nil
		}
	}

	// Clean up the entry at the end of the operation.
	defer func() {
		r.mutex.Lock()
		defer r.mutex.Unlock()
		delete(r.dialing, key)
		de.createdAt = ret
		close(de.created)
	}()

	conn, err := r.dialer.Dial(ctx, addr, r.ser)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to dial")
	}

	if err := ExchangeAddrsActive(ctx, r.id, addr, conn); err != nil {
		conn.Close()
		return nil, errors.WithMessage(err, "ExchangeAddrs failed")
	}

	return r.addEndpoint(addr, conn, true), nil
}

// dialingEndpoint retrieves or creates a dialingEndpoint for the passed address.
func (r *EndpointRegistry) dialingEndpoint(a map[int]wire.Address) (_ *dialingEndpoint, created bool) {
	key := wire.Keys(a)
	entry, ok := r.dialing[key]
	if !ok {
		entry = newDialingEndpoint(a)
		r.dialing[key] = entry
	}

	return entry, !ok
}

// NumPeers returns the current number of peers in the registry including
// placeholder peers (cf. Registry.Get).
func (r *EndpointRegistry) NumPeers() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return len(r.endpoints)
}

// Has return true if and only if there is a peer with the given address in the
// registry. The function does not differentiate between regular and
// placeholder peers.
func (r *EndpointRegistry) Has(addr map[int]wire.Address) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	_, ok := r.endpoints[wire.Keys(addr)]

	return ok
}

// addEndpoint adds a new peer to the registry.
func (r *EndpointRegistry) addEndpoint(addr map[int]wire.Address, conn Conn, dialer bool) *Endpoint {
	r.Log().WithField("peer", addr).Trace("EndpointRegistry.addEndpoint")

	e := newEndpoint(addr, conn)
	fe, created := r.fullEndpoint(addr, e)
	if !created {
		if e, closed := fe.replace(e, r.id.Address(), dialer); closed {
			return e
		}
	}

	consumer := r.onNewEndpoint(addr)
	// Start receiving messages.
	go func() {
		if err := e.recvLoop(consumer); err != nil {
			r.Log().WithError(err).Error("recvLoop finished unexpectedly")
		}
		fe.delete(e)
	}()

	return e
}

// fullEndpoint retrieves or creates a fullEndpoint for the passed address.
func (r *EndpointRegistry) fullEndpoint(addr map[int]wire.Address, e *Endpoint) (_ *fullEndpoint, created bool) {
	key := wire.Keys(addr)
	r.mutex.Lock()
	defer r.mutex.Unlock()
	entry, ok := r.endpoints[key]
	if !ok {
		entry = newFullEndpoint(e)
		r.endpoints[key] = entry
	}
	return entry, !ok
}

// replace sets a new endpoint and resolves ties when both parties dial each
// other concurrently. It returns the endpoint that is selected after potential
// tie resolving, and whether the supplied endpoint was closed in the process.
func (p *fullEndpoint) replace(newValue *Endpoint, self map[int]wire.Address, dialer bool) (updated *Endpoint, closed bool) {
	// If there was no previous endpoint, just set the new one.
	wasNil := atomic.CompareAndSwapPointer(&p.endpoint, nil, unsafe.Pointer(newValue))
	if wasNil {
		return newValue, false
	}

	// If an endpoint already exists, we are in a race where both parties dialed
	// each other concurrently. Deterministically select the same connection to
	// close on both sides. Close the endpoint that is created by the dialer
	// with the lesser Perun address and return the previously existing
	// endpoint.
	for key, selfAddr := range self {
		// Check if the same key exists in newValue.Address
		newAddr, exists := newValue.Address[key]

		// If the key does not exist in newValue.Address, you might skip it or handle it
		if !exists {
			continue // or handle this scenario according to your requirements
		}

		// Compare the addresses
		if dialer == (selfAddr.Cmp(newAddr) < 0) {
			// If selfAddr is "lesser", close the new value
			if err := newValue.Close(); err != nil {
				log.Warn("newValue dialer already closed")
			}
			// Return the existing endpoint associated with this key
			return p.Endpoint(), true
		}
	}

	// Otherwise, install the new endpoint and close the old endpoint.
	old := atomic.SwapPointer(&p.endpoint, unsafe.Pointer(newValue))
	if old != nil {
		// It may be possible that in the meanwhile, the peer might have been
		// replaced by another goroutine.
		if err := (*Endpoint)(old).Close(); err != nil {
			log.Warn("Old Endpoint was already closed")
		}
	}

	return newValue, false
}

// delete deletes an endpoint if it was not replaced previously.
func (p *fullEndpoint) delete(expectedOldValue *Endpoint) {
	atomic.CompareAndSwapPointer(&p.endpoint, unsafe.Pointer(expectedOldValue), nil)
}

func (r *EndpointRegistry) find(addr map[int]wire.Address) *Endpoint {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	if e, ok := r.endpoints[wire.Keys(addr)]; ok {
		return e.Endpoint()
	}
	return nil
}

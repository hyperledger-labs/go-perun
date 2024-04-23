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

package simple

import (
	"crypto/tls"
	"time"

	"github.com/pkg/errors"

	"perun.network/go-perun/wire"
	"polycry.pt/poly-go/sync"
)

// ConnHub is a factory for creating and connecting test dialers and listeners.
type ConnHub struct {
	listeners      map[wire.Address]*Listener
	listenersMutex sync.Mutex // Mutex for managing listeners
	dialers        []*Dialer
	dialersMutex   sync.Mutex // Mutex for managing dialers

	sync.Closer
}

// NewConnHub initializes and returns a new ConnHub instance.
func NewConnHub() *ConnHub {
	return &ConnHub{
		listeners: make(map[wire.Address]*Listener),
	}
}

// NewNetListener creates a new listener for the given address.
// Registers the new listener in the hub. Panics if the address was already
// entered or the hub is closed.
func (h *ConnHub) NewNetListener(addr wire.Address, host string, config *tls.Config) *Listener {
	h.listenersMutex.Lock()
	defer h.listenersMutex.Unlock()

	if h.IsClosed() {
		panic("ConnHub already closed")
	}

	listener, err := NewTCPListener(host, config)
	if err != nil {
		panic(errors.WithMessage(err, "failed to create listener"))
	}

	if _, exists := h.listeners[addr]; exists {
		panic("double registration")
	}

	h.listeners[addr] = listener
	return listener
}

// NewNetDialer creates a new dialer.
// Registers the new dialer in the hub. Panics if the hub is closed.
func (h *ConnHub) NewNetDialer(defaultTimeout time.Duration, tlsConfig *tls.Config) *Dialer {
	h.dialersMutex.Lock()
	defer h.dialersMutex.Unlock()

	if h.IsClosed() {
		panic("ConnHub already closed")
	}

	dialer := NewTCPDialer(defaultTimeout, tlsConfig)
	h.dialers = append(h.dialers, dialer)
	dialer.hub = h

	return dialer
}

// Close closes the ConnHub and all its listeners.
func (h *ConnHub) Close() (err error) {
	h.listenersMutex.Lock()
	defer h.listenersMutex.Unlock()

	h.dialersMutex.Lock()
	defer h.dialersMutex.Unlock()

	if h.IsClosed() {
		return errors.New("ConnHub already closed")
	}

	for _, d := range h.dialers {
		if cerr := d.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}

	h.dialers = nil

	for _, l := range h.listeners {
		if cerr := l.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}

	h.listeners = nil

	if err := h.Closer.Close(); err != nil {
		return errors.WithMessage(err, "ConnHub already closed")
	}

	return
}

// findEntry is not mutexed, and is only to be called from within the type's
// other functions.
func (h *ConnHub) findListenerEntry(key wire.Address) (*Listener, wire.Address, bool) {
	if h.IsClosed() {
		panic("ConnHub already closed")
	}

	for k, v := range h.listeners {
		if k.Equal(key) {
			return v, k, true
		}
	}

	return nil, nil, false
}

func (h *ConnHub) findListener(key wire.Address) (*Listener, bool) {
	if h.IsClosed() {
		panic("ConnHub already closed")
	}

	if e, _, ok := h.findListenerEntry(key); ok {
		return e, true
	}
	return nil, false
}

func (h *ConnHub) eraseDialer(dialer *Dialer) error {
	if h.IsClosed() {
		panic("ConnHub already closed")
	}
	for i, d := range h.dialers {
		if d == dialer {
			h.dialers[i] = h.dialers[len(h.dialers)-1]
			h.dialers = h.dialers[:len(h.dialers)-1]
			return nil
		}
	}

	return errors.New("dialer does not exist")
}

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
	gosync "sync"
	"time"

	"github.com/pkg/errors"

	"perun.network/go-perun/wire"
	"polycry.pt/poly-go/sync"
)

// DefaultTimeout is the default timeout for dialers.
const DefaultTimeout = 10 * time.Millisecond

// listenerMapEntry is a key-value entry inside a listener map.
type listenerMapEntry struct {
	key   wire.Address
	value *Listener
}

// ConnHub is a factory for creating and connecting test dialers and listeners.
type ConnHub struct {
	mutex     gosync.RWMutex
	listeners []listenerMapEntry
	dialers   []*Dialer

	sync.Closer
}

// NewNetListener creates a new listener for the given address.
// Registers the new listener in the hub. Panics if the address was already
// entered or the hub is closed.
func (h *ConnHub) NewNetListener(addr wire.Address, host string, config *tls.Config) *Listener {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.IsClosed() {
		panic("ConnHub already closed")
	}

	listener, err := NewTCPListener(host, config)
	if err != nil {
		panic(errors.WithMessage(err, "failed to create listener"))
	}

	if err := h.insertListener(addr, listener); err != nil {
		panic("double registration")
	}

	listener.OnClose(func() {
		h.eraseListener(addr) //nolint:errcheck
	})

	return listener
}

// NewNetDialer creates a new dialer.
// Registers the new dialer in the hub. Panics if the hub is closed.
func (h *ConnHub) NewNetDialer(defaultTimeout time.Duration, tlsConfig *tls.Config) *Dialer {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.IsClosed() {
		panic("ConnHub already closed")
	}

	dialer := NewTCPDialer(defaultTimeout, tlsConfig)
	h.insertDialer(dialer)
	dialer.hub = h

	return dialer
}

// Close closes the ConnHub and all its listeners.
func (h *ConnHub) Close() (err error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	for _, l := range h.listeners {
		if cerr := l.value.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}

	h.listeners = nil

	for _, d := range h.dialers {
		if cerr := d.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}

	h.dialers = nil

	if err := h.Closer.Close(); err != nil {
		return errors.WithMessage(err, "ConnHub already closed")
	}

	return
}

// findEntry is not mutexed, and is only to be called from within the type's
// other functions.
func (h *ConnHub) findListenerEntry(key wire.Address) (listenerMapEntry, int, bool) {
	if h.IsClosed() {
		panic("ConnHub already closed")
	}

	for i, v := range h.listeners {
		if v.key.Equal(key) {
			return v, i, true
		}
	}

	return listenerMapEntry{}, -1, false
}

func (h *ConnHub) findListener(key wire.Address) (*Listener, bool) {
	if h.IsClosed() {
		panic("ConnHub already closed")
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if e, _, ok := h.findListenerEntry(key); ok {
		return e.value, true
	}
	return nil, false
}

func (h *ConnHub) insertListener(key wire.Address, value *Listener) error {
	if h.IsClosed() {
		panic("ConnHub already closed")
	}
	if _, _, ok := h.findListenerEntry(key); ok {
		return errors.New("tried to re-insert existing key")
	}
	h.listeners = append(h.listeners, listenerMapEntry{key, value})
	return nil
}

func (h *ConnHub) eraseListener(key wire.Address) error {
	if h.IsClosed() {
		panic("ConnHub already closed")
	}
	if _, i, ok := h.findListenerEntry(key); ok {
		h.listeners[i] = h.listeners[len(h.listeners)-1]
		h.listeners = h.listeners[:len(h.listeners)-1]
		return nil
	}
	return errors.New("tried to erase nonexistent entry")
}

func (h *ConnHub) insertDialer(dialer *Dialer) {
	if h.IsClosed() {
		panic("ConnHub already closed")
	}
	h.dialers = append(h.dialers, dialer)
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

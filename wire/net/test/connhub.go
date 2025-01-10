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

package test

import (
	gosync "sync"

	"perun.network/go-perun/wallet"

	"github.com/pkg/errors"

	"perun.network/go-perun/wire"
	"polycry.pt/poly-go/sync"
)

// ConnHub is a factory for creating and connecting test dialers and listeners.
type ConnHub struct {
	mutex gosync.RWMutex
	listenerMap
	dialers dialerList

	sync.Closer
}

// NewNetListener creates a new test listener for the given address.
// Registers the new listener in the hub. Panics if the address was already
// entered or the hub is closed.
func (h *ConnHub) NewNetListener(addr map[wallet.BackendID]wire.Address) *Listener {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if h.IsClosed() {
		panic("ConnHub already closed")
	}

	listener := NewNetListener()
	if err := h.insert(addr, listener); err != nil {
		panic("double registration")
	}

	// Remove the listener from the hub after it's closed.
	listener.OnClose(func() {
		h.erase(addr) //nolint:errcheck
	})
	return listener
}

// NewNetDialer creates a new test dialer.
// Registers the new dialer in the hub. Panics if the hub is closed.
func (h *ConnHub) NewNetDialer() *Dialer {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if h.IsClosed() {
		panic("ConnHub already closed")
	}

	dialer := NewDialer(h)
	h.dialers.insert(dialer)
	dialer.OnClose(func() {
		h.dialers.erase(dialer) //nolint:errcheck
	})

	return dialer
}

// Close closes the ConnHub and all its listeners.
func (h *ConnHub) Close() (err error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if err := h.Closer.Close(); err != nil {
		return errors.WithMessage(err, "ConnHub already closed")
	}

	for _, l := range h.clear() {
		if cerr := l.value.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}

	for _, d := range h.dialers.clear() {
		if cerr := d.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}

	return
}

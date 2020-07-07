// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package test

import (
	gosync "sync"

	"github.com/pkg/errors"

	"perun.network/go-perun/pkg/sync"
	"perun.network/go-perun/wire"
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
func (h *ConnHub) NewNetListener(addr wire.Address) *Listener {
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
	listener.OnClose(func() { h.erase(addr) })

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

	dialer := &Dialer{hub: h}
	h.dialers.insert(dialer)
	dialer.OnClose(func() { h.dialers.erase(dialer) })

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

// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package test

import (
	gosync "sync"

	"github.com/pkg/errors"

	"perun.network/go-perun/peer"
	"perun.network/go-perun/pkg/sync"
)

// ConnHub is a factory for creating and connecting test dialers and listeners.
type ConnHub struct {
	mutex gosync.RWMutex
	listenerMap
	dialers dialerList

	sync.Closer
}

// Create creates a new test dialer and test listener for the given identity.
// Registers the new listener in the hub. Fails if the address was already
// entered or the hub is closed.
func (h *ConnHub) Create(addr peer.Address) (peer.Dialer, peer.Listener, error) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if h.IsClosed() {
		return nil, nil, errors.WithMessage(h.Closer.Close(), "ConnHub already closed")
	}

	listener := NewListener()
	if err := h.insert(addr, listener); err != nil {
		return nil, nil, errors.New("double registration")
	}

	// Remove the listener from the hub after it's closed.
	listener.OnClose(func() { h.erase(addr) })

	dialer := &Dialer{hub: h}
	h.dialers.insert(dialer)
	dialer.OnClose(func() { h.dialers.erase(dialer) })
	return dialer, listener, nil
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

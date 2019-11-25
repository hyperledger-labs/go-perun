// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package test

import (
	"github.com/pkg/errors"

	"perun.network/go-perun/peer"
)

// ConnHub is a factory for creating and connecting test dialers and listeners.
type ConnHub struct {
	listenerMap
}

// Creates a new test dialer and test listener for the given identity.
// Registers the new listener in the hub. Fails if the address was already
// entered.
func (h *ConnHub) Create(id peer.Identity) (peer.Dialer, peer.Listener, error) {
	listener := NewListener()
	if err := h.insert(id.Address(), listener); err != nil {
		return nil, nil, errors.New("double registration")
	}

	// Remove the listener from the hub after it's closed.
	listener.OnClose(func() { h.erase(id.Address()) })

	dialer := &Dialer{hub: h, identity: id}
	return dialer, listener, nil
}

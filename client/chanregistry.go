// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package client

import (
	"sync"

	"perun.network/go-perun/channel"
	psync "perun.network/go-perun/pkg/sync"
)

// ChanRegistry is a registry for channels.
// You can safely look up channels via their ID and concurrently modify the
// registry. Always initialize instances of this type with MakeChanRegistry().
type chanRegistry struct {
	mutex  sync.RWMutex
	values map[channel.ID]*Channel
}

// makeChanRegistry creates a new empty channel registry.
func makeChanRegistry() chanRegistry {
	return chanRegistry{values: make(map[channel.ID]*Channel)}
}

// Put puts a new channel into the registry.
// If an entry with the same ID already existed, this call does nothing and
// returns false. Otherwise, it adds the new channel into the registry and
// returns true.
func (r *chanRegistry) Put(id channel.ID, value *Channel) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, ok := r.values[id]; ok {
		return false
	}
	r.values[id] = value
	return true
}

// Has checks whether a channel with the requested ID is registered.
func (r *chanRegistry) Has(id channel.ID) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	_, ok := r.values[id]
	return ok
}

// Get retrieves a channel from the registry.
// If the channel exists, returns the channel, and true. Otherwise, returns nil,
// false.
func (r *chanRegistry) Get(id channel.ID) (*Channel, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	v, ok := r.values[id]
	return v, ok
}

// Delete deletes a channel from the registry.
// If the channel did not exist, does nothing. Returns whether the channel
// existed.
func (r *chanRegistry) Delete(id channel.ID) (deleted bool) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, deleted = r.values[id]; deleted {
		delete(r.values, id)
	}
	return
}

func (r *chanRegistry) CloseAll() (err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for _, c := range r.values {
		if cerr := c.Close(); err == nil && !psync.IsAlreadyClosedError(cerr) {
			err = cerr
		}
	}

	return err
}

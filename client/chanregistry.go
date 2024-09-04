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

package client

import (
	"sync"

	"perun.network/go-perun/channel"
	psync "polycry.pt/poly-go/sync"
)

// chanRegistry is a registry for channels.
// You can safely look up channels via their ID and concurrently modify the
// registry. Always initialize instances of this type with MakeChanRegistry().
type chanRegistry struct {
	mutex             sync.RWMutex
	values            map[string]*Channel
	newChannelHandler func(*Channel)
}

// makeChanRegistry creates a new empty channel registry.
func makeChanRegistry() chanRegistry {
	return chanRegistry{values: make(map[string]*Channel)}
}

// Put puts a new channel into the registry.
// If an entry with the same ID already existed, this call does nothing and
// returns false. Otherwise, it adds the new channel into the registry and
// returns true.
func (r *chanRegistry) Put(id map[int]channel.ID, value *Channel) bool {
	r.mutex.Lock()

	if _, ok := r.values[channel.IDKey(id)]; ok {
		r.mutex.Unlock()
		return false
	}
	r.values[channel.IDKey(id)] = value
	handler := r.newChannelHandler
	r.mutex.Unlock()
	value.OnCloseAlways(func() { r.Delete(id) })
	if handler != nil {
		handler(value)
	}
	return true
}

// OnNewChannel sets a callback to be called whenever a new channel is added to
// the registry via Put. Only one such handler can be set at a time, and
// repeated calls to this function will overwrite the currently existing
// handler. This function may be safely called at any time.
func (r *chanRegistry) OnNewChannel(handler func(*Channel)) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.newChannelHandler = handler
}

// Has checks whether a channel with the requested ID is registered.
func (r *chanRegistry) Has(id map[int]channel.ID) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	_, ok := r.values[channel.IDKey(id)]
	return ok
}

// Channel retrieves a channel from the registry.
// If the channel exists, returns the channel, and true. Otherwise, returns nil,
// false.
func (r *chanRegistry) Channel(id map[int]channel.ID) (*Channel, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	v, ok := r.values[channel.IDKey(id)]
	return v, ok
}

// Delete deletes a channel from the registry.
// If the channel did not exist, does nothing. Returns whether the channel
// existed.
func (r *chanRegistry) Delete(id map[int]channel.ID) (deleted bool) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, deleted = r.values[channel.IDKey(id)]; deleted {
		delete(r.values, channel.IDKey(id))
	}
	return
}

func (r *chanRegistry) CloseAll() (err error) {
	r.mutex.Lock()
	values := r.values
	r.values = make(map[string]*Channel)
	r.mutex.Unlock()

	for _, c := range values {
		if cerr := c.Close(); err == nil && !psync.IsAlreadyClosedError(cerr) {
			err = cerr
		}
	}

	return err
}

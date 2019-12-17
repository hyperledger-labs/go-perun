// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package test

import (
	"sync"

	"github.com/pkg/errors"

	"perun.network/go-perun/peer"
)

// listenerMapEntry is a key-value entry inside a listener map.
type listenerMapEntry struct {
	key   peer.Address
	value *Listener
}

// listenerMap is a peer.Address -> *Listener mapping.
type listenerMap struct {
	mutex   sync.RWMutex
	entries []listenerMapEntry
}

// findEntry is not mutexed, and is only to be called from within the type's
// other functions.
func (m *listenerMap) findEntry(key peer.Address) (listenerMapEntry, int, bool) {
	for i, v := range m.entries {
		if v.key.Equals(key) {
			return v, i, true
		}
	}

	return listenerMapEntry{}, -1, false
}

func (m *listenerMap) find(key peer.Address) (*Listener, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if e, _, ok := m.findEntry(key); ok {
		return e.value, true
	}
	return nil, false
}

func (m *listenerMap) insert(key peer.Address, value *Listener) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if _, _, ok := m.findEntry(key); ok {
		return errors.New("tried to re-insert existing key")
	}
	m.entries = append(m.entries, listenerMapEntry{key, value})
	return nil
}

func (m *listenerMap) erase(key peer.Address) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, i, ok := m.findEntry(key); ok {
		m.entries[i] = m.entries[len(m.entries)-1]
		m.entries = m.entries[:len(m.entries)-1]
		return nil
	} else {
		return errors.New("Tried to erase nonexistent entry.")
	}
}

func (m *listenerMap) clear() []listenerMapEntry {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	ret := m.entries
	m.entries = nil
	return ret
}

// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

// Package memorydb provides an implementation of the sortedkv interfaces. The main
// type, Database, is an in-memory key-value store. Since the database is not
// persistent, the package is not suited for production use, and more suited
// for simplifying tests and mockups. The database is thread-safe.
//
// Constructors
//
// The NewDatabase() constructor creates a new empty database. The FromData()
// constructor takes a key-value mapping and uses that as the database's
// contents.
package memorydb // import "perun.network/go-perun/pkg/sortedkv/memorydb"

import (
	"perun.network/go-perun/pkg/sortedkv"

	"sort"
	"strings"
	"sync"
)

// Database implements the Database interface and stores the values in memory.
type Database struct {
	mutex sync.RWMutex
	data  map[string]string
}

// NewDatabase creates a new, empty Database.
func NewDatabase() sortedkv.Database {
	return &Database{
		data: make(map[string]string),
	}
}

// FromData creates a Database from a map of values.
// The provided data will not be cloned. If data is nil, an empty database is
// created.
func FromData(data map[string]string) sortedkv.Database {
	if data == nil {
		data = make(map[string]string)
	}

	return &Database{
		data: data,
	}
}

// Reader interface.

// Has returns true if the memorydb contains a key.
func (d *Database) Has(key string) (bool, error) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	_, exists := d.data[key]
	return exists, nil
}

// Get returns a value to a key.
func (d *Database) Get(key string) (string, error) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	value, exists := d.data[key]
	if !exists {
		return "", &sortedkv.ErrNotFound{Key: key}
	}
	return value, nil
}

// GetBytes returns a value to a key in bytes.
func (d *Database) GetBytes(key string) ([]byte, error) {
	value, err := d.Get(key)
	return []byte(value), err
}

// Writer interface.

// Put saves a value under a key.
func (d *Database) Put(key string, value string) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.data[key] = value
	return nil
}

// PutBytes saves a bytes value under a key.
func (d *Database) PutBytes(key string, value []byte) error {
	return d.Put(key, string(value))
}

// Delete deletes a key from the database.
func (d *Database) Delete(key string) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if _, has := d.data[key]; !has {
		return &sortedkv.ErrNotFound{Key: key}
	}
	delete(d.data, key)
	return nil
}

// Batcher interface.

// NewBatch creates a new batch.
func (d *Database) NewBatch() sortedkv.Batch {
	batch := Batch{db: d}
	batch.Reset()
	return &batch
}

// Iterateable interface.

// NewIterator creates a new iterator.
func (d *Database) NewIterator() sortedkv.Iterator {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	keys := make([]string, 0, len(d.data))
	for key := range d.data {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return &Iterator{
		keys:   keys,
		values: d.readValues(keys),
	}
}

// NewIteratorWithRange creates a new iterator based on a given range.
func (d *Database) NewIteratorWithRange(start string, end string) sortedkv.Iterator {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	var keys []string
	// No need to check for start == "", as all strings >= "".
	if end == "" {
		for key := range d.data {
			if key >= start {
				keys = append(keys, key)
			}
		}
	} else {
		for key := range d.data {
			if key >= start && key < end {
				keys = append(keys, key)
			}
		}
	}

	sort.Strings(keys)
	return &Iterator{
		keys:   keys,
		values: d.readValues(keys),
	}
}

// NewIteratorWithPrefix creates a new iterator for a given prefix.
func (d *Database) NewIteratorWithPrefix(prefix string) sortedkv.Iterator {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	var keys []string
	for key := range d.data {
		if strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
		}
	}

	sort.Strings(keys)
	return &Iterator{
		keys:   keys,
		values: d.readValues(keys),
	}
}

// readValues reads the values matched to a set of keys from a database.
// The database must be readlocked already.
func (d *Database) readValues(keys []string) []string {
	data := make([]string, 0, len(keys))
	for key := range keys {
		data = append(data, d.data[keys[key]])
	}

	return data
}

// Closer interface

// Close clears the database.
func (d *Database) Close() error {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.data = nil
	return nil
}

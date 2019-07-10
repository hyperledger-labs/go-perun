// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

/*
	Package memorydb provides an implementation of the db interfaces. The main
	type, Database, is an in-memory key-value store. Since the database is not
	persistent, the package is not suited for production use, and more suited
	for simplifying tests and mockups. The database is thread-safe.

	Constructors

	The NewDatabase() constructor creates a new empty database. The FromData()
	constructor takes a key-value mapping and uses that as the database's
	contents.
*/
package memorydb // import "perun.network/go-perun/db/memorydb"

import (
	"perun.network/go-perun/db"

	"sort"
	"strings"
	"sync"
)

// Implementation of the Database interface that stores the values in memory.
type Database struct {
	mutex sync.RWMutex
	data  map[string]string
}

// Creates a new, empty Database.
func NewDatabase() db.Database {
	return &Database{
		data: make(map[string]string),
	}
}

/*
	FromData creates a Database from a map of values.
	The provided data will not be cloned. If data is nil, an empty database is
	created.
*/
func FromData(data map[string]string) db.Database {
	if data == nil {
		data = make(map[string]string)
	}

	return &Database{
		data: data,
	}
}

// Reader interface.

func (this *Database) Has(key string) (bool, error) {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	_, exists := this.data[key]
	return exists, nil
}

func (this *Database) Get(key string) (string, error) {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	value, exists := this.data[key]
	if !exists {
		return "", &db.ErrNotFound{Key: key}
	} else {
		return value, nil
	}
}

func (this *Database) GetBytes(key string) ([]byte, error) {
	value, err := this.Get(key)
	return []byte(value), err
}

// Writer interface.

func (this *Database) Put(key string, value string) error {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.data[key] = value
	return nil
}

func (this *Database) PutBytes(key string, value []byte) error {
	return this.Put(key, string(value))
}

func (this *Database) Delete(key string) error {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if _, has := this.data[key]; !has {
		return &db.ErrNotFound{Key: key}
	} else {
		delete(this.data, key)
		return nil
	}
}

// Batcher interface.

func (this *Database) NewBatch() db.Batch {
	batch := Batch{db: this}
	batch.Reset()
	return &batch
}

// Iterateable interface.

func (this *Database) NewIterator() db.Iterator {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	keys := make([]string, 0, len(this.data))
	for key := range this.data {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return &Iterator{
		keys:   keys,
		values: this.readValues(keys),
	}
}

func (this *Database) NewIteratorWithRange(start string, end string) db.Iterator {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	var keys []string
	// No need to check for start == "", as all strings >= "".
	if end == "" {
		for key := range this.data {
			if key >= start {
				keys = append(keys, key)
			}
		}
	} else {
		for key := range this.data {
			if key >= start && key < end {
				keys = append(keys, key)
			}
		}
	}

	sort.Strings(keys)
	return &Iterator{
		keys:   keys,
		values: this.readValues(keys),
	}
}

func (this *Database) NewIteratorWithPrefix(prefix string) db.Iterator {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	prefixString := string(prefix)
	var keys []string
	for key := range this.data {
		if strings.HasPrefix(key, prefixString) {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return &Iterator{
		keys:   keys,
		values: this.readValues(keys),
	}
}

/*
	Reads the values matched to a set of keys from a database.
	The database must be readlocked already.
*/
func (this *Database) readValues(keys []string) []string {
	data := make([]string, 0, len(keys))
	for key := range keys {
		data = append(data, this.data[keys[key]])
	}

	return data
}

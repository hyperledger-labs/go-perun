// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package memorydb

import (
	"github.com/perun-network/go-perun/db"
	"github.com/pkg/errors"

	"sort"
	"strconv"
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

	if has, _ := this.Has(key); has {
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

func (this *Database) NewIteratorWithStart(start string) db.Iterator {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	startString := string(start)
	keys := make([]string, 0, len(this.data))
	for key := range this.data {
		if key >= startString {
			keys = append(keys, key)
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
	keys := make([]string, 0, len(this.data))
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

func (this *Database) readValues(keys []string) []string {
	data := make([]string, 0, len(keys))
	for key := range keys {
		data = append(data, this.data[keys[key]])
	}

	return data
}

// PropertyProvider interface.

func (this *Database) Property(property string) (string, error) {
	switch property {
	case "count":
		return strconv.Itoa(len(this.data)), nil
	case "valuesize":
		{
			size := 0
			for key := range this.data {
				size += len(this.data[key])
			}
			return strconv.Itoa(len(this.data)), nil
		}
	case "type":
		return "memorydb", nil
	default:
		return "", errors.New("Property(): Unknown property '" + property + "'")
	}
}

func (this *Database) DefaultProperties() (map[string]string, error) {
	return db.Properties(this, []string{"count", "valuesize", "type"})
}

// Compacter interface.

func (this *Database) Compact(start, end string) error {
	return nil
}

// io.Closer interface.

func (this *Database) Close() error {
	this.data = nil
	return nil
}

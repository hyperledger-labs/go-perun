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
	vmutex sync.RWMutex
	data   map[string][]byte
}

// Creates a new, empty Database.
func NewDatabase() db.Database {
	return &Database{
		data: make(map[string][]byte),
	}
}

/*
	FromData creates a Database from a map of values.
	The provided data will not be cloned. If data is nil, an empty database is
	created.
*/
func FromData(data map[string][]byte) db.Database {
	if data == nil {
		data = make(map[string][]byte)
	}

	return &Database{
		data: data,
	}
}

// Reader interface.

func (this *Database) Has(key []byte) (bool, error) {
	this.vmutex.RLock()
	defer this.vmutex.RUnlock()

	_, exists := this.data[string(key)]
	return exists, nil
}

func (this *Database) Get(key []byte) ([]byte, error) {
	this.vmutex.RLock()
	defer this.vmutex.RUnlock()

	value, exists := this.data[string(key)]
	if !exists {
		return nil, errors.New("Requested nonexistent entry.")
	} else {
		return value, nil
	}
}

// Writer interface.

func (this *Database) Put(key []byte, value []byte) error {
	this.vmutex.Lock()
	defer this.vmutex.Unlock()

	this.data[string(key)] = value
	return nil
}

func (this *Database) Delete(key []byte) error {
	this.vmutex.Lock()
	defer this.vmutex.Unlock()

	if has, _ := this.Has(key); has {
		return errors.New("Tried to delete nonexistent entry.")
	} else {
		delete(this.data, string(key))
		return nil
	}
}

// Batcher interface.

func (this *Database) NewBatch() db.Batch {
	return &Batch{db: this}
}

// Iterateable interface.

func (this *Database) NewIterator() db.Iterator {
	this.vmutex.RLock()
	defer this.vmutex.RUnlock()

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

func (this *Database) NewIteratorWithStart(start []byte) db.Iterator {
	this.vmutex.RLock()
	defer this.vmutex.RUnlock()

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

func (this *Database) NewIteratorWithPrefix(prefix []byte) db.Iterator {
	this.vmutex.RLock()
	defer this.vmutex.RUnlock()

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

func (this *Database) readValues(keys []string) [][]byte {
	data := make([][]byte, 0, len(keys))
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

func (this *Database) Compact(start []byte, end []byte) error {
	return nil
}

// io.Closer interface.

func (this *Database) Close() error {
	this.data = nil
	return nil
}

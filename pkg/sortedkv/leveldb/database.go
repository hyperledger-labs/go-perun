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

package leveldb

import (
	"sync"

	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"

	"perun.network/go-perun/pkg/sortedkv"
)

// Database implements the Database interface and stores the values in memory.
type Database struct {
	*leveldb.DB
	path string
}

// LoadDatabase creates a new, empty Database.
func LoadDatabase(path string) (*Database, error) {
	db, err := leveldb.OpenFile(path, nil)

	if err != nil {
		return nil, errors.Wrap(err, "Database.LoadDatabase(path) could not open/create file")
	}

	return &Database{
		db,
		path,
	}, nil
}

// interface Reader

// Has returns true iff the memorydb contains a key.
func (d *Database) Has(key string) (bool, error) {
	has, err := d.DB.Has([]byte(key), nil)
	return has, errors.Wrap(err, "Database.Has(key) error")
}

// Get returns the value as string for given key if it is present in the store.
func (d *Database) Get(key string) (string, error) {
	val, err := d.GetBytes(key)
	return string(val), err
}

// GetBytes returns the value as []byte for given key if it is present in the store.
func (d *Database) GetBytes(key string) ([]byte, error) {
	val, err := d.DB.Get([]byte(key), nil)
	return val, errors.Wrap(err, "Database.Get(key) error")
}

// interface Writer

// Put inserts the given value into the key-value store.
// If the key is already present, it is overwritten and no error is returned.
func (d *Database) Put(key string, value string) error {
	return d.PutBytes(key, []byte(value))
}

// PutBytes inserts the given value into the key-value store.
// If the key is already present, it is overwritten and no error is returned.
func (d *Database) PutBytes(key string, value []byte) error {
	err := d.DB.Put([]byte(key), value, nil)
	return errors.Wrap(err, "Database.Put(key, value) error")
}

// Delete removes the key from the key-value store.
// If the key is not present, an error is returned.
func (d *Database) Delete(key string) error {
	has, err := d.DB.Has([]byte(key), nil)

	if err != nil {
		return errors.Wrap(err, "Database.Delete(key) error")
	}

	if !has {
		return errors.New("Database.Delete(key) error")
	}

	err = d.DB.Delete([]byte(key), nil)
	return errors.Wrap(err, "Database.Delete(key) error")
}

// Batcher interface.

// NewBatch creates a new batch.
func (d *Database) NewBatch() sortedkv.Batch {
	return &Batch{&leveldb.Batch{}, d.DB}
}

// Iterateable interface.

// NewIterator creates a new iterator.
func (d *Database) NewIterator() sortedkv.Iterator {
	return &Iterator{d.DB.NewIterator(&util.Range{Start: nil, Limit: nil}, nil), sync.Mutex{}}
}

// NewIteratorWithRange creates a new iterator based on a given range.
func (d *Database) NewIteratorWithRange(start string, end string) sortedkv.Iterator {
	var Start []byte
	var End []byte

	if len(start) != 0 {
		Start = []byte(start)
	}

	if len(end) != 0 {
		End = []byte(end)
	}

	return &Iterator{d.DB.NewIterator(&util.Range{Start: Start, Limit: End}, nil), sync.Mutex{}}
}

// NewIteratorWithPrefix creates a new iterator for a given prefix.
func (d *Database) NewIteratorWithPrefix(prefix string) sortedkv.Iterator {
	var slice *util.Range

	if len(prefix) != 0 {
		slice = util.BytesPrefix([]byte(prefix))
	}

	return &Iterator{d.DB.NewIterator(slice, nil), sync.Mutex{}}
}

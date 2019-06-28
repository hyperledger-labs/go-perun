// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package db

import "sync"

// Table is a wrapper around a database with a key prefix. All key access is
// automatically prefixed. Close() is a noop and properties are forwarded
// from the database.
type table struct {
	Database
	prefix      string
	compactOnce sync.Once // Lazily caclulate nextPrefix on first Compact call
	nextPrefix  string    // Compacter needs to know the next prefix as upper bound
}

func NewTable(db Database, prefix string) *table {
	return &table{
		Database: db,
		prefix:   prefix,
	}
}

func (t *table) pkey(key string) string {
	return t.prefix + key
}

// Has calls db.Has with the prefixed key
func (t *table) Has(key string) (bool, error) {
	return t.Database.Has(t.pkey(key))
}

// Get calls db.Get with the prefixed key
func (t *table) Get(key string) (string, error) {
	return t.Database.Get(t.pkey(key))
}

// GetBytes calls db.GetBytes with the prefixed key
func (t *table) GetBytes(key string) ([]byte, error) {
	return t.Database.GetBytes(t.pkey(key))
}

// Put calls db.Put with the prefixed key
func (t *table) Put(key, value string) error {
	return t.Database.Put(t.pkey(key), value)
}

// PutBytes calls db.PutBytes with the prefixed key
func (t *table) PutBytes(key string, value []byte) error {
	return t.Database.PutBytes(t.pkey(key), value)
}

// Delete calls db.Delete with the prefixed key
func (t *table) Delete(key string) error {
	return t.Database.Delete(t.pkey(key))
}

func (t *table) NewBatch() Batch {
	return &tableBatch{t.Database.NewBatch(), t.prefix}
}

func (t *table) NewIterator() Iterator {
	return t.Database.NewIteratorWithPrefix(t.prefix)
}

func (t *table) NewIteratorWithStart(start string) Iterator {
	return t.Database.NewIteratorWithStart(t.pkey(start))
}

func (t *table) NewIteratorWithPrefix(prefix string) Iterator {
	return t.Database.NewIteratorWithPrefix(t.pkey(prefix))
}

func (t *table) Compact(start, end string) error {
	// lazily caclulate next key once Compact is called for the first time
	t.compactOnce.Do(func() {
		t.nextPrefix = nextKey(t.prefix)
	})
	// if no end is specified, we need to set it to the key after the prefix
	if end == "" {
		end = t.nextPrefix
	} else {
		end = t.pkey(end)
	}

	return t.Database.Compact(t.pkey(start), end)
}

func (t *table) Close() error {
	return nil
}

func nextKey(key string) string {
	keyb := []byte(key)
	for i := len(keyb) - 1; i >= 0; i-- {
		// Increment current byte, stop if it doesn't overflow
		keyb[i]++
		if keyb[i] > 0 {
			break
		}
		// Character overflown, proceed to next or return "" if last
		if i == 0 {
			return ""
		}
	}
	return string(keyb)
}

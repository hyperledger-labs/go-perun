// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package db

// Table is a wrapper around a database with a key prefix. All key access is
// automatically prefixed. Close() is a noop and properties are forwarded
// from the database.
type table struct {
	Database
	prefix []byte
}

func NewTable(db Database, prefix []byte) *table {
	return &table{
		Database: db,
		prefix:   prefix,
	}
}

func (t *table) pkey(key []byte) []byte {
	return append(t.prefix, key...)
}

// Has calls db.Has with the prefixed key
func (t *table) Has(key []byte) (bool, error) {
	return t.Database.Has(t.pkey(key))
}

// Get calls db.Get with the prefixed key
func (t *table) Get(key []byte) ([]byte, error) {
	return t.Database.Get(t.pkey(key))
}

// Put calls db.Put with the prefixed key
func (t *table) Put(key, value []byte) error {
	return t.Database.Put(t.pkey(key), value)
}

// Delete calls db.Delete with the prefixed key
func (t *table) Delete(key []byte) error {
	return t.Database.Delete(t.pkey(key))
}

func (t *table) NewBatch() Batch {
	return &tableBatch{t.Database.NewBatch(), t.prefix}
}

func (t *table) NewIterator() Iterator {
	return t.Database.NewIteratorWithPrefix(t.prefix)
}

func (t *table) NewIteratorWithStart(start []byte) Iterator {
	return t.Database.NewIteratorWithStart(t.pkey(start))
}

func (t *table) NewIteratorWithPrefix(prefix []byte) Iterator {
	return t.Database.NewIteratorWithPrefix(t.pkey(prefix))
}

func (t *table) Compact(start, end []byte) error {
	// if no limit is specified, we need to set the first key after the
	// prefix
	if end == nil {
		end = []byte(t.prefix)
		for i := len(end) - 1; i >= 0; i-- {
			// Bump the current character, stopping if it doesn't overflow
			end[i]++
			if end[i] > 0 {
				break
			}
			// Character overflown, proceed to the next or nil if the last
			if i == 0 {
				end = nil
			}
		}
	} else {
		end = t.pkey(end)
	}

	return t.Database.Compact(t.pkey(start), end)
}

func (t *table) Close() error {
	return nil
}

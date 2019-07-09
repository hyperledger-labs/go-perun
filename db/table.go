// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package db

// Table is a wrapper around a database with a key prefix. All key access is
// automatically prefixed. Close() is a noop and properties are forwarded
// from the database.
type table struct {
	Database
	prefix string
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
	return newTableIterator(t.Database.NewIteratorWithPrefix(t.prefix), t)
}

func (t *table) NewIteratorWithRange(start string, end string) Iterator {
	start = t.pkey(start)
	if end == "" {
		end = IncrementPrefix(t.prefix)
	} else {
		end = t.pkey(end)
	}

	return newTableIterator(t.Database.NewIteratorWithRange(start, end), t)
}

func (t *table) NewIteratorWithPrefix(prefix string) Iterator {
	return newTableIterator(t.Database.NewIteratorWithPrefix(t.pkey(prefix)), t)
}

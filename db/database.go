// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

// Package db defines a key-value store abstraction.
// It is used by other persistence packages (like channeldb) to continuously
// save the state of the running program or load it upon startup.
// It is based on go-ethereum's https://github.com/ethereum/go-ethereum/ethdb
// and https://github.com/ethereum/go-ethereum/core/rawdb
// and is also inspired by perkeep.org/pkg/sorted
package db

import (
	"io"

	"github.com/pkg/errors"
)

var (
	ErrNotFound = errors.New("db: key not found")
)

// Reader wraps the Had and Get methods of a key-value store.
type Reader interface {
	// Has checks if a key is present in the store.
	Has(key string) (bool, error)

	// Get returns the value as string for given key if it is present in the store.
	Get(key string) (string, error)

	// Get returns the value as []byte for given key if it is present in the store.
	GetBytes(key string) ([]byte, error)
}

// Writer wraps the Put and Delete methods of a key-value store.
type Writer interface {
	// Put inserts the given value into the key-value store.
	// If the key is already present, it is overwritten and no error is returned.
	Put(key string, value string) error

	// PutBytes inserts the given value into the key-value store.
	// If the key is already present, it is overwritten and no error is returned.
	PutBytes(key string, value []byte) error

	// Delete removes the key from the key-value store.
	// If the key is not present, an error is returned
	Delete(key string) error
}

// PropertyProvider wraps the Property and Properties method of a database.
type PropertyProvider interface {
	/*
		Property looks up a property in the database.
		Requesting unknown properties results in an error.
	*/
	Property(property string) (string, error)

	/*
		DefaultProperties returns an implementation specific set of stats of the database.
		These stats can be useful for verbose logging.

		Example implementation (see memorydb):
			func (this *Database) DefaultProperties() (map[string]string, error) {
			    return db.Properties(this, []string{"count", "valuesize", "type"})
			}
	*/
	DefaultProperties() (map[string]string, error)
}

// Helper function to look up multiple properties at once.
func Properties(this PropertyProvider, names []string) (props map[string]string, err error) {
	props = make(map[string]string)
	for _, name := range names {
		props[name], err = this.Property(name)
		if err != nil {
			err = errors.Wrap(err, "Error retrieving property '"+name+"'")
			return
		}
	}

	return
}

// Compacter wraps the Compact method of a key-value store.
type Compacter interface {
	// Compact flattens the underlying key-value store for the given key range.
	//
	// In essence, deleted and overwritten versions are discarded, and the data
	// is rearranged to reduce the cost of operations needed to access them.
	//
	// A "" start is treated as a key before all keys in the data store; a ""
	// end is treated as a key after all keys in the data store. If both are ""
	// then it will compact the entire data store.
	Compact(start, end string) error
}

// Database is a key-value store (not to be confused with SQL-like databases).
type Database interface {
	Reader
	Writer
	Batcher
	Iterateable
	PropertyProvider
	Compacter
	io.Closer
}

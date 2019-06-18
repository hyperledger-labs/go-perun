// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

// Package db defines a key-value store abstraction.
// It is used by other persistence packages (like channeldb) to continuously
// save the state of the running program or load it upon startup.
// It is based on go-ethereum's https://github.com/ethereum/go-ethereum/ethdb
// and https://github.com/ethereum/go-ethereum/core/rawdb

package db

import "io"

// Reader wraps the Had and Get methods of a key-value store.
type Reader interface {
	// Has checks if a key is present in the store.
	Has(key []byte) (bool, error)

	// Get retrieves the given key if it is present in the store.
	Get(key []byte) ([]byte, error)
}

// Writer wraps the Put and Delete methods of a key-value store.
type Writer interface {
	// Put inserts the given value into the key-value store.
	// If the key is already present, it is overwritten and no error is returned.
	Put(key []byte, value []byte) error

	// Delete removes the key from the key-value store.
	// If the key is not present, an error is returned
	Delete(key []byte) error
}

// PropertyProvider wraps the Property and Properties method of a database.
type PropertyProvider interface {
	// Stat returns a particular internal stat of the database.
	Property(property string) (string, error)

	// Stats returns a default set of common stats of the database.
	// Good for debugging purposes
	Properties() (map[string]string, error)
}

// Compacter wraps the Compact method of a key-value store.
type Compacter interface {
	// Compact flattens the underlying key-value store for the given key range.
	// In essence, deleted and overwritten versions are discarded, and the data
	// is rearranged to reduce the cost of operations needed to access them.
	//
	// A nil start is treated as a key before all keys in the data store; a nil
	// end is treated as a key after all keys in the data store. If both is nil
	// then it will compact the entire data store.
	Compact(start []byte, end []byte) error
}

type Database interface {
	Reader
	Writer
	Batcher
	Iterateable
	PropertyProvider
	Compacter
	io.Closer
}

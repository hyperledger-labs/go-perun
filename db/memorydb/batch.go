// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package memorydb

import (
	"github.com/pkg/errors"
)

// Batch represents a batch and implements the batch interface.
type Batch struct {
	db      *Database
	writes  map[string]string
	deletes map[string]struct{}
}

// Put puts a new value in the batch.
func (b *Batch) Put(key string, value string) error {
	delete(b.deletes, key)
	b.writes[key] = value
	return nil
}

// PutBytes puts a new byte slice into the batch.
func (b *Batch) PutBytes(key string, value []byte) error {
	return b.Put(key, string(value))
}

// Delete deletes a value from the batch.
func (b *Batch) Delete(key string) error {
	delete(b.writes, key)
	b.deletes[key] = struct{}{}
	return nil
}

// Apply applies the batch to the database.
func (b *Batch) Apply() error {
	for key, value := range b.writes {
		err := b.db.Put(key, value)
		if err != nil {
			return errors.Wrap(err, "Failed to put entry.")
		}
	}

	for key := range b.deletes {
		err := b.db.Delete(key)
		if err != nil {
			return errors.Wrap(err, "Failed to delete entry.")
		}
	}
	return nil
}

// Reset resets the batch.
func (b *Batch) Reset() {
	b.writes = make(map[string]string)
	b.deletes = make(map[string]struct{})
}

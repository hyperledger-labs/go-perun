// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package leveldb

import (
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

// Batch represents a batch and implements the batch interface.
type Batch struct {
	*leveldb.Batch
	db *leveldb.DB
}

// Put puts a new value in the batch.
func (b *Batch) Put(key string, value string) error {
	return b.PutBytes(key, []byte(value))
}

// PutBytes puts a new byte slice into the batch.
func (b *Batch) PutBytes(key string, value []byte) error {
	b.Batch.Put([]byte(key), []byte(value))
	return nil
}

// Delete deletes a value from the batch.
func (b *Batch) Delete(key string) error {
	b.Batch.Delete([]byte(key))
	return nil
}

// Apply applies the batch to the database.
func (b *Batch) Apply() error {
	err := b.db.Write(b.Batch, nil)
	return errors.Wrap(err, "leveldb batch apply error")
}

// Reset resets the batch.
func (b *Batch) Reset() {
	b.Batch.Reset()
}

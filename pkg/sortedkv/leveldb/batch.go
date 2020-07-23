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

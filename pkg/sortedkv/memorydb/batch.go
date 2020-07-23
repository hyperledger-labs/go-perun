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
			return errors.Wrap(err, "failed to put entry")
		}
	}

	for key := range b.deletes {
		err := b.db.Delete(key)
		if err != nil {
			return errors.Wrap(err, "failed to delete entry")
		}
	}
	return nil
}

// Reset resets the batch.
func (b *Batch) Reset() {
	b.writes = make(map[string]string)
	b.deletes = make(map[string]struct{})
}

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

	"github.com/syndtr/goleveldb/leveldb/iterator"
	"perun.network/go-perun/log"
)

// Iterator provides an iterator over a key range.
type Iterator struct {
	iterator.Iterator
	mu sync.Mutex
}

// Next returns true if the iterator has a next element.
func (i *Iterator) Next() bool {
	i.mu.Lock()
	defer i.mu.Unlock()

	// Was the iterator closed?
	if i.Iterator == nil {
		return false
	}

	return i.Iterator.Next()
}

// Key returns the key of the current element.
func (i *Iterator) Key() string {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.Iterator == nil || !i.Iterator.Valid() {
		log.Panic("Iterator.Key() called on invalid iterator")
	}

	return string(i.Iterator.Key())
}

// Value returns the value of the current element.
func (i *Iterator) Value() string {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.Iterator == nil || !i.Iterator.Valid() {
		log.Panic("Iterator.Value() called on invalid iterator")
	}

	return string(i.Iterator.Value())
}

// ValueBytes returns the value converted to bytes of the current element.
func (i *Iterator) ValueBytes() []byte {
	return []byte(i.Value())
}

// Close closes this iterator.
func (i *Iterator) Close() error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.Iterator == nil {
		return nil
	}

	// The accumulated errors are only returned on the first call to Close()
	err := i.Iterator.Error()
	i.Iterator.Release()
	i.Iterator = nil
	return err
}

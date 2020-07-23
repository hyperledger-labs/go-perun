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

package sortedkv

// Iterator iterates over a data store's key/value pairs in ascending key order.
//
// When it encounters an error, any Next() will return false and will yield no key/
// value pairs. The error can be queried by calling the Error method. Calling
// Release is still necessary.
//
// An iterator must be released after use, but it is not necessary to read an
// iterator until exhaustion. An iterator is not safe for concurrent use, but it
// is safe to use multiple iterators concurrently.
type Iterator interface {
	// Next moves the iterator to the next key/value pair. It returns false if the
	// iterator is exhausted or closed, and true otherwise.
	Next() bool

	// Key returns the key of the current key/value pair, or "" if done.
	Key() string

	// Value returns the value of the current key/value pair, or "" if done.
	Value() string

	// ValueBytes returns the value of the current key/value pair, or nil if done. The
	// caller should not modify the contents of the returned slice, and its contents
	// may change on the next call to Next.
	ValueBytes() []byte

	// Close releases associated resources. It returns any accumulated error.
	// Exhausting all the key/value pairs is not considered to be an error.
	// Close can be called multiple times.
	Close() error
}

// Iterable wraps the NewIterator methods of a backing data store.
type Iterable interface {
	// NewIterator creates a binary-alphabetical iterator over the entire keyspace
	// contained within the key-value database.
	NewIterator() Iterator

	// NewIteratorWithStart creates a binary-alphabetical iterator over a subset of
	// database content over a key range of [start, end). If start is empty, then
	// the iterator starts with the database's first entry. If end is empty, then
	// the iterator ends with the database's last entry.
	NewIteratorWithRange(start string, end string) Iterator

	// NewIteratorWithPrefix creates a binary-alphabetical iterator over a subset
	// of database content with a particular key prefix.
	NewIteratorWithPrefix(prefix string) Iterator
}

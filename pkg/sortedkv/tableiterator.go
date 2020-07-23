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

// tableIterator is a wrapper around the Iterator interface.
type tableIterator struct {
	Iterator
	prefix int
}

// newTableIterator creates a new table iterator
func newTableIterator(it Iterator, table *table) Iterator {
	return &tableIterator{
		Iterator: it,
		prefix:   len(table.prefix),
	}
}

// Key returns the value that is iterated over, but without the table's prefix.
func (it *tableIterator) Key() string {
	return it.Iterator.Key()[it.prefix:]
}

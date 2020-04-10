// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

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

// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package test

import (
	"testing"

	"perun.network/go-perun/db"
)

// IteratorTest provides the values needed for the generic tests.
type IteratorTest struct {
	*testing.T
	Iterator db.Iterator
}

// GenericIteratorTest provides generic tests for iterator implementations.
func GenericIteratorTest(t *testing.T, database db.Database) {
	dbtest := DatabaseTest{T: t, Database: database}
	dbtest.Put("2b", "2bv")
	dbtest.Put("3", "3v")
	dbtest.Put("1", "1v")
	dbtest.Put("2a", "2av")

	// Test all elements.
	it := IteratorTest{T: t, Iterator: database.NewIterator()}
	it.NextMustEqual("1", "1v")
	it.NextMustEqual("2a", "2av")
	it.NextMustEqual("2b", "2bv")
	it.NextMustEqual("3", "3v")
	it.MustEnd()

	// Test that [..., ...] encompasses all elements.
	it.Iterator = database.NewIteratorWithRange("", "")
	it.NextMustEqual("1", "1v")
	it.NextMustEqual("2a", "2av")
	it.NextMustEqual("2b", "2bv")
	it.NextMustEqual("3", "3v")
	it.MustEnd()

	// Test [..., "2b")
	it.Iterator = database.NewIteratorWithRange("", "2b")
	it.NextMustEqual("1", "1v")
	it.NextMustEqual("2a", "2av")

	// Test ["2", ...]
	it.Iterator = database.NewIteratorWithRange("2", "")
	it.NextMustEqual("2a", "2av")
	it.NextMustEqual("2b", "2bv")
	it.NextMustEqual("3", "3v")
	it.MustEnd()

	// Test ["2", "2b")
	it.Iterator = database.NewIteratorWithRange("2", "2b")
	it.NextMustEqual("2a", "2av")
	it.MustEnd()

	// Test "2"+...
	it.Iterator = database.NewIteratorWithPrefix("2")
	it.NextMustEqual("2a", "2av")
	it.NextMustEqual("2b", "2bv")
	it.MustEnd()

	// Test whether closing really ends the iterator.
	it.Iterator = database.NewIterator()
	it.NextMustEqual("1", "1v")
	it.Close()
	it.MustEnd()

	return
}

// NextMustEqual tests the next method.
func (i *IteratorTest) NextMustEqual(key, value string) {
	if !i.Iterator.Next() {
		i.Errorf("Next(): Expected [%q] = %q, but iterator ended.\n", key, value)
		return
	}

	if actual := i.Iterator.Value(); actual != value {
		i.Errorf("Value(): Expected %q, but got %q.\n", value, actual)
	}
	if actual := i.Iterator.ValueBytes(); string(actual) != value {
		i.Errorf("ValueBytes(): Expected %q, but got %q.\n", value, string(actual))
	}
	if actual := i.Iterator.Key(); actual != key {
		i.Errorf("Key(): Expected %q, but got %q.\n", key, actual)
	}
	return
}

// MustEnd tests the next method.
func (i *IteratorTest) MustEnd() {
	if i.Iterator.Next() {
		i.Errorf(
			"Next(): Expected end, but got [%q] = %q.\n",
			i.Iterator.Key(),
			i.Iterator.Value())
	}

	i.Close()
	return
}

// Close tests the close method.
func (i *IteratorTest) Close() {
	if err := i.Iterator.Close(); err != nil {
		i.Errorf("Close(): failed with error: %v\n", err)
	}
	return
}

// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

// Package test of the go-perun/sortedkv package implements a generic test for all
// implementations of the Database interface.
// Test your implementation by passing an empty database to the
// GenericDatabaseTest() function
package test // import "perun.network/go-perun/pkg/sortedkv/test"

import (
	"bytes"
	"testing"

	"perun.network/go-perun/pkg/sortedkv"
)

// GenericDatabaseTest provides generic sortedkv tests.
func GenericDatabaseTest(t *testing.T, database sortedkv.Database) {
	test := DatabaseTest{T: t}
	t.Run("Generic database test", func(t *testing.T) {
		test.Database = database
		test.test()
		GenericTableTest(t, test.Database)
	})
	t.Run("Empty prefix table test", func(t *testing.T) {
		test.Database = sortedkv.NewTable(database, "")
		test.test()
		GenericTableTest(t, test.Database)
	})
	t.Run("Normal table test", func(t *testing.T) {
		test.Database = sortedkv.NewTable(database, "Table.")
		test.test()
		GenericTableTest(t, test.Database)
	})
}

// test Tests a generic database.
func (d *DatabaseTest) test() {
	if d.T == nil {
		panic("No tester provided!")
	}

	if d.Database == nil {
		d.Fatalf("Did not supply a database!")
	}

	// Test that the database does not have 1234.
	d.MustNotHave("1234")
	// Test that get fails if Has() returns false.
	d.MustFailGet("1234")
	// Put() must work for inserting entries.
	d.Put("1234", "qwer")
	d.PutBytes("5678", []byte("5678 value"))
	// Put() must work for inserting a second entry.
	d.Put("asdf", "yxcv")
	// Has() must return true for inserted elements.
	d.MustHave("1234")
	d.MustHave("5678")
	d.MustHave("asdf")
	// Get() must return the correct value.
	d.MustGetEqual("1234", "qwer")
	d.MustGetBytesEqual("5678", []byte("5678 value"))
	d.MustGetEqual("asdf", "yxcv")
	// Remove 1234
	d.Delete("1234")
	d.MustFailDelete("1234")
	// Has() must be false for deleted entries and Get() must fail.
	d.MustNotHave("1234")
	d.MustFailGet("1234")
	// Only the intended entry must be deleted.
	d.MustHave("asdf")
	// Overwrites must work correctly.
	d.Put("asdf", "YXCV")
	d.MustGetEqual("asdf", "YXCV")
	d.Delete("asdf")
}

// DatabaseTest is a sortedkv testing struct.
type DatabaseTest struct {
	*testing.T
	Database sortedkv.Database
}

// Has tests the has functionality.
func (d *DatabaseTest) Has(key string) bool {
	has, err := d.Database.Has(key)
	if err != nil {
		d.Fatalf("Has(): Failed to query %q: %v\n", key, err)
	}
	return has
}

// MustHave tests the has functionality.
func (d *DatabaseTest) MustHave(key string) {
	if !d.Has(key) {
		d.Errorf("Database does not have entry %q but it should.\n", key)
	}
}

// MustNotHave tests the has functionality.
func (d *DatabaseTest) MustNotHave(key string) {
	if d.Has(key) {
		d.Errorf("Database has entry %q but it shouldn't.\n", key)
	}
}

// Put tests the put functionality.
func (d *DatabaseTest) Put(key string, value string) {
	err := d.Database.Put(key, value)
	if err != nil {
		d.Fatalf("Failed to put [%q]=%q.\n", key, value)
	}
}

// PutBytes tests the putBytes functionality.
func (d *DatabaseTest) PutBytes(key string, value []byte) {
	if err := d.Database.PutBytes(key, value); err != nil {
		d.Fatalf("PutBytes(): Failed to put [%q]=%q.\n", key, value)
	}
}

// Get tests the get functionality.
func (d *DatabaseTest) Get(key string) string {
	value, err := d.Database.Get(key)
	if err != nil {
		d.Fatalf("Failed to put [%q]=%q.\n", key, value)
	}
	return value
}

// GetBytes tests the getBytes functionality.
func (d *DatabaseTest) GetBytes(key string) []byte {
	value, err := d.Database.GetBytes(key)
	if err != nil {
		d.Fatalf("Failed to get bytes [%q]\n", key)
	}
	return value
}

// MustFailGet tests the get functionality.
func (d *DatabaseTest) MustFailGet(key string) {
	_, err := d.Database.Get(key)
	if err == nil {
		d.Errorf("Get() did not fail when expected to ([%q]).\n", key)
	}
}

// MustGetEqual tests the get functionality.
func (d *DatabaseTest) MustGetEqual(key string, expected string) {
	if value := d.Get(key); value != expected {
		d.Errorf(
			"Get() returned the wrong value: [%q] (is %q, expected %q)\n",
			key,
			value,
			expected)
	}
}

// MustGetBytesEqual tests the getBytes functionality.
func (d *DatabaseTest) MustGetBytesEqual(key string, expected []byte) {
	if value := d.GetBytes(key); !bytes.Equal(value, expected) {
		d.Errorf(
			"Get() returned the wrong value: [%q] (is %q, expected %q)\n",
			key,
			value,
			expected)
	}
}

// Delete tests the delete functionality.
func (d *DatabaseTest) Delete(key string) {
	if err := d.Database.Delete(key); err != nil {
		d.Errorf("Delete() [%q] failed: %v", key, err)
	}
}

// MustFailDelete tests the delete functionality.
func (d *DatabaseTest) MustFailDelete(key string) {
	if err := d.Database.Delete(key); err == nil {
		d.Errorf("Delete() [%q] should have failed, but did not.\n", key)
	}
}

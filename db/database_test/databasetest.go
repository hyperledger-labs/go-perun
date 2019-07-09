package database_test

import (
	"bytes"
	"testing"

	"github.com/perun-network/go-perun/db"
)

func GenericDatabaseTest(t *testing.T, database db.Database) {
	test := DatabaseTest{T: t}
	t.Run("Generic database test", func(t *testing.T) {
		test.Database = database
		test.test()
		GenericTableTest(t, test.Database)
	})
	t.Run("Empty prefix table test", func(t *testing.T) {
		test.Database = db.NewTable(database, "")
		test.test()
		GenericTableTest(t, test.Database)
	})
	t.Run("Normal table test", func(t *testing.T) {
		test.Database = db.NewTable(database, "Table.")
		test.test()
		GenericTableTest(t, test.Database)
	})
	return
}

// Tests a generic database.
func (this *DatabaseTest) test() {
	if this.T == nil {
		panic("No tester provided!")
	}

	if this.Database == nil {
		this.Fatalf("Did not supply a database!")
	}

	// Test that the database does not have 1234.
	this.MustNotHave("1234")
	// Test that get fails if Has() returns false.
	this.MustFailGet("1234")
	// Put() must work for inserting entries.
	this.Put("1234", "qwer")
	this.PutBytes("5678", []byte("5678 value"))
	// Put() must work for inserting a second entry.
	this.Put("asdf", "yxcv")
	// Has() must return true for inserted elements.
	this.MustHave("1234")
	this.MustHave("5678")
	this.MustHave("asdf")
	// Get() must return the correct value.
	this.MustGetEqual("1234", "qwer")
	this.MustGetBytesEqual("5678", []byte("5678 value"))
	this.MustGetEqual("asdf", "yxcv")
	// Remove 1234
	this.Delete("1234")
	this.MustFailDelete("1234")
	// Has() must be false for deleted entries and Get() must fail.
	this.MustNotHave("1234")
	this.MustFailGet("1234")
	// Only the intended entry must be deleted.
	this.MustHave("asdf")
	// Overwrites must work correctly.
	this.Put("asdf", "YXCV")
	this.MustGetEqual("asdf", "YXCV")
	this.Delete("asdf")

	return
}

type DatabaseTest struct {
	*testing.T
	Database db.Database
}

func (this *DatabaseTest) Has(key string) bool {
	has, err := this.Database.Has(key)
	if err != nil {
		this.Fatalf("Has(): Failed to query %q: %v\n", key, err)
	}
	return has
}

func (this *DatabaseTest) MustHave(key string) {
	if !this.Has(key) {
		this.Errorf("Database does not have entry %q but it should.\n", key)
	}
	return
}

func (this *DatabaseTest) MustNotHave(key string) {
	if this.Has(key) {
		this.Errorf("Database has entry %q but it shouldn't.\n", key)
	}
	return
}

func (this *DatabaseTest) Put(key string, value string) {
	err := this.Database.Put(key, value)
	if err != nil {
		this.Fatalf("Failed to put [%q]=%q.\n", key, value)
	}
	return
}

func (this *DatabaseTest) PutBytes(key string, value []byte) {
	if err := this.Database.PutBytes(key, value); err != nil {
		this.Fatalf("PutBytes(): Failed to put [%q]=%q.\n", key, value)
	}
	return
}

func (this *DatabaseTest) Get(key string) string {
	value, err := this.Database.Get(key)
	if err != nil {
		this.Fatalf("Failed to put [%q]=%q.\n", key, value)
	}
	return value
}

func (this *DatabaseTest) GetBytes(key string) []byte {
	value, err := this.Database.GetBytes(key)
	if err != nil {
		this.Fatalf("Failed to get bytes [%q]\n", key)
	}
	return value
}

func (this *DatabaseTest) MustFailGet(key string) {
	_, err := this.Database.Get(key)
	if err == nil {
		this.Errorf("Get() did not fail when expected to ([%q]).\n", key)
	}
	return
}

func (this *DatabaseTest) MustGetEqual(key string, expected string) {
	if value := this.Get(key); value != expected {
		this.Errorf(
			"Get() returned the wrong value: [%q] (is %q, expected %q)\n",
			key,
			value,
			expected)
	}
	return
}

func (this *DatabaseTest) MustGetBytesEqual(key string, expected []byte) {
	if value := this.GetBytes(key); !bytes.Equal(value, expected) {
		this.Errorf(
			"Get() returned the wrong value: [%q] (is %q, expected %q)\n",
			key,
			value,
			expected)
	}
	return
}

func (this *DatabaseTest) Delete(key string) {
	if err := this.Database.Delete(key); err != nil {
		this.Errorf("Delete() [%q] failed: %v", key, err)
	}
	return
}

func (this *DatabaseTest) MustFailDelete(key string) {
	if err := this.Database.Delete(key); err == nil {
		this.Errorf("Delete() [%q] should have failed, but did not.\n", key)
	}
	return
}

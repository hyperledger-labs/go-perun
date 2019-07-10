package test

import (
	"testing"

	"perun.network/go-perun/db"
)

type IteratorTest struct {
	*testing.T
	Iterator db.Iterator
}

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

func (this *IteratorTest) NextMustEqual(key, value string) {
	if !this.Iterator.Next() {
		this.Errorf("Next(): Expected [%q] = %q, but iterator ended.\n", key, value)
		return
	}

	if actual := this.Iterator.Value(); actual != value {
		this.Errorf("Value(): Expected %q, but got %q.\n", value, actual)
	}
	if actual := this.Iterator.ValueBytes(); string(actual) != value {
		this.Errorf("ValueBytes(): Expected %q, but got %q.\n", value, string(actual))
	}
	if actual := this.Iterator.Key(); actual != key {
		this.Errorf("Key(): Expected %q, but got %q.\n", key, actual)
	}
	return
}

func (this *IteratorTest) MustEnd() {
	if this.Iterator.Next() {
		this.Errorf(
			"Next(): Expected end, but got [%q] = %q.\n",
			this.Iterator.Key(),
			this.Iterator.Value())
	}

	this.Close()
	return
}

func (this *IteratorTest) Close() {
	if err := this.Iterator.Close(); err != nil {
		this.Errorf("Close(): failed with error: %v\n", err)
	}
	return
}

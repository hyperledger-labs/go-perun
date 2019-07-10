package test

import (
	"testing"

	"perun.network/go-perun/db"
)

func GenericTableTest(t *testing.T, database db.Database) {
	dbtest := DatabaseTest{
		T:        t,
		Database: database,
	}
	dbtest.Put("KeyA", "ValueA")
	dbtest.Put("KeyB", "ValueB")
	dbtest.Put("Table.Inner.KeyA", "Table.Inner.ValueA")
	dbtest.Put("Table.Inner.KeyB", "Table.Inner.ValueB")
	dbtest.Put("Table.KeyA", "Table.ValueA")
	dbtest.Put("Table.KeyB", "Table.ValueB")
	dbtest.Put("Table.KeyC", "Table.ValueC")

	table := DatabaseTest{
		T:        t,
		Database: db.NewTable(dbtest.Database, "Table."),
	}

	t.Run(`All values`, func(t *testing.T) {
		it := IteratorTest{T: t, Iterator: table.Database.NewIterator()}
		it.NextMustEqual("Inner.KeyA", "Table.Inner.ValueA")
		it.NextMustEqual("Inner.KeyB", "Table.Inner.ValueB")
		it.NextMustEqual("KeyA", "Table.ValueA")
		it.NextMustEqual("KeyB", "Table.ValueB")
		it.NextMustEqual("KeyC", "Table.ValueC")
		it.MustEnd()
	})

	t.Run(`["KeyA", "KeyC"]`, func(t *testing.T) {
		it := IteratorTest{
			T: t,
			Iterator: table.Database.NewIteratorWithRange(
				"KeyA", db.IncrementKey("KeyC")),
		}
		it.NextMustEqual("KeyA", "Table.ValueA")
		it.NextMustEqual("KeyB", "Table.ValueB")
		it.NextMustEqual("KeyC", "Table.ValueC")
		it.MustEnd()
	})

	t.Run(`["KeyA", "KeyC")`, func(t *testing.T) {
		it := IteratorTest{
			T:        t,
			Iterator: table.Database.NewIteratorWithRange("KeyB", "KeyC"),
		}
		it.NextMustEqual("KeyB", "Table.ValueB")
		it.MustEnd()
	})

	t.Run(`"Inner."+`, func(t *testing.T) {
		it := IteratorTest{
			T:        t,
			Iterator: table.Database.NewIteratorWithPrefix("Inner."),
		}
		it.NextMustEqual("Inner.KeyA", "Table.Inner.ValueA")
		it.NextMustEqual("Inner.KeyB", "Table.Inner.ValueB")
	})
}

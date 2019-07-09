package memorydb

import (
	"testing"

	"github.com/perun-network/go-perun/db/database_test"
)

func TestDatabase(t *testing.T) {
	t.Run("Generic Database test", func(t *testing.T) {
		database_test.GenericDatabaseTest(t, NewDatabase())
	})

	dbtest := database_test.DatabaseTest{
		T: t,
		Database: FromData(map[string]string{
			"k2": "v2",
			"k3": "v3",
			"k1": "v1",
		}),
	}

	dbtest.MustGetEqual("k1", "v1")
	dbtest.MustGetEqual("k2", "v2")
	dbtest.MustGetEqual("k3", "v3")
	ittest := database_test.IteratorTest{
		T:        t,
		Iterator: dbtest.Database.NewIterator(),
	}

	ittest.NextMustEqual("k1", "v1")
	ittest.NextMustEqual("k2", "v2")
	ittest.NextMustEqual("k3", "v3")
	ittest.MustEnd()

	return
}

package memorydb

import (
	"testing"

	"perun.network/go-perun/db/database_test"
)

func TestIterator(t *testing.T) {
	t.Run("Generic iterator test", func(t *testing.T) {
		database_test.GenericIteratorTest(t, NewDatabase())
	})
	return
}

package memorydb

import (
	"testing"

	"perun.network/go-perun/db/test"
)

func TestIterator(t *testing.T) {
	t.Run("Generic iterator test", func(t *testing.T) {
		test.GenericIteratorTest(t, NewDatabase())
	})
	return
}

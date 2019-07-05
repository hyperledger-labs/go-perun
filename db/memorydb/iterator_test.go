package memorydb

import (
	"testing"

	"github.com/perun-network/go-perun/db/iterator_test"
)

func TestIterator(t *testing.T) {
	t.Run("Generic iterator test", func(t *testing.T) {
		iterator_test.GenericIteratorTest(t, NewDatabase())
	})
	return
}

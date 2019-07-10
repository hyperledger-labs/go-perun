package memorydb

import (
	"testing"

	"perun.network/go-perun/db/test"
)

func TestBatch(t *testing.T) {
	t.Run("Generic Batch test", func(t *testing.T) {
		test.GenericBatchTest(t, NewDatabase())
	})
	return
}

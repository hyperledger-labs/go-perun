package memorydb

import (
	"testing"

	"github.com/perun-network/go-perun/db/batch_test"
)

func TestBatch(t *testing.T) {
	t.Run("Generic Batch test", func(t *testing.T) {
		batch_test.GenericBatchTest(t, NewDatabase())
	})
	return
}

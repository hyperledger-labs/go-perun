package memorydb

import (
	"testing"

	"github.com/perun-network/go-perun/db/database_test"
)

func TestBatch(t *testing.T) {
	t.Run("Generic Batch test", func(t *testing.T) {
		database_test.GenericBatchTest(t, NewDatabase())
	})
	return
}

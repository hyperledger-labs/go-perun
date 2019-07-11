package test

import (
	"testing"

	"perun.network/go-perun/db"
)

// GenericBatchTest is to be called from the batch implementation tests.
func GenericBatchTest(t *testing.T, database db.Database) {
	this := BatchTest{T: t, Batch: database.NewBatch()}

	dbtest := DatabaseTest{T: t, Database: database}

	dbtest.Put("1234", "1234 initial value")
	dbtest.Put("5678", "5678 initial value")

	const strLen1 = "Test Put() tracking."
	const strLen2 = "Test Put() tracking override."

	this.Batch.Reset()
	// Batch must be empty after reset.
	this.MustHaveLength(0)
	this.MustHaveValueSize(0)
	// Test that deleting works on empty and full batches.
	this.MustDelete("1234")
	this.MustDelete("5678")
	this.Batch.Reset()
	// Put must work on empty and full batches.
	// New Put() must increase length.
	this.MustPut("1234", strLen1)
	this.MustHaveLength(1)
	this.MustHaveValueSize(uint(len(strLen1)))
	// Test PutBytes() overwrite.
	this.MustPutBytes("1234", []byte(strLen1))
	this.MustHaveLength(1)
	this.MustHaveValueSize(uint(len(strLen1)))
	// Overwrite Put() must keep length.
	this.MustPut("1234", strLen2)
	this.MustHaveLength(1)
	this.MustHaveValueSize(uint(len(strLen2)))
	// New Delete() must increase length.
	this.MustDelete("5678")
	this.MustHaveLength(2)
	// Delete() of existing Put() must not change length.
	this.MustDelete("1234")
	this.MustHaveLength(2)
	this.MustHaveValueSize(0)
	// Put() of existing Delete() must not change length.
	this.MustPut("5678", "ghjk")
	this.MustHaveLength(2)
	this.MustHaveValueSize(4)

	this.MustApply()

	dbtest.MustNotHave("1234")
	dbtest.MustGetEqual("5678", "ghjk")

	return
}

// BatchTest tests a batch.
type BatchTest struct {
	*testing.T
	Batch db.Batch
}

// MustPut tests the put functionality.
func (bt *BatchTest) MustPut(key, value string) {
	if err := bt.Batch.Put(key, value); err != nil {
		bt.Fatalf("Put(): Failed to put [%q] = %q: %v.\n", key, value, err)
	}
	return
}

// MustPutBytes tests the putBytes functionality.
func (bt *BatchTest) MustPutBytes(key string, value []byte) {
	if err := bt.Batch.PutBytes(key, value); err != nil {
		bt.Fatalf("PutBytes(): Failed to put [%q] = '%v': %v\n", key, value, err)
	}
	return
}

// MustDelete tests the delete functionality.
func (bt *BatchTest) MustDelete(key string) {
	if err := bt.Batch.Delete(key); err != nil {
		bt.Fatalf("Put(): Failed to delete [%q]: %v.\n", key, err)
	}
	return
}

// MustHaveLength tests the len functionality.
func (bt *BatchTest) MustHaveLength(length uint) {
	if actual := bt.Batch.Len(); actual != length {
		bt.Errorf("Len(): Batch has %d elements, expected %d.\n", actual, length)
	}
	return
}

// MustHaveValueSize tests the valueSize functionality.
func (bt *BatchTest) MustHaveValueSize(size uint) {
	if actual := bt.Batch.ValueSize(); actual != size {
		bt.Errorf("ValueSize(): Batch has size %d, expected %d.\n", actual, size)
	}
	return
}

// MustApply tests the apply functionality.
func (bt *BatchTest) MustApply() {
	if err := bt.Batch.Apply(); err != nil {
		bt.Errorf("Apply(): Failed with reason %v.\n", err)
	}
}

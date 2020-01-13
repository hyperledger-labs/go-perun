// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

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
	// Test that deleting works on empty and full batches.
	this.MustDelete("1234")
	this.MustDelete("5678")
	this.Batch.Reset()
	// Put must work on empty and full batches.
	this.MustPut("1234", strLen1)
	this.MustPutBytes("1234", []byte(strLen1))
	this.MustPut("1234", strLen2)
	this.MustDelete("5678")
	this.MustDelete("1234")
	this.MustPut("5678", "ghjk")

	this.MustApply()

	dbtest.MustNotHave("1234")
	dbtest.MustGetEqual("5678", "ghjk")
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
}

// MustPutBytes tests the putBytes functionality.
func (bt *BatchTest) MustPutBytes(key string, value []byte) {
	if err := bt.Batch.PutBytes(key, value); err != nil {
		bt.Fatalf("PutBytes(): Failed to put [%q] = '%v': %v\n", key, value, err)
	}
}

// MustDelete tests the delete functionality.
func (bt *BatchTest) MustDelete(key string) {
	if err := bt.Batch.Delete(key); err != nil {
		bt.Fatalf("Put(): Failed to delete [%q]: %v.\n", key, err)
	}
}

// MustApply tests the apply functionality.
func (bt *BatchTest) MustApply() {
	if err := bt.Batch.Apply(); err != nil {
		bt.Errorf("Apply(): Failed with reason %v.\n", err)
	}
}

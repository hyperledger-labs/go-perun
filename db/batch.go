// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package db

// Batch is a write-only database that buffers changes to the underlying
// database until Apply() is called.
type Batch interface {
	Writer // Put and Delete

	// Apply performs all batched actions on the database.
	Apply() error

	// Reset resets the batch so that it doesn't contain any items and can be reused.
	Reset()
}

// Batcher wraps the NewBatch method of a backing data store.
type Batcher interface {
	// NewBatch creates a Batch that will write to the Batcher.
	NewBatch() Batch
}

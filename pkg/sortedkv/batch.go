// Copyright 2019 - See NOTICE file for copyright holders.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sortedkv

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

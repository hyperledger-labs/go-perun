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

// tableBatch is a wrapper around a Database Batch with a key prefix. All
// Writer operations are automatically prefixed.
type tableBatch struct {
	Batch
	prefix string
}

func (b *tableBatch) pkey(key string) string {
	return b.prefix + key
}

// Put puts a value into a table batch.
func (b *tableBatch) Put(key, value string) error {
	return b.Batch.Put(b.pkey(key), value)
}

// Put puts a value into a table batch.
func (b *tableBatch) PutBytes(key string, value []byte) error {
	return b.Batch.PutBytes(b.pkey(key), value)
}

// Delete deletes a value from a table batch.
func (b *tableBatch) Delete(key string) error {
	return b.Batch.Delete(b.pkey(key))
}

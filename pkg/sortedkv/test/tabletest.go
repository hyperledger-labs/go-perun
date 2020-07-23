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

package test

import (
	"testing"

	"perun.network/go-perun/pkg/sortedkv"
	"perun.network/go-perun/pkg/sortedkv/key"
)

// GenericTableTest provides generic tests for table implementations.
func GenericTableTest(t *testing.T, database sortedkv.Database) {
	dbtest := DatabaseTest{
		T:        t,
		Database: database,
	}
	dbtest.Put("KeyA", "ValueA")
	dbtest.Put("KeyB", "ValueB")
	dbtest.Put("Table.Inner.KeyA", "Table.Inner.ValueA")
	dbtest.Put("Table.Inner.KeyB", "Table.Inner.ValueB")
	dbtest.Put("Table.KeyA", "Table.ValueA")
	dbtest.Put("Table.KeyB", "Table.ValueB")
	dbtest.Put("Table.KeyC", "Table.ValueC")

	table := DatabaseTest{
		T:        t,
		Database: sortedkv.NewTable(dbtest.Database, "Table."),
	}

	t.Run(`All values`, func(t *testing.T) {
		it := IteratorTest{T: t, Iterator: table.Database.NewIterator()}
		it.NextMustEqual("Inner.KeyA", "Table.Inner.ValueA")
		it.NextMustEqual("Inner.KeyB", "Table.Inner.ValueB")
		it.NextMustEqual("KeyA", "Table.ValueA")
		it.NextMustEqual("KeyB", "Table.ValueB")
		it.NextMustEqual("KeyC", "Table.ValueC")
		it.MustEnd()
	})

	t.Run(`["KeyA", "KeyC"]`, func(t *testing.T) {
		it := IteratorTest{
			T: t,
			Iterator: table.Database.NewIteratorWithRange(
				"KeyA", key.Next("KeyC")),
		}
		it.NextMustEqual("KeyA", "Table.ValueA")
		it.NextMustEqual("KeyB", "Table.ValueB")
		it.NextMustEqual("KeyC", "Table.ValueC")
		it.MustEnd()
	})

	t.Run(`["KeyA", "KeyC")`, func(t *testing.T) {
		it := IteratorTest{
			T:        t,
			Iterator: table.Database.NewIteratorWithRange("KeyB", "KeyC"),
		}
		it.NextMustEqual("KeyB", "Table.ValueB")
		it.MustEnd()
	})

	t.Run(`"Inner."+`, func(t *testing.T) {
		it := IteratorTest{
			T:        t,
			Iterator: table.Database.NewIteratorWithPrefix("Inner."),
		}
		it.NextMustEqual("Inner.KeyA", "Table.Inner.ValueA")
		it.NextMustEqual("Inner.KeyB", "Table.Inner.ValueB")
	})
}

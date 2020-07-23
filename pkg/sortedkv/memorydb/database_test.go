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

package memorydb

import (
	"testing"

	"perun.network/go-perun/pkg/sortedkv/test"
)

func TestDatabase(t *testing.T) {
	t.Run("Generic Database test", func(t *testing.T) {
		test.GenericDatabaseTest(t, NewDatabase())
	})

	dbtest := test.DatabaseTest{
		T: t,
		Database: FromData(map[string]string{
			"k2": "v2",
			"k3": "v3",
			"k1": "v1",
		}),
	}

	dbtest.MustGetEqual("k1", "v1")
	dbtest.MustGetEqual("k2", "v2")
	dbtest.MustGetEqual("k3", "v3")
	ittest := test.IteratorTest{
		T:        t,
		Iterator: dbtest.Database.NewIterator(),
	}

	ittest.NextMustEqual("k1", "v1")
	ittest.NextMustEqual("k2", "v2")
	ittest.NextMustEqual("k3", "v3")
	ittest.MustEnd()
}

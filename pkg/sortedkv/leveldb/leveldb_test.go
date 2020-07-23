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

package leveldb

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"perun.network/go-perun/pkg/sortedkv"
	"perun.network/go-perun/pkg/sortedkv/test"
)

func TestBatch(t *testing.T) {
	runTestOnTempDatabase(t, func(db *Database) {
		test.GenericBatchTest(t, db)
	})
	runTestOnTempDatabase(t, func(db *Database) {
		test.GenericBatchTest(t, sortedkv.NewTable(db, "table"))
	})
}

func TestDatabase(t *testing.T) {
	runTestOnTempDatabase(t, func(db *Database) {
		test.GenericDatabaseTest(t, db)
	})
}

func TestIterator(t *testing.T) {
	runTestOnTempDatabase(t, func(db *Database) {
		test.GenericIteratorTest(t, db)
	})
}

func runTestOnTempDatabase(t *testing.T, tester func(*Database)) {
	// Create a temporary directory and delete it when done
	path, err := ioutil.TempDir("", "perun_testdb_")
	require.Nil(t, err, "Could not create temporary directory for database")
	defer func() { require.Nil(t, os.RemoveAll(path)) }()

	// Create a database in the directory and close it when done
	db, err := LoadDatabase(path)
	require.Nil(t, err, "Could not load database")
	defer func() { assert.Nil(t, db.DB.Close(), "Could not close database") }()

	tester(db)
}

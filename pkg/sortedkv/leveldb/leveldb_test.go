// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package leveldb

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"perun.network/go-perun/pkg/sortedkv/test"
)

func TestBatch(t *testing.T) {
	runTestOnTempDatabase(t, func(db *Database) {
		test.GenericBatchTest(t, db)
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

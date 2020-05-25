// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package keyvalue

import (
	"context"
	"io/ioutil"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	_ "perun.network/go-perun/backend/sim"
	"perun.network/go-perun/channel/persistence/test"
	"perun.network/go-perun/pkg/sortedkv"
	"perun.network/go-perun/pkg/sortedkv/leveldb"
	"perun.network/go-perun/pkg/sortedkv/memorydb"
)

func TestPersistRestorer_Generic(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "perun-test-kvpersistrestorer-db-*")
	require.NoError(t, err)
	lvldb, err := leveldb.LoadDatabase(tmpdir)
	require.NoError(t, err)

	dbs := []sortedkv.Database{
		lvldb,
		memorydb.NewDatabase(),
	}

	for _, db := range dbs {
		func() {
			defer func() { require.NoError(t, db.Close()) }()
			pr, err := NewPersistRestorer(db)
			require.NoError(t, err)
			test.GenericPersistRestorerTest(
				context.Background(),
				t,
				rand.New(rand.NewSource(0xC00FED)),
				pr,
				4,
				16)
		}()
	}
}

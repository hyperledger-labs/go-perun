// Copyright 2020 - See NOTICE file for copyright holders.
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

package keyvalue

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "perun.network/go-perun/backend/sim" // backend init
	"perun.network/go-perun/channel/persistence/test"
	"polycry.pt/poly-go/sortedkv"
	"polycry.pt/poly-go/sortedkv/leveldb"
	"polycry.pt/poly-go/sortedkv/memorydb"
	pkgtest "polycry.pt/poly-go/test"
)

func TestPersistRestorer_Generic(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "perun-test-kvpersistrestorer-db-*")
	require.NoError(t, err)
	lvldb, err := leveldb.LoadDatabase(tmpdir)
	require.NoError(t, err)

	dbs := []sortedkv.Database{
		lvldb,
		memorydb.NewDatabase(),
	}

	for i, db := range dbs {
		func(i int64) {
			defer func() { require.NoError(t, db.Close()) }()
			pr := NewPersistRestorer(db)
			rng := pkgtest.Prng(t, i)
			test.GenericPersistRestorerTest(
				context.Background(),
				t,
				rng,
				pr,
				4,
				16)
		}(int64(i))
	}
}

func TestChannelIterator_Next_Empty(t *testing.T) {
	var it ChannelIterator
	var success bool
	assert.NotPanics(t, func() { success = it.Next(context.Background()) })
	assert.False(t, success)
	assert.NoError(t, it.err)
}

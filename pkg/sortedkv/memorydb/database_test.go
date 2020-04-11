// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

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

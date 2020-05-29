// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package memorydb

import (
	"testing"

	"perun.network/go-perun/pkg/sortedkv"
	"perun.network/go-perun/pkg/sortedkv/test"
)

func TestBatch(t *testing.T) {
	t.Run("Generic Batch test", func(t *testing.T) {
		test.GenericBatchTest(t, NewDatabase())
	})

	t.Run("Generic table batch test", func(t *testing.T) {
		test.GenericBatchTest(t, sortedkv.NewTable(NewDatabase(), "table"))
	})
}

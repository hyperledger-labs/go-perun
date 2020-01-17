// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package memorydb

import (
	"testing"

	"perun.network/go-perun/db/test"
)

func TestIterator(t *testing.T) {
	t.Run("Generic iterator test", func(t *testing.T) {
		test.GenericIteratorTest(t, NewDatabase())
	})
}

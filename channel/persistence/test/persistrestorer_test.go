// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package test_test

import (
	"context"
	"math/rand"
	"testing"

	_ "perun.network/go-perun/backend/sim" // backend init
	"perun.network/go-perun/channel/persistence/test"
)

func TestPersistRestorer_Generic(t *testing.T) {
	pr := test.NewPersistRestorer(t)
	test.GenericPersistRestorerTest(
		context.Background(),
		t,
		rand.New(rand.NewSource(20200525)),
		pr,
		8,
		8,
	)
}

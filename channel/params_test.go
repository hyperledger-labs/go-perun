// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package channel_test

import (
	"math/rand"
	"testing"

	"perun.network/go-perun/channel/test"
	"perun.network/go-perun/pkg/io"
	iotest "perun.network/go-perun/pkg/io/test"
	pkgtest "perun.network/go-perun/pkg/test"
)

func TestParams_Clone(t *testing.T) {
	rng := rand.New(rand.NewSource(0xDDDDD))
	params := test.NewRandomParams(rng)
	pkgtest.VerifyClone(t, params)
}

func TestParams_Serializer(t *testing.T) {
	rng := rand.New(rand.NewSource(0xC00FED))
	params := make([]io.Serializer, 10)
	for i := range params {
		params[i] = test.NewRandomParams(rng)
	}

	iotest.GenericSerializerTest(t, params...)
}

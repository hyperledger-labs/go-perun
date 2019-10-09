// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package test // import "perun.network/go-perun/channel/test"

import (
	"math/rand"
	"testing"

	iotest "perun.network/go-perun/pkg/io/test"
)

func TestAsset(t *testing.T) {
	rng := rand.New(rand.NewSource(1337))
	asset := newRandomAsset(rng)
	iotest.GenericSerializableTest(t, asset)
}

func TestNoAppData(t *testing.T) {
	rng := rand.New(rand.NewSource(1337))
	data := newRandomNoAppData(rng)
	iotest.GenericSerializableTest(t, data)
}

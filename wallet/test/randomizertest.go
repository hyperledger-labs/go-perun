// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package test // import "perun.network/go-perun/wallet/test"

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func SetRandomizerTest(t *testing.T) {
	assert.Panics(t, func() { SetRandomizer(nil) }, "nil backend set should panic")
	require.NotNil(t, randomizer, "backend should be already set by init()")
	assert.Panics(t, func() { SetRandomizer(randomizer) }, "setting a backend twice should panic")
}

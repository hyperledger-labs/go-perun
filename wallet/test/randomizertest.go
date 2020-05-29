// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package test // import "perun.network/go-perun/wallet/test"

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SetRandomizerTest is a generic test to test that the wallet randomizer is set correctly.
func SetRandomizerTest(t *testing.T) {
	assert.Panics(t, func() { SetRandomizer(nil) }, "nil backend set should panic")
	require.NotNil(t, randomizer, "backend should be already set by init()")
	assert.Panics(t, func() { SetRandomizer(randomizer) }, "setting a backend twice should panic")
}

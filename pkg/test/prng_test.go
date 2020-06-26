// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type namer string

func (s namer) Name() string { return string(s) }

func TestSeedGeneration(t *testing.T) {
	s := Seed("123")
	assert.Equal(t, s, Seed("123"))
	assert.NotEqual(t, s, Seed("1234"))

	backup := rootSeed
	defer func() { rootSeed = backup }()
	rootSeed = 465
	assert.NotEqual(t, s, Seed("123"))
	assert.Equal(t, int64(-6222279139776267227), Seed("123"))
}

func TestPrng(t *testing.T) {
	pA := Prng(namer("123"))
	pB := Prng(namer("123"))
	// Same seeds should produce same random values.
	for i := 0; i < 100; i++ {
		require.Equal(t, pA.Int31(), pB.Int31())
	}
	// Different seeds should produce different random values.
	pA = Prng(namer("123"))
	pC := Prng(namer("1234"))
	i := 0
	for ; i < 100; i++ {
		if pA.Int31() != pC.Int31() {
			break
		}
	}
	assert.NotEqual(t, i, 100)
}

func TestGenRootSeed(t *testing.T) {
	// Unset GOTESTSEED and reset in defer.
	val, ok := os.LookupEnv(envTestSeed)
	if ok {
		require.NoError(t, os.Unsetenv(envTestSeed))
		defer require.NoError(t, os.Setenv(envTestSeed, val))
	}
	backup := rootSeed
	defer func() { rootSeed = backup }()

	// Seb promised to buy a beer crate if this test ever fails
	// in any pipeline (as long as the root seed is time.Now().UnixNano()).
	t.Run("Root-Changes", func(t *testing.T) {
		assert.NotEqual(t, genRootSeed(), genRootSeed())
	})
	t.Run("Root-Environment", testRootEnv)
}

func testRootEnv(t *testing.T) {
	seed := "123"
	require.NoError(t, os.Setenv(envTestSeed, seed))
	assert.Equal(t, genRootSeed(), genRootSeed())

	rootSeed = genRootSeed()
	assert.Equal(t, int64(8490245241735582052), Prng(namer("hi")).Int63())
}

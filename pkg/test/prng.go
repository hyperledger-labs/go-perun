// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package test

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"math/rand"
	"os"
	"strconv"
	"time"
)

const envTestSeed = "GOTESTSEED"

var rootSeed int64

func init() {
	rootSeed = genRootSeed()
	fmt.Printf("pkg/test: using rootSeed %d\n", rootSeed)
}

func genRootSeed() (rootSeed int64) {
	if val, ok := os.LookupEnv(envTestSeed); ok {
		rootSeed, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			panic("Could not parse GOTESTSEED as int64")
		}
		return rootSeed
	}
	return time.Now().UnixNano()
}

// Prng returns a pseudo-RNG that is seeded with the output of the `Seed`
// function by passing it `t.Name()`.
// Use it in tests with: rng := pkgtest.Prng(t)
func Prng(t interface{ Name() string }) *rand.Rand {
	return rand.New(rand.NewSource(Seed(t.Name())))
}

// Seed generates a seed that is dependent on the rootSeed and the passed
// seed argument.
// To fix this seed, set the GOTESTSEED environment variable.
// Example: GOTESTSEED=123 go test ./...
func Seed(seed string) int64 {
	hasher := fnv.New64a()
	if _, err := hasher.Write([]byte(seed)); err != nil {
		panic("Could not hash the seed")
	}
	if err := binary.Write(hasher, binary.LittleEndian, rootSeed); err != nil {
		panic("Could not hash the root seed")
	}
	return int64(hasher.Sum64())
}

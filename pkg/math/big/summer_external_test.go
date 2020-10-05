// Copyright 2020 - See NOTICE file for copyright holders.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package big_test

import (
	crand "crypto/rand"
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pkgbig "perun.network/go-perun/pkg/math/big"
	"perun.network/go-perun/pkg/test"
)

func TestSummer(t *testing.T) {
	rng := test.Prng(t)

	// Same summers do not error and are equal.
	for i := 0; i < 10; i++ {
		a := randomSummer(i, rng)
		equal, err := pkgbig.EqualSum(a, a)
		require.NoError(t, err)
		assert.True(t, equal)
	}

	// Same size summers do not error and are different.
	for i := 1; i < 10; i++ {
		a, b := randomSummer(i, rng), randomSummer(i, rng)
		equal, err := pkgbig.EqualSum(a, b)
		require.NoError(t, err)
		assert.False(t, equal)
	}

	// Different size summers do error.
	for i := 0; i < 10; i++ {
		a, b := randomSummer(i, rng), randomSummer(i+1, rng)
		equal, err := pkgbig.EqualSum(a, b)
		require.EqualError(t, err, "dimension mismatch")
		assert.False(t, equal)
	}
}

func TestAddSums(t *testing.T) {
	rng := test.Prng(t)

	// Calculates the correct sum.
	a, b := pkgbig.Sum{big.NewInt(4)}, pkgbig.Sum{big.NewInt(5)}
	sum, err := pkgbig.AddSums(a, b)
	require.NoError(t, err)
	require.Len(t, sum, 1)
	assert.Equal(t, sum[0], big.NewInt(9)) // 5 + 4 = 9

	// Different size summers do error.
	for i := 0; i < 10; i++ {
		a, b := randomSummer(i, rng), randomSummer(i+1, rng)
		_, err := pkgbig.AddSums(a, b)
		require.EqualError(t, err, "dimension mismatch")
	}

	// No input must return nil, nil.
	sum, err = pkgbig.AddSums()
	assert.Nil(t, err)
	assert.Nil(t, sum)
}

func randomSummer(len int, rng *rand.Rand) *pkgbig.Sum {
	data := make([]*big.Int, len)
	for i := range data {
		d, err := crand.Int(rng, new(big.Int).Lsh(big.NewInt(1), 255))
		if err != nil {
			panic(fmt.Sprintf("Creating random big.Int: %v", err))
		}
		data[i] = d
	}
	ret := pkgbig.Sum(data)
	return &ret
}

// TestAddSums_Const checks that `AddSums` does not modify the input.
func TestAddSums_Const(t *testing.T) {
	a, b := pkgbig.Sum{big.NewInt(4)}, pkgbig.Sum{big.NewInt(5)}

	_, err := pkgbig.AddSums(a, b)
	require.NoError(t, err)

	assert.Equal(t, a[0], big.NewInt(4))
	assert.Equal(t, b[0], big.NewInt(5))
}

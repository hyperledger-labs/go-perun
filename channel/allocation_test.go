// Copyright 2025 - See NOTICE file for copyright holders.
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

package channel_test

import (
	"math/big"
	"math/rand"
	"testing"

	"perun.network/go-perun/wallet"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/test"
	"perun.network/go-perun/wire/perunio"
	peruniotest "perun.network/go-perun/wire/perunio/test"
	pkgbig "polycry.pt/poly-go/math/big"
	pkgtest "polycry.pt/poly-go/test"
)

func TestAllocationNumParts(t *testing.T) {
	rng := pkgtest.Prng(t)
	tests := []struct {
		name  string
		alloc *channel.Allocation
		want  int
	}{
		{
			"empty balances",
			test.NewRandomAllocation(rng, test.WithNumParts(0), test.WithNumAssets(0)),
			-1,
		},
		{
			"single asset/three participants",
			test.NewRandomAllocation(rng, test.WithNumAssets(1), test.WithNumParts(3)),
			3,
		},
		{
			"three assets/three participants",
			test.NewRandomAllocation(rng, test.WithNumAssets(3), test.WithNumParts(3)),
			3,
		},
	}

	for _, _tt := range tests {
		tt := _tt
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.alloc.NumParts(); got != tt.want {
				t.Errorf("Allocation.NumParts() = %v, want valid = %v", got, tt.want)
			}
		})
	}
}

func randomBalancesWithMismatchingNumAssets(rng *rand.Rand, rngBase int) (b1, b2 channel.Balances) {
	numParts := 2 + rng.Intn(rngBase)

	randomNumAssets := func() int {
		return 1 + rng.Intn(rngBase)
	}
	numAssets1 := randomNumAssets()
	numAssets2 := randomNumAssets()
	for numAssets2 == numAssets1 {
		numAssets2 = randomNumAssets()
	}

	b1 = test.NewRandomBalances(rng, test.WithNumAssets(numAssets1), test.WithNumParts(numParts))
	b2 = test.NewRandomBalances(rng, test.WithNumAssets(numAssets2), test.WithNumParts(numParts))

	return
}

func randomBalancesWithMismatchingNumParts(rng *rand.Rand, rngBase int) (b1, b2 channel.Balances) {
	numAssets := 1 + rng.Intn(rngBase)

	randomNumParts := func() int {
		return 2 + rng.Intn(rngBase)
	}
	numParts1 := randomNumParts()
	numParts2 := randomNumParts()
	for numParts2 == numParts1 {
		numParts2 = randomNumParts()
	}

	b1 = test.NewRandomBalances(rng, test.WithNumAssets(numAssets), test.WithNumParts(numParts1))
	b2 = test.NewRandomBalances(rng, test.WithNumAssets(numAssets), test.WithNumParts(numParts2))

	return
}

func TestBalancesEqualAndAssertEqual(t *testing.T) {
	assert := assert.New(t)
	rng := pkgtest.Prng(t)
	const rngBase = 10

	t.Run("fails with mismatching number of assets", func(t *testing.T) {
		b1, b2 := randomBalancesWithMismatchingNumAssets(rng, rngBase)
		assert.False(b1.Equal(b2))
		assert.Error(b1.AssertEqual(b2))
	})

	t.Run("fails with mismatching number of parts", func(t *testing.T) {
		b1, b2 := randomBalancesWithMismatchingNumParts(rng, rngBase)
		assert.False(b1.Equal(b2))
		assert.Error(b1.AssertEqual(b2))
	})

	t.Run("compares correctly", func(t *testing.T) {
		numAssets := 1 + rng.Intn(rngBase)
		numParts := 2 + rng.Intn(rngBase)

		b1 := test.NewRandomBalances(rng, test.WithNumAssets(numAssets), test.WithNumParts(numParts), test.WithBalancesInRange(big.NewInt(0), big.NewInt(rngBase)))
		b2 := test.NewRandomBalances(rng, test.WithNumAssets(numAssets), test.WithNumParts(numParts), test.WithBalancesInRange(big.NewInt(rngBase), big.NewInt(2*rngBase)))

		assert.False(b1.Equal(b2))
		assert.Error(b1.AssertEqual(b2))

		assert.True(b1.Equal(b1)) //nolint:gocritic
		assert.NoError(b1.AssertEqual(b1))
	})
}

func TestBalancesGreaterOrEqual(t *testing.T) {
	assert := assert.New(t)
	rng := pkgtest.Prng(t)
	const rngBase = 10

	t.Run("fails with mismatching number of assets", func(t *testing.T) {
		b1, b2 := randomBalancesWithMismatchingNumAssets(rng, rngBase)
		assert.Error(b1.AssertGreaterOrEqual(b2))
	})

	t.Run("fails with mismatching number of parts", func(t *testing.T) {
		b1, b2 := randomBalancesWithMismatchingNumParts(rng, rngBase)
		assert.Error(b1.AssertGreaterOrEqual(b2))
	})

	t.Run("compares correctly", func(t *testing.T) {
		numAssets := 1 + rng.Intn(rngBase)
		numParts := 2 + rng.Intn(rngBase)

		b1 := test.NewRandomBalances(rng, test.WithNumAssets(numAssets), test.WithNumParts(numParts), test.WithBalancesInRange(big.NewInt(0), big.NewInt(rngBase)))
		b2 := test.NewRandomBalances(rng, test.WithNumAssets(numAssets), test.WithNumParts(numParts), test.WithBalancesInRange(big.NewInt(rngBase), big.NewInt(2*rngBase)))

		assert.Error(b1.AssertGreaterOrEqual(b2))
		assert.NoError(b2.AssertGreaterOrEqual(b1))
	})
}

func TestBalancesEqualSum(t *testing.T) {
	rng := pkgtest.Prng(t)
	for i := 0; i < 10; i++ {
		// Two random balances are different by chance.
		a, b := test.NewRandomBalances(rng), test.NewRandomBalances(rng)
		ok, err := pkgbig.EqualSum(a, b)
		if len(a) != len(b) {
			require.Error(t, err, "Sum with different dimensions should error")
		} else {
			assert.False(t, ok, "a and b should have different sums")
		}
		// Shuffling must not change the sum.
		s := test.ShuffleBalances(rng, a)
		ok, err = pkgbig.EqualSum(a, s)
		require.NoError(t, err)
		assert.True(t, ok)
	}
}

func TestBalancesAdd(t *testing.T) {
	testBalancesOperation(
		t,
		func(b1, b2 channel.Balances) channel.Balances { return b1.Add(b2) },
		func(b1, b2 channel.Bal) channel.Bal { return new(big.Int).Add(b1, b2) },
	)
}

func TestBalancesSub(t *testing.T) {
	testBalancesOperation(
		t,
		func(b1, b2 channel.Balances) channel.Balances { return b1.Sub(b2) },
		func(b1, b2 channel.Bal) channel.Bal { return new(big.Int).Sub(b1, b2) },
	)
}

func testBalancesOperation(t *testing.T, op func(channel.Balances, channel.Balances) channel.Balances, elementOp func(channel.Bal, channel.Bal) channel.Bal) {
	t.Helper()
	assert := assert.New(t)
	rng := pkgtest.Prng(t)
	const rngBase = 10

	t.Run("fails with mismatching number of assets", func(t *testing.T) {
		b1, b2 := randomBalancesWithMismatchingNumAssets(rng, rngBase)
		assert.Panics(func() { op(b1, b2) })
	})

	t.Run("fails with mismatching number of parts", func(t *testing.T) {
		b1, b2 := randomBalancesWithMismatchingNumParts(rng, rngBase)
		assert.Panics(func() { op(b1, b2) })
	})

	t.Run("calculates correctly", func(t *testing.T) {
		numAssets := 1 + rng.Intn(rngBase)
		numParts := 2 + rng.Intn(rngBase)

		b1 := test.NewRandomBalances(rng, test.WithNumAssets(numAssets), test.WithNumParts(numParts))
		b2 := test.NewRandomBalances(rng, test.WithNumAssets(numAssets), test.WithNumParts(numParts))

		b0 := op(b1, b2)

		for a := range b0 {
			for u := range b0[a] {
				expected := elementOp(b1[a][u], b2[a][u])
				got := b0[a][u]
				assert.Zero(got.Cmp(expected), "value mismatch at index [%d, %d]: expected %v, got %v", a, u, expected, got)
			}
		}
	})
}

func TestBalancesSerialization(t *testing.T) {
	rng := pkgtest.Prng(t)
	for n := 0; n < 10; n++ {
		alloc := test.NewRandomAllocation(rng)
		if alloc.Valid() == nil {
			peruniotest.GenericSerializerTest(t, alloc)
		}
	}
}

func TestAllocationSerialization(t *testing.T) {
	rng := pkgtest.Prng(t)
	inputs := []perunio.Serializer{
		test.NewRandomAllocation(rng, test.WithNumParts(1), test.WithNumAssets(1), test.WithNumLocked(0)),
		test.NewRandomAllocation(rng, test.WithNumParts(1), test.WithNumAssets(1), test.WithNumLocked(1)),
		test.NewRandomAllocation(rng, test.WithNumParts(3), test.WithNumAssets(2), test.WithNumLocked(1)),
	}

	peruniotest.GenericSerializerTest(t, inputs...)
}

func TestAllocationValidLimits(t *testing.T) {
	rng := pkgtest.Prng(t)
	inputs := []struct {
		numAssets         int
		numParts          int
		numSuballocations int
	}{
		{channel.MaxNumAssets + 1, 1, 0},
		{1, channel.MaxNumParts + 1, 0},
		{1, 1, channel.MaxNumSubAllocations + 1},
		{
			channel.MaxNumAssets + 2,
			2 * channel.MaxNumParts,
			4 * channel.MaxNumSubAllocations,
		},
	}

	for ti, x := range inputs {
		allocation := &channel.Allocation{
			Assets:   make([]channel.Asset, x.numAssets),
			Backends: make([]wallet.BackendID, x.numAssets),
			Balances: make(channel.Balances, x.numAssets),
			Locked:   make([]channel.SubAlloc, x.numSuballocations),
		}

		allocation.Assets = test.NewRandomAssets(rng, test.WithNumAssets(x.numAssets))
		for i := range allocation.Assets {
			allocation.Backends[i] = channel.TestBackendID
		}

		for i := range allocation.Balances {
			for j := range allocation.Balances[i] {
				bal := big.NewInt(int64(x.numAssets)*int64(i) + int64(j))
				allocation.Balances[i][j] = bal
			}
		}

		for i := range allocation.Locked {
			allocation.Locked[i] = *channel.NewSubAlloc(
				channel.ID{byte(i), byte(i >> 8), byte(i >> 16), byte(i >> 24)},
				make([]channel.Bal, x.numAssets),
				nil,
			)
			allocation.Locked[i] = *channel.NewSubAlloc(
				channel.ID{byte(i), byte(i >> 8), byte(i >> 16), byte(i >> 24)},
				make([]channel.Bal, x.numAssets),
				nil,
			)

			for j := range allocation.Locked[i].Bals {
				bal := big.NewInt(int64(x.numAssets)*int64(i) + int64(j) + 1)
				allocation.Locked[i].Bals[j] = bal
			}
		}

		assert.Errorf(t, allocation.Valid(), "[%d] expected error for parameters %v", ti, x)
	}
}

func TestAllocation_Clone(t *testing.T) {
	rng := pkgtest.Prng(t)
	tests := []struct {
		name  string
		alloc channel.Allocation
	}{
		{
			"assets-1,parts-1,locks-nil",
			*test.NewRandomAllocation(rng, test.WithNumAssets(1), test.WithNumParts(1)),
		},
		{
			"assets-1,parts-1,locks",
			*test.NewRandomAllocation(rng, test.WithNumAssets(1), test.WithNumParts(1)),
		},
		{
			"assets-2,parties-4,locks-nil",
			*test.NewRandomAllocation(rng, test.WithNumAssets(2), test.WithNumParts(4)),
		},
		{
			"assets-2,parties-4,locks",
			*test.NewRandomAllocation(rng, test.WithNumAssets(2), test.WithNumParts(4)),
		},
	}

	for _, _tt := range tests {
		tt := _tt
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.alloc.Valid(); err != nil {
				t.Fatal(err.Error())
			}
			pkgtest.VerifyClone(t, tt.alloc)
		})
	}
}

func TestAllocation_Sum(t *testing.T) {
	rng := pkgtest.Prng(t)

	// note: different invalid allocations are tested in TestAllocation_valid

	// valid Allocations
	tests := []struct {
		name  string
		alloc channel.Allocation
		want  []channel.Bal
	}{
		{
			"single asset/one participant",
			*test.NewRandomAllocation(rng, test.WithNumAssets(1), test.WithNumParts(1), test.WithBalancesInRange(big.NewInt(1), big.NewInt(1))),
			[]channel.Bal{big.NewInt(1)},
		},

		{
			"single asset/one participant/empty locked slice",
			*test.NewRandomAllocation(rng, test.WithNumAssets(1), test.WithNumParts(1), test.WithBalancesInRange(big.NewInt(1), big.NewInt(1))),
			[]channel.Bal{big.NewInt(1)},
		},

		{
			"single asset/three participants",
			*test.NewRandomAllocation(rng, test.WithNumAssets(1), test.WithNumParts(1), test.WithBalances([]channel.Bal{big.NewInt(1), big.NewInt(2), big.NewInt(4)})),
			[]channel.Bal{big.NewInt(7)},
		},

		{
			"three assets/three participants",
			*test.NewRandomAllocation(rng, test.WithNumAssets(3), test.WithNumParts(3), test.WithBalances(channel.Balances{
				{big.NewInt(1), big.NewInt(2), big.NewInt(4)},
				{big.NewInt(8), big.NewInt(16), big.NewInt(32)},
				{big.NewInt(64), big.NewInt(128), big.NewInt(256)},
			}...)),
			[]channel.Bal{big.NewInt(7), big.NewInt(56), big.NewInt(448)},
		},

		{
			"single asset/one participants/one locked",
			*test.NewRandomAllocation(rng, test.WithNumAssets(1), test.WithNumParts(1), test.WithLocked(*channel.NewSubAlloc(channel.ID{}, []channel.Bal{big.NewInt(2)}, nil)), test.WithBalancesInRange(big.NewInt(1), big.NewInt(1))),
			[]channel.Bal{big.NewInt(3)},
		},

		{
			"three assets/two participants/three locked",
			*test.NewRandomAllocation(rng, test.WithNumAssets(3), test.WithNumParts(3), test.WithLocked(
				*test.NewRandomSubAlloc(rng, test.WithLockedBals(big.NewInt(4), big.NewInt(0x80), big.NewInt(0x1000))),
				*test.NewRandomSubAlloc(rng, test.WithLockedBals(big.NewInt(8), big.NewInt(0x100), big.NewInt(0x2000))),
				*test.NewRandomSubAlloc(rng, test.WithLockedBals(big.NewInt(0x10), big.NewInt(0x200), big.NewInt(0x4000))),
			), test.WithBalances(channel.Balances{
				{big.NewInt(1), big.NewInt(2)},
				{big.NewInt(0x20), big.NewInt(0x40)},
				{big.NewInt(0x400), big.NewInt(0x800)},
			}...)),
			[]channel.Bal{big.NewInt(0x1f), big.NewInt(0x3e0), big.NewInt(0x7c00)},
		},
	}

	for _, _tt := range tests {
		tt := _tt
		t.Run(tt.name, func(t *testing.T) {
			for i, got := range tt.alloc.Sum() {
				if got.Cmp(tt.want[i]) != 0 {
					t.Errorf("Allocation.Sum()[%d] = %v, want %v", i, got, tt.want[i])
				}
			}
		})
	}
}

func TestAllocation_Valid(t *testing.T) {
	rng := pkgtest.Prng(t)
	// note that all valid branches are already indirectly tested in TestAllocation_Sum
	tests := []struct {
		name  string
		alloc channel.Allocation
		valid bool
	}{
		{
			"one participant/no locked valid",
			channel.Allocation{
				Assets:   test.NewRandomAssets(rng, test.WithNumAssets(1)),
				Balances: channel.Balances{{big.NewInt(1)}},
				Locked:   nil,
			},
			true,
		},

		{
			"nil asset/nil participant",
			channel.Allocation{
				Assets:   nil,
				Balances: nil,
				Locked:   nil,
			},
			false,
		},

		{
			"nil participant/no locked",
			channel.Allocation{
				Assets:   test.NewRandomAssets(rng, test.WithNumAssets(1)),
				Balances: nil,
				Locked:   nil,
			},
			false,
		},

		{
			"no participant/no locked",
			channel.Allocation{
				Assets:   test.NewRandomAssets(rng, test.WithNumAssets(1)),
				Balances: make(channel.Balances, 0),
			},
			false,
		},

		{
			"two participants/wrong number of asset",
			channel.Allocation{
				Assets: test.NewRandomAssets(rng, test.WithNumAssets(3)),
				Balances: channel.Balances{
					{big.NewInt(1), big.NewInt(8)},
					{big.NewInt(2), big.NewInt(16)},
				},
			},
			false,
		},

		{
			"three assets/wrong number of participants",
			channel.Allocation{
				Assets: test.NewRandomAssets(rng, test.WithNumAssets(3)),
				Balances: channel.Balances{
					{big.NewInt(1), big.NewInt(8)},
					{big.NewInt(2), big.NewInt(16)},
					{big.NewInt(64)},
				},
			},
			false,
		},

		{
			"two participants/one locked wrong dimension",
			channel.Allocation{
				Assets: test.NewRandomAssets(rng, test.WithNumAssets(3)),
				Balances: channel.Balances{
					{big.NewInt(1), big.NewInt(2)},
					{big.NewInt(8), big.NewInt(16)},
					{big.NewInt(64), big.NewInt(128)},
				},
				Locked: []channel.SubAlloc{
					*channel.NewSubAlloc(channel.Zero, []channel.Bal{big.NewInt(4)}, nil),
				},
			},
			false,
		},

		{
			"three assets/one locked invalid dimension",
			channel.Allocation{
				Assets: test.NewRandomAssets(rng, test.WithNumAssets(3)),
				Balances: channel.Balances{
					{big.NewInt(1), big.NewInt(2)},
					{big.NewInt(8), big.NewInt(16)},
					{big.NewInt(64), big.NewInt(128)},
				},
				Locked: []channel.SubAlloc{
					*channel.NewSubAlloc(channel.Zero, []channel.Bal{big.NewInt(-1)}, nil),
				},
			},
			false,
		},
		{
			"two participants/negative balance",
			channel.Allocation{
				Assets: test.NewRandomAssets(rng, test.WithNumAssets(3)),
				Balances: channel.Balances{
					{big.NewInt(1), big.NewInt(2)},
					{big.NewInt(8), big.NewInt(-1)},
					{big.NewInt(64), big.NewInt(128)},
				},
				Locked: nil,
			},
			false,
		},

		{
			"two participants/one locked negative balance",
			channel.Allocation{
				Assets: test.NewRandomAssets(rng, test.WithNumAssets(2)),
				Balances: channel.Balances{
					{big.NewInt(1), big.NewInt(8)},
					{big.NewInt(2), big.NewInt(16)},
				},
				Locked: []channel.SubAlloc{
					*channel.NewSubAlloc(channel.Zero, []channel.Bal{big.NewInt(4), big.NewInt(-1)}, nil),
				},
			},
			false,
		},
	}

	for _, _tt := range tests {
		tt := _tt
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.alloc.Valid(); (got == nil) != tt.valid {
				t.Errorf("Allocation.valid() = %v, want valid = %v", got, tt.valid)
			}
		})
	}
}

// suballocation serialization.
func TestSuballocSerialization(t *testing.T) {
	ss := []perunio.Serializer{
		channel.NewSubAlloc(channel.ID{2}, []channel.Bal{}, nil),
		channel.NewSubAlloc(channel.ID{3}, []channel.Bal{big.NewInt(0)}, nil),
		channel.NewSubAlloc(channel.ID{4}, []channel.Bal{big.NewInt(5), big.NewInt(1 << 62)}, nil),
	}

	peruniotest.GenericSerializerTest(t, ss...)
}

func TestRemoveSubAlloc(t *testing.T) {
	assert := assert.New(t)
	rng := pkgtest.Prng(t)

	alloc := test.NewRandomAllocation(rng, test.WithNumLocked(3))
	lenBefore := len(alloc.Locked)
	subAlloc := alloc.Locked[rng.Intn(lenBefore)]

	require.NoError(t, alloc.RemoveSubAlloc(subAlloc), "removing contained element should not fail")

	assert.Equal(lenBefore-1, len(alloc.Locked), "length should decrease by 1")

	_, ok := alloc.SubAlloc(subAlloc.ID)
	assert.False(ok, "element should not be found after removal") // this could potentially fail because duplicates are currently not removed

	assert.Error(alloc.RemoveSubAlloc(subAlloc), "removing not-contained element should fail")
}

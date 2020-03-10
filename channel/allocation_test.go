// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package channel_test

import (
	"math/big"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/test"
	perunio "perun.network/go-perun/pkg/io"
	iotest "perun.network/go-perun/pkg/io/test"
	pkgtest "perun.network/go-perun/pkg/test"
)

func TestAllocationSerialization(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	inputs := []perunio.Serializer{
		&channel.Allocation{
			Assets:   test.NewRandomAssets(rng, 1),
			Balances: test.NewRandomBalances(rng, 1, 1),
			Locked:   []channel.SubAlloc{},
		},
		&channel.Allocation{
			Assets:   test.NewRandomAssets(rng, 1),
			Balances: test.NewRandomBalances(rng, 1, 1),
			Locked: []channel.SubAlloc{
				{
					ID:   channel.ID{0},
					Bals: test.NewRandomBals(rng, 1)},
			},
		},
		&channel.Allocation{
			Assets:   test.NewRandomAssets(rng, 3),
			Balances: test.NewRandomBalances(rng, 3, 2),
			Locked: []channel.SubAlloc{
				{
					ID:   channel.ID{0},
					Bals: test.NewRandomBals(rng, 3),
				},
			},
		},
	}

	iotest.GenericSerializerTest(t, inputs...)
}

func TestAllocationValidLimits(t *testing.T) {
	rng := rand.New(rand.NewSource(1337))
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
			4 * channel.MaxNumSubAllocations},
	}

	for ti, x := range inputs {
		allocation := &channel.Allocation{
			Assets:   make([]channel.Asset, x.numAssets),
			Balances: make([][]channel.Bal, x.numAssets),
			Locked:   make([]channel.SubAlloc, x.numSuballocations),
		}

		allocation.Assets = test.NewRandomAssets(rng, x.numAssets)

		for i := range allocation.Balances {
			for j := range allocation.Balances[i] {
				bal := big.NewInt(int64(x.numAssets)*int64(i) + int64(j))
				allocation.Balances[i][j] = bal
			}
		}

		for i := range allocation.Locked {
			allocation.Locked[i] = channel.SubAlloc{
				ID:   channel.ID{byte(i), byte(i >> 8), byte(i >> 16), byte(i >> 24)},
				Bals: make([]channel.Bal, x.numAssets)}

			for j := range allocation.Locked[i].Bals {
				bal := big.NewInt(int64(x.numAssets)*int64(i) + int64(j) + 1)
				allocation.Locked[i].Bals[j] = bal
			}
		}

		assert.Errorf(t, allocation.Valid(), "[%d] expected error for parameters %v", ti, x)
	}
}

func TestAllocation_Clone(t *testing.T) {
	rng := rand.New(rand.NewSource(1337))
	tests := []struct {
		name  string
		alloc channel.Allocation
	}{
		{
			"test.NewRandomAssets-1,parts-1,locks-nil",
			channel.Allocation{
				test.NewRandomAssets(rng, 1),
				test.NewRandomBalances(rng, 1, 1),
				nil,
			},
		},

		{
			"test.NewRandomAssets-1,parts-1,locks",
			channel.Allocation{
				test.NewRandomAssets(rng, 1),
				test.NewRandomBalances(rng, 1, 1),
				[]channel.SubAlloc{{channel.ID{123}, test.NewRandomBals(rng, 1)}},
			},
		},

		{
			"test.NewRandomAssets-2,parties-4,locks-nil",
			channel.Allocation{
				test.NewRandomAssets(rng, 2),
				test.NewRandomBalances(rng, 2, 4),
				nil,
			},
		},

		{
			"test.NewRandomAssets-2,parties-4,locks",
			channel.Allocation{
				test.NewRandomAssets(rng, 2),
				test.NewRandomBalances(rng, 2, 4),
				[]channel.SubAlloc{
					{channel.ID{1}, test.NewRandomBals(rng, 2)},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.alloc.Valid(); err != nil {
				t.Fatal(err.Error())
			}
			pkgtest.VerifyClone(t, tt.alloc)
		})
	}
}

func TestAllocation_Sum(t *testing.T) {
	rng := rand.New(rand.NewSource(1337))

	// note: different invalid allocations are tested in TestAllocation_valid

	// valid Allocations
	tests := []struct {
		name  string
		alloc channel.Allocation
		want  []channel.Bal
	}{
		{
			"single asset/one participant",
			channel.Allocation{
				Assets:   test.NewRandomAssets(rng, 1),
				Balances: [][]channel.Bal{{big.NewInt(1)}},
			},
			[]channel.Bal{big.NewInt(1)},
		},

		{
			"single asset/one participant/empty locked slice",
			channel.Allocation{
				Assets:   test.NewRandomAssets(rng, 1),
				Balances: [][]channel.Bal{{big.NewInt(1)}},
				Locked:   make([]channel.SubAlloc, 0),
			},
			[]channel.Bal{big.NewInt(1)},
		},

		{
			"single asset/three participants",
			channel.Allocation{
				Assets: test.NewRandomAssets(rng, 1),
				Balances: [][]channel.Bal{
					{big.NewInt(1), big.NewInt(2), big.NewInt(4)},
				},
			},
			[]channel.Bal{big.NewInt(7)},
		},

		{
			"three test.NewRandomAssets/three participants",
			channel.Allocation{
				Assets: test.NewRandomAssets(rng, 3),
				Balances: [][]channel.Bal{
					{big.NewInt(1), big.NewInt(2), big.NewInt(4)},
					{big.NewInt(8), big.NewInt(16), big.NewInt(32)},
					{big.NewInt(64), big.NewInt(128), big.NewInt(256)},
				},
			},
			[]channel.Bal{big.NewInt(7), big.NewInt(56), big.NewInt(448)},
		},

		{
			"single test.NewRandomAssets/one participants/one locked",
			channel.Allocation{
				Assets: test.NewRandomAssets(rng, 1),
				Balances: [][]channel.Bal{
					{big.NewInt(1)},
				},
				Locked: []channel.SubAlloc{
					{channel.Zero, []channel.Bal{big.NewInt(2)}},
				},
			},
			[]channel.Bal{big.NewInt(3)},
		},

		{
			"three test.NewRandomAssets/two participants/three locked",
			channel.Allocation{
				Assets: test.NewRandomAssets(rng, 3),
				Balances: [][]channel.Bal{
					{big.NewInt(1), big.NewInt(2)},
					{big.NewInt(0x20), big.NewInt(0x40)},
					{big.NewInt(0x400), big.NewInt(0x800)},
				},
				Locked: []channel.SubAlloc{
					{channel.Zero, []channel.Bal{big.NewInt(4), big.NewInt(0x80), big.NewInt(0x1000)}},
					{channel.Zero, []channel.Bal{big.NewInt(8), big.NewInt(0x100), big.NewInt(0x2000)}},
					{channel.Zero, []channel.Bal{big.NewInt(0x10), big.NewInt(0x200), big.NewInt(0x4000)}},
				},
			},
			[]channel.Bal{big.NewInt(0x1f), big.NewInt(0x3e0), big.NewInt(0x7c00)},
		},
	}

	for _, tt := range tests {
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
	rng := rand.New(rand.NewSource(1337))
	// note that all valid branches are already indirectly tested in TestAllocation_Sum
	tests := []struct {
		name  string
		alloc channel.Allocation
		valid bool
	}{
		{
			"one participant/no locked valid",
			channel.Allocation{
				Assets:   test.NewRandomAssets(rng, 1),
				Balances: [][]channel.Bal{{big.NewInt(1)}},
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
				Assets:   test.NewRandomAssets(rng, 1),
				Balances: nil,
				Locked:   nil,
			},
			false,
		},

		{
			"no participant/no locked",
			channel.Allocation{
				Assets:   test.NewRandomAssets(rng, 1),
				Balances: make([][]channel.Bal, 0),
			},
			false,
		},

		{
			"two participants wrong dimension",
			channel.Allocation{
				Assets: test.NewRandomAssets(rng, 3),
				Balances: [][]channel.Bal{
					{big.NewInt(1), big.NewInt(8), big.NewInt(64)},
					{big.NewInt(2), big.NewInt(16)},
				},
			},
			false,
		},

		{
			"two participants/one locked wrong dimension",
			channel.Allocation{
				Assets: test.NewRandomAssets(rng, 3),
				Balances: [][]channel.Bal{
					{big.NewInt(1), big.NewInt(2)},
					{big.NewInt(8), big.NewInt(16)},
					{big.NewInt(64), big.NewInt(128)},
				},
				Locked: []channel.SubAlloc{
					{channel.Zero, []channel.Bal{big.NewInt(4)}},
				},
			},
			false,
		},

		{
			"two participants/negative balance",
			channel.Allocation{
				Assets: test.NewRandomAssets(rng, 3),
				Balances: [][]channel.Bal{
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
				Assets: test.NewRandomAssets(rng, 2),
				Balances: [][]channel.Bal{
					{big.NewInt(1), big.NewInt(8)},
					{big.NewInt(2), big.NewInt(16)},
				},
				Locked: []channel.SubAlloc{
					{channel.Zero, []channel.Bal{big.NewInt(4), big.NewInt(-1)}},
				},
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.alloc.Valid(); (got == nil) != tt.valid {
				t.Errorf("Allocation.valid() = %v, want valid = %v", got, tt.valid)
			}
		})
	}
}

// suballocation serialization
func TestSuballocSerialization(t *testing.T) {
	ss := []perunio.Serializer{
		&channel.SubAlloc{channel.ID{2}, []channel.Bal{}},
		&channel.SubAlloc{channel.ID{3}, []channel.Bal{big.NewInt(0)}},
		&channel.SubAlloc{channel.ID{4}, []channel.Bal{big.NewInt(5), big.NewInt(1 << 62)}},
	}

	iotest.GenericSerializerTest(t, ss...)
}

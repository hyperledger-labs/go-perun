// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

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

func assets(rng *rand.Rand, n uint) []channel.Asset {
	as := make([]channel.Asset, n)
	for i := uint(0); i < n; i++ {
		as[i] = test.NewRandomAsset(rng)
	}
	return as
}

func TestAllocationSerialization(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	inputs := []perunio.Serializable{
		&channel.Allocation{
			Assets:  assets(rng, 1),
			OfParts: [][]channel.Bal{[]channel.Bal{big.NewInt(123)}},
			Locked:  []channel.SubAlloc{},
		},
		&channel.Allocation{
			Assets: assets(rng, 1),
			OfParts: [][]channel.Bal{
				[]channel.Bal{big.NewInt(1)},
			},
			Locked: []channel.SubAlloc{
				channel.SubAlloc{
					ID:   channel.ID{0},
					Bals: []channel.Bal{big.NewInt(2)}},
			},
		},
		&channel.Allocation{
			Assets: assets(rng, 3),
			OfParts: [][]channel.Bal{
				[]channel.Bal{big.NewInt(1), big.NewInt(10), big.NewInt(100)},
				[]channel.Bal{big.NewInt(7), big.NewInt(11), big.NewInt(13)},
			},
			Locked: []channel.SubAlloc{
				channel.SubAlloc{
					ID: channel.ID{0},
					Bals: []channel.Bal{
						big.NewInt(1), big.NewInt(3), big.NewInt(5),
					},
				},
			},
		},
	}

	iotest.GenericSerializableTest(t, inputs...)
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
			Assets:  make([]channel.Asset, x.numAssets),
			OfParts: make([][]channel.Bal, x.numParts),
			Locked:  make([]channel.SubAlloc, x.numSuballocations),
		}

		for i := range allocation.Assets {
			allocation.Assets[i] = test.NewRandomAsset(rng)
		}

		for i := range allocation.OfParts {
			allocation.OfParts[i] = make([]channel.Bal, x.numAssets)

			for j := range allocation.OfParts[i] {
				bal := big.NewInt(int64(x.numAssets)*int64(i) + int64(j))
				allocation.OfParts[i][j] = bal
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
			"assets-1,parts-1,locks-nil",
			channel.Allocation{assets(rng, 1), [][]channel.Bal{[]channel.Bal{big.NewInt(-1)}}, nil},
		},

		{
			"assets-1,parts-1,locks",
			channel.Allocation{
				assets(rng, 1),
				[][]channel.Bal{[]channel.Bal{big.NewInt(0)}},
				[]channel.SubAlloc{channel.SubAlloc{channel.ID{123}, []*big.Int{big.NewInt(0)}}},
			},
		},

		{
			"assets-2,parties-4,locks-nil",
			channel.Allocation{
				assets(rng, 2),
				[][]channel.Bal{
					[]channel.Bal{big.NewInt(1), big.NewInt(11)},
					[]channel.Bal{big.NewInt(2), big.NewInt(2)},
					[]channel.Bal{big.NewInt(3), big.NewInt(5)},
					[]channel.Bal{big.NewInt(10), big.NewInt(2)},
				},
				nil,
			},
		},

		{
			"assets-2,parties-4,locks",
			channel.Allocation{
				assets(rng, 2),
				[][]channel.Bal{
					[]channel.Bal{big.NewInt(1), big.NewInt(11)},
					[]channel.Bal{big.NewInt(2), big.NewInt(2)},
					[]channel.Bal{big.NewInt(3), big.NewInt(5)},
					[]channel.Bal{big.NewInt(10), big.NewInt(2)},
				},
				[]channel.SubAlloc{
					channel.SubAlloc{channel.ID{1}, []channel.Bal{big.NewInt(1), big.NewInt(2)}},
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
	// invalid Allocation
	invalidAllocation := channel.Allocation{}
	assert.Panics(t, func() { invalidAllocation.Sum() })

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
				Assets:  assets(rng, 1),
				OfParts: [][]channel.Bal{[]channel.Bal{big.NewInt(1)}},
			},
			[]channel.Bal{big.NewInt(1)},
		},

		{
			"single asset/one participant/empty locked slice",
			channel.Allocation{
				Assets:  assets(rng, 1),
				OfParts: [][]channel.Bal{[]channel.Bal{big.NewInt(1)}},
				Locked:  make([]channel.SubAlloc, 0),
			},
			[]channel.Bal{big.NewInt(1)},
		},

		{
			"single asset/three participants",
			channel.Allocation{
				Assets: assets(rng, 1),
				OfParts: [][]channel.Bal{
					[]channel.Bal{big.NewInt(1)},
					[]channel.Bal{big.NewInt(2)},
					[]channel.Bal{big.NewInt(4)},
				},
			},
			[]channel.Bal{big.NewInt(7)},
		},

		{
			"three assets/three participants",
			channel.Allocation{
				Assets: assets(rng, 3),
				OfParts: [][]channel.Bal{
					[]channel.Bal{big.NewInt(1), big.NewInt(8), big.NewInt(64)},
					[]channel.Bal{big.NewInt(2), big.NewInt(16), big.NewInt(128)},
					[]channel.Bal{big.NewInt(4), big.NewInt(32), big.NewInt(256)},
				},
			},
			[]channel.Bal{big.NewInt(7), big.NewInt(56), big.NewInt(448)},
		},

		{
			"single assets/one participants/one locked",
			channel.Allocation{
				Assets: assets(rng, 1),
				OfParts: [][]channel.Bal{
					[]channel.Bal{big.NewInt(1)},
				},
				Locked: []channel.SubAlloc{
					channel.SubAlloc{channel.Zero, []channel.Bal{big.NewInt(2)}},
				},
			},
			[]channel.Bal{big.NewInt(3)},
		},

		{
			"three assets/two participants/three locked",
			channel.Allocation{
				Assets: assets(rng, 3),
				OfParts: [][]channel.Bal{
					[]channel.Bal{big.NewInt(1), big.NewInt(0x20), big.NewInt(0x400)},
					[]channel.Bal{big.NewInt(2), big.NewInt(0x40), big.NewInt(0x800)},
				},
				Locked: []channel.SubAlloc{
					channel.SubAlloc{channel.Zero, []channel.Bal{big.NewInt(4), big.NewInt(0x80), big.NewInt(0x1000)}},
					channel.SubAlloc{channel.Zero, []channel.Bal{big.NewInt(8), big.NewInt(0x100), big.NewInt(0x2000)}},
					channel.SubAlloc{channel.Zero, []channel.Bal{big.NewInt(0x10), big.NewInt(0x200), big.NewInt(0x4000)}},
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
				Assets:  assets(rng, 1),
				OfParts: [][]channel.Bal{[]channel.Bal{big.NewInt(1)}},
				Locked:  nil,
			},
			true,
		},

		{
			"nil asset/nil participant",
			channel.Allocation{
				Assets:  nil,
				OfParts: nil,
				Locked:  nil,
			},
			false,
		},

		{
			"nil participant/no locked",
			channel.Allocation{
				Assets:  assets(rng, 1),
				OfParts: nil,
				Locked:  nil,
			},
			false,
		},

		{
			"no participant/no locked",
			channel.Allocation{
				Assets:  assets(rng, 1),
				OfParts: make([][]channel.Bal, 0),
			},
			false,
		},

		{
			"two participants wrong dimension",
			channel.Allocation{
				Assets: assets(rng, 3),
				OfParts: [][]channel.Bal{
					[]channel.Bal{big.NewInt(1), big.NewInt(8), big.NewInt(64)},
					[]channel.Bal{big.NewInt(2), big.NewInt(16)},
				},
			},
			false,
		},

		{
			"two participants/one locked wrong dimension",
			channel.Allocation{
				Assets: assets(rng, 3),
				OfParts: [][]channel.Bal{
					[]channel.Bal{big.NewInt(1), big.NewInt(8), big.NewInt(64)},
					[]channel.Bal{big.NewInt(2), big.NewInt(16), big.NewInt(128)},
				},
				Locked: []channel.SubAlloc{
					channel.SubAlloc{channel.Zero, []channel.Bal{big.NewInt(4)}},
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
	ss := []perunio.Serializable{
		&channel.SubAlloc{channel.ID{2}, []channel.Bal{}},
		&channel.SubAlloc{channel.ID{3}, []channel.Bal{big.NewInt(0)}},
		&channel.SubAlloc{channel.ID{4}, []channel.Bal{big.NewInt(5), big.NewInt(1 << 62)}},
	}

	iotest.GenericSerializableTest(t, ss...)
}

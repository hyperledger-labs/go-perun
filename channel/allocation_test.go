// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel

import (
	"io"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	perunIo "perun.network/go-perun/pkg/io"
	ioTest "perun.network/go-perun/pkg/io/test"
	"perun.network/go-perun/pkg/test"
)

// asset is a test asset
type asset struct{}

// pkg/io.Serializable interface

func (a asset) Decode(io.Reader) error {
	return nil
}

func (a asset) Encode(io.Writer) error {
	return nil
}

func assets(n uint) []Asset {
	as := make([]Asset, n)
	for i := uint(0); i < n; i++ {
		as[i] = new(asset)
	}
	return as
}

func TestAllocation_Clone(t *testing.T) {
	tests := []struct {
		name  string
		alloc Allocation
	}{
		{
			"assets-1,parts-1,locks-nil",
			Allocation{assets(1), [][]Bal{[]Bal{big.NewInt(-1)}}, nil},
		},

		{
			"assets-1,parts-1,locks",
			Allocation{
				assets(1),
				[][]Bal{[]Bal{big.NewInt(0)}},
				[]SubAlloc{SubAlloc{ID{123}, []*big.Int{big.NewInt(0)}}},
			},
		},

		{
			"assets-2,parties-4,locks-nil",
			Allocation{
				assets(2),
				[][]Bal{
					[]Bal{big.NewInt(1), big.NewInt(11)},
					[]Bal{big.NewInt(2), big.NewInt(2)},
					[]Bal{big.NewInt(3), big.NewInt(5)},
					[]Bal{big.NewInt(10), big.NewInt(2)},
				},
				nil,
			},
		},

		{
			"assets-2,parties-4,locks",
			Allocation{
				assets(2),
				[][]Bal{
					[]Bal{big.NewInt(1), big.NewInt(11)},
					[]Bal{big.NewInt(2), big.NewInt(2)},
					[]Bal{big.NewInt(3), big.NewInt(5)},
					[]Bal{big.NewInt(10), big.NewInt(2)},
				},
				[]SubAlloc{
					SubAlloc{ID{1}, []Bal{big.NewInt(1), big.NewInt(2)}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.alloc.valid(); err != nil {
				t.Fatal(err.Error())
			}

			test.VerifyClone(t, tt.alloc)
		})
	}
}

func TestAllocationSerialization(t *testing.T) {
	inputs := []perunIo.Serializable{
		&Allocation{
			Assets:  []Asset{&DummyAsset{0}},
			OfParts: [][]Bal{[]Bal{big.NewInt(123)}},
			Locked:  []SubAlloc{},
		},
		&Allocation{
			Assets: []Asset{&DummyAsset{1}},
			OfParts: [][]Bal{
				[]Bal{big.NewInt(1)},
			},
			Locked: []SubAlloc{
				SubAlloc{ID{0}, []Bal{big.NewInt(2)}},
			},
		},
		&Allocation{
			Assets: []Asset{&DummyAsset{1}, &DummyAsset{2}, &DummyAsset{3}},
			OfParts: [][]Bal{
				[]Bal{big.NewInt(1), big.NewInt(10), big.NewInt(100)},
				[]Bal{big.NewInt(7), big.NewInt(11), big.NewInt(13)},
			},
			Locked: []SubAlloc{
				SubAlloc{
					ID: ID{0},
					Bals: []Bal{
						big.NewInt(1), big.NewInt(3), big.NewInt(5),
					},
				},
			},
		},
	}

	ioTest.GenericSerializableTest(t, inputs...)
}

func TestAllocation_Sum(t *testing.T) {
	// invalid Allocation
	invalidAllocation := Allocation{}
	assert.Panics(t, func() { invalidAllocation.Sum() })

	// note: different invalid allocations are tested in TestAllocation_valid

	// valid Allocations
	tests := []struct {
		name  string
		alloc Allocation
		want  []Bal
	}{
		{
			"single asset/one participant",
			Allocation{
				Assets:  assets(1),
				OfParts: [][]Bal{[]Bal{big.NewInt(1)}},
			},
			[]Bal{big.NewInt(1)},
		},

		{
			"single asset/one participant/empty locked slice",
			Allocation{
				Assets:  assets(1),
				OfParts: [][]Bal{[]Bal{big.NewInt(1)}},
				Locked:  make([]SubAlloc, 0),
			},
			[]Bal{big.NewInt(1)},
		},

		{
			"single asset/three participants",
			Allocation{
				Assets: assets(1),
				OfParts: [][]Bal{
					[]Bal{big.NewInt(1)},
					[]Bal{big.NewInt(2)},
					[]Bal{big.NewInt(4)},
				},
			},
			[]Bal{big.NewInt(7)},
		},

		{
			"three assets/three participants",
			Allocation{
				Assets: assets(3),
				OfParts: [][]Bal{
					[]Bal{big.NewInt(1), big.NewInt(8), big.NewInt(64)},
					[]Bal{big.NewInt(2), big.NewInt(16), big.NewInt(128)},
					[]Bal{big.NewInt(4), big.NewInt(32), big.NewInt(256)},
				},
			},
			[]Bal{big.NewInt(7), big.NewInt(56), big.NewInt(448)},
		},

		{
			"single assets/one participants/one locked",
			Allocation{
				Assets: assets(1),
				OfParts: [][]Bal{
					[]Bal{big.NewInt(1)},
				},
				Locked: []SubAlloc{
					SubAlloc{Zero, []Bal{big.NewInt(2)}},
				},
			},
			[]Bal{big.NewInt(3)},
		},

		{
			"three assets/two participants/three locked",
			Allocation{
				Assets: assets(3),
				OfParts: [][]Bal{
					[]Bal{big.NewInt(1), big.NewInt(0x20), big.NewInt(0x400)},
					[]Bal{big.NewInt(2), big.NewInt(0x40), big.NewInt(0x800)},
				},
				Locked: []SubAlloc{
					SubAlloc{Zero, []Bal{big.NewInt(4), big.NewInt(0x80), big.NewInt(0x1000)}},
					SubAlloc{Zero, []Bal{big.NewInt(8), big.NewInt(0x100), big.NewInt(0x2000)}},
					SubAlloc{Zero, []Bal{big.NewInt(0x10), big.NewInt(0x200), big.NewInt(0x4000)}},
				},
			},
			[]Bal{big.NewInt(0x1f), big.NewInt(0x3e0), big.NewInt(0x7c00)},
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

func TestAllocation_valid(t *testing.T) {
	// note that all valid branches are already indirectly tested in TestAllocation_Sum
	tests := []struct {
		name  string
		alloc Allocation
		valid bool
	}{
		{
			"one participant/no locked valid",
			Allocation{
				Assets:  assets(1),
				OfParts: [][]Bal{[]Bal{big.NewInt(1)}},
				Locked:  nil,
			},
			true,
		},

		{
			"nil asset/nil participant",
			Allocation{
				Assets:  nil,
				OfParts: nil,
				Locked:  nil,
			},
			false,
		},

		{
			"nil participant/no locked",
			Allocation{
				Assets:  assets(1),
				OfParts: nil,
				Locked:  nil,
			},
			false,
		},

		{
			"no participant/no locked",
			Allocation{
				Assets:  assets(1),
				OfParts: make([][]Bal, 0),
			},
			false,
		},

		{
			"two participants wrong dimension",
			Allocation{
				Assets: assets(3),
				OfParts: [][]Bal{
					[]Bal{big.NewInt(1), big.NewInt(8), big.NewInt(64)},
					[]Bal{big.NewInt(2), big.NewInt(16)},
				},
			},
			false,
		},

		{
			"two participants/one locked wrong dimension",
			Allocation{
				Assets: assets(3),
				OfParts: [][]Bal{
					[]Bal{big.NewInt(1), big.NewInt(8), big.NewInt(64)},
					[]Bal{big.NewInt(2), big.NewInt(16), big.NewInt(128)},
				},
				Locked: []SubAlloc{
					SubAlloc{Zero, []Bal{big.NewInt(4)}},
				},
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.alloc.valid(); (got == nil) != tt.valid {
				t.Errorf("Allocation.valid() = %v, want valid = %v", got, tt.valid)
			}
		})
	}
}

// simple summer for testing
type balsum struct {
	b []Bal
}

func (b balsum) Sum() []Bal {
	return b.b
}

func TestEqualBalance(t *testing.T) {
	empty := balsum{make([]Bal, 0)}
	one1 := balsum{[]Bal{big.NewInt(1)}}
	one2 := balsum{[]Bal{big.NewInt(2)}}
	two12 := balsum{[]Bal{big.NewInt(1), big.NewInt(2)}}
	two48 := balsum{[]Bal{big.NewInt(4), big.NewInt(8)}}

	assert := assert.New(t)

	_, err := equalSum(empty, one1)
	assert.NotNil(err)

	eq, err := equalSum(empty, empty)
	assert.Nil(err)
	assert.True(eq)

	eq, err = equalSum(one1, one1)
	assert.Nil(err)
	assert.True(eq)

	eq, err = equalSum(one1, one2)
	assert.Nil(err)
	assert.False(eq)

	_, err = equalSum(one1, two12)
	assert.NotNil(err)

	eq, err = equalSum(two12, two12)
	assert.Nil(err)
	assert.True(eq)

	eq, err = equalSum(two12, two48)
	assert.Nil(err)
	assert.False(eq)
}

// suballocation serialization
func TestSuballocSerialization(t *testing.T) {
	ss := []perunIo.Serializable{
		&SubAlloc{ID{2}, []Bal{}},
		&SubAlloc{ID{3}, []Bal{big.NewInt(0)}},
		&SubAlloc{ID{4}, []Bal{big.NewInt(5), big.NewInt(1 << 62)}},
	}

	ioTest.GenericSerializableTest(t, ss...)
}

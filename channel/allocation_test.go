// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel

import (
	"math/big"
	"testing"
)

func TestAllocation_Sum(t *testing.T) {
	// invalid Allocation
	invalidAllocation := Allocation{
		OfParts: make([][]Bal, 0),
		Locked:  make([]Alloc, 0),
	}
	assert.Panics(t, func() { invalidAllocation.Sum() })

	// note: different invalid allocations are testes in TestAllocation_valid

	// valid Allocations
	tests := []struct {
		name  string
		alloc Allocation
		want  []Bal
	}{
		{
			"single asset/one participant",
			Allocation{
				OfParts: [][]Bal{[]Bal{big.NewInt(1)}},
				Locked:  make([]Alloc, 0),
			},
			[]Bal{big.NewInt(1)},
		},

		{
			"single asset/three participants",
			Allocation{
				OfParts: [][]Bal{
					[]Bal{big.NewInt(1)},
					[]Bal{big.NewInt(2)},
					[]Bal{big.NewInt(4)},
				},
				Locked: make([]Alloc, 0),
			},
			[]Bal{big.NewInt(7)},
		},

		{
			"three assets/three participants",
			Allocation{
				OfParts: [][]Bal{
					[]Bal{big.NewInt(1), big.NewInt(8), big.NewInt(64)},
					[]Bal{big.NewInt(2), big.NewInt(16), big.NewInt(128)},
					[]Bal{big.NewInt(4), big.NewInt(32), big.NewInt(256)},
				},
				Locked: make([]Alloc, 0),
			},
			[]Bal{big.NewInt(7), big.NewInt(56), big.NewInt(448)},
		},

		{
			"single assets/one participants/one locked",
			Allocation{
				OfParts: [][]Bal{
					[]Bal{big.NewInt(1)},
				},
				Locked: []Alloc{
					Alloc{Zero, []Bal{big.NewInt(2)}},
				},
			},
			[]Bal{big.NewInt(3)},
		},

		{
			"three assets/two participants/three locked",
			Allocation{
				OfParts: [][]Bal{
					[]Bal{big.NewInt(1), big.NewInt(0x20), big.NewInt(0x400)},
					[]Bal{big.NewInt(2), big.NewInt(0x40), big.NewInt(0x800)},
				},
				Locked: []Alloc{
					Alloc{Zero, []Bal{big.NewInt(4), big.NewInt(0x80), big.NewInt(0x1000)}},
					Alloc{Zero, []Bal{big.NewInt(8), big.NewInt(0x100), big.NewInt(0x2000)}},
					Alloc{Zero, []Bal{big.NewInt(0x10), big.NewInt(0x200), big.NewInt(0x4000)}},
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
	type fields struct {
		OfParts [][]Bal
		Locked  []Alloc
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Allocation{
				OfParts: tt.fields.OfParts,
				Locked:  tt.fields.Locked,
			}
			if got := a.valid(); got != tt.want {
				t.Errorf("Allocation.valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel

import (
	"math/big"
)

type Bal = *big.Int

// Allocation is the distribution of assets were the channel to be finalized.
// OfParts holds the balance allocations to the participants.
// Its outer dimension must match the size of the Params.parts slice.
// Its inner dimension must match the size of the Params.assets slice.
// All asset distributions could have been saved as a single []Alloc, but this
// would have saved the participants slice twice, wasting space.
type Allocation struct {
	// OfParts is the allocation of assets to the Params.parts
	OfParts [][]Bal
	// Locked is the locked allocation to sub-app-channels
	Locked []Alloc
}

// valid checks that the asset-dimensions match and slices are not nil
func (a Allocation) valid() bool {
	if a.OfParts == nil || a.Locked == nil || len(a.OfParts) == 0 {
		return false
	}

	n := len(a.OfParts[0])
	for _, pa := range a.OfParts {
		if len(pa) != n {
			return false
		}
	}

	for _, l := range a.Locked {
		if len(l.Bals) != n {
			return false
		}
	}

	return true
}

// Sum returns the sum of each asset over all participant and locked allocations
// It runs an internal check that the dimensions of all slices are valid and
// panics if not.
func (a Allocation) Sum() []Bal {
	if !a.valid() {
		log.Panic("invalid dimensions in Allocation slices")
	}

	n := len(a.OfParts[0])
	sum := make([]*big.Int, n)
	for i := 0; i < n; i++ {
		sum[i] = new(big.Int)
	}

	for _, bals := range a.OfParts {
		for i, bal := range bals {
			sum[i].Add(sum[i], bal)
		}
	}

	for _, a := range a.Locked {
		for i, bal := range a.Bals {
			sum[i].Add(sum[i], bal)
		}
	}

	return sum
}

// Alloc is the allocation of assets to a single receiver recv.
// The size of the balances slice must be of the same size as the assets slice
// of the channel Params
type Alloc struct {
	AppID ID
	Bals  []Bal
}

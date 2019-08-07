// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel

import (
	"math/big"

	"perun.network/go-perun/log"
	"perun.network/go-perun/pkg/io"

	"github.com/pkg/errors"
)

// Allocation and associated types
type (
	// Allocation is the distribution of assets, were the channel to be finalized.
	//
	// Assets identify the assets held in the channel, like an address to the
	// deposit holder contract for this asset.
	//
	// OfParts holds the balance allocations to the participants.
	// Its outer dimension must match the size of the Params.parts slice.
	// Its inner dimension must match the size of Assets.
	// All asset distributions could have been saved as a single []Alloc, but this
	// would have saved the participants slice twice, wasting space.
	//
	// Locked holds the locked allocations to sub-app-channels.
	Allocation struct {
		// Assets are the asset types held in this channel
		Assets []io.Serializable
		// OfParts is the allocation of assets to the Params.parts
		OfParts [][]Bal
		// Locked is the locked allocation to sub-app-channels. It is allowed to be
		// nil, in which case there's nothing locked.
		Locked []Alloc
	}

	// Alloc is the allocation of assets to a single receiver recv.
	// The size of the balances slice must be of the same size as the assets slice
	// of the channel Params.
	Alloc struct {
		AppID ID
		Bals  []Bal
	}

	// Bal is a single asset's balance
	Bal = *big.Int
)

// valid checks that the asset-dimensions match and slices are not nil
func (a Allocation) valid() bool {
	if a.Assets == nil || a.OfParts == nil || len(a.OfParts) == 0 {
		return false
	}

	n := len(a.Assets)
	for _, pa := range a.OfParts {
		if len(pa) != n {
			return false
		}
	}

	// Locked is allowd to be nil, in which case there's nothing locked.
	if a.Locked == nil {
		return true
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

	n := len(a.Assets)
	totals := make([]*big.Int, n)
	for i := 0; i < n; i++ {
		totals[i] = new(big.Int)
	}

	for _, bals := range a.OfParts {
		for i, bal := range bals {
			totals[i].Add(totals[i], bal)
		}
	}

	// Locked is allowd to be nil, in which case there's nothing locked.
	if a.Locked == nil {
		return totals
	}

	for _, a := range a.Locked {
		for i, bal := range a.Bals {
			totals[i].Add(totals[i], bal)
		}
	}

	return totals
}

// summer returns sums of balances
type summer interface {
	Sum() []Bal
}

func equalSum(b0, b1 summer) (bool, error) {
	s0, s1 := b0.Sum(), b1.Sum()
	n := len(s0)
	if n != len(s1) {
		return false, errors.New("dimension mismatch")
	}

	for i := 0; i < n; i++ {
		if s0[i].Cmp(s1[i]) != 0 {
			return false, nil
		}
	}
	return true, nil
}

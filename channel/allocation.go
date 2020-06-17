// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package channel

import (
	"io"
	"math/big"

	perunio "perun.network/go-perun/pkg/io"

	"github.com/pkg/errors"
)

// MaxNumAssets is an artificial limit on the number of serialized assets in an
// Allocation to avoid having users run out of memory when a malicious peer
// pretends to send a large number of assets.
const MaxNumAssets = 1024

// MaxNumParts is an artificial limit on the number participant assets in an
// Allocation to avoid having users run out of memory when a malicious peer
// pretends to send balances for a large number of participants.
// Keep in mind that an Allocation contains information about every
// participant's balance for every asset, i.e., there are num-assets times
// num-participants balances in an Allocation.
const MaxNumParts = 1024

// MaxNumSubAllocations is an artificial limit on the number of suballocations
// in an Allocation to avoid having users run out of memory when a malicious
// peer pretends to send a large number of suballocations.
// Keep in mind that an Allocation contains information about every
// asset for every suballocation, i.e., there are num-assets times
// num-suballocations items of information in an Allocation.
const MaxNumSubAllocations = 1024

// Allocation and associated types
type (
	// Allocation is the distribution of assets, were the channel to be finalized.
	//
	// Assets identify the assets held in the channel, like an address to the
	// deposit holder contract for this asset.
	//
	// Balances holds the balance allocations to the participants.
	// Its outer dimension must match the size of Assets.
	// Its inner dimension must match the size of the Params.parts slice.
	// All asset distributions could have been saved as a single []SubAlloc, but this
	// would have saved the participants slice twice, wasting space.
	//
	// Locked holds the locked allocations to sub-app-channels.
	Allocation struct {
		// Assets are the asset types held in this channel
		Assets []Asset
		// Balances is the allocation of assets to the Params.Parts
		Balances [][]Bal
		// Locked is the locked allocation to sub-app-channels. It is allowed to be
		// nil, in which case there's nothing locked.
		Locked []SubAlloc
	}

	// SubAlloc is the allocation of assets to a single receiver channel ID.
	// The size of the balances slice must be of the same size as the assets slice
	// of the channel Params.
	SubAlloc struct {
		ID   ID
		Bals []Bal
	}

	// Bal is a single asset's balance
	Bal = *big.Int

	// Asset identifies an asset. E.g., it may be the address of the multi-sig
	// where all participants' assets are deposited.
	// The same Asset should be shareable by multiple Allocation instances.
	// Decoding happens with AppBackend.DecodeAsset.
	Asset = perunio.Encoder
)

var _ perunio.Serializer = new(Allocation)

// NumParts returns the number of participants of this Allocation. It returns -1 if
// there are no Balances, i.e., if the Allocation is invalid.
func (a *Allocation) NumParts() int {
	if len(a.Balances) == 0 {
		return -1
	}
	return len(a.Balances[0])
}

// Clone returns a deep copy of the Allocation object.
// If it is nil, it returns nil.
func (a Allocation) Clone() (clone Allocation) {
	if a.Assets != nil {
		clone.Assets = make([]Asset, len(a.Assets))
		for i, asset := range a.Assets {
			clone.Assets[i] = asset
		}
	}

	if a.Balances != nil {
		clone.Balances = make([][]Bal, len(a.Balances))
		for i, pa := range a.Balances {
			clone.Balances[i] = CloneBals(pa)
		}
	}

	if a.Locked != nil {
		clone.Locked = make([]SubAlloc, len(a.Locked))
		for i, sa := range a.Locked {
			clone.Locked[i] = SubAlloc{
				ID:   sa.ID,
				Bals: CloneBals(sa.Bals),
			}
		}
	}

	return clone
}

// Encode encodes this allocation into an io.Writer.
func (a Allocation) Encode(w io.Writer) error {
	if err := a.Valid(); err != nil {
		return errors.WithMessagef(
			err, "invalid allocations cannot be encoded, got %v", a)
	}
	// encode dimensions
	if err := perunio.Encode(w, Index(len(a.Assets)), Index(len(a.Balances[0])), Index(len(a.Locked))); err != nil {
		return err
	}
	// encode assets
	for i, a := range a.Assets {
		if err := a.Encode(w); err != nil {
			return errors.WithMessagef(err, "encoding asset %d", i)
		}
	}
	// encode participant allocations
	for i := range a.Balances {
		for j := range a.Balances[i] {
			if err := perunio.Encode(w, a.Balances[i][j]); err != nil {
				return errors.WithMessagef(
					err, "encoding balance of asset %d of participant %d", i, j)
			}
		}
	}
	// encode suballocations
	for i, s := range a.Locked {
		if err := s.Encode(w); err != nil {
			return errors.WithMessagef(
				err, "encoding suballocation %d", i)
		}
	}

	return nil
}

// Decode decodes an allocation from an io.Reader.
func (a *Allocation) Decode(r io.Reader) error {
	// decode dimensions
	var numAssets, numParts, numLocked Index
	if err := perunio.Decode(r, &numAssets, &numParts, &numLocked); err != nil {
		return errors.WithMessage(err, "decoding numAssets, numParts or numLocked")
	}
	if numAssets > MaxNumAssets || numParts > MaxNumParts || numLocked > MaxNumSubAllocations {
		return errors.New("numAssets, numParts or numLocked too big")
	}
	// decode assets
	a.Assets = make([]Asset, numAssets)
	for i := range a.Assets {
		asset, err := DecodeAsset(r)
		if err != nil {
			return errors.WithMessagef(err, "decoding asset %d", i)
		}
		a.Assets[i] = asset
	}
	// decode participant allocations
	a.Balances = make([][]Bal, numAssets)
	for i := range a.Balances {
		a.Balances[i] = make([]Bal, numParts)
		for j := range a.Balances[i] {
			a.Balances[i][j] = new(big.Int)
			if err := perunio.Decode(r, &a.Balances[i][j]); err != nil {
				return errors.WithMessagef(
					err, "decoding balance of asset %d of participant %d", i, j)
			}
		}
	}
	// decode locked allocations
	a.Locked = make([]SubAlloc, numLocked)
	for i := range a.Locked {
		if err := a.Locked[i].Decode(r); err != nil {
			return errors.WithMessagef(
				err, "decoding suballocation %d", i)
		}
	}

	return a.Valid()
}

// CloneBals creates a deep copy of a balance array.
func CloneBals(orig []Bal) []Bal {
	if orig == nil {
		return nil
	}

	clone := make([]Bal, len(orig))
	for i, bal := range orig {
		clone[i] = new(big.Int).Set(bal)
	}
	return clone
}

// Valid checks that the asset-dimensions match and slices are not nil.
// Assets and Balances cannot be of zero length.
func (a Allocation) Valid() error {
	if len(a.Assets) == 0 || len(a.Balances) == 0 {
		return errors.New("assets and participant balances must not be of length zero (or nil)")
	}
	if len(a.Assets) > MaxNumAssets || len(a.Locked) > MaxNumSubAllocations {
		return errors.New("too many assets or sub-allocations")
	}

	n := len(a.Assets)

	if len(a.Balances) != n {
		return errors.Errorf("dimension mismatch: number of Assets: %d vs Balances: %d", n, len(a.Balances))
	}

	numParts := len(a.Balances[0])
	if numParts <= 0 || numParts > MaxNumParts {
		return errors.Errorf("number of participants is zero or too large")
	}

	for i, asset := range a.Balances {
		if len(asset) != numParts {
			return errors.Errorf("%d participants for asset %d, expected %d", len(asset), i, numParts)
		}
		for j, bal := range asset {
			if bal.Sign() == -1 {
				return errors.Errorf("balance[%d][%d] is negative: got %v", i, j, bal)
			}
		}
	}

	// Locked is allowed to have zero length, in which case there's nothing locked
	// and the loop is empty.
	for _, l := range a.Locked {
		if err := l.Valid(); err != nil {
			return errors.WithMessage(err, "invalid sub-allocation")
		}
		if len(l.Bals) != n {
			return errors.Errorf("dimension mismatch of app-channel balance vector (ID: %x): got %d, expected %d", l.ID, l.Bals, n)
		}
	}

	return nil
}

// Sum returns the sum of each asset over all participants and locked
// allocations.
func (a Allocation) Sum() []Bal {
	n := len(a.Assets)
	totals := make([]*big.Int, n)
	for i := 0; i < n; i++ {
		totals[i] = new(big.Int)
	}

	for i, asset := range a.Balances {
		for _, bal := range asset {
			totals[i].Add(totals[i], bal)
		}
	}

	// Locked is allowed to have zero length, in which case there's nothing locked
	// and the loop is empty.
	for _, a := range a.Locked {
		for i, bal := range a.Bals {
			totals[i].Add(totals[i], bal)
		}
	}

	return totals
}

// Equal returns whether two `Allocation` objects are equal.
func (a *Allocation) Equal(b *Allocation) error {
	if a == b {
		return nil
	}
	// Compare Assets
	if len(a.Assets) != len(b.Assets) {
		return errors.New("different number of assets")
	}
	for i, asset := range a.Assets {
		if ok, err := perunio.EqualEncoding(asset, b.Assets[i]); err != nil {
			return errors.WithMessagef(err, "comparing asset[%d] encoding", i)
		} else if !ok {
			return errors.Errorf("different asset[%d]", i)
		}
	}
	// Compare Balances
	if len(a.Balances) != len(b.Balances) {
		return errors.New("different number of balances")
	}
	for i, bals := range a.Balances {
		if len(bals) != len(b.Balances[i]) {
			return errors.Errorf("different number of parts for %d'th balances", i)
		}
		for j, bal := range bals {
			if bal.Cmp(b.Balances[i][j]) != 0 {
				return errors.Errorf("different balance[%d][%d]", i, j)
			}
		}
	}
	// Compare Locked
	if len(a.Locked) != len(b.Locked) {
		return errors.New("different number of sub allocations")
	}
	for i, locked := range a.Locked {
		if err := locked.Equal(&b.Locked[i]); err != nil {
			return errors.WithMessagef(err, "different sub allocation[%d]", i)
		}
	}
	return nil
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

var _ perunio.Serializer = new(SubAlloc)

// Valid checks if this suballocation is valid.
func (s SubAlloc) Valid() error {
	if len(s.Bals) > MaxNumAssets {
		return errors.New("too many bals")
	}
	for j, bal := range s.Bals {
		if bal.Sign() == -1 {
			return errors.Errorf("suballoc[%d] of ID %d is negative: got %v", j, s.ID, bal)
		}
	}
	return nil
}

// Encode encodes the SubAlloc s into w and returns an error if it failed.
func (s SubAlloc) Encode(w io.Writer) error {
	if err := s.Valid(); err != nil {
		return errors.WithMessagef(
			err, "invalid sub-allocations cannot be encoded, got %v", s)
	}
	// encode ID and dimension
	if err := perunio.Encode(w, s.ID, Index(len(s.Bals))); err != nil {
		return errors.WithMessagef(
			err, "encoding sub-allocation ID or dimension, id %v", s.ID)
	}
	// encode bals
	for i, bal := range s.Bals {
		if err := perunio.Encode(w, bal); err != nil {
			return errors.WithMessagef(
				err, "encoding balance of participant %d", i)
		}
	}

	return nil
}

// Decode decodes the SubAlloc s encoded in r and returns an error if it
// failed.
func (s *SubAlloc) Decode(r io.Reader) error {
	var numAssets Index
	// decode ID and dimension
	if err := perunio.Decode(r, &s.ID, &numAssets); err != nil {
		return errors.WithMessage(err, "decoding sub-allocation ID or dimension")
	}
	if numAssets > MaxNumAssets {
		return errors.Errorf("numAssets too big, got: %d max: %d", numAssets, MaxNumAssets)
	}
	// decode bals
	s.Bals = make([]Bal, numAssets)
	for i := range s.Bals {
		if err := perunio.Decode(r, &s.Bals[i]); err != nil {
			return errors.WithMessagef(
				err, "encoding participant balance %d", i)
		}
	}

	return s.Valid()
}

// Equal returns whether two `SubAlloc` objects are equal.
func (s *SubAlloc) Equal(t *SubAlloc) error {
	if s == t {
		return nil
	}
	if s.ID != t.ID {
		return errors.New("different ID")
	}
	if len(s.Bals) != len(t.Bals) {
		return errors.New("different number of bals")
	}
	for i, bal := range s.Bals {
		if bal.Cmp(t.Bals[i]) != 0 {
			return errors.Errorf("different balance[%d]", i)
		}
	}
	return nil
}

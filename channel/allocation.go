// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel

import (
	"io"
	"math"
	"math/big"

	"perun.network/go-perun/log"
	perunio "perun.network/go-perun/pkg/io"
	"perun.network/go-perun/wire"

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
	// All asset distributions could have been saved as a single []SubAlloc, but this
	// would have saved the participants slice twice, wasting space.
	//
	// Locked holds the locked allocations to sub-app-channels.
	Allocation struct {
		// Assets are the asset types held in this channel
		Assets []Asset
		// OfParts is the allocation of assets to the Params.parts
		OfParts [][]Bal
		// Locked is the locked allocation to sub-app-channels. It is allowed to be
		// nil, in which case there's nothing locked.
		Locked []SubAlloc
	}

	// SubAlloc is the allocation of assets to a single receiver channel `ID`.
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
	Asset = perunio.Serializable
)

// Clone returns a deep copy of the Allocation object.
// If it is nil, it returns nil.
func (orig Allocation) Clone() (clone Allocation) {
	if orig.Assets != nil {
		clone.Assets = make([]Asset, len(orig.Assets))
		for i, asset := range orig.Assets {
			clone.Assets[i] = asset
		}
	}

	if orig.OfParts != nil {
		clone.OfParts = make([][]Bal, len(orig.OfParts))
		for i, pa := range orig.OfParts {
			clone.OfParts[i] = CloneBals(pa)
		}
	}

	if orig.Locked != nil {
		clone.Locked = make([]SubAlloc, len(orig.Locked))
		for i, sa := range orig.Locked {
			clone.Locked[i] = SubAlloc{
				ID:   sa.ID,
				Bals: CloneBals(sa.Bals),
			}
		}
	}

	return clone
}

func (alloc Allocation) Encode(w io.Writer) error {
	if err := alloc.valid(); err != nil {
		return errors.WithMessagef(
			err, "invalid allocations cannot be encoded, got %v", alloc)
	}

	numAssets := len(alloc.Assets)
	if numAssets > math.MaxInt32 {
		return errors.Errorf(
			"expected at most %d assets, got %d", math.MaxInt32, numAssets)
	}
	if err := wire.Encode(w, int32(numAssets)); err != nil {
		return err
	}

	numParts := len(alloc.OfParts)
	if numParts > math.MaxInt32 {
		return errors.Errorf(
			"expected at most %d participants, got %d", math.MaxInt32, numParts)
	}
	if err := wire.Encode(w, int32(numParts)); err != nil {
		return err
	}

	numLocks := len(alloc.Locked)
	if numLocks > math.MaxInt32 {
		return errors.Errorf(
			"expected at most %d suballocations, got %d", math.MaxInt32, numLocks)
	}
	if err := wire.Encode(w, int32(numLocks)); err != nil {
		return err
	}

	// encode assets
	if err := perunio.Encode(w, alloc.Assets...); err != nil {
		return err
	}

	// encode participant allocations
	for i := 0; i < len(alloc.OfParts); i++ {
		for j := 0; j < len(alloc.OfParts[i]); j++ {
			if err := wire.Encode(w, alloc.OfParts[i][j]); err != nil {
				return errors.WithMessagef(
					err, "encoding error for balance %d of participant %d", j, i)
			}
		}
	}

	// encode suballocations
	for i, s := range alloc.Locked {
		if err := s.Encode(w); err != nil {
			return errors.WithMessagef(
				err, "encoding error for suballocation %d", i)
		}
	}

	return nil
}

func (alloc *Allocation) Decode(r io.Reader) error {
	// decode dimensions
	var numAssets int32
	if err := wire.Decode(r, &numAssets); err != nil {
		return err
	}
	if numAssets < 0 {
		return errors.Errorf(
			"expected non-negative number of assets, got %d", numAssets)
	}

	var numParts int32
	if err := wire.Decode(r, &numParts); err != nil {
		return err
	}
	if numParts < 0 {
		return errors.Errorf(
			"expected non-negative number of participants, got %d", numParts)
	}

	var numLocked int32
	if err := wire.Decode(r, &numLocked); err != nil {
		return err
	}
	if numLocked < 0 {
		return errors.Errorf(
			"expected non-negative number of participants, got %d", numLocked)
	}

	// decode assets
	alloc.Assets = make([]Asset, numAssets)
	for i := 0; i < len(alloc.Assets); i++ {
		if asset, err := appBackend.DecodeAsset(r); err != nil {
			return errors.WithMessagef(err, "decoding error for asset %d", i)
		} else {
			alloc.Assets[i] = asset
		}
	}

	// decode participant allocations
	alloc.OfParts = make([][]Bal, numParts)
	for i := 0; i < len(alloc.OfParts); i++ {
		alloc.OfParts[i] = make([]Bal, len(alloc.Assets))
		for j := range alloc.OfParts[i] {
			alloc.OfParts[i][j] = new(big.Int)
			if err := wire.Decode(r, &alloc.OfParts[i][j]); err != nil {
				return errors.WithMessagef(
					err, "decoding error for balance %d of participant %d", j, i)
			}
		}
	}

	// decode locked allocations
	alloc.Locked = make([]SubAlloc, numLocked)
	for i := 0; i < len(alloc.Locked); i++ {
		if err := alloc.Locked[i].Decode(r); err != nil {
			return errors.WithMessagef(
				err, "decoding error for suballocation %d", i)
		}
	}

	return alloc.valid()
}

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

// valid checks that the asset-dimensions match and slices are not nil.
// Assets and OfParts cannot be of zero length.
func (a Allocation) valid() error {
	if len(a.Assets) == 0 || len(a.OfParts) == 0 {
		return errors.New("assets and participant balances must not be of length zero")
	}

	n := len(a.Assets)
	for i, pa := range a.OfParts {
		if len(pa) != n {
			return errors.Errorf("dimension mismatch of participant %d's balance vector", i)
		}
	}

	// Locked is allowed to have zero length, in which case there's nothing locked
	// and the loop is empty.
	for _, l := range a.Locked {
		if len(l.Bals) != n {
			return errors.Errorf("dimension mismatch of app-channel balance vector (ID: %x)", l.ID)
		}
	}

	return nil
}

// Sum returns the sum of each asset over all participants and locked
// allocations.  It runs an internal check that the dimensions of all slices are
// valid and panics if not.
func (a Allocation) Sum() []Bal {
	if err := a.valid(); err != nil {
		log.Panic(err)
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

	// Locked is allowed to have zero length, in which case there's nothing locked
	// and the loop is empty.
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

// suballocation serialization
func (s *SubAlloc) Encode(w io.Writer) error {
	if err := s.ID.Encode(w); err != nil {
		return errors.WithMessagef(
			err, "error encoding suballocation id %v", s.ID)
	}

	numAssets := 0
	if s.Bals != nil {
		numAssets = len(s.Bals)
	}
	if err := wire.Encode(w, int32(numAssets)); err != nil {
		return err
	}

	for i, bal := range s.Bals {
		if err := wire.Encode(w, bal); err != nil {
			return errors.WithMessagef(
				err, "encoding error for balance of participant %d", i)
		}
	}

	return nil
}

func (s *SubAlloc) Decode(r io.Reader) error {
	if err := s.ID.Decode(r); err != nil {
		return errors.WithMessage(err, "error when decoding suballocation ID")
	}

	var numAssets int32
	if err := wire.Decode(r, &numAssets); err != nil {
		return err
	}

	s.Bals = make([]Bal, numAssets)
	for i := range s.Bals {
		s.Bals[i] = new(big.Int)
		if err := wire.Decode(r, &s.Bals[i]); err != nil {
			return errors.WithMessagef(
				err, "encoding error for participant balance %d", i)
		}
	}

	return nil
}

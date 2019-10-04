// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel

import (
	"encoding/binary"
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

	DummyAsset struct {
		Value uint64
	}
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
	if err := binary.Write(w, binary.LittleEndian, int32(numAssets)); err != nil {
		return err
	}

	numParts := len(alloc.OfParts)
	if numParts > math.MaxInt32 {
		return errors.Errorf(
			"expected at most %d participants, got %d", math.MaxInt32, numParts)
	}
	if err := binary.Write(w, binary.LittleEndian, int32(numParts)); err != nil {
		return err
	}

	numLocks := 0
	if alloc.Locked != nil {
		numLocks = len(alloc.Locked)
	}
	if numLocks > math.MaxInt32 {
		return errors.Errorf(
			"expected at most %d suballocations, got %d", math.MaxInt32, numLocks)
	}
	if err := binary.Write(w, binary.LittleEndian, int32(numLocks)); err != nil {
		return err
	}

	// encode assets
	for i := 0; i < numAssets; i++ {
		if err := alloc.Assets[i].Encode(w); err != nil {
			return errors.WithMessagef(err, "encoding error for asset %d", i)
		}
	}

	// encode participant allocations
	for i := 0; i < len(alloc.OfParts); i++ {
		if len(alloc.OfParts[i]) != numAssets {
			return errors.Errorf(
				"expected %d asset allocations for participant %d, got %d",
				numAssets, i, len(alloc.OfParts[i]))
		}

		balances := make([]interface{}, numAssets)
		for j := 0; j < numAssets; j++ {
			balances[j] = alloc.OfParts[i][j]
		}

		if err := wire.Encode(w, balances...); err != nil {
			return errors.WithMessagef(
				err, "encoding error for participant balances")
		}
	}

	// encode suballocations
	ss := make([]perunio.Serializable, numLocks)
	for i := 0; i < numLocks; i++ {
		ss[i] = &alloc.Locked[i]
	}
	if err := perunio.Encode(w, ss...); err != nil {
		return errors.WithMessagef(
			err, "suballocations encoding error")
	}

	return nil
}

func (alloc *Allocation) Decode(r io.Reader) error {
	// decode dimensions
	var numAssets int32
	if err := binary.Read(r, binary.LittleEndian, &numAssets); err != nil {
		return err
	}
	if numAssets < 0 {
		return errors.Errorf(
			"expected non-negative number of assets, got %d", numAssets)
	}

	var numParts int32
	if err := binary.Read(r, binary.LittleEndian, &numParts); err != nil {
		return err
	}
	if numParts < 0 {
		return errors.Errorf(
			"expected non-negative number of participants, got %d", numParts)
	}

	var numLocked int32
	if err := binary.Read(r, binary.LittleEndian, &numLocked); err != nil {
		return err
	}
	if numLocked < 0 {
		return errors.Errorf(
			"expected non-negative number of participants, got %d", numLocked)
	}

	// decode assets
	alloc.Assets = make([]Asset, numAssets)
	for i := 0; i < len(alloc.Assets); i++ {
		alloc.Assets[i] = &DummyAsset{}
		if err := alloc.Assets[i].Decode(r); err != nil {
			return errors.WithMessagef(err, "decoding error for asset %d", i)
		}
	}

	// decode participant allocations
	alloc.OfParts = make([][]Bal, numParts)
	for i := 0; i < len(alloc.OfParts); i++ {
		alloc.OfParts[i] = make([]Bal, len(alloc.Assets))
		balances := make([]interface{}, len(alloc.Assets))
		for j := 0; j < len(alloc.Assets); j++ {
			alloc.OfParts[i][j] = new(big.Int)
			balances[j] = &alloc.OfParts[i][j]
		}

		if err := wire.Decode(r, balances...); err != nil {
			return errors.WithMessagef(
				err, "decoding error for participant balances")
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
	if _, err := w.Write(s.ID[:]); err != nil {
		return errors.WithMessagef(
			err, "error encoding suballocation id %v", s.ID)
	}

	numAssets := 0
	if s.Bals != nil {
		numAssets = len(s.Bals)
	}
	if err := binary.Write(w, binary.LittleEndian, int32(numAssets)); err != nil {
		return err
	}

	balances := make([]interface{}, numAssets)
	for i := 0; i < numAssets; i++ {
		balances[i] = s.Bals[i]
	}
	if err := wire.Encode(w, balances...); err != nil {
		return errors.WithMessagef(
			err, "encoding error for participant balances")
	}

	return nil
}

func (s *SubAlloc) Decode(r io.Reader) error {
	if n, err := io.ReadFull(r, s.ID[:]); n != len(s.ID) || err != nil {
		if n != len(s.ID) {
			return errors.Errorf(
				"expected to read %d bytes of ID, got %d", len(s.ID), n)
		}
		if err != nil {
			return errors.WithMessage(err, "error when reading suballocation ID")
		}
	}

	var numAssets int32
	if err := binary.Read(r, binary.LittleEndian, &numAssets); err != nil {
		return err
	}

	s.Bals = make([]Bal, numAssets)
	balances := make([]interface{}, numAssets)
	for i := 0; i < int(numAssets); i++ {
		s.Bals[i] = new(big.Int)
		balances[i] = &s.Bals[i]
	}
	if err := wire.Decode(r, balances...); err != nil {
		return errors.WithMessagef(
			err, "encoding error for participant balances")
	}

	return nil
}

func (d *DummyAsset) Encode(w io.Writer) error {
	return binary.Write(w, binary.LittleEndian, d.Value)
}

func (d *DummyAsset) Decode(r io.Reader) error {
	return binary.Read(r, binary.LittleEndian, &d.Value)
}

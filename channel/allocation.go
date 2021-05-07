// Copyright 2019 - See NOTICE file for copyright holders.
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

package channel

import (
	"io"
	"log"
	"math/big"

	perunio "perun.network/go-perun/pkg/io"
	perunbig "perun.network/go-perun/pkg/math/big"
	"perun.network/go-perun/wire"

	"github.com/ethereum/go-ethereum/accounts/abi"
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

// MaxBalance is the maximum amount of funds per asset that a user can possess.
// It is set to 2 ^ 256 - 1.
var MaxBalance = abi.MaxUint256

// Allocation and associated types.
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
		// Participants of the corresponding channel.
		participants []wire.Address
		// Assets are the asset types held in this channel
		Assets []Asset
		// Balances is the allocation of assets to the Params.Parts
		Balances
		// Locked describes the funds locked in subchannels. There is one entry
		// per subchannel.
		Locked []SubAlloc
	}

	// Balances two dimensional slice of `Bal`. Is a `Summer`.
	Balances [][]Bal

	// SubAlloc is the allocation of assets to a single receiver channel ID.
	// The size of the balances slice must be of the same size as the assets slice
	// of the channel Params.
	SubAlloc struct {
		ID   ID
		Bals []Bal
	}

	// Bal is a single asset's balance.
	Bal = *big.Int

	// Asset identifies an asset. E.g., it may be the address of the multi-sig
	// where all participants' assets are deposited.
	// The same Asset should be shareable by multiple Allocation instances.
	// Decoding happens with AppBackend.DecodeAsset.
	Asset = perunio.Encoder
)

var _ perunio.Serializer = (*Allocation)(nil)
var _ perunio.Serializer = (*Balances)(nil)
var _ perunbig.Summer = (*Allocation)(nil)
var _ perunbig.Summer = (*Balances)(nil)

// NewBalances returns a new Balances object of the specified size.
func NewBalances(numAssets int, numParticipants int) Balances {
	balances := make([][]*big.Int, numAssets)
	for i := range balances {
		balances[i] = make([]*big.Int, numParticipants)
	}
	return balances
}

// NewAllocation returns a new allocation for the given participants and assets.
func NewAllocation(participants []wire.Address, assets ...Asset) Allocation {
	return Allocation{
		participants: participants,
		Assets:       assets,
		Balances:     NewBalances(len(assets), len(participants)),
	}
}

// AssetIndex returns the index of the asset in the allocation.
func (a Allocation) AssetIndex(asset Asset) (int, bool) {
	for idx, _asset := range a.Assets {
		if ok, err := perunio.EqualEncoding(asset, _asset); ok && err == nil {
			return idx, true
		}
	}
	return 0, false
}

// ParticipantIndex returns the index of the participant in the allocation.
func (a Allocation) ParticipantIndex(participant wire.Address) (int, bool) {
	for idx, _participant := range a.participants {
		if participant.Equals(_participant) {
			return idx, true
		}
	}
	return 0, false
}

// SetBalance sets the balance for the given asset and participant.
func (a Allocation) SetBalance(asset Asset, participant wire.Address, val *big.Int) error {
	assetIdx, ok := a.AssetIndex(asset)
	if !ok {
		return errors.New("failed to determine asset index")
	}

	partIdx, ok := a.ParticipantIndex(participant)
	if !ok {
		return errors.New("failed to determine participant index")
	}

	a.Balances[assetIdx][partIdx] = val
	return nil
}

// GetBalance gets the balance for the given asset and participant.
func (a Allocation) GetBalance(asset Asset, participant wire.Address) (val *big.Int, err error) {
	assetIdx, ok := a.AssetIndex(asset)
	if !ok {
		err = errors.New("failed to determine asset index")
		return
	}

	partIdx, ok := a.ParticipantIndex(participant)
	if !ok {
		err = errors.New("failed to determine participant index")
		return
	}

	val = a.Balances[assetIdx][partIdx]
	return
}

// AddToBalance adds a given amount to the balance of the specified participant
// for the given asset.
func (a Allocation) AddToBalance(asset Asset, participant wire.Address, val *big.Int) (err error) {
	bal, err := a.GetBalance(asset, participant)
	if err != nil {
		return
	}

	bal.Add(bal, val)
	return nil
}

// SubFromBalance subtracts a given amount from the balance of the specified
// participant for the given asset.
func (a Allocation) SubFromBalance(asset Asset, participant wire.Address, val *big.Int) (err error) {
	bal, err := a.GetBalance(asset, participant)
	if err != nil {
		return
	}

	bal.Sub(bal, val)
	return nil
}

// SetSubAllocBalance sets the sub-allocation balance for the given asset.
func (a Allocation) SetSubAllocBalance(id ID, asset Asset, val *big.Int) error {
	subAlloc, ok := a.SubAlloc(id)
	if !ok {
		subAlloc = SubAlloc{ID: id, Bals: make([]*big.Int, len(a.Assets))}
		a.Locked = append(a.Locked, subAlloc)
	}

	assetIdx, ok := a.AssetIndex(asset)
	if !ok {
		return errors.New("failed to determine asset index")
	}

	subAlloc.Bals[assetIdx] = val
	return nil
}

// GetSubAllocBalance gets the sub-allocation balance for the given asset.
func (a Allocation) GetSubAllocBalance(id ID, asset Asset) (val *big.Int, err error) {
	subAlloc, ok := a.SubAlloc(id)
	if !ok {
		err = errors.New("sub-allocation not found")
		return
	}

	assetIdx, ok := a.AssetIndex(asset)
	if !ok {
		err = errors.New("failed to determine asset index")
		return
	}

	val = subAlloc.Bals[assetIdx]
	return
}

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

	clone.Balances = a.Balances.Clone()

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

// Clone returns a deep copy of the Balance object.
// If it is nil, it returns nil.
func (b Balances) Clone() Balances {
	if b == nil {
		return nil
	}
	clone := make([][]Bal, len(b))
	for i, pa := range b {
		clone[i] = CloneBals(pa)
	}
	return clone
}

// Equal returns whether the two balances are equal.
func (b Balances) Equal(bals Balances) bool {
	return b.AssertEqual(bals) == nil
}

// AssertEqual returns an error if the two balances are not equal.
func (b Balances) AssertEqual(bals Balances) error {
	if len(bals) != len(b) {
		return errors.New("outer length mismatch")
	}
	for i := range bals {
		if len(bals[i]) != len(b[i]) {
			return errors.Errorf("inner length mismatch at index %d", i)
		}
		for j := range bals[i] {
			if bals[i][j].Cmp(b[i][j]) != 0 {
				return errors.Errorf("value mismatch at position [%d, %d]", i, j)
			}
		}
	}

	return nil
}

// AssertGreaterOrEqual returns whether each entry of b is greater or equal to
// the corresponding entry of bals.
func (b Balances) AssertGreaterOrEqual(bals Balances) error {
	if len(bals) != len(b) {
		return errors.New("outer length mismatch")
	}
	for i := range bals {
		if len(bals[i]) != len(b[i]) {
			return errors.Errorf("inner length mismatch at index %d", i)
		}
		for j := range bals[i] {
			if b[i][j].Cmp(bals[i][j]) < 0 {
				return errors.Errorf("value not greater or equal at position [%d, %d]", i, j)
			}
		}
	}

	return nil
}

// Add returns the sum b + a. It panics if the dimensions do not match.
func (b Balances) Add(a Balances) Balances {
	return b.operate(
		a,
		func(b1, b2 Bal) Bal { return new(big.Int).Add(b1, b2) },
	)
}

// Sub returns the difference b - a. It panics if the dimensions do not match.
func (b Balances) Sub(a Balances) Balances {
	return b.operate(
		a,
		func(b1, b2 Bal) Bal { return new(big.Int).Sub(b1, b2) },
	)
}

// operate returns op(b, a). It panics if the dimensions do not match.
func (b Balances) operate(a Balances, op func(Bal, Bal) Bal) Balances {
	if len(a) != len(b) {
		log.Panic("outer length mismatch")
	}

	c := make([][]Bal, len(a))
	for i := range a {
		if len(a[i]) != len(b[i]) {
			log.Panicf("inner length mismatch at index %d", i)
		}
		c[i] = make([]Bal, len(a[i]))
		for j := range a[i] {
			c[i][j] = op(b[i][j], a[i][j])
		}
	}

	return c
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
	if err := a.Balances.Encode(w); err != nil {
		return errors.WithMessage(err, "encoding balances")
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
	if err := perunio.Decode(r, &a.Balances); err != nil {
		return errors.WithMessage(err, "decoding balances")
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

// Decode decodes a Balances from an io.Reader.
func (b *Balances) Decode(r io.Reader) error {
	var numAssets, numParts Index
	if err := perunio.Decode(r, &numAssets, &numParts); err != nil {
		return errors.WithMessage(err, "decoding dimensions")
	}
	if numAssets > MaxNumAssets {
		return errors.Errorf("expected maximum number of assets %d, got %d", MaxNumAssets, numAssets)
	}
	if numParts > MaxNumParts {
		return errors.Errorf("expected maximum number of parts %d, got %d", MaxNumParts, numParts)
	}

	*b = make(Balances, numAssets)
	for i := range *b {
		(*b)[i] = make([]Bal, numParts)
		for j := range (*b)[i] {
			(*b)[i][j] = new(big.Int)
			if err := perunio.Decode(r, &(*b)[i][j]); err != nil {
				return errors.WithMessagef(
					err, "decoding balance of asset %d of participant %d", i, j)
			}
		}
	}
	return nil
}

// Encode encodes these balances into an io.Writer.
func (b Balances) Encode(w io.Writer) error {
	numAssets := len(b)
	numParts := 0

	if numAssets > 0 {
		numParts = len(b[0])
	}
	if numAssets > MaxNumAssets {
		return errors.Errorf("expected maximum number of assets %d, got %d", MaxNumAssets, numAssets)
	}
	if numParts > MaxNumParts {
		return errors.Errorf("expected maximum number of parts %d, got %d", MaxNumParts, numParts)
	}

	if err := perunio.Encode(w, Index(numAssets), Index(numParts)); err != nil {
		return errors.WithMessage(err, "encoding dimensions")
	}
	for i := range b {
		for j := range b[i] {
			if err := perunio.Encode(w, b[i][j]); err != nil {
				return errors.WithMessagef(
					err, "encoding balance of asset %d of participant %d", i, j)
			}
		}
	}
	return nil
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
	totals := a.Balances.Sum()
	// Locked is allowed to have zero length, in which case there's nothing locked
	// and the loop is empty.
	for _, a := range a.Locked {
		for i, bal := range a.Bals {
			totals[i].Add(totals[i], bal)
		}
	}

	return totals
}

// Sum returns the sum of each asset over all participants.
func (b Balances) Sum() []Bal {
	n := len(b)
	totals := make([]*big.Int, n)
	for i := 0; i < n; i++ {
		totals[i] = new(big.Int)
	}

	for i, asset := range b {
		for _, bal := range asset {
			totals[i].Add(totals[i], bal)
		}
	}
	return totals
}

// NewSubAlloc creates a new sub-allocation.
func NewSubAlloc(id ID, bals []Bal) *SubAlloc {
	return &SubAlloc{ID: id, Bals: bals}
}

// SubAlloc tries to return the sub-allocation for the given subchannel.
// The second return value indicates success.
func (a Allocation) SubAlloc(subchannel ID) (subAlloc SubAlloc, ok bool) {
	for _, subAlloc = range a.Locked {
		if subAlloc.ID == subchannel {
			ok = true
			return
		}
	}
	ok = false
	return
}

// AddSubAlloc adds the given sub-allocation.
func (a *Allocation) AddSubAlloc(subAlloc SubAlloc) {
	a.Locked = append(a.Locked, subAlloc)
}

// RemoveSubAlloc removes the given sub-allocation.
func (a *Allocation) RemoveSubAlloc(subAlloc SubAlloc) error {
	for i := range a.Locked {
		if subAlloc.Equal(&a.Locked[i]) == nil {
			// remove element at index i
			b := a.Locked
			copy(b[i:], b[i+1:])
			b[len(b)-1] = SubAlloc{}
			a.Locked = b[:len(b)-1]
			return nil
		}
	}
	return errors.New("not found")
}

// Equal returns nil if the `Allocation` objects are equal and an error if they
// are not equal.
func (a *Allocation) Equal(b *Allocation) error {
	if a == b {
		return nil
	}
	// Compare Assets
	if err := AssetsAssertEqual(a.Assets, b.Assets); err != nil {
		return errors.WithMessage(err, "comparing assets")
	}

	// Compare Balances
	if err := a.Balances.AssertEqual(b.Balances); err != nil {
		return errors.WithMessage(err, "comparing balances")
	}

	// Compare Locked
	return errors.WithMessage(SubAllocsAssertEqual(a.Locked, b.Locked), "comparing sub-allocations")
}

// AssetsAssertEqual returns an error if the given assets are not equal.
func AssetsAssertEqual(a []Asset, b []Asset) error {
	if len(a) != len(b) {
		return errors.New("length mismatch")
	}

	for i, asset := range a {
		if ok, err := perunio.EqualEncoding(asset, b[i]); err != nil {
			return errors.WithMessagef(err, "comparing encoding at index %d", i)
		} else if !ok {
			return errors.Errorf("value mismatch at index %d", i)
		}
	}

	return nil
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

// Equal returns nil if the `SubAlloc` objects are equal and an error if they
// are not equal.
func (s *SubAlloc) Equal(t *SubAlloc) error {
	if s == t {
		return nil
	}
	if s.ID != t.ID {
		return errors.New("different ID")
	}
	if !s.BalancesEqual(t.Bals) {
		return errors.New("balances unequal")
	}
	return nil
}

// BalancesEqual returns whether balances b equal s.Bals.
func (s *SubAlloc) BalancesEqual(b []Bal) bool {
	if len(s.Bals) != len(b) {
		return false
	}
	for i, bal := range s.Bals {
		if bal.Cmp(b[i]) != 0 {
			return false
		}
	}
	return true
}

// SubAllocsAssertEqual asserts that the two sub-allocations are equal. If they
// are unequal, an error with additional information is thrown.
func SubAllocsAssertEqual(a []SubAlloc, b []SubAlloc) error {
	if len(a) != len(b) {
		return errors.New("length mismatch")
	}
	for i := range a {
		if a[i].Equal(&b[i]) != nil {
			return errors.Errorf("value mismatch at index %d", i)
		}
	}
	return nil
}

// SubAllocsEqual returns whether the two sub-allocations are equal.
func SubAllocsEqual(a []SubAlloc, b []SubAlloc) bool {
	return SubAllocsAssertEqual(a, b) == nil
}

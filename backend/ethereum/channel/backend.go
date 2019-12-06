// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel // import "perun.network/go-perun/backend/ethereum/channel"

import (
	"bytes"
	"io"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"perun.network/go-perun/backend/ethereum/bindings/adjudicator"
	"perun.network/go-perun/backend/ethereum/wallet"
	"perun.network/go-perun/channel"
	perunwallet "perun.network/go-perun/wallet"
)

// Backend implements the interface defined in channel/Backend.go.
type Backend struct{}

var (
	// compile time check that we implement the channel backend interface.
	_ channel.Backend = new(Backend)
	// Definiton of ABI datatypes.
	abiUint256, _       = abi.NewType("uint256", nil)
	abiUint256Arr, _    = abi.NewType("uint256[]", nil)
	abiUint256ArrArr, _ = abi.NewType("uint256[][]", nil)
	abiAddress, _       = abi.NewType("address", nil)
	abiAddressArr, _    = abi.NewType("address[]", nil)
	abiBytes, _         = abi.NewType("bytes", nil)
	abiBytes32, _       = abi.NewType("bytes32", nil)
	abiUint64, _        = abi.NewType("uint64", nil)
	abiBool, _          = abi.NewType("bool", nil)
)

// ChannelID calculates the channelID as needed by the ethereum smart contracts.
func (*Backend) ChannelID(p *channel.Params) (id channel.ID) {
	params := channelParamsToEthParams(p)
	bytes, err := encodeParams(&params)
	if err != nil {
		log.Panicf("could not encode parameters: %v", err)
	}
	// Hash encoded params.
	return crypto.Keccak256Hash(bytes)
}

// Sign signs the channel state as needed by the ethereum smart contracts.
func (*Backend) Sign(acc perunwallet.Account, p *channel.Params, s *channel.State) (perunwallet.Sig, error) {
	if acc == nil || p == nil || s == nil {
		return nil, errors.New("Sign called with invalid parameters")
	}
	state := channelStateToEthState(s)
	enc, err := encodeState(&state)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to encode state")
	}
	return acc.SignData(enc)
}

// Verify verifies that a state was signed correctly.
func (*Backend) Verify(addr perunwallet.Address, p *channel.Params, s *channel.State, sig perunwallet.Sig) (bool, error) {
	if err := s.Valid(); err != nil {
		return false, errors.Wrap(err, "Can not verify invalid state")
	}
	state := channelStateToEthState(s)
	enc, err := encodeState(&state)
	if err != nil {
		return false, errors.Wrap(err, "Failed to encode state")
	}
	return perunwallet.VerifySignature(enc, sig, addr)
}

// DecodeAsset decodes an asset from a stream.
func (*Backend) DecodeAsset(r io.Reader) (channel.Asset, error) {
	var asset Asset
	return asset, asset.Decode(r)
}

// channelParamsToEthParams converts a channel.Params to a PerunTypesParams struct.
func channelParamsToEthParams(p *channel.Params) adjudicator.PerunTypesParams {
	app := p.App.Def().(*wallet.Address)
	return adjudicator.PerunTypesParams{
		ChallengeDuration: new(big.Int).SetUint64(p.ChallengeDuration),
		Nonce:             p.Nonce,
		App:               app.Address,
		Participants:      pwToCommonAddresses(p.Parts),
	}
}

// channelStateToEthState converts a channel.State to a PerunTypesState struct.
func channelStateToEthState(s *channel.State) adjudicator.PerunTypesState {
	var locked []adjudicator.PerunTypesSubAlloc
	for _, sub := range s.Locked {
		locked = append(
			locked,
			adjudicator.PerunTypesSubAlloc{ID: sub.ID, Balances: sub.Bals},
		)
	}
	outcome := adjudicator.PerunTypesAllocation{
		Assets:   assetToCommonAddresses(s.Allocation.Assets),
		Balances: transformPartBals(s.OfParts),
		Locked:   locked,
	}
	// Check allocation dimensions
	if len(outcome.Assets) != len(outcome.Balances) {
		panic("invalid allocation dimensions")
	}
	appData := new(bytes.Buffer)
	s.Data.Encode(appData)
	return adjudicator.PerunTypesState{
		ChannelID: s.ID,
		Version:   s.Version,
		Outcome:   outcome,
		AppData:   appData.Bytes(),
		IsFinal:   s.IsFinal,
	}
}

// encodeParams encodes the parameters as with abi.encode() in the smart contracts.
func encodeParams(params *adjudicator.PerunTypesParams) ([]byte, error) {
	args := abi.Arguments{
		{Type: abiUint256},
		{Type: abiUint256},
		{Type: abiAddress},
		{Type: abiAddressArr},
	}
	return args.Pack(
		params.ChallengeDuration,
		params.Nonce,
		params.App,
		params.Participants,
	)
}

// encodeState encodes the state as with abi.encode() in the smart contracts.
func encodeState(state *adjudicator.PerunTypesState) ([]byte, error) {
	args := abi.Arguments{
		{Type: abiBytes32},
		{Type: abiUint64},
		{Type: abiBytes},
		{Type: abiBytes},
		{Type: abiBool},
	}
	alloc, err := encodeAllocation(&state.Outcome)
	if err != nil {
		return nil, err
	}
	return args.Pack(
		state.ChannelID,
		state.Version,
		alloc,
		state.AppData,
		state.IsFinal,
	)
}

// encodeAllocation encodes the allocation as with abi.encode() in the smart contracts.
func encodeAllocation(alloc *adjudicator.PerunTypesAllocation) ([]byte, error) {
	args := abi.Arguments{
		{Type: abiAddressArr},
		{Type: abiUint256ArrArr},
		{Type: abiBytes},
	}
	var subAllocs []byte
	for _, sub := range alloc.Locked {
		subAlloc, err := encodeSubAlloc(&sub)
		if err != nil {
			return nil, err
		}
		subAllocs = append(subAllocs, subAlloc...)
	}
	return args.Pack(
		alloc.Assets,
		alloc.Balances,
		subAllocs,
	)
}

// encodeSubAlloc encodes the suballoc as with abi.encode() in the smart contracts.
func encodeSubAlloc(sub *adjudicator.PerunTypesSubAlloc) ([]byte, error) {
	args := abi.Arguments{
		{Type: abiBytes32},
		{Type: abiUint256Arr},
	}
	return args.Pack(
		sub.ID,
		sub.Balances,
	)
}

// assetToCommonAddresses converts an array of io.Encoder's to common.Address's.
func assetToCommonAddresses(addr []channel.Asset) []common.Address {
	cAddrs := make([]common.Address, len(addr))
	for i, part := range addr {
		asset := part.(*Asset)
		cAddrs[i] = asset.Address
	}
	return cAddrs
}

// pwToCommonAddresses converts an array of perun/wallet.Address's to common.Address's.
func pwToCommonAddresses(addr []perunwallet.Address) []common.Address {
	cAddrs := make([]common.Address, len(addr))
	for i, part := range addr {
		cAddrs[i] = part.(*wallet.Address).Address
	}
	return cAddrs
}

// transformPartBals turns valid channel.Allocation.OfParts into adjudicator.Allocation.balances.
// Currently the channel.Allocation.OfParts are encoded as following:
// OfParts[i][k] is the balance for the i-th participant on the k-th asset.
// For ethereum it is cheaper to translate this to the following.
// balances[i][k] is the balance for the i-th asset of participant k.
func transformPartBals(ofBals [][]*big.Int) [][]*big.Int {
	if len(ofBals) == 0 || len(ofBals[0]) == 0 {
		return [][]*big.Int{}
	}
	trans := make([][]*big.Int, len(ofBals[0]))
	for k := 0; k < len(ofBals[0]); k++ {
		trans[k] = make([]*big.Int, len(ofBals))
	}
	// Fill with balances.
	for i := 0; i < len(ofBals); i++ {
		for k := 0; k < len(ofBals[i]); k++ {
			trans[k][i] = ofBals[i][k]
		}
	}
	return trans
}

// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package channel

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
	ethwallet "perun.network/go-perun/backend/ethereum/wallet"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"
)

var (
	// compile time check that we implement the channel backend interface.
	_ channel.Backend = new(Backend)
	// Definition of ABI datatypes.
	abiUint256, _       = abi.NewType("uint256", "", nil)
	abiUint256Arr, _    = abi.NewType("uint256[]", "", nil)
	abiUint256ArrArr, _ = abi.NewType("uint256[][]", "", nil)
	abiAddress, _       = abi.NewType("address", "", nil)
	abiAddressArr, _    = abi.NewType("address[]", "", nil)
	abiBytes, _         = abi.NewType("bytes", "", nil)
	abiBytes32, _       = abi.NewType("bytes32", "", nil)
	abiUint64, _        = abi.NewType("uint64", "", nil)
	abiBool, _          = abi.NewType("bool", "", nil)
	abiString, _        = abi.NewType("string", "", nil)
)

// Backend implements the interface defined in channel/Backend.go.
type Backend struct{}

// CalcID calculates the channelID as needed by the ethereum smart contracts.
func (*Backend) CalcID(p *channel.Params) (id channel.ID) {
	return CalcID(p)
}

// Sign signs the channel state as needed by the ethereum smart contracts.
func (*Backend) Sign(acc wallet.Account, p *channel.Params, s *channel.State) (wallet.Sig, error) {
	return Sign(acc, p, s)
}

// Verify verifies that a state was signed correctly.
func (*Backend) Verify(addr wallet.Address, p *channel.Params, s *channel.State, sig wallet.Sig) (bool, error) {
	return Verify(addr, p, s, sig)
}

// DecodeAsset decodes an asset from a stream.
func (*Backend) DecodeAsset(r io.Reader) (channel.Asset, error) {
	return DecodeAsset(r)
}

// CalcID calculates the channelID as needed by the ethereum smart contracts.
func CalcID(p *channel.Params) (id channel.ID) {
	params := channelParamsToEthParams(p)
	bytes, err := encodeParams(&params)
	if err != nil {
		log.Panicf("could not encode parameters: %v", err)
	}
	// Hash encoded params.
	return crypto.Keccak256Hash(bytes)
}

// Sign signs the channel state as needed by the ethereum smart contracts.
func Sign(acc wallet.Account, p *channel.Params, s *channel.State) (wallet.Sig, error) {
	state := channelStateToEthState(s)
	enc, err := encodeState(&state)
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to encode state")
	}
	return acc.SignData(enc)
}

// Verify verifies that a state was signed correctly.
func Verify(addr wallet.Address, p *channel.Params, s *channel.State, sig wallet.Sig) (bool, error) {
	if err := s.Valid(); err != nil {
		return false, errors.WithMessage(err, "invalid state")
	}
	state := channelStateToEthState(s)
	enc, err := encodeState(&state)
	if err != nil {
		return false, errors.WithMessage(err, "Failed to encode state")
	}
	return ethwallet.VerifySignature(enc, sig, addr)
}

// DecodeAsset decodes an asset from a stream.
func DecodeAsset(r io.Reader) (channel.Asset, error) {
	var asset Asset
	return &asset, asset.Decode(r)
}

// channelParamsToEthParams converts a channel.Params to a ChannelParams struct.
func channelParamsToEthParams(p *channel.Params) adjudicator.ChannelParams {
	app := p.App.Def().(*ethwallet.Address)
	return adjudicator.ChannelParams{
		ChallengeDuration: new(big.Int).SetUint64(p.ChallengeDuration),
		Nonce:             p.Nonce,
		App:               common.Address(*app),
		Participants:      pwToCommonAddresses(p.Parts),
	}
}

// channelStateToEthState converts a channel.State to a ChannelState struct.
func channelStateToEthState(s *channel.State) adjudicator.ChannelState {
	locked := make([]adjudicator.ChannelSubAlloc, len(s.Locked))
	for i, sub := range s.Locked {
		locked[i] = adjudicator.ChannelSubAlloc{ID: sub.ID, Balances: sub.Bals}
	}
	outcome := adjudicator.ChannelAllocation{
		Assets:   assetToCommonAddresses(s.Allocation.Assets),
		Balances: s.Balances,
		Locked:   locked,
	}
	// Check allocation dimensions
	if len(outcome.Assets) != len(outcome.Balances) || len(s.Balances) != len(outcome.Balances) {
		log.Panic("invalid allocation dimensions")
	}
	appData := new(bytes.Buffer)
	if err := s.Data.Encode(appData); err != nil {
		log.Panicf("error encoding app data: %v", err)
	}
	return adjudicator.ChannelState{
		ChannelID: s.ID,
		Version:   s.Version,
		Outcome:   outcome,
		AppData:   appData.Bytes(),
		IsFinal:   s.IsFinal,
	}
}

// encodeParams encodes the parameters as with abi.encode() in the smart contracts.
func encodeParams(params *adjudicator.ChannelParams) ([]byte, error) {
	args := abi.Arguments{
		{Type: abiUint256},
		{Type: abiUint256},
		{Type: abiAddress},
		{Type: abiAddressArr},
	}
	enc, err := args.Pack(
		params.ChallengeDuration,
		params.Nonce,
		params.App,
		params.Participants,
	)
	return enc, errors.WithStack(err)
}

// encodeState encodes the state as with abi.encode() in the smart contracts.
func encodeState(state *adjudicator.ChannelState) ([]byte, error) {
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
	enc, err := args.Pack(
		state.ChannelID,
		state.Version,
		alloc,
		state.AppData,
		state.IsFinal,
	)
	return enc, errors.WithStack(err)
}

// encodeAllocation encodes the allocation as with abi.encode() in the smart contracts.
func encodeAllocation(alloc *adjudicator.ChannelAllocation) ([]byte, error) {
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
	enc, err := args.Pack(
		alloc.Assets,
		alloc.Balances,
		subAllocs,
	)
	return enc, errors.WithStack(err)
}

// encodeSubAlloc encodes the suballoc as with abi.encode() in the smart contracts.
func encodeSubAlloc(sub *adjudicator.ChannelSubAlloc) ([]byte, error) {
	args := abi.Arguments{
		{Type: abiBytes32},
		{Type: abiUint256Arr},
	}
	enc, err := args.Pack(
		sub.ID,
		sub.Balances,
	)
	return enc, errors.WithStack(err)
}

// assetToCommonAddresses converts an array of Assets to common.Addresses.
func assetToCommonAddresses(addr []channel.Asset) []common.Address {
	cAddrs := make([]common.Address, len(addr))
	for i, part := range addr {
		asset := part.(*Asset)
		cAddrs[i] = common.Address(*asset)
	}
	return cAddrs
}

// pwToCommonAddresses converts an array of perun/ethwallet.Addresses to common.Addresses.
func pwToCommonAddresses(addr []wallet.Address) []common.Address {
	cAddrs := make([]common.Address, len(addr))
	for i, part := range addr {
		cAddrs[i] = ethwallet.AsEthAddr(part)
	}
	return cAddrs
}

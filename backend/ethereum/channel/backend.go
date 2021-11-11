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
	"bytes"
	"io"
	"log"
	"math/big"
	"strings"

	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"perun.network/go-perun/backend/ethereum/bindings/adjudicator"
	ethwallet "perun.network/go-perun/backend/ethereum/wallet"
	"perun.network/go-perun/channel"
	perunio "perun.network/go-perun/pkg/io"
	"perun.network/go-perun/wallet"
)

const (
	phaseDispute = iota
	phaseForceExec
	phaseConcluded
)

var (
	// compile time check that we implement the channel backend interface.
	_ channel.Backend = new(Backend)
	// Definition of ABI datatypes.
	abiUint256, _ = abi.NewType("uint256", "", nil)
	abiAddress, _ = abi.NewType("address", "", nil)
	abiBytes32, _ = abi.NewType("bytes32", "", nil)
	abiParams     abi.Type
	abiState      abi.Type
	abiProgress   abi.Method
	abiRegister   abi.Method
)

func init() {
	// The ABI type parser is unable to parse the correct params and state types,
	// therefore we fetch them from the function signatures.
	adj, err := abi.JSON(strings.NewReader(adjudicator.AdjudicatorABI))
	if err != nil {
		panic("decoding ABI json")
	}
	// Get the Params type.
	chID, ok := adj.Methods["channelID"]
	if !ok || len(chID.Inputs) != 1 {
		panic("channelID not found")
	}
	abiParams = chID.Inputs[0].Type
	// Get the State type.
	hashState, ok := adj.Methods["hashState"]
	if !ok || len(hashState.Inputs) != 1 {
		panic("hashState not found")
	}
	abiState = hashState.Inputs[0].Type

	if abiProgress, ok = adj.Methods["progress"]; !ok {
		panic("Could not find method progress in adjudicator contract.")
	}

	if abiRegister, ok = adj.Methods["register"]; !ok {
		panic("Could not find method register in adjudicator contract.")
	}
}

// Backend implements the interface defined in channel/Backend.go.
type Backend struct{}

// CalcID calculates the channelID as needed by the ethereum smart contracts.
func (*Backend) CalcID(p *channel.Params) (id channel.ID) {
	return CalcID(p)
}

// Sign signs the channel state as needed by the ethereum smart contracts.
func (*Backend) Sign(acc wallet.Account, s *channel.State) (wallet.Sig, error) {
	return Sign(acc, s)
}

// Verify verifies that a state was signed correctly.
func (*Backend) Verify(addr wallet.Address, s *channel.State, sig wallet.Sig) (bool, error) {
	return Verify(addr, s, sig)
}

// DecodeAsset decodes an asset from a stream.
func (*Backend) DecodeAsset(r io.Reader) (channel.Asset, error) {
	return DecodeAsset(r)
}

// CalcID calculates the channelID as needed by the ethereum smart contracts.
func CalcID(p *channel.Params) (id channel.ID) {
	params := ToEthParams(p)
	bytes, err := EncodeParams(&params)
	if err != nil {
		log.Panicf("could not encode parameters: %v", err)
	}
	// Hash encoded params.
	return crypto.Keccak256Hash(bytes)
}

// HashState calculates the hash of a state as needed by the ethereum smart contracts.
func HashState(s *channel.State) (id channel.ID) {
	state := ToEthState(s)
	bytes, err := EncodeState(&state)
	if err != nil {
		log.Panicf("could not encode parameters: %v", err)
	}
	return crypto.Keccak256Hash(bytes)
}

// Sign signs the channel state as needed by the ethereum smart contracts.
func Sign(acc wallet.Account, s *channel.State) (wallet.Sig, error) {
	state := ToEthState(s)
	enc, err := EncodeState(&state)
	if err != nil {
		return nil, errors.WithMessage(err, "encoding state")
	}
	return acc.SignData(enc)
}

// Verify verifies that a state was signed correctly.
func Verify(addr wallet.Address, s *channel.State, sig wallet.Sig) (bool, error) {
	state := ToEthState(s)
	enc, err := EncodeState(&state)
	if err != nil {
		return false, errors.WithMessage(err, "encoding state")
	}
	return ethwallet.VerifySignature(enc, sig, addr)
}

// DecodeAsset decodes an asset from a stream.
func DecodeAsset(r io.Reader) (channel.Asset, error) {
	var asset Asset
	return &asset, perunio.Decode(r, &asset)
}

// ToEthParams converts a channel.Params to a ChannelParams struct.
func ToEthParams(p *channel.Params) adjudicator.ChannelParams {
	var app common.Address
	if p.App != nil && !channel.IsNoApp(p.App) {
		app = ethwallet.AsEthAddr(p.App.Def())
	}

	return adjudicator.ChannelParams{
		ChallengeDuration: new(big.Int).SetUint64(p.ChallengeDuration),
		Nonce:             p.Nonce,
		App:               app,
		Participants:      pwToCommonAddresses(p.Parts),
		LedgerChannel:     p.LedgerChannel,
		VirtualChannel:    p.VirtualChannel,
	}
}

// ToEthState converts a channel.State to a ChannelState struct.
func ToEthState(s *channel.State) adjudicator.ChannelState {
	locked := make([]adjudicator.ChannelSubAlloc, len(s.Locked))
	for i, sub := range s.Locked {
		// Create index map.
		indexMap := make([]uint16, s.NumParts())
		if len(sub.IndexMap) == 0 {
			for i := range indexMap {
				indexMap[i] = uint16(i)
			}
		} else {
			for i, x := range sub.IndexMap {
				indexMap[i] = uint16(x)
			}
		}

		locked[i] = adjudicator.ChannelSubAlloc{ID: sub.ID, Balances: sub.Bals, IndexMap: indexMap}
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

// EncodeParams encodes the parameters as with abi.encode() in the smart contracts.
func EncodeParams(params *adjudicator.ChannelParams) ([]byte, error) {
	args := abi.Arguments{{Type: abiParams}}
	enc, err := args.Pack(*params)
	return enc, errors.WithStack(err)
}

// EncodeState encodes the state as with abi.encode() in the smart contracts.
func EncodeState(state *adjudicator.ChannelState) ([]byte, error) {
	args := abi.Arguments{{Type: abiState}}
	enc, err := args.Pack(*state)
	return enc, errors.WithStack(err)
}

// assetToCommonAddresses converts an array of Assets to common.Addresses.
func assetToCommonAddresses(addr []channel.Asset) []common.Address {
	cAddrs := make([]common.Address, len(addr))
	for i, part := range addr {
		asset, ok := part.(*Asset)
		if !ok {
			log.Panicf("wrong address type: %T", part)
		}
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

// FromEthState converts a ChannelState to a channel.State struct.
func FromEthState(app channel.App, s *adjudicator.ChannelState) channel.State {
	locked := make([]channel.SubAlloc, len(s.Outcome.Locked))
	for i, sub := range s.Outcome.Locked {
		indexMap := makeIndexMap(sub.IndexMap)
		locked[i] = *channel.NewSubAlloc(sub.ID, sub.Balances, indexMap)
	}
	alloc := channel.Allocation{
		Assets:   fromEthAssets(s.Outcome.Assets),
		Balances: s.Outcome.Balances,
		Locked:   locked,
	}
	// Check allocation dimensions
	if len(alloc.Assets) != len(alloc.Balances) || len(s.Outcome.Balances) != len(alloc.Balances) {
		log.Panic("invalid allocation dimensions")
	}

	data, err := app.DecodeData(bytes.NewReader(s.AppData))
	if err != nil {
		log.Panicf("decoding app data: %v", err)
	}

	return channel.State{
		ID:         s.ChannelID,
		Version:    s.Version,
		Allocation: alloc,
		App:        app,
		Data:       data,
		IsFinal:    s.IsFinal,
	}
}

func makeIndexMap(m []uint16) []channel.Index {
	_m := make([]channel.Index, len(m))
	for i, x := range m {
		_m[i] = channel.Index(x)
	}
	return _m
}

func fromEthAssets(assets []common.Address) []channel.Asset {
	_assets := make([]channel.Asset, len(assets))
	for i, a := range assets {
		_assets[i] = ethwallet.AsWalletAddr(a)
	}
	return _assets
}

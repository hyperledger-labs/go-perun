// Copyright 2022 - See NOTICE file for copyright holders.
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

package protobuf

import (
	"math/big"

	"github.com/pkg/errors"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/client"
	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wire"
)

func toLedgerChannelProposal(in *LedgerChannelProposalMsg) (*client.LedgerChannelProposalMsg, error) {
	baseChannelProposal, err := toBaseChannelProposal(in.BaseChannelProposal)
	if err != nil {
		return nil, err
	}
	participant, err := toWalletAddr(in.Participant)
	if err != nil {
		return nil, err
	}
	peers, err := toWireAddrs(in.Peers)
	if err != nil {
		return nil, err
	}

	return &client.LedgerChannelProposalMsg{
		BaseChannelProposal: baseChannelProposal,
		Participant:         participant,
		Peers:               peers,
	}, nil
}

func toSubChannelProposal(in *SubChannelProposalMsg) (*client.SubChannelProposalMsg, error) {
	baseChannelProposal, err := toBaseChannelProposal(in.BaseChannelProposal)
	if err != nil {
		return nil, err
	}
	out := &client.SubChannelProposalMsg{
		BaseChannelProposal: baseChannelProposal,
	}
	copy(out.Parent[:], in.Parent)
	return out, nil
}

func toVirtualChannelProposal(in *VirtualChannelProposalMsg) (*client.VirtualChannelProposalMsg, error) {
	baseChannelProposal, err := toBaseChannelProposal(in.BaseChannelProposal)
	if err != nil {
		return nil, err
	}
	proposer, err := toWalletAddr(in.Proposer)
	if err != nil {
		return nil, err
	}
	peers, err := toWireAddrs(in.Peers)
	if err != nil {
		return nil, err
	}
	parents := make([]channel.ID, len(in.Parents))
	for i := range in.Parents {
		copy(parents[i][:], in.Parents[i])
	}
	indexMaps := make([][]channel.Index, len(in.IndexMaps))
	for i := range in.Parents {
		indexMaps[i] = toIndexMap(in.IndexMaps[i].IndexMap)
	}

	return &client.VirtualChannelProposalMsg{
		BaseChannelProposal: baseChannelProposal,
		Proposer:            proposer,
		Peers:               peers,
		Parents:             parents,
		IndexMaps:           indexMaps,
	}, nil
}

func toLedgerChannelProposalAcc(in *LedgerChannelProposalAccMsg) (*client.LedgerChannelProposalAccMsg, error) {
	participant, err := toWalletAddr(in.Participant)
	if err != nil {
		return nil, err
	}

	return &client.LedgerChannelProposalAccMsg{
		BaseChannelProposalAcc: toBaseChannelProposalAcc(in.BaseChannelProposalAcc),
		Participant:            participant,
	}, nil
}

func toSubChannelProposalAcc(in *SubChannelProposalAccMsg) *client.SubChannelProposalAccMsg {
	return &client.SubChannelProposalAccMsg{
		BaseChannelProposalAcc: toBaseChannelProposalAcc(in.BaseChannelProposalAcc),
	}
}

func toVirtualChannelProposalAcc(in *VirtualChannelProposalAccMsg) (*client.VirtualChannelProposalAccMsg, error) {
	responder, err := toWalletAddr(in.Responder)
	if err != nil {
		return nil, err
	}

	return &client.VirtualChannelProposalAccMsg{
		BaseChannelProposalAcc: toBaseChannelProposalAcc(in.BaseChannelProposalAcc),
		Responder:              responder,
	}, nil
}

func toWalletAddr(in []byte) (wallet.Address, error) {
	out := wallet.NewAddress()
	return out, out.UnmarshalBinary(in)
}

func toWireAddrs(in [][]byte) ([]wire.Address, error) {
	out := make([]wire.Address, len(in))
	for i := range in {
		out[i] = wire.NewAddress()
		err := out[i].UnmarshalBinary(in[i])
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func toBaseChannelProposal(in *BaseChannelProposal) (client.BaseChannelProposal, error) {
	initBals, err := toAllocation(in.InitBals)
	if err != nil {
		return client.BaseChannelProposal{}, errors.WithMessage(err, "encoding init bals")
	}
	fundingAgreement := toBalances(in.FundingAgreement)
	if err != nil {
		return client.BaseChannelProposal{}, errors.WithMessage(err, "encoding init bals")
	}
	app, initData, err := toAppAndData(in.App, in.InitData)
	if err != nil {
		return client.BaseChannelProposal{}, errors.WithMessage(err, "encoding and and init data")
	}
	out := client.BaseChannelProposal{
		ChallengeDuration: in.ChallengeDuration,
		App:               app,
		InitData:          initData,
		InitBals:          &initBals,
		FundingAgreement:  fundingAgreement,
	}
	copy(out.ProposalID[:], in.ProposalId)
	copy(out.NonceShare[:], in.NonceShare)
	return out, nil
}

func toBaseChannelProposalAcc(in *BaseChannelProposalAcc) (out client.BaseChannelProposalAcc) {
	copy(out.ProposalID[:], in.ProposalId)
	copy(out.NonceShare[:], in.NonceShare)
	return
}

func toAppAndData(appIn, dataIn []byte) (appOut channel.App, dataOut channel.Data, err error) {
	if len(appIn) == 0 {
		appOut = channel.NoApp()
		dataOut = channel.NoData()
		return
	}
	appDef := wallet.NewAddress()
	err = appDef.UnmarshalBinary(appIn)
	if err != nil {
		return
	}
	appOut, err = channel.Resolve(appDef)
	if err != nil {
		return
	}
	dataOut = appOut.NewData()
	err = dataOut.UnmarshalBinary(dataIn)
	return
}

func toAllocation(in *Allocation) (channel.Allocation, error) {
	var err error
	assets := make([]channel.Asset, len(in.Assets))
	for i := range in.Assets {
		assets[i] = channel.NewAsset()
		err = assets[i].UnmarshalBinary(in.Assets[i])
		if err != nil {
			return channel.Allocation{}, errors.WithMessagef(err, "marshalling %d'th asset", i)
		}
	}

	locked := make([]channel.SubAlloc, len(in.Locked))
	for i := range in.Locked {
		locked[i], err = toSubAlloc(in.Locked[i])
		if err != nil {
			return channel.Allocation{}, errors.WithMessagef(err, "marshalling %d'th sub alloc", i)
		}
	}

	return channel.Allocation{
		Assets:   assets,
		Balances: toBalances(in.Balances),
		Locked:   locked,
	}, nil
}

func toBalances(in *Balances) (out channel.Balances) {
	out = make([][]channel.Bal, len(in.Balances))
	for i := range in.Balances {
		out[i] = toBalance(in.Balances[i])
	}
	return out
}

func toBalance(in *Balance) []channel.Bal {
	out := make([]channel.Bal, len(in.Balance))
	for j := range in.Balance {
		out[j] = new(big.Int).SetBytes(in.Balance[j])
	}
	return out
}

func toSubAlloc(in *SubAlloc) (channel.SubAlloc, error) {
	subAlloc := channel.SubAlloc{
		Bals:     toBalance(in.Bals),
		IndexMap: toIndexMap(in.IndexMap.IndexMap),
	}
	if len(in.Id) != len(subAlloc.ID) {
		return channel.SubAlloc{}, errors.New("sub alloc id has incorrect length")
	}
	copy(subAlloc.ID[:], in.Id)

	return subAlloc, nil
}

func toIndexMap(in []uint32) []channel.Index {
	// In tests, include a test for type of index map is uint16
	indexMap := make([]channel.Index, len(in))
	for i := range in {
		indexMap[i] = channel.Index(uint16(in[i]))
	}
	return indexMap
}

func toChannelProposalRej(in *ChannelProposalRejMsg) (out *client.ChannelProposalRejMsg) {
	out = &client.ChannelProposalRejMsg{}
	copy(out.ProposalID[:], in.ProposalId)
	out.Reason = in.Reason
	return
}

func fromLedgerChannelProposal(in *client.LedgerChannelProposalMsg) (*LedgerChannelProposalMsg, error) {
	baseChannelProposal, err := fromBaseChannelProposal(in.BaseChannelProposal)
	if err != nil {
		return nil, err
	}
	participant, err := fromWalletAddr(in.Participant)
	if err != nil {
		return nil, err
	}
	peers, err := fromWireAddrs(in.Peers)
	if err != nil {
		return nil, err
	}

	return &LedgerChannelProposalMsg{
		BaseChannelProposal: baseChannelProposal,
		Participant:         participant,
		Peers:               peers,
	}, nil
}

func fromSubChannelProposal(in *client.SubChannelProposalMsg) (*SubChannelProposalMsg, error) {
	baseChannelProposal, err := fromBaseChannelProposal(in.BaseChannelProposal)
	if err != nil {
		return nil, err
	}

	out := &SubChannelProposalMsg{
		BaseChannelProposal: baseChannelProposal,
		Parent:              make([]byte, len(in.Parent)),
	}
	copy(out.Parent, in.Parent[:])
	return out, nil
}

func fromVirtualChannelProposal(in *client.VirtualChannelProposalMsg) (*VirtualChannelProposalMsg, error) {
	baseChannelProposal, err := fromBaseChannelProposal(in.BaseChannelProposal)
	if err != nil {
		return nil, err
	}
	proposer, err := fromWalletAddr(in.Proposer)
	if err != nil {
		return nil, err
	}
	peers, err := fromWireAddrs(in.Peers)
	if err != nil {
		return nil, err
	}
	parents := make([][]byte, len(in.Parents))
	for i := range in.Parents {
		parents[i] = make([]byte, len(in.Parents[i]))
		copy(parents[i], in.Parents[i][:])
	}
	indexMaps := make([]*IndexMap, len(in.IndexMaps))
	for i := range in.IndexMaps {
		indexMaps[i] = &IndexMap{IndexMap: fromIndexMap(in.IndexMaps[i])}
	}

	return &VirtualChannelProposalMsg{
		BaseChannelProposal: baseChannelProposal,
		Proposer:            proposer,
		Peers:               peers,
		Parents:             parents,
		IndexMaps:           indexMaps,
	}, nil
}

func fromLedgerChannelProposalAcc(in *client.LedgerChannelProposalAccMsg) (*LedgerChannelProposalAccMsg, error) {
	baseChannelProposalAcc := fromBaseChannelProposalAcc(in.BaseChannelProposalAcc)
	participant, err := fromWalletAddr(in.Participant)
	if err != nil {
		return nil, err
	}

	return &LedgerChannelProposalAccMsg{
		BaseChannelProposalAcc: baseChannelProposalAcc,
		Participant:            participant,
	}, nil
}

func fromSubChannelProposalAcc(in *client.SubChannelProposalAccMsg) *SubChannelProposalAccMsg {
	return &SubChannelProposalAccMsg{
		BaseChannelProposalAcc: fromBaseChannelProposalAcc(in.BaseChannelProposalAcc),
	}
}

func fromVirtualChannelProposalAcc(in *client.VirtualChannelProposalAccMsg) (*VirtualChannelProposalAccMsg, error) {
	baseChannelProposalAcc := fromBaseChannelProposalAcc(in.BaseChannelProposalAcc)
	responder, err := fromWalletAddr(in.Responder)
	if err != nil {
		return nil, err
	}

	return &VirtualChannelProposalAccMsg{
		BaseChannelProposalAcc: baseChannelProposalAcc,
		Responder:              responder,
	}, nil
}

func fromWalletAddr(in wallet.Address) ([]byte, error) {
	return in.MarshalBinary()
}

func fromWireAddrs(in []wire.Address) (out [][]byte, err error) {
	out = make([][]byte, len(in))
	for i := range in {
		out[i], err = in[i].MarshalBinary()
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func fromBaseChannelProposal(in client.BaseChannelProposal) (*BaseChannelProposal, error) {
	initBals, err := fromAllocation(*in.InitBals)
	if err != nil {
		return nil, errors.WithMessage(err, "encoding init bals")
	}
	fundingAgreement, err := fromBalances(in.FundingAgreement)
	if err != nil {
		return nil, errors.WithMessage(err, "encoding init bals")
	}
	app, initData, err := fromAppAndData(in.App, in.InitData)
	if err != nil {
		return nil, err
	}
	return &BaseChannelProposal{
		ProposalId:        in.ProposalID[:],
		ChallengeDuration: in.ChallengeDuration,
		NonceShare:        in.NonceShare[:],
		App:               app,
		InitData:          initData,
		InitBals:          initBals,
		FundingAgreement:  fundingAgreement,
	}, nil
}

func fromBaseChannelProposalAcc(in client.BaseChannelProposalAcc) (out *BaseChannelProposalAcc) {
	out = &BaseChannelProposalAcc{}
	out.ProposalId = make([]byte, len(in.ProposalID))
	out.NonceShare = make([]byte, len(in.NonceShare))
	copy(out.ProposalId, in.ProposalID[:])
	copy(out.NonceShare, in.NonceShare[:])
	return
}

func fromAppAndData(appIn channel.App, dataIn channel.Data) (appOut, dataOut []byte, err error) {
	if channel.IsNoApp(appIn) {
		return
	}
	appOut, err = appIn.Def().MarshalBinary()
	if err != nil {
		err = errors.WithMessage(err, "marshalling app")
		return
	}
	dataOut, err = dataIn.MarshalBinary()
	err = errors.WithMessage(err, "marshalling app")
	return
}

func fromAllocation(in channel.Allocation) (*Allocation, error) {
	assets := make([][]byte, len(in.Assets))
	var err error
	for i := range in.Assets {
		assets[i], err = in.Assets[i].MarshalBinary()
		if err != nil {
			return nil, errors.WithMessagef(err, "marshalling %d'th asset", i)
		}
	}
	bals, err := fromBalances(in.Balances)
	if err != nil {
		return nil, errors.WithMessage(err, "marshalling balances")
	}
	locked := make([]*SubAlloc, len(in.Locked))
	for i := range in.Locked {
		locked[i], err = fromSubAlloc(in.Locked[i])
		if err != nil {
			return nil, errors.WithMessagef(err, "marshalling %d'th sub alloc", i)
		}
	}

	return &Allocation{
		Assets:   assets,
		Balances: bals,
		Locked:   locked,
	}, nil
}

func fromBalances(in channel.Balances) (out *Balances, err error) {
	out = &Balances{
		Balances: make([]*Balance, len(in)),
	}
	for i := range in {
		out.Balances[i], err = fromBalance(in[i])
		if err != nil {
			return nil, err
		}
	}
	return
}

func fromBalance(in []channel.Bal) (*Balance, error) {
	out := &Balance{
		Balance: make([][]byte, len(in)),
	}
	for j := range in {
		if in[j] == nil {
			return nil, errors.New("logic error: tried to encode nil big.Int")
		}
		if in[j].Sign() == -1 {
			return nil, errors.New("encoding of negative big.Int not implemented")
		}
		out.Balance[j] = in[j].Bytes()
	}
	return out, nil
}

func fromSubAlloc(in channel.SubAlloc) (*SubAlloc, error) {
	bals, err := fromBalance(in.Bals)
	if err != nil {
		return nil, err
	}
	return &SubAlloc{
		Id:   in.ID[:],
		Bals: bals,
		IndexMap: &IndexMap{
			IndexMap: fromIndexMap(in.IndexMap),
		},
	}, nil
}

func fromIndexMap(in []channel.Index) []uint32 {
	out := make([]uint32, len(in))
	for i := range in {
		out[i] = uint32(in[i])
	}
	return out
}

func fromChannelProposalRej(in *client.ChannelProposalRejMsg) (out *ChannelProposalRejMsg) {
	out = &ChannelProposalRejMsg{}
	out.ProposalId = make([]byte, len(in.ProposalID))
	copy(out.ProposalId, in.ProposalID[:])
	out.Reason = in.Reason
	return
}

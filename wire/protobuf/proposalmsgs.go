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
	"fmt"
	"math"
	"math/big"

	"github.com/pkg/errors"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/client"
	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wire"
)

// ToLedgerChannelProposalMsg converts a protobuf Envelope_LedgerChannelProposalMsg to a client
// LedgerChannelProposalMsg.
func ToLedgerChannelProposalMsg(protoEnvMsg *Envelope_LedgerChannelProposalMsg) (*client.LedgerChannelProposalMsg, error) {
	protoMsg := protoEnvMsg.LedgerChannelProposalMsg

	var err error
	msg := &client.LedgerChannelProposalMsg{}
	msg.BaseChannelProposal, err = ToBaseChannelProposal(protoMsg.GetBaseChannelProposal())
	if err != nil {
		return nil, err
	}
	msg.Participant, err = ToWalletAddr(protoMsg.GetParticipant())
	if err != nil {
		return nil, errors.WithMessage(err, "participant address")
	}
	msg.Peers, err = ToWireAddrs(protoMsg.GetPeers())
	return msg, errors.WithMessage(err, "peers")
}

// ToSubChannelProposalMsg converts a protobuf Envelope_SubChannelProposalMsg to a client SubChannelProposalMsg.
func ToSubChannelProposalMsg(protoEnvMsg *Envelope_SubChannelProposalMsg) (*client.SubChannelProposalMsg, error) {
	protoMsg := protoEnvMsg.SubChannelProposalMsg

	var err error
	msg := &client.SubChannelProposalMsg{}
	copy(msg.Parent[:], protoMsg.GetParent())
	msg.BaseChannelProposal, err = ToBaseChannelProposal(protoMsg.GetBaseChannelProposal())
	return msg, err
}

// ToVirtualChannelProposalMsg converts a protobuf Envelope_VirtualChannelProposalMsg to a client
// VirtualChannelProposalMsg.
//
//nolint:forbidigo
func ToVirtualChannelProposalMsg(protoEnvMsg *Envelope_VirtualChannelProposalMsg) (*client.VirtualChannelProposalMsg, error) {
	protoMsg := protoEnvMsg.VirtualChannelProposalMsg

	var err error
	msg := &client.VirtualChannelProposalMsg{}
	msg.BaseChannelProposal, err = ToBaseChannelProposal(protoMsg.GetBaseChannelProposal())
	if err != nil {
		return nil, err
	}
	msg.Proposer, err = ToWalletAddr(protoMsg.GetProposer())
	if err != nil {
		return nil, errors.WithMessage(err, "proposer")
	}
	msg.Parents = make([]channel.ID, len(protoMsg.GetParents()))
	for i := range protoMsg.GetParents() {
		copy(msg.Parents[i][:], protoMsg.GetParents()[i])
	}
	msg.IndexMaps = make([][]channel.Index, len(protoMsg.GetIndexMaps()))
	for i := range protoMsg.GetIndexMaps() {
		msg.IndexMaps[i], err = ToIndexMap(protoMsg.GetIndexMaps()[i].GetIndexMap())
		if err != nil {
			return nil, err
		}
	}
	msg.Peers, err = ToWireAddrs(protoMsg.GetPeers())
	return msg, errors.WithMessage(err, "peers")
}

// ToLedgerChannelProposalAccMsg converts a protobuf Envelope_LedgerChannelProposalAccMsg to a client
// LedgerChannelProposalAccMsg.
func ToLedgerChannelProposalAccMsg(protoEnvMsg *Envelope_LedgerChannelProposalAccMsg) (*client.LedgerChannelProposalAccMsg, error) {
	protoMsg := protoEnvMsg.LedgerChannelProposalAccMsg

	var err error
	msg := &client.LedgerChannelProposalAccMsg{}
	msg.BaseChannelProposalAcc = ToBaseChannelProposalAcc(protoMsg.GetBaseChannelProposalAcc())
	msg.Participant, err = ToWalletAddr(protoMsg.GetParticipant())
	return msg, errors.WithMessage(err, "participant")
}

// ToSubChannelProposalAccMsg converts a protobuf Envelope_SubChannelProposalAccMsg to a client
// SubChannelProposalAccMsg.
func ToSubChannelProposalAccMsg(protoEnvMsg *Envelope_SubChannelProposalAccMsg) *client.SubChannelProposalAccMsg {
	protoMsg := protoEnvMsg.SubChannelProposalAccMsg

	msg := &client.SubChannelProposalAccMsg{}
	msg.BaseChannelProposalAcc = ToBaseChannelProposalAcc(protoMsg.GetBaseChannelProposalAcc())
	return msg
}

// ToVirtualChannelProposalAccMsg converts a protobuf Envelope_VirtualChannelProposalAccMsg to a client
// VirtualChannelProposalAccMsg.
func ToVirtualChannelProposalAccMsg(protoEnvMsg *Envelope_VirtualChannelProposalAccMsg) (*client.VirtualChannelProposalAccMsg, error) {
	protoMsg := protoEnvMsg.VirtualChannelProposalAccMsg

	var err error
	msg := &client.VirtualChannelProposalAccMsg{}
	msg.BaseChannelProposalAcc = ToBaseChannelProposalAcc(protoMsg.GetBaseChannelProposalAcc())
	msg.Responder, err = ToWalletAddr(protoMsg.GetResponder())
	return msg, errors.WithMessage(err, "responder")
}

// ToChannelProposalRejMsg converts a protobuf Envelope_ChannelProposalRejMsg to a client ChannelProposalRejMsg.
func ToChannelProposalRejMsg(protoEnvMsg *Envelope_ChannelProposalRejMsg) *client.ChannelProposalRejMsg {
	protoMsg := protoEnvMsg.ChannelProposalRejMsg

	msg := &client.ChannelProposalRejMsg{}
	copy(msg.ProposalID[:], protoMsg.GetProposalId())
	msg.Reason = protoMsg.GetReason()
	return msg
}

// ToWalletAddr converts a protobuf wallet address to a wallet.Address.
func ToWalletAddr(protoAddr []byte) (wallet.Address, error) {
	addr := wallet.NewAddress()
	return addr, addr.UnmarshalBinary(protoAddr)
}

// ToWalletAddrs converts protobuf wallet addresses to a slice of wallet.Address.
func ToWalletAddrs(protoAddrs [][]byte) ([]wallet.Address, error) {
	addrs := make([]wallet.Address, len(protoAddrs))
	for i := range protoAddrs {
		addrs[i] = wallet.NewAddress()
		err := addrs[i].UnmarshalBinary(protoAddrs[i])
		if err != nil {
			return nil, errors.WithMessagef(err, "%d'th address", i)
		}
	}
	return addrs, nil
}

// ToWireAddrs converts protobuf wire addresses to a slice of wire.Address.
func ToWireAddrs(protoAddrs [][]byte) ([]wire.Address, error) {
	addrs := make([]wire.Address, len(protoAddrs))
	for i := range protoAddrs {
		addrs[i] = wire.NewAddress()
		err := addrs[i].UnmarshalBinary(protoAddrs[i])
		if err != nil {
			return nil, errors.WithMessagef(err, "%d'th address", i)
		}
	}
	return addrs, nil
}

// ToBaseChannelProposal converts a protobuf BaseChannelProposal to a client BaseChannelProposal.
func ToBaseChannelProposal(protoProp *BaseChannelProposal) (client.BaseChannelProposal, error) {
	var prop client.BaseChannelProposal
	var err error
	prop.ChallengeDuration = protoProp.GetChallengeDuration()
	copy(prop.ProposalID[:], protoProp.GetProposalId())
	copy(prop.NonceShare[:], protoProp.GetNonceShare())
	prop.InitBals, err = ToAllocation(protoProp.GetInitBals())
	if err != nil {
		return prop, errors.WithMessage(err, "init bals")
	}
	prop.FundingAgreement = ToBalances(protoProp.GetFundingAgreement())
	if err != nil {
		return prop, errors.WithMessage(err, "funding agreement")
	}
	prop.App, prop.InitData, err = ToAppAndData(protoProp.GetApp(), protoProp.GetInitData())
	return prop, err
}

// ToBaseChannelProposalAcc converts a protobuf BaseChannelProposalAcc to a client BaseChannelProposalAcc.
func ToBaseChannelProposalAcc(protoPropAcc *BaseChannelProposalAcc) client.BaseChannelProposalAcc {
	var propAcc client.BaseChannelProposalAcc
	copy(propAcc.ProposalID[:], protoPropAcc.GetProposalId())
	copy(propAcc.NonceShare[:], protoPropAcc.GetNonceShare())
	return propAcc
}

// ToApp converts a protobuf app to a channel.App.
func ToApp(protoApp []byte) (channel.App, error) {
	var app channel.App
	if len(protoApp) == 0 {
		app = channel.NoApp()
		return app, nil
	}
	appDef := channel.NewAppID()
	err := appDef.UnmarshalBinary(protoApp)
	if err != nil {
		return app, err
	}
	app, err = channel.Resolve(appDef)
	return app, err
}

// ToAppAndData converts protobuf app and data to a channel.App and channel.Data.
func ToAppAndData(protoApp, protoData []byte) (channel.App, channel.Data, error) {
	var app channel.App
	var data channel.Data
	if len(protoApp) == 0 {
		app = channel.NoApp()
		data = channel.NoData()
		return app, data, nil
	}
	appDef := channel.NewAppID()
	err := appDef.UnmarshalBinary(protoApp)
	if err != nil {
		return nil, nil, err
	}
	app, err = channel.Resolve(appDef)
	if err != nil {
		return app, data, err
	}
	data = app.NewData()
	return app, data, data.UnmarshalBinary(protoData)
}

// ToAllocation converts a protobuf allocation to a channel.Allocation.
func ToAllocation(protoAlloc *Allocation) (*channel.Allocation, error) {
	var err error
	alloc := &channel.Allocation{}
	alloc.Assets = make([]channel.Asset, len(protoAlloc.GetAssets()))
	for i := range protoAlloc.GetAssets() {
		alloc.Assets[i] = channel.NewAsset()
		err = alloc.Assets[i].UnmarshalBinary(protoAlloc.GetAssets()[i])
		if err != nil {
			return nil, errors.WithMessagef(err, "%d'th asset", i)
		}
	}
	alloc.Locked = make([]channel.SubAlloc, len(protoAlloc.GetLocked()))
	for i := range protoAlloc.GetLocked() {
		alloc.Locked[i], err = ToSubAlloc(protoAlloc.GetLocked()[i])
		if err != nil {
			return nil, errors.WithMessagef(err, "%d'th sub alloc", i)
		}
	}
	alloc.Balances = ToBalances(protoAlloc.GetBalances())
	return alloc, nil
}

// ToBalances converts a protobuf Balances to a channel.Balances.
func ToBalances(protoBalances *Balances) channel.Balances {
	balances := make([][]channel.Bal, len(protoBalances.GetBalances()))
	for i := range protoBalances.GetBalances() {
		balances[i] = ToBalance(protoBalances.GetBalances()[i])
	}
	return balances
}

// ToBalance converts a protobuf Balance to a channel.Bal.
func ToBalance(protoBalance *Balance) []channel.Bal {
	balance := make([]channel.Bal, len(protoBalance.GetBalance()))
	for j := range protoBalance.GetBalance() {
		balance[j] = new(big.Int).SetBytes(protoBalance.GetBalance()[j])
	}
	return balance
}

// ToSubAlloc converts a protobuf SubAlloc to a channel.SubAlloc.
//
//nolint:forbidigo
func ToSubAlloc(protoSubAlloc *SubAlloc) (channel.SubAlloc, error) {
	var err error
	subAlloc := channel.SubAlloc{}
	subAlloc.Bals = ToBalance(protoSubAlloc.GetBals())
	if len(protoSubAlloc.GetId()) != len(subAlloc.ID) {
		return subAlloc, errors.New("sub alloc id has incorrect length")
	}
	copy(subAlloc.ID[:], protoSubAlloc.GetId())
	subAlloc.IndexMap, err = ToIndexMap(protoSubAlloc.GetIndexMap().GetIndexMap())
	return subAlloc, err
}

// ToIndexMap converts a protobuf IndexMap to a channel.IndexMap.
func ToIndexMap(protoIndexMap []uint32) ([]channel.Index, error) {
	indexMap := make([]channel.Index, len(protoIndexMap))
	for i := range protoIndexMap {
		if protoIndexMap[i] > math.MaxUint16 {
			return nil, fmt.Errorf("%d'th index is invalid", i) //nolint:goerr113  // We do not want to define this as constant error.
		}
		indexMap[i] = channel.Index(uint16(protoIndexMap[i]))
	}
	return indexMap, nil
}

// FromLedgerChannelProposalMsg converts a client LedgerChannelProposalMsg to a protobuf
// Envelope_LedgerChannelProposalMsg.
func FromLedgerChannelProposalMsg(msg *client.LedgerChannelProposalMsg) (*Envelope_LedgerChannelProposalMsg, error) {
	var err error
	protoMsg := &LedgerChannelProposalMsg{}
	protoMsg.BaseChannelProposal, err = FromBaseChannelProposal(msg.BaseChannelProposal)
	if err != nil {
		return nil, err
	}
	protoMsg.Participant, err = FromWalletAddr(msg.Participant)
	if err != nil {
		return nil, errors.WithMessage(err, "participant address")
	}
	protoMsg.Peers, err = FromWireAddrs(msg.Peers)
	return &Envelope_LedgerChannelProposalMsg{protoMsg}, errors.WithMessage(err, "peers")
}

// FromSubChannelProposalMsg converts a client SubChannelProposalMsg to a protobuf Envelope_SubChannelProposalMsg.
//
//nolint:protogetter
func FromSubChannelProposalMsg(msg *client.SubChannelProposalMsg) (*Envelope_SubChannelProposalMsg, error) {
	var err error
	protoMsg := &SubChannelProposalMsg{}
	protoMsg.Parent = make([]byte, len(msg.Parent))
	copy(protoMsg.Parent, msg.Parent[:])
	protoMsg.BaseChannelProposal, err = FromBaseChannelProposal(msg.BaseChannelProposal)
	return &Envelope_SubChannelProposalMsg{protoMsg}, err
}

// FromVirtualChannelProposalMsg converts a client VirtualChannelProposalMsg to a protobuf
// Envelope_VirtualChannelProposalMsg.
//
//nolint:protogetter
func FromVirtualChannelProposalMsg(msg *client.VirtualChannelProposalMsg) (*Envelope_VirtualChannelProposalMsg, error) {
	var err error
	protoMsg := &VirtualChannelProposalMsg{}
	protoMsg.BaseChannelProposal, err = FromBaseChannelProposal(msg.BaseChannelProposal)
	if err != nil {
		return nil, err
	}
	protoMsg.Proposer, err = FromWalletAddr(msg.Proposer)
	if err != nil {
		return nil, err
	}
	protoMsg.Parents = make([][]byte, len(msg.Parents))
	for i := range msg.Parents {
		protoMsg.Parents[i] = make([]byte, len(msg.Parents[i]))
		copy(protoMsg.Parents[i], msg.Parents[i][:])
	}
	protoMsg.IndexMaps = make([]*IndexMap, len(msg.IndexMaps))
	for i := range msg.IndexMaps {
		protoMsg.IndexMaps[i] = &IndexMap{IndexMap: FromIndexMap(msg.IndexMaps[i])}
	}
	protoMsg.Peers, err = FromWireAddrs(msg.Peers)
	return &Envelope_VirtualChannelProposalMsg{protoMsg}, errors.WithMessage(err, "peers")
}

// FromLedgerChannelProposalAccMsg converts a client LedgerChannelProposalAccMsg to a protobuf
// Envelope_LedgerChannelProposalAccMsg.
func FromLedgerChannelProposalAccMsg(msg *client.LedgerChannelProposalAccMsg) (*Envelope_LedgerChannelProposalAccMsg, error) {
	var err error
	protoMsg := &LedgerChannelProposalAccMsg{}
	protoMsg.BaseChannelProposalAcc = FromBaseChannelProposalAcc(msg.BaseChannelProposalAcc)
	protoMsg.Participant, err = FromWalletAddr(msg.Participant)
	return &Envelope_LedgerChannelProposalAccMsg{protoMsg}, errors.WithMessage(err, "participant")
}

// FromSubChannelProposalAccMsg converts a client SubChannelProposalAccMsg to a protobuf
// Envelope_SubChannelProposalAccMsg.
func FromSubChannelProposalAccMsg(msg *client.SubChannelProposalAccMsg) *Envelope_SubChannelProposalAccMsg {
	protoMsg := &SubChannelProposalAccMsg{}
	protoMsg.BaseChannelProposalAcc = FromBaseChannelProposalAcc(msg.BaseChannelProposalAcc)
	return &Envelope_SubChannelProposalAccMsg{protoMsg}
}

// FromVirtualChannelProposalAccMsg converts a client VirtualChannelProposalAccMsg to a protobuf
// Envelope_VirtualChannelProposalAccMsg.
func FromVirtualChannelProposalAccMsg(msg *client.VirtualChannelProposalAccMsg) (*Envelope_VirtualChannelProposalAccMsg, error) {
	var err error
	protoMsg := &VirtualChannelProposalAccMsg{}
	protoMsg.BaseChannelProposalAcc = FromBaseChannelProposalAcc(msg.BaseChannelProposalAcc)
	protoMsg.Responder, err = FromWalletAddr(msg.Responder)
	return &Envelope_VirtualChannelProposalAccMsg{protoMsg}, errors.WithMessage(err, "responder")
}

// FromChannelProposalRejMsg converts a client ChannelProposalRejMsg to a protobuf Envelope_ChannelProposalRejMsg.
//
//nolint:protogetter
func FromChannelProposalRejMsg(msg *client.ChannelProposalRejMsg) *Envelope_ChannelProposalRejMsg {
	protoMsg := &ChannelProposalRejMsg{}
	protoMsg.ProposalId = make([]byte, len(msg.ProposalID))
	copy(protoMsg.ProposalId, msg.ProposalID[:])
	protoMsg.Reason = msg.Reason
	return &Envelope_ChannelProposalRejMsg{protoMsg}
}

// FromWalletAddr converts a wallet.Address to a protobuf wallet address.
func FromWalletAddr(addr wallet.Address) ([]byte, error) {
	return addr.MarshalBinary()
}

// FromWalletAddrs converts a slice of wallet.Address to protobuf wallet addresses.
func FromWalletAddrs(addrs []wallet.Address) ([][]byte, error) {
	var err error
	protoAddrs := make([][]byte, len(addrs))
	for i := range addrs {
		protoAddrs[i], err = addrs[i].MarshalBinary()
		if err != nil {
			return nil, errors.WithMessagef(err, "%d'th address", i)
		}
	}
	return protoAddrs, nil
}

// FromWireAddrs converts a slice of wire.Address to protobuf wire addresses.
func FromWireAddrs(addrs []wire.Address) ([][]byte, error) {
	var err error
	protoAddrs := make([][]byte, len(addrs))
	for i := range addrs {
		protoAddrs[i], err = addrs[i].MarshalBinary()
		if err != nil {
			return nil, errors.WithMessagef(err, "%d'th address", i)
		}
	}
	return protoAddrs, nil
}

// FromBaseChannelProposal converts a client BaseChannelProposal to a protobuf BaseChannelProposal.
//
//nolint:protogetter
func FromBaseChannelProposal(prop client.BaseChannelProposal) (*BaseChannelProposal, error) {
	protoProp := &BaseChannelProposal{}
	var err error

	protoProp.ProposalId = make([]byte, len(prop.ProposalID))
	copy(protoProp.ProposalId, prop.ProposalID[:])

	protoProp.NonceShare = make([]byte, len(prop.NonceShare))
	copy(protoProp.NonceShare, prop.NonceShare[:])

	protoProp.ChallengeDuration = prop.ChallengeDuration

	protoProp.InitBals, err = FromAllocation(*prop.InitBals)
	if err != nil {
		return nil, errors.WithMessage(err, "init bals")
	}
	protoProp.FundingAgreement, err = FromBalances(prop.FundingAgreement)
	if err != nil {
		return nil, errors.WithMessage(err, "funding agreement")
	}
	protoProp.App, protoProp.InitData, err = FromAppAndData(prop.App, prop.InitData)
	return protoProp, err
}

// FromBaseChannelProposalAcc converts a client BaseChannelProposalAcc to a protobuf BaseChannelProposalAcc.
//
//nolint:protogetter
func FromBaseChannelProposalAcc(propAcc client.BaseChannelProposalAcc) *BaseChannelProposalAcc {
	protoPropAcc := &BaseChannelProposalAcc{}
	protoPropAcc.ProposalId = make([]byte, len(propAcc.ProposalID))
	protoPropAcc.NonceShare = make([]byte, len(propAcc.NonceShare))
	copy(protoPropAcc.ProposalId, propAcc.ProposalID[:])
	copy(protoPropAcc.NonceShare, propAcc.NonceShare[:])
	return protoPropAcc
}

// FromApp converts a channel.App to a protobuf app.
func FromApp(app channel.App) ([]byte, error) {
	if channel.IsNoApp(app) {
		return []byte{}, nil
	}
	protoApp, err := app.Def().MarshalBinary()
	return protoApp, err
}

// FromAppAndData converts channel.App and channel.Data to protobuf app and data.
func FromAppAndData(app channel.App, data channel.Data) ([]byte, []byte, error) {
	if channel.IsNoApp(app) {
		return []byte{}, []byte{}, nil
	}
	protoApp, err := app.Def().MarshalBinary()
	if err != nil {
		return []byte{}, []byte{}, err
	}
	protoData, err := data.MarshalBinary()
	return protoApp, protoData, err
}

// FromAllocation converts a channel.Allocation to a protobuf Allocation.
func FromAllocation(alloc channel.Allocation) (*Allocation, error) {
	var err error
	protoAlloc := &Allocation{}
	protoAlloc.Assets = make([][]byte, len(alloc.Assets))
	for i := range alloc.Assets {
		protoAlloc.Assets[i], err = alloc.Assets[i].MarshalBinary()
		if err != nil {
			return nil, errors.WithMessagef(err, "%d'th asset", i)
		}
	}
	locked := make([]*SubAlloc, len(alloc.Locked))
	for i := range alloc.Locked {
		locked[i], err = FromSubAlloc(alloc.Locked[i])
		if err != nil {
			return nil, errors.WithMessagef(err, "%d'th sub alloc", i)
		}
	}
	protoAlloc.Balances, err = FromBalances(alloc.Balances)
	return protoAlloc, err
}

// FromBalances converts a channel.Balances to a protobuf Balances.
func FromBalances(balances channel.Balances) (*Balances, error) {
	var err error
	protoBalances := &Balances{}
	protoBalances.Balances = make([]*Balance, len(balances))

	for i := range balances {
		protoBalances.Balances[i], err = FromBalance(balances[i])
		if err != nil {
			return nil, errors.WithMessagef(err, "%d'th balance", i)
		}
	}
	return protoBalances, nil
}

// FromBalance converts a slice of channel.Bal to a protobuf Balance.
func FromBalance(balance []channel.Bal) (*Balance, error) {
	protoBalance := &Balance{
		Balance: make([][]byte, len(balance)),
	}
	for i := range balance {
		if balance[i] == nil {
			return nil, fmt.Errorf("%d'th amount is nil", i) //nolint:goerr113  // We do not want to define this as constant error.
		}
		if balance[i].Sign() == -1 {
			return nil, fmt.Errorf("%d'th amount is negative", i) //nolint:goerr113  // We do not want to define this as constant error.
		}
		protoBalance.Balance[i] = balance[i].Bytes()
	}
	return protoBalance, nil
}

// FromSubAlloc converts a channel.SubAlloc to a protobuf SubAlloc.
//
//nolint:protogetter
func FromSubAlloc(subAlloc channel.SubAlloc) (*SubAlloc, error) {
	protoSubAlloc := &SubAlloc{}
	var err error
	protoSubAlloc.Id = make([]byte, len(subAlloc.ID))
	copy(protoSubAlloc.Id, subAlloc.ID[:])
	protoSubAlloc.IndexMap = &IndexMap{IndexMap: FromIndexMap(subAlloc.IndexMap)}
	protoSubAlloc.Bals, err = FromBalance(subAlloc.Bals)
	return protoSubAlloc, err
}

// FromIndexMap converts a channel.IndexMap to a protobuf index map.
func FromIndexMap(indexMap []channel.Index) []uint32 {
	protoIndexMap := make([]uint32, len(indexMap))
	for i := range indexMap {
		protoIndexMap[i] = uint32(indexMap[i])
	}
	return protoIndexMap
}

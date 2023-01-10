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

func toLedgerChannelProposalMsg(protoEnvMsg *Envelope_LedgerChannelProposalMsg) (msg *client.LedgerChannelProposalMsg, err error) {
	protoMsg := protoEnvMsg.LedgerChannelProposalMsg

	msg = &client.LedgerChannelProposalMsg{}
	msg.BaseChannelProposal, err = toBaseChannelProposal(protoMsg.BaseChannelProposal)
	if err != nil {
		return nil, err
	}
	msg.Participant, err = toWalletAddr(protoMsg.Participant)
	if err != nil {
		return nil, errors.WithMessage(err, "participant address")
	}
	msg.Peers, err = toWireAddrs(protoMsg.Peers)
	return msg, errors.WithMessage(err, "peers")
}

func toSubChannelProposalMsg(protoEnvMsg *Envelope_SubChannelProposalMsg) (msg *client.SubChannelProposalMsg, err error) {
	protoMsg := protoEnvMsg.SubChannelProposalMsg

	msg = &client.SubChannelProposalMsg{}
	copy(msg.Parent[:], protoMsg.Parent)
	msg.BaseChannelProposal, err = toBaseChannelProposal(protoMsg.BaseChannelProposal)
	return msg, err
}

func toVirtualChannelProposalMsg(protoEnvMsg *Envelope_VirtualChannelProposalMsg) (msg *client.VirtualChannelProposalMsg, err error) {
	protoMsg := protoEnvMsg.VirtualChannelProposalMsg

	msg = &client.VirtualChannelProposalMsg{}
	msg.BaseChannelProposal, err = toBaseChannelProposal(protoMsg.BaseChannelProposal)
	if err != nil {
		return nil, err
	}
	msg.Proposer, err = toWalletAddr(protoMsg.Proposer)
	if err != nil {
		return nil, errors.WithMessage(err, "proposer")
	}
	msg.Parents = make([]channel.ID, len(protoMsg.Parents))
	for i := range protoMsg.Parents {
		copy(msg.Parents[i][:], protoMsg.Parents[i])
	}
	msg.IndexMaps = make([][]channel.Index, len(protoMsg.IndexMaps))
	for i := range protoMsg.IndexMaps {
		msg.IndexMaps[i], err = toIndexMap(protoMsg.IndexMaps[i].IndexMap)
		if err != nil {
			return nil, err
		}
	}
	msg.Peers, err = toWireAddrs(protoMsg.Peers)
	return msg, errors.WithMessage(err, "peers")
}

func toLedgerChannelProposalAccMsg(protoEnvMsg *Envelope_LedgerChannelProposalAccMsg) (msg *client.LedgerChannelProposalAccMsg, err error) {
	protoMsg := protoEnvMsg.LedgerChannelProposalAccMsg

	msg = &client.LedgerChannelProposalAccMsg{}
	msg.BaseChannelProposalAcc = toBaseChannelProposalAcc(protoMsg.BaseChannelProposalAcc)
	msg.Participant, err = toWalletAddr(protoMsg.Participant)
	return msg, errors.WithMessage(err, "participant")
}

func toSubChannelProposalAccMsg(protoEnvMsg *Envelope_SubChannelProposalAccMsg) (msg *client.SubChannelProposalAccMsg) {
	protoMsg := protoEnvMsg.SubChannelProposalAccMsg

	msg = &client.SubChannelProposalAccMsg{}
	msg.BaseChannelProposalAcc = toBaseChannelProposalAcc(protoMsg.BaseChannelProposalAcc)
	return msg
}

func toVirtualChannelProposalAccMsg(protoEnvMsg *Envelope_VirtualChannelProposalAccMsg) (msg *client.VirtualChannelProposalAccMsg, err error) {
	protoMsg := protoEnvMsg.VirtualChannelProposalAccMsg

	msg = &client.VirtualChannelProposalAccMsg{}
	msg.BaseChannelProposalAcc = toBaseChannelProposalAcc(protoMsg.BaseChannelProposalAcc)
	msg.Responder, err = toWalletAddr(protoMsg.Responder)
	return msg, errors.WithMessage(err, "responder")
}

func toChannelProposalRejMsg(protoEnvMsg *Envelope_ChannelProposalRejMsg) (msg *client.ChannelProposalRejMsg) {
	protoMsg := protoEnvMsg.ChannelProposalRejMsg

	msg = &client.ChannelProposalRejMsg{}
	copy(msg.ProposalID[:], protoMsg.ProposalId)
	msg.Reason = protoMsg.Reason
	return msg
}

func toWalletAddr(protoAddr []byte) (wallet.Address, error) {
	addr := wallet.NewAddress()
	return addr, addr.UnmarshalBinary(protoAddr)
}

func toWalletAddrs(protoAddrs [][]byte) ([]wallet.Address, error) {
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

func toWireAddrs(protoAddrs [][]byte) ([]wire.Address, error) {
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

func toBaseChannelProposal(protoProp *BaseChannelProposal) (prop client.BaseChannelProposal, err error) {
	prop.ChallengeDuration = protoProp.ChallengeDuration
	copy(prop.ProposalID[:], protoProp.ProposalId)
	copy(prop.NonceShare[:], protoProp.NonceShare)
	prop.InitBals, err = toAllocation(protoProp.InitBals)
	if err != nil {
		return prop, errors.WithMessage(err, "init bals")
	}
	prop.FundingAgreement = ToBalances(protoProp.FundingAgreement)
	if err != nil {
		return prop, errors.WithMessage(err, "funding agreement")
	}
	prop.App, prop.InitData, err = toAppAndData(protoProp.App, protoProp.InitData)
	return prop, err
}

func toBaseChannelProposalAcc(protoPropAcc *BaseChannelProposalAcc) (propAcc client.BaseChannelProposalAcc) {
	copy(propAcc.ProposalID[:], protoPropAcc.ProposalId)
	copy(propAcc.NonceShare[:], protoPropAcc.NonceShare)
	return
}

func toApp(protoApp []byte) (app channel.App, err error) {
	if len(protoApp) == 0 {
		app = channel.NoApp()
		return app, nil
	}
	appDef := channel.NewAppID()
	err = appDef.UnmarshalBinary(protoApp)
	if err != nil {
		return app, err
	}
	app, err = channel.Resolve(appDef)
	return app, err
}

func toAppAndData(protoApp, protoData []byte) (app channel.App, data channel.Data, err error) {
	if len(protoApp) == 0 {
		app = channel.NoApp()
		data = channel.NoData()
		return app, data, nil
	}
	appDef := channel.NewAppID()
	err = appDef.UnmarshalBinary(protoApp)
	if err != nil {
		return nil, nil, err
	}
	app, err = channel.Resolve(appDef)
	if err != nil {
		return
	}
	data = app.NewData()
	return app, data, data.UnmarshalBinary(protoData)
}

func toAllocation(protoAlloc *Allocation) (alloc *channel.Allocation, err error) {
	alloc = &channel.Allocation{}
	alloc.Assets = make([]channel.Asset, len(protoAlloc.Assets))
	for i := range protoAlloc.Assets {
		alloc.Assets[i] = channel.NewAsset()
		err = alloc.Assets[i].UnmarshalBinary(protoAlloc.Assets[i])
		if err != nil {
			return nil, errors.WithMessagef(err, "%d'th asset", i)
		}
	}
	alloc.Locked = make([]channel.SubAlloc, len(protoAlloc.Locked))
	for i := range protoAlloc.Locked {
		alloc.Locked[i], err = toSubAlloc(protoAlloc.Locked[i])
		if err != nil {
			return nil, errors.WithMessagef(err, "%d'th sub alloc", i)
		}
	}
	alloc.Balances = ToBalances(protoAlloc.Balances)
	return alloc, nil
}

// ToBalances parses protobuf balances.
func ToBalances(protoBalances *Balances) (balances channel.Balances) {
	balances = make([][]channel.Bal, len(protoBalances.Balances))
	for i := range protoBalances.Balances {
		balances[i] = toBalance(protoBalances.Balances[i])
	}
	return balances
}

func toBalance(protoBalance *Balance) (balance []channel.Bal) {
	balance = make([]channel.Bal, len(protoBalance.Balance))
	for j := range protoBalance.Balance {
		balance[j] = new(big.Int).SetBytes(protoBalance.Balance[j])
	}
	return balance
}

func toSubAlloc(protoSubAlloc *SubAlloc) (subAlloc channel.SubAlloc, err error) {
	subAlloc = channel.SubAlloc{}
	subAlloc.Bals = toBalance(protoSubAlloc.Bals)
	if len(protoSubAlloc.Id) != len(subAlloc.ID) {
		return subAlloc, errors.New("sub alloc id has incorrect length")
	}
	copy(subAlloc.ID[:], protoSubAlloc.Id)
	subAlloc.IndexMap, err = toIndexMap(protoSubAlloc.IndexMap.IndexMap)
	return subAlloc, err
}

func toIndexMap(protoIndexMap []uint32) (indexMap []channel.Index, err error) {
	indexMap = make([]channel.Index, len(protoIndexMap))
	for i := range protoIndexMap {
		if protoIndexMap[i] > math.MaxUint16 {
			return nil, fmt.Errorf("%d'th index is invalid", i) //nolint:goerr113  // We do not want to define this as constant error.
		}
		indexMap[i] = channel.Index(uint16(protoIndexMap[i]))
	}
	return indexMap, nil
}

func fromLedgerChannelProposalMsg(msg *client.LedgerChannelProposalMsg) (_ *Envelope_LedgerChannelProposalMsg, err error) {
	protoMsg := &LedgerChannelProposalMsg{}
	protoMsg.BaseChannelProposal, err = fromBaseChannelProposal(msg.BaseChannelProposal)
	if err != nil {
		return nil, err
	}
	protoMsg.Participant, err = fromWalletAddr(msg.Participant)
	if err != nil {
		return nil, errors.WithMessage(err, "participant address")
	}
	protoMsg.Peers, err = fromWireAddrs(msg.Peers)
	return &Envelope_LedgerChannelProposalMsg{protoMsg}, errors.WithMessage(err, "peers")
}

func fromSubChannelProposalMsg(msg *client.SubChannelProposalMsg) (_ *Envelope_SubChannelProposalMsg, err error) {
	protoMsg := &SubChannelProposalMsg{}
	protoMsg.Parent = make([]byte, len(msg.Parent))
	copy(protoMsg.Parent, msg.Parent[:])
	protoMsg.BaseChannelProposal, err = fromBaseChannelProposal(msg.BaseChannelProposal)
	return &Envelope_SubChannelProposalMsg{protoMsg}, err
}

func fromVirtualChannelProposalMsg(msg *client.VirtualChannelProposalMsg) (_ *Envelope_VirtualChannelProposalMsg, err error) {
	protoMsg := &VirtualChannelProposalMsg{}
	protoMsg.BaseChannelProposal, err = fromBaseChannelProposal(msg.BaseChannelProposal)
	if err != nil {
		return nil, err
	}
	protoMsg.Proposer, err = fromWalletAddr(msg.Proposer)
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
		protoMsg.IndexMaps[i] = &IndexMap{IndexMap: fromIndexMap(msg.IndexMaps[i])}
	}
	protoMsg.Peers, err = fromWireAddrs(msg.Peers)
	return &Envelope_VirtualChannelProposalMsg{protoMsg}, errors.WithMessage(err, "peers")
}

func fromLedgerChannelProposalAccMsg(msg *client.LedgerChannelProposalAccMsg) (_ *Envelope_LedgerChannelProposalAccMsg, err error) {
	protoMsg := &LedgerChannelProposalAccMsg{}
	protoMsg.BaseChannelProposalAcc = fromBaseChannelProposalAcc(msg.BaseChannelProposalAcc)
	protoMsg.Participant, err = fromWalletAddr(msg.Participant)
	return &Envelope_LedgerChannelProposalAccMsg{protoMsg}, errors.WithMessage(err, "participant")
}

func fromSubChannelProposalAccMsg(msg *client.SubChannelProposalAccMsg) (_ *Envelope_SubChannelProposalAccMsg) {
	protoMsg := &SubChannelProposalAccMsg{}
	protoMsg.BaseChannelProposalAcc = fromBaseChannelProposalAcc(msg.BaseChannelProposalAcc)
	return &Envelope_SubChannelProposalAccMsg{protoMsg}
}

func fromVirtualChannelProposalAccMsg(msg *client.VirtualChannelProposalAccMsg) (_ *Envelope_VirtualChannelProposalAccMsg, err error) {
	protoMsg := &VirtualChannelProposalAccMsg{}
	protoMsg.BaseChannelProposalAcc = fromBaseChannelProposalAcc(msg.BaseChannelProposalAcc)
	protoMsg.Responder, err = fromWalletAddr(msg.Responder)
	return &Envelope_VirtualChannelProposalAccMsg{protoMsg}, errors.WithMessage(err, "responder")
}

func fromChannelProposalRejMsg(msg *client.ChannelProposalRejMsg) (_ *Envelope_ChannelProposalRejMsg) {
	protoMsg := &ChannelProposalRejMsg{}
	protoMsg.ProposalId = make([]byte, len(msg.ProposalID))
	copy(protoMsg.ProposalId, msg.ProposalID[:])
	protoMsg.Reason = msg.Reason
	return &Envelope_ChannelProposalRejMsg{protoMsg}
}

func fromWalletAddr(addr wallet.Address) ([]byte, error) {
	return addr.MarshalBinary()
}

func fromWalletAddrs(addrs []wallet.Address) (protoAddrs [][]byte, err error) {
	protoAddrs = make([][]byte, len(addrs))
	for i := range addrs {
		protoAddrs[i], err = addrs[i].MarshalBinary()
		if err != nil {
			return nil, errors.WithMessagef(err, "%d'th address", i)
		}
	}
	return protoAddrs, nil
}

func fromWireAddrs(addrs []wire.Address) (protoAddrs [][]byte, err error) {
	protoAddrs = make([][]byte, len(addrs))
	for i := range addrs {
		protoAddrs[i], err = addrs[i].MarshalBinary()
		if err != nil {
			return nil, errors.WithMessagef(err, "%d'th address", i)
		}
	}
	return protoAddrs, nil
}

func fromBaseChannelProposal(prop client.BaseChannelProposal) (protoProp *BaseChannelProposal, err error) {
	protoProp = &BaseChannelProposal{}

	protoProp.ProposalId = make([]byte, len(prop.ProposalID))
	copy(protoProp.ProposalId, prop.ProposalID[:])

	protoProp.NonceShare = make([]byte, len(prop.NonceShare))
	copy(protoProp.NonceShare, prop.NonceShare[:])

	protoProp.ChallengeDuration = prop.ChallengeDuration

	protoProp.InitBals, err = fromAllocation(*prop.InitBals)
	if err != nil {
		return nil, errors.WithMessage(err, "init bals")
	}
	protoProp.FundingAgreement, err = fromBalances(prop.FundingAgreement)
	if err != nil {
		return nil, errors.WithMessage(err, "funding agreement")
	}
	protoProp.App, protoProp.InitData, err = fromAppAndData(prop.App, prop.InitData)
	return protoProp, err
}

func fromBaseChannelProposalAcc(propAcc client.BaseChannelProposalAcc) (protoPropAcc *BaseChannelProposalAcc) {
	protoPropAcc = &BaseChannelProposalAcc{}
	protoPropAcc.ProposalId = make([]byte, len(propAcc.ProposalID))
	protoPropAcc.NonceShare = make([]byte, len(propAcc.NonceShare))
	copy(protoPropAcc.ProposalId, propAcc.ProposalID[:])
	copy(protoPropAcc.NonceShare, propAcc.NonceShare[:])
	return protoPropAcc
}

func fromApp(app channel.App) (protoApp []byte, err error) {
	if channel.IsNoApp(app) {
		return []byte{}, nil
	}
	protoApp, err = app.Def().MarshalBinary()
	return protoApp, err
}

func fromAppAndData(app channel.App, data channel.Data) (protoApp, protoData []byte, err error) {
	if channel.IsNoApp(app) {
		return []byte{}, []byte{}, nil
	}
	protoApp, err = app.Def().MarshalBinary()
	if err != nil {
		return []byte{}, []byte{}, err
	}
	protoData, err = data.MarshalBinary()
	return protoApp, protoData, err
}

func fromAllocation(alloc channel.Allocation) (protoAlloc *Allocation, err error) {
	protoAlloc = &Allocation{}
	protoAlloc.Assets = make([][]byte, len(alloc.Assets))
	for i := range alloc.Assets {
		protoAlloc.Assets[i], err = alloc.Assets[i].MarshalBinary()
		if err != nil {
			return nil, errors.WithMessagef(err, "%d'th asset", i)
		}
	}
	locked := make([]*SubAlloc, len(alloc.Locked))
	for i := range alloc.Locked {
		locked[i], err = fromSubAlloc(alloc.Locked[i])
		if err != nil {
			return nil, errors.WithMessagef(err, "%d'th sub alloc", i)
		}
	}
	protoAlloc.Balances, err = fromBalances(alloc.Balances)
	return protoAlloc, err
}

func fromBalances(balances channel.Balances) (protoBalances *Balances, err error) {
	protoBalances = &Balances{
		Balances: make([]*Balance, len(balances)),
	}
	for i := range balances {
		protoBalances.Balances[i], err = fromBalance(balances[i])
		if err != nil {
			return nil, errors.WithMessagef(err, "%d'th balance", i)
		}
	}
	return protoBalances, nil
}

func fromBalance(balance []channel.Bal) (protoBalance *Balance, err error) {
	protoBalance = &Balance{
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

func fromSubAlloc(subAlloc channel.SubAlloc) (protoSubAlloc *SubAlloc, err error) {
	protoSubAlloc = &SubAlloc{}
	protoSubAlloc.Id = make([]byte, len(subAlloc.ID))
	copy(protoSubAlloc.Id, subAlloc.ID[:])
	protoSubAlloc.IndexMap = &IndexMap{IndexMap: fromIndexMap(subAlloc.IndexMap)}
	protoSubAlloc.Bals, err = fromBalance(subAlloc.Bals)
	return protoSubAlloc, err
}

func fromIndexMap(indexMap []channel.Index) (protoIndexMap []uint32) {
	protoIndexMap = make([]uint32, len(indexMap))
	for i := range indexMap {
		protoIndexMap[i] = uint32(indexMap[i])
	}
	return protoIndexMap
}

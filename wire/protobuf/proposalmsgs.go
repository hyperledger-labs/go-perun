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
	"bytes"
	"encoding/binary"
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
func ToLedgerChannelProposalMsg(protoEnvMsg *Envelope_LedgerChannelProposalMsg) (msg *client.LedgerChannelProposalMsg, err error) {
	protoMsg := protoEnvMsg.LedgerChannelProposalMsg

	msg = &client.LedgerChannelProposalMsg{}
	msg.BaseChannelProposal, err = ToBaseChannelProposal(protoMsg.BaseChannelProposal)
	if err != nil {
		return nil, err
	}
	msg.Participant, err = ToWalletAddr(protoMsg.Participant)
	if err != nil {
		return nil, errors.WithMessage(err, "participant address")
	}
	msg.Peers, err = ToWireAddrs(protoMsg.Peers)
	return msg, errors.WithMessage(err, "peers")
}

// ToSubChannelProposalMsg converts a protobuf Envelope_SubChannelProposalMsg to a client SubChannelProposalMsg.
func ToSubChannelProposalMsg(protoEnvMsg *Envelope_SubChannelProposalMsg) (msg *client.SubChannelProposalMsg, err error) {
	protoMsg := protoEnvMsg.SubChannelProposalMsg

	msg = &client.SubChannelProposalMsg{}
	copy(msg.Parent[:], protoMsg.Parent)
	msg.BaseChannelProposal, err = ToBaseChannelProposal(protoMsg.BaseChannelProposal)
	return msg, err
}

// ToVirtualChannelProposalMsg converts a protobuf Envelope_VirtualChannelProposalMsg to a client
// VirtualChannelProposalMsg.
func ToVirtualChannelProposalMsg(protoEnvMsg *Envelope_VirtualChannelProposalMsg) (msg *client.VirtualChannelProposalMsg, err error) {
	protoMsg := protoEnvMsg.VirtualChannelProposalMsg

	msg = &client.VirtualChannelProposalMsg{}
	msg.BaseChannelProposal, err = ToBaseChannelProposal(protoMsg.BaseChannelProposal)
	if err != nil {
		return nil, err
	}
	msg.Proposer, err = ToWalletAddr(protoMsg.Proposer)
	if err != nil {
		return nil, errors.WithMessage(err, "proposer")
	}
	msg.Parents = make([]channel.ID, len(protoMsg.Parents))
	for i := range protoMsg.Parents {
		copy(msg.Parents[i][:], protoMsg.Parents[i])
	}
	msg.IndexMaps = make([][]channel.Index, len(protoMsg.IndexMaps))
	for i := range protoMsg.IndexMaps {
		msg.IndexMaps[i], err = ToIndexMap(protoMsg.IndexMaps[i].IndexMap)
		if err != nil {
			return nil, err
		}
	}
	msg.Peers, err = ToWireAddrs(protoMsg.Peers)
	return msg, errors.WithMessage(err, "peers")
}

// ToLedgerChannelProposalAccMsg converts a protobuf Envelope_LedgerChannelProposalAccMsg to a client
// LedgerChannelProposalAccMsg.
func ToLedgerChannelProposalAccMsg(protoEnvMsg *Envelope_LedgerChannelProposalAccMsg) (msg *client.LedgerChannelProposalAccMsg, err error) {
	protoMsg := protoEnvMsg.LedgerChannelProposalAccMsg

	msg = &client.LedgerChannelProposalAccMsg{}
	msg.BaseChannelProposalAcc = ToBaseChannelProposalAcc(protoMsg.BaseChannelProposalAcc)
	msg.Participant, err = ToWalletAddr(protoMsg.Participant)
	return msg, errors.WithMessage(err, "participant")
}

// ToSubChannelProposalAccMsg converts a protobuf Envelope_SubChannelProposalAccMsg to a client
// SubChannelProposalAccMsg.
func ToSubChannelProposalAccMsg(protoEnvMsg *Envelope_SubChannelProposalAccMsg) (msg *client.SubChannelProposalAccMsg) {
	protoMsg := protoEnvMsg.SubChannelProposalAccMsg

	msg = &client.SubChannelProposalAccMsg{}
	msg.BaseChannelProposalAcc = ToBaseChannelProposalAcc(protoMsg.BaseChannelProposalAcc)
	return msg
}

// ToVirtualChannelProposalAccMsg converts a protobuf Envelope_VirtualChannelProposalAccMsg to a client
// VirtualChannelProposalAccMsg.
func ToVirtualChannelProposalAccMsg(protoEnvMsg *Envelope_VirtualChannelProposalAccMsg) (msg *client.VirtualChannelProposalAccMsg, err error) {
	protoMsg := protoEnvMsg.VirtualChannelProposalAccMsg

	msg = &client.VirtualChannelProposalAccMsg{}
	msg.BaseChannelProposalAcc = ToBaseChannelProposalAcc(protoMsg.BaseChannelProposalAcc)
	msg.Responder, err = ToWalletAddr(protoMsg.Responder)
	return msg, errors.WithMessage(err, "responder")
}

// ToChannelProposalRejMsg converts a protobuf Envelope_ChannelProposalRejMsg to a client ChannelProposalRejMsg.
func ToChannelProposalRejMsg(protoEnvMsg *Envelope_ChannelProposalRejMsg) (msg *client.ChannelProposalRejMsg) {
	protoMsg := protoEnvMsg.ChannelProposalRejMsg

	msg = &client.ChannelProposalRejMsg{}
	copy(msg.ProposalID[:], protoMsg.ProposalId)
	msg.Reason = protoMsg.Reason
	return msg
}

// ToWalletAddr converts a protobuf wallet address to a wallet.Address.
func ToWalletAddr(protoAddr *Address) (map[wallet.BackendID]wallet.Address, error) {
	addrMap := make(map[wallet.BackendID]wallet.Address)
	for i := range protoAddr.AddressMapping {
		var k int32
		if err := binary.Read(bytes.NewReader(protoAddr.AddressMapping[i].Key), binary.BigEndian, &k); err != nil {
			return nil, fmt.Errorf("failed to read key: %w", err)
		}
		addr := wallet.NewAddress(wallet.BackendID(k))
		if err := addr.UnmarshalBinary(protoAddr.AddressMapping[i].Address); err != nil {
			return nil, fmt.Errorf("failed to unmarshal address for key %d: %w", k, err)
		}

		addrMap[wallet.BackendID(k)] = addr
	}
	return addrMap, nil
}

// ToWireAddr converts a protobuf wallet address to a wallet.Address.
func ToWireAddr(protoAddr *Address) (map[wallet.BackendID]wire.Address, error) {
	addrMap := make(map[wallet.BackendID]wire.Address)
	for i := range protoAddr.AddressMapping {
		var k int32
		if err := binary.Read(bytes.NewReader(protoAddr.AddressMapping[i].Key), binary.BigEndian, &k); err != nil {
			return nil, fmt.Errorf("failed to read key: %w", err)
		}
		addr := wire.NewAddress()
		if err := addr.UnmarshalBinary(protoAddr.AddressMapping[i].Address); err != nil {
			return nil, fmt.Errorf("failed to unmarshal address for key %d: %w", k, err)
		}

		addrMap[wallet.BackendID(k)] = addr
	}
	return addrMap, nil
}

// ToWalletAddrs converts protobuf wallet addresses to a slice of wallet.Address.
func ToWalletAddrs(protoAddrs []*Address) ([]map[wallet.BackendID]wallet.Address, error) {
	addrs := make([]map[wallet.BackendID]wallet.Address, len(protoAddrs))
	for i := range protoAddrs {
		addrMap, err := ToWalletAddr(protoAddrs[i])
		if err != nil {
			return nil, errors.WithMessagef(err, "%d'th address", i)
		}
		addrs[i] = addrMap
	}
	return addrs, nil
}

// ToWireAddrs converts protobuf wire addresses to a slice of wire.Address.
func ToWireAddrs(protoAddrs []*Address) ([]map[wallet.BackendID]wire.Address, error) {
	addrMap := make([]map[wallet.BackendID]wire.Address, len(protoAddrs))
	var err error
	for i, addMap := range protoAddrs {
		addrMap[i], err = ToWireAddr(addMap)
		if err != nil {
			return nil, errors.WithMessagef(err, "%d'th address", i)
		}
	}
	return addrMap, nil
}

// ToBaseChannelProposal converts a protobuf BaseChannelProposal to a client BaseChannelProposal.
func ToBaseChannelProposal(protoProp *BaseChannelProposal) (prop client.BaseChannelProposal, err error) {
	prop.ChallengeDuration = protoProp.ChallengeDuration
	copy(prop.ProposalID[:], protoProp.ProposalId)
	copy(prop.NonceShare[:], protoProp.NonceShare)
	prop.InitBals, err = ToAllocation(protoProp.InitBals)
	if err != nil {
		return prop, errors.WithMessage(err, "init bals")
	}
	prop.FundingAgreement = ToBalances(protoProp.FundingAgreement)
	prop.App, prop.InitData, err = ToAppAndData(protoProp.App, protoProp.InitData)
	return prop, err
}

// ToBaseChannelProposalAcc converts a protobuf BaseChannelProposalAcc to a client BaseChannelProposalAcc.
func ToBaseChannelProposalAcc(protoPropAcc *BaseChannelProposalAcc) (propAcc client.BaseChannelProposalAcc) {
	copy(propAcc.ProposalID[:], protoPropAcc.ProposalId)
	copy(propAcc.NonceShare[:], protoPropAcc.NonceShare)
	return
}

// ToApp converts a protobuf app to a channel.App.
func ToApp(protoApp []byte) (app channel.App, err error) {
	if len(protoApp) == 0 {
		app = channel.NoApp()
		return app, nil
	}
	appDef, _ := channel.NewAppID()
	err = appDef.UnmarshalBinary(protoApp)
	if err != nil {
		return app, err
	}
	app, err = channel.Resolve(appDef)
	return app, err
}

// ToAppAndData converts protobuf app and data to a channel.App and channel.Data.
func ToAppAndData(protoApp, protoData []byte) (app channel.App, data channel.Data, err error) {
	if len(protoApp) == 0 {
		app = channel.NoApp()
		data = channel.NoData()
		return app, data, nil
	}
	appDef, _ := channel.NewAppID()
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

// ToIntSlice converts a [][]byte field from a protobuf message to a []int.
func ToIntSlice(backends [][]byte) ([]wallet.BackendID, error) {
	ints := make([]wallet.BackendID, len(backends))

	for i, backend := range backends {
		if len(backend) != 4 { //nolint:gomnd
			return nil, fmt.Errorf("backend %d length is not 4 bytes", i)
		}

		var value int32
		err := binary.Read(bytes.NewReader(backend), binary.BigEndian, &value)
		if err != nil {
			return nil, fmt.Errorf("failed to convert backend %d bytes to int: %w", i, err)
		}

		ints[i] = wallet.BackendID(value)
	}

	return ints, nil
}

// ToAllocation converts a protobuf allocation to a channel.Allocation.
func ToAllocation(protoAlloc *Allocation) (alloc *channel.Allocation, err error) {
	alloc = &channel.Allocation{}
	alloc.Backends, err = ToIntSlice(protoAlloc.Backends)
	if err != nil {
		return nil, errors.WithMessage(err, "backends")
	}
	alloc.Assets = make([]channel.Asset, len(protoAlloc.Assets))
	for i := range protoAlloc.Assets {
		alloc.Assets[i] = channel.NewAsset(alloc.Backends[i])
		err = alloc.Assets[i].UnmarshalBinary(protoAlloc.Assets[i])
		if err != nil {
			return nil, errors.WithMessagef(err, "%d'th asset", i)
		}
	}
	alloc.Locked = make([]channel.SubAlloc, len(protoAlloc.Locked))
	for i := range protoAlloc.Locked {
		alloc.Locked[i], err = ToSubAlloc(protoAlloc.Locked[i])
		if err != nil {
			return nil, errors.WithMessagef(err, "%d'th sub alloc", i)
		}
	}
	alloc.Balances = ToBalances(protoAlloc.Balances)
	return alloc, nil
}

// ToBalances converts a protobuf Balances to a channel.Balances.
func ToBalances(protoBalances *Balances) (balances channel.Balances) {
	balances = make([][]channel.Bal, len(protoBalances.Balances))
	for i := range protoBalances.Balances {
		balances[i] = ToBalance(protoBalances.Balances[i])
	}
	return balances
}

// ToBalance converts a protobuf Balance to a channel.Bal.
func ToBalance(protoBalance *Balance) (balance []channel.Bal) {
	balance = make([]channel.Bal, len(protoBalance.Balance))
	for j := range protoBalance.Balance {
		balance[j] = new(big.Int).SetBytes(protoBalance.Balance[j])
	}
	return balance
}

// ToSubAlloc converts a protobuf SubAlloc to a channel.SubAlloc.
func ToSubAlloc(protoSubAlloc *SubAlloc) (subAlloc channel.SubAlloc, err error) {
	subAlloc = channel.SubAlloc{}
	subAlloc.Bals = ToBalance(protoSubAlloc.Bals)
	if len(protoSubAlloc.Id) != len(subAlloc.ID) {
		return subAlloc, errors.New("sub alloc id has incorrect length")
	}
	copy(subAlloc.ID[:], protoSubAlloc.Id)
	subAlloc.IndexMap, err = ToIndexMap(protoSubAlloc.IndexMap.IndexMap)
	return subAlloc, err
}

// ToIndexMap converts a protobuf IndexMap to a channel.IndexMap.
func ToIndexMap(protoIndexMap []uint32) (indexMap []channel.Index, err error) {
	indexMap = make([]channel.Index, len(protoIndexMap))
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
func FromLedgerChannelProposalMsg(msg *client.LedgerChannelProposalMsg) (_ *Envelope_LedgerChannelProposalMsg, err error) {
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
func FromSubChannelProposalMsg(msg *client.SubChannelProposalMsg) (_ *Envelope_SubChannelProposalMsg, err error) {
	protoMsg := &SubChannelProposalMsg{}
	protoMsg.Parent = make([]byte, len(msg.Parent))
	copy(protoMsg.Parent, msg.Parent[:])
	protoMsg.BaseChannelProposal, err = FromBaseChannelProposal(msg.BaseChannelProposal)
	return &Envelope_SubChannelProposalMsg{protoMsg}, err
}

// FromVirtualChannelProposalMsg converts a client VirtualChannelProposalMsg to a protobuf
// Envelope_VirtualChannelProposalMsg.
func FromVirtualChannelProposalMsg(msg *client.VirtualChannelProposalMsg) (_ *Envelope_VirtualChannelProposalMsg, err error) {
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
func FromLedgerChannelProposalAccMsg(msg *client.LedgerChannelProposalAccMsg) (_ *Envelope_LedgerChannelProposalAccMsg, err error) {
	protoMsg := &LedgerChannelProposalAccMsg{}
	protoMsg.BaseChannelProposalAcc = FromBaseChannelProposalAcc(msg.BaseChannelProposalAcc)
	protoMsg.Participant, err = FromWalletAddr(msg.Participant)
	return &Envelope_LedgerChannelProposalAccMsg{protoMsg}, errors.WithMessage(err, "participant")
}

// FromSubChannelProposalAccMsg converts a client SubChannelProposalAccMsg to a protobuf
// Envelope_SubChannelProposalAccMsg.
func FromSubChannelProposalAccMsg(msg *client.SubChannelProposalAccMsg) (_ *Envelope_SubChannelProposalAccMsg) {
	protoMsg := &SubChannelProposalAccMsg{}
	protoMsg.BaseChannelProposalAcc = FromBaseChannelProposalAcc(msg.BaseChannelProposalAcc)
	return &Envelope_SubChannelProposalAccMsg{protoMsg}
}

// FromVirtualChannelProposalAccMsg converts a client VirtualChannelProposalAccMsg to a protobuf
// Envelope_VirtualChannelProposalAccMsg.
func FromVirtualChannelProposalAccMsg(msg *client.VirtualChannelProposalAccMsg) (_ *Envelope_VirtualChannelProposalAccMsg, err error) {
	protoMsg := &VirtualChannelProposalAccMsg{}
	protoMsg.BaseChannelProposalAcc = FromBaseChannelProposalAcc(msg.BaseChannelProposalAcc)
	protoMsg.Responder, err = FromWalletAddr(msg.Responder)
	return &Envelope_VirtualChannelProposalAccMsg{protoMsg}, errors.WithMessage(err, "responder")
}

// FromChannelProposalRejMsg converts a client ChannelProposalRejMsg to a protobuf Envelope_ChannelProposalRejMsg.
func FromChannelProposalRejMsg(msg *client.ChannelProposalRejMsg) (_ *Envelope_ChannelProposalRejMsg) {
	protoMsg := &ChannelProposalRejMsg{}
	protoMsg.ProposalId = make([]byte, len(msg.ProposalID))
	copy(protoMsg.ProposalId, msg.ProposalID[:])
	protoMsg.Reason = msg.Reason
	return &Envelope_ChannelProposalRejMsg{protoMsg}
}

// FromWalletAddr converts a wallet.Address to a protobuf wallet address.
func FromWalletAddr(addr map[wallet.BackendID]wallet.Address) (*Address, error) {
	var addressMappings []*AddressMapping //nolint:prealloc

	for key, address := range addr {
		keyBytes := make([]byte, 4) //nolint:gomnd
		binary.BigEndian.PutUint32(keyBytes, uint32(key))

		addressBytes, err := address.MarshalBinary()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal address for key %d: %w", key, err)
		}

		addressMappings = append(addressMappings, &AddressMapping{
			Key:     keyBytes,
			Address: addressBytes,
		})
	}

	return &Address{
		AddressMapping: addressMappings,
	}, nil
}

// FromWireAddr converts a wallet.Address to a protobuf wire address.
func FromWireAddr(addr map[wallet.BackendID]wire.Address) (*Address, error) {
	var addressMappings []*AddressMapping //nolint:prealloc

	for key, address := range addr {
		keyBytes := make([]byte, 4) //nolint:gomnd
		binary.BigEndian.PutUint32(keyBytes, uint32(key))

		addressBytes, err := address.MarshalBinary()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal address for key %d: %w", key, err)
		}

		addressMappings = append(addressMappings, &AddressMapping{
			Key:     keyBytes,
			Address: addressBytes,
		})
	}

	return &Address{
		AddressMapping: addressMappings,
	}, nil
}

// FromWalletAddrs converts a slice of wallet.Address to protobuf wallet addresses.
func FromWalletAddrs(addrs []map[wallet.BackendID]wallet.Address) (protoAddrs []*Address, err error) {
	protoAddrs = make([]*Address, len(addrs))
	for i := range addrs {
		protoAddrs[i], err = FromWalletAddr(addrs[i])
		if err != nil {
			return nil, errors.WithMessagef(err, "%d'th address", i)
		}
	}
	return protoAddrs, nil
}

// FromWireAddrs converts a slice of wire.Address to protobuf wire addresses.
func FromWireAddrs(addrs []map[wallet.BackendID]wire.Address) (protoAddrs []*Address, err error) {
	protoAddrs = make([]*Address, len(addrs))
	for i := range addrs {
		protoAddrs[i], err = FromWireAddr(addrs[i])
		if err != nil {
			return nil, errors.WithMessagef(err, "%d'th address", i)
		}
	}
	return protoAddrs, nil
}

// FromBaseChannelProposal converts a client BaseChannelProposal to a protobuf BaseChannelProposal.
func FromBaseChannelProposal(prop client.BaseChannelProposal) (protoProp *BaseChannelProposal, err error) {
	protoProp = &BaseChannelProposal{}

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
func FromBaseChannelProposalAcc(propAcc client.BaseChannelProposalAcc) (protoPropAcc *BaseChannelProposalAcc) {
	protoPropAcc = &BaseChannelProposalAcc{}
	protoPropAcc.ProposalId = make([]byte, len(propAcc.ProposalID))
	protoPropAcc.NonceShare = make([]byte, len(propAcc.NonceShare))
	copy(protoPropAcc.ProposalId, propAcc.ProposalID[:])
	copy(protoPropAcc.NonceShare, propAcc.NonceShare[:])
	return protoPropAcc
}

// FromApp converts a channel.App to a protobuf app.
func FromApp(app channel.App) (protoApp []byte, err error) {
	if channel.IsNoApp(app) {
		return []byte{}, nil
	}
	protoApp, err = app.Def().MarshalBinary()
	return protoApp, err
}

// FromAppAndData converts channel.App and channel.Data to protobuf app and data.
func FromAppAndData(app channel.App, data channel.Data) (protoApp, protoData []byte, err error) {
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

// FromAllocation converts a channel.Allocation to a protobuf Allocation.
func FromAllocation(alloc channel.Allocation) (protoAlloc *Allocation, err error) {
	protoAlloc = &Allocation{}
	protoAlloc.Backends = make([][]byte, len(alloc.Backends))
	for i := range alloc.Backends {
		protoAlloc.Backends[i] = make([]byte, 4) //nolint:gomnd
		binary.BigEndian.PutUint32(protoAlloc.Backends[i], uint32(alloc.Backends[i]))
	}
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
func FromBalances(balances channel.Balances) (protoBalances *Balances, err error) {
	protoBalances = &Balances{
		Balances: make([]*Balance, len(balances)),
	}
	for i := range balances {
		protoBalances.Balances[i], err = FromBalance(balances[i])
		if err != nil {
			return nil, errors.WithMessagef(err, "%d'th balance", i)
		}
	}
	return protoBalances, nil
}

// FromBalance converts a slice of channel.Bal to a protobuf Balance.
func FromBalance(balance []channel.Bal) (protoBalance *Balance, err error) {
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

// FromSubAlloc converts a channel.SubAlloc to a protobuf SubAlloc.
func FromSubAlloc(subAlloc channel.SubAlloc) (protoSubAlloc *SubAlloc, err error) {
	protoSubAlloc = &SubAlloc{}
	protoSubAlloc.Id = make([]byte, len(subAlloc.ID))
	copy(protoSubAlloc.Id, subAlloc.ID[:])
	protoSubAlloc.IndexMap = &IndexMap{IndexMap: FromIndexMap(subAlloc.IndexMap)}
	protoSubAlloc.Bals, err = FromBalance(subAlloc.Bals)
	return protoSubAlloc, err
}

// FromIndexMap converts a channel.IndexMap to a protobuf index map.
func FromIndexMap(indexMap []channel.Index) (protoIndexMap []uint32) {
	protoIndexMap = make([]uint32, len(indexMap))
	for i := range indexMap {
		protoIndexMap[i] = uint32(indexMap[i])
	}
	return protoIndexMap
}

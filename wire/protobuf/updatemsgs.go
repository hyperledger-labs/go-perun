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
	"math"
	"math/big"

	"github.com/pkg/errors"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/client"
)

// ToChannelUpdateMsg converts a protobuf Envelope_ChannelUpdateMsg to a client.ChannelUpdateMsg.
func ToChannelUpdateMsg(protoEnvMsg *Envelope_ChannelUpdateMsg) (*client.ChannelUpdateMsg, error) {
	update, err := ToChannelUpdate(protoEnvMsg.ChannelUpdateMsg)
	return &update, err
}

// ToVirtualChannelFundingProposalMsg converts a protobuf Envelope_VirtualChannelFundingProposalMsg to a
// client.VirtualChannelFundingProposalMsg.
func ToVirtualChannelFundingProposalMsg(protoEnvMsg *Envelope_VirtualChannelFundingProposalMsg) (
	msg *client.VirtualChannelFundingProposalMsg,
	err error,
) {
	protoMsg := protoEnvMsg.VirtualChannelFundingProposalMsg

	msg = &client.VirtualChannelFundingProposalMsg{}
	msg.Initial, err = ToSignedState(protoMsg.Initial)
	if err != nil {
		return nil, errors.WithMessage(err, "initial state")
	}
	msg.IndexMap, err = ToIndexMap(protoMsg.IndexMap.IndexMap)
	if err != nil {
		return nil, err
	}
	msg.ChannelUpdateMsg, err = ToChannelUpdate(protoMsg.ChannelUpdateMsg)
	return msg, err
}

// ToVirtualChannelSettlementProposalMsg converts a protobuf Envelope_VirtualChannelSettlementProposalMsg to a
// client.VirtualChannelSettlementProposalMsg.
func ToVirtualChannelSettlementProposalMsg(protoEnvMsg *Envelope_VirtualChannelSettlementProposalMsg) (
	msg *client.VirtualChannelSettlementProposalMsg,
	err error,
) {
	protoMsg := protoEnvMsg.VirtualChannelSettlementProposalMsg

	msg = &client.VirtualChannelSettlementProposalMsg{}
	msg.Final, err = ToSignedState(protoMsg.Final)
	if err != nil {
		return nil, errors.WithMessage(err, "final state")
	}
	msg.ChannelUpdateMsg, err = ToChannelUpdate(protoMsg.ChannelUpdateMsg)
	return msg, err
}

// ToChannelUpdateAccMsg converts a protobuf Envelope_ChannelUpdateAccMsg to a client.ChannelUpdateAccMsg.
func ToChannelUpdateAccMsg(protoEnvMsg *Envelope_ChannelUpdateAccMsg) (msg *client.ChannelUpdateAccMsg) {
	protoMsg := protoEnvMsg.ChannelUpdateAccMsg

	msg = &client.ChannelUpdateAccMsg{}
	msg.ChannelID, _ = ToIDs(protoEnvMsg.ChannelUpdateAccMsg.ChannelId)
	msg.Version = protoMsg.Version
	msg.Sig = make([]byte, len(protoMsg.Sig))
	copy(msg.Sig, protoMsg.Sig)
	return msg
}

// ToChannelUpdateRejMsg converts a protobuf Envelope_ChannelUpdateRejMsg to a client.ChannelUpdateRejMsg.
func ToChannelUpdateRejMsg(protoEnvMsg *Envelope_ChannelUpdateRejMsg) (msg *client.ChannelUpdateRejMsg) {
	protoMsg := protoEnvMsg.ChannelUpdateRejMsg

	msg = &client.ChannelUpdateRejMsg{}
	msg.ChannelID, _ = ToIDs(protoEnvMsg.ChannelUpdateRejMsg.ChannelId)
	msg.Version = protoMsg.Version
	msg.Reason = protoMsg.Reason
	return msg
}

// ToChannelUpdate parses protobuf channel updates.
func ToChannelUpdate(protoUpdate *ChannelUpdateMsg) (update client.ChannelUpdateMsg, err error) {
	if protoUpdate.ChannelUpdate.ActorIdx > math.MaxUint16 {
		return update, errors.New("actor index is invalid")
	}
	update.ActorIdx = channel.Index(protoUpdate.ChannelUpdate.ActorIdx)
	update.Sig = make([]byte, len(protoUpdate.Sig))
	copy(update.Sig, protoUpdate.Sig)
	update.State, err = ToState(protoUpdate.ChannelUpdate.State)
	return update, err
}

// ToSignedState converts a protobuf SignedState to a channel.SignedState.
func ToSignedState(protoSignedState *SignedState) (signedState channel.SignedState, err error) {
	signedState.Params, err = ToParams(protoSignedState.Params)
	if err != nil {
		return signedState, err
	}
	signedState.Sigs = make([][]byte, len(protoSignedState.Sigs))
	for i := range protoSignedState.Sigs {
		signedState.Sigs[i] = make([]byte, len(protoSignedState.Sigs[i]))
		copy(signedState.Sigs[i], protoSignedState.Sigs[i])
	}
	signedState.State, err = ToState(protoSignedState.State)
	return signedState, err
}

// ToParams converts a protobuf Params to a channel.Params.
func ToParams(protoParams *Params) (*channel.Params, error) {
	app, err := ToApp(protoParams.App)
	if err != nil {
		return nil, err
	}
	parts, err := ToWalletAddrs(protoParams.Parts)
	if err != nil {
		return nil, errors.WithMessage(err, "parts")
	}
	params := channel.NewParamsUnsafe(
		protoParams.ChallengeDuration,
		parts,
		app,
		(new(big.Int)).SetBytes(protoParams.Nonce),
		protoParams.LedgerChannel,
		protoParams.VirtualChannel)

	return params, nil
}

// ToState converts a protobuf State to a channel.State.
func ToState(protoState *State) (state *channel.State, err error) {
	state = &channel.State{}
	state.ID, err = ToIDs(protoState.Id)
	state.Version = protoState.Version
	state.IsFinal = protoState.IsFinal
	allocation, err := ToAllocation(protoState.Allocation)
	if err != nil {
		return nil, errors.WithMessage(err, "allocation")
	}
	state.Allocation = *allocation
	state.App, state.Data, err = ToAppAndData(protoState.App, protoState.Data)
	return state, err
}

// FromChannelUpdateMsg converts a client.ChannelUpdateMsg to a protobuf Envelope_ChannelUpdateMsg.
func FromChannelUpdateMsg(msg *client.ChannelUpdateMsg) (*Envelope_ChannelUpdateMsg, error) {
	protoMsg, err := FromChannelUpdate(msg)
	return &Envelope_ChannelUpdateMsg{protoMsg}, err
}

// FromVirtualChannelFundingProposalMsg converts a client.VirtualChannelFundingProposalMsg to a protobuf
// Envelope_VirtualChannelFundingProposalMsg.
func FromVirtualChannelFundingProposalMsg(msg *client.VirtualChannelFundingProposalMsg) (
	_ *Envelope_VirtualChannelFundingProposalMsg,
	err error,
) {
	protoMsg := &VirtualChannelFundingProposalMsg{}

	protoMsg.Initial, err = FromSignedState(&msg.Initial)
	if err != nil {
		return nil, errors.WithMessage(err, "initial state")
	}
	protoMsg.IndexMap = &IndexMap{IndexMap: FromIndexMap(msg.IndexMap)}
	protoMsg.ChannelUpdateMsg, err = FromChannelUpdate(&msg.ChannelUpdateMsg)
	return &Envelope_VirtualChannelFundingProposalMsg{protoMsg}, err
}

// FromVirtualChannelSettlementProposalMsg converts a client.VirtualChannelSettlementProposalMsg to a protobuf
// Envelope_VirtualChannelSettlementProposalMsg.
func FromVirtualChannelSettlementProposalMsg(msg *client.VirtualChannelSettlementProposalMsg) (
	_ *Envelope_VirtualChannelSettlementProposalMsg,
	err error,
) {
	protoMsg := &VirtualChannelSettlementProposalMsg{}

	protoMsg.ChannelUpdateMsg, err = FromChannelUpdate(&msg.ChannelUpdateMsg)
	if err != nil {
		return nil, err
	}
	protoMsg.Final, err = FromSignedState(&msg.Final)
	return &Envelope_VirtualChannelSettlementProposalMsg{protoMsg}, err
}

// FromChannelUpdateAccMsg converts a client.ChannelUpdateAccMsg to a protobuf Envelope_ChannelUpdateAccMsg.
func FromChannelUpdateAccMsg(msg *client.ChannelUpdateAccMsg) *Envelope_ChannelUpdateAccMsg {
	protoMsg := &ChannelUpdateAccMsg{}

	protoMsg.ChannelId, _ = FromIDs(msg.ChannelID)
	protoMsg.Sig = make([]byte, len(msg.Sig))
	copy(protoMsg.Sig, msg.Sig)
	protoMsg.Version = msg.Version
	return &Envelope_ChannelUpdateAccMsg{protoMsg}
}

// FromChannelUpdateRejMsg converts a client.ChannelUpdateRejMsg to a protobuf Envelope_ChannelUpdateRejMsg.
func FromChannelUpdateRejMsg(msg *client.ChannelUpdateRejMsg) *Envelope_ChannelUpdateRejMsg {
	protoMsg := &ChannelUpdateRejMsg{}
	protoMsg.ChannelId, _ = FromIDs(msg.ChannelID)
	protoMsg.Version = msg.Version
	protoMsg.Reason = msg.Reason
	return &Envelope_ChannelUpdateRejMsg{protoMsg}
}

// FromChannelUpdate converts a client.ChannelUpdateMsg to a protobuf ChannelUpdateMsg.
func FromChannelUpdate(update *client.ChannelUpdateMsg) (protoUpdate *ChannelUpdateMsg, err error) {
	protoUpdate = &ChannelUpdateMsg{}
	protoUpdate.ChannelUpdate = &ChannelUpdate{}

	protoUpdate.ChannelUpdate.ActorIdx = uint32(update.ChannelUpdate.ActorIdx)
	protoUpdate.Sig = make([]byte, len(update.Sig))
	copy(protoUpdate.Sig, update.Sig)
	protoUpdate.ChannelUpdate.State, err = FromState(update.ChannelUpdate.State)
	return protoUpdate, err
}

// FromSignedState converts a channel.SignedState to a protobuf SignedState.
func FromSignedState(signedState *channel.SignedState) (protoSignedState *SignedState, err error) {
	protoSignedState = &SignedState{}
	protoSignedState.Sigs = make([][]byte, len(signedState.Sigs))
	for i := range signedState.Sigs {
		protoSignedState.Sigs[i] = make([]byte, len(signedState.Sigs[i]))
		copy(protoSignedState.Sigs[i], signedState.Sigs[i])
	}
	protoSignedState.Params, err = FromParams(signedState.Params)
	if err != nil {
		return nil, err
	}
	protoSignedState.State, err = FromState(signedState.State)
	return protoSignedState, err
}

// FromParams converts a channel.Params to a protobuf Params.
func FromParams(params *channel.Params) (protoParams *Params, err error) {
	protoParams = &Params{}

	protoParams.Nonce = params.Nonce.Bytes()
	protoParams.ChallengeDuration = params.ChallengeDuration
	protoParams.LedgerChannel = params.LedgerChannel
	protoParams.VirtualChannel = params.VirtualChannel
	protoParams.Parts, err = FromWalletAddrs(params.Parts)
	if err != nil {
		return nil, errors.WithMessage(err, "parts")
	}
	protoParams.App, err = FromApp(params.App)
	return protoParams, err
}

// FromState converts a channel.State to a protobuf State.
func FromState(state *channel.State) (protoState *State, err error) {
	protoState = &State{}
	protoState.Id, _ = FromIDs(state.ID)
	protoState.Version = state.Version
	protoState.IsFinal = state.IsFinal
	protoState.Allocation, err = FromAllocation(state.Allocation)
	if err != nil {
		return nil, errors.WithMessage(err, "allocation")
	}
	protoState.App, protoState.Data, err = FromAppAndData(state.App, state.Data)
	return protoState, err
}

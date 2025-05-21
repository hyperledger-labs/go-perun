// Copyright 2025 - See NOTICE file for copyright holders.
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
	msg.Initial, err = ToSignedState(protoMsg.GetInitial())
	if err != nil {
		return nil, errors.WithMessage(err, "initial state")
	}
	msg.IndexMap, err = ToIndexMap(protoMsg.GetIndexMap().GetIndexMap())
	if err != nil {
		return nil, err
	}
	msg.ChannelUpdateMsg, err = ToChannelUpdate(protoMsg.GetChannelUpdateMsg())
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
	msg.Final, err = ToSignedState(protoMsg.GetFinal())
	if err != nil {
		return nil, errors.WithMessage(err, "final state")
	}
	msg.ChannelUpdateMsg, err = ToChannelUpdate(protoMsg.GetChannelUpdateMsg())
	return msg, err
}

// ToChannelUpdateAccMsg converts a protobuf Envelope_ChannelUpdateAccMsg to a client.ChannelUpdateAccMsg.
func ToChannelUpdateAccMsg(protoEnvMsg *Envelope_ChannelUpdateAccMsg) (msg *client.ChannelUpdateAccMsg) {
	protoMsg := protoEnvMsg.ChannelUpdateAccMsg

	msg = &client.ChannelUpdateAccMsg{}
	copy(msg.ChannelID[:], protoMsg.GetChannelId())
	msg.Version = protoMsg.GetVersion()
	msg.Sig = make([]byte, len(protoMsg.GetSig()))
	copy(msg.Sig, protoMsg.GetSig())
	return msg
}

// ToChannelUpdateRejMsg converts a protobuf Envelope_ChannelUpdateRejMsg to a client.ChannelUpdateRejMsg.
func ToChannelUpdateRejMsg(protoEnvMsg *Envelope_ChannelUpdateRejMsg) (msg *client.ChannelUpdateRejMsg) {
	protoMsg := protoEnvMsg.ChannelUpdateRejMsg

	msg = &client.ChannelUpdateRejMsg{}
	copy(msg.ChannelID[:], protoMsg.GetChannelId())
	msg.Version = protoMsg.GetVersion()
	msg.Reason = protoMsg.GetReason()
	return msg
}

// ToChannelUpdate parses protobuf channel updates.
func ToChannelUpdate(protoUpdate *ChannelUpdateMsg) (update client.ChannelUpdateMsg, err error) {
	if protoUpdate.GetChannelUpdate().GetActorIdx() > math.MaxUint16 {
		return update, errors.New("actor index is invalid")
	}
	idx := protoUpdate.GetChannelUpdate().GetActorIdx()
	if idx > math.MaxUint16 {
		return update, fmt.Errorf("invalid actor index: %d", idx)
	}
	update.ActorIdx = channel.Index(uint16(idx))

	update.Sig = make([]byte, len(protoUpdate.GetSig()))
	copy(update.Sig, protoUpdate.GetSig())
	update.State, err = ToState(protoUpdate.GetChannelUpdate().GetState())
	return update, err
}

// ToSignedState converts a protobuf SignedState to a channel.SignedState.
func ToSignedState(protoSignedState *SignedState) (signedState channel.SignedState, err error) {
	signedState.Params, err = ToParams(protoSignedState.GetParams())
	if err != nil {
		return signedState, err
	}
	signedState.Sigs = make([][]byte, len(protoSignedState.GetSigs()))
	for i := range protoSignedState.GetSigs() {
		signedState.Sigs[i] = make([]byte, len(protoSignedState.GetSigs()[i]))
		copy(signedState.Sigs[i], protoSignedState.GetSigs()[i])
	}
	signedState.State, err = ToState(protoSignedState.GetState())
	return signedState, err
}

// ToParams converts a protobuf Params to a channel.Params.
func ToParams(protoParams *Params) (*channel.Params, error) {
	app, err := ToApp(protoParams.GetApp())
	if err != nil {
		return nil, err
	}
	parts, err := ToWalletAddrs(protoParams.GetParts())
	if err != nil {
		return nil, errors.WithMessage(err, "parts")
	}

	var aux channel.Aux
	copy(aux[:], protoParams.GetAux())
	params := channel.NewParamsUnsafe(
		protoParams.GetChallengeDuration(),
		parts,
		app,
		(new(big.Int)).SetBytes(protoParams.GetNonce()),
		protoParams.GetLedgerChannel(),
		protoParams.GetVirtualChannel(),
		aux,
	)

	return params, nil
}

// ToState converts a protobuf State to a channel.State.
func ToState(protoState *State) (state *channel.State, err error) {
	state = &channel.State{}
	copy(state.ID[:], protoState.GetId())
	state.Version = protoState.GetVersion()
	state.IsFinal = protoState.GetIsFinal()
	allocation, err := ToAllocation(protoState.GetAllocation())
	if err != nil {
		return nil, errors.WithMessage(err, "allocation")
	}
	state.Allocation = *allocation
	state.App, state.Data, err = ToAppAndData(protoState.GetApp(), protoState.GetData())
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

	protoMsg.ChannelId = make([]byte, len(msg.ChannelID))
	copy(protoMsg.GetChannelId(), msg.ChannelID[:])
	protoMsg.Sig = make([]byte, len(msg.Sig))
	copy(protoMsg.GetSig(), msg.Sig)
	protoMsg.Version = msg.Version
	return &Envelope_ChannelUpdateAccMsg{protoMsg}
}

// FromChannelUpdateRejMsg converts a client.ChannelUpdateRejMsg to a protobuf Envelope_ChannelUpdateRejMsg.
func FromChannelUpdateRejMsg(msg *client.ChannelUpdateRejMsg) *Envelope_ChannelUpdateRejMsg {
	protoMsg := &ChannelUpdateRejMsg{}
	protoMsg.ChannelId = make([]byte, len(msg.ChannelID))
	copy(protoMsg.GetChannelId(), msg.ChannelID[:])
	protoMsg.Version = msg.Version
	protoMsg.Reason = msg.Reason
	return &Envelope_ChannelUpdateRejMsg{protoMsg}
}

// FromChannelUpdate converts a client.ChannelUpdateMsg to a protobuf ChannelUpdateMsg.
func FromChannelUpdate(update *client.ChannelUpdateMsg) (protoUpdate *ChannelUpdateMsg, err error) {
	protoUpdate = &ChannelUpdateMsg{}
	protoUpdate.ChannelUpdate = &ChannelUpdate{}

	protoUpdate.ChannelUpdate.ActorIdx = uint32(update.ActorIdx)
	protoUpdate.Sig = make([]byte, len(update.Sig))
	copy(protoUpdate.GetSig(), update.Sig)
	protoUpdate.ChannelUpdate.State, err = FromState(update.State)
	return protoUpdate, err
}

// FromSignedState converts a channel.SignedState to a protobuf SignedState.
func FromSignedState(signedState *channel.SignedState) (protoSignedState *SignedState, err error) {
	protoSignedState = &SignedState{}
	protoSignedState.Sigs = make([][]byte, len(signedState.Sigs))
	for i := range signedState.Sigs {
		protoSignedState.Sigs[i] = make([]byte, len(signedState.Sigs[i]))
		copy(protoSignedState.GetSigs()[i], signedState.Sigs[i])
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
	protoParams.Aux = params.Aux[:]
	return protoParams, err
}

// FromState converts a channel.State to a protobuf State.
func FromState(state *channel.State) (protoState *State, err error) {
	protoState = &State{}

	protoState.Id = make([]byte, len(state.ID))
	copy(protoState.GetId(), state.ID[:])
	protoState.Version = state.Version
	protoState.IsFinal = state.IsFinal
	protoState.Allocation, err = FromAllocation(state.Allocation)
	if err != nil {
		return nil, errors.WithMessage(err, "allocation")
	}
	protoState.App, protoState.Data, err = FromAppAndData(state.App, state.Data)
	return protoState, err
}

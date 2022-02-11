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

func toChannelUpdateMsg(protoEnvMsg *Envelope_ChannelUpdateMsg) (*client.ChannelUpdateMsg, error) {
	update, err := toChannelUpdate(protoEnvMsg.ChannelUpdateMsg)
	return &update, err
}

func toVirtualChannelFundingProposalMsg(protoEnvMsg *Envelope_VirtualChannelFundingProposalMsg) (
	msg *client.VirtualChannelFundingProposalMsg,
	err error,
) {
	protoMsg := protoEnvMsg.VirtualChannelFundingProposalMsg

	msg = &client.VirtualChannelFundingProposalMsg{}
	msg.Initial, err = toSignedState(protoMsg.Initial)
	if err != nil {
		return nil, errors.WithMessage(err, "initial state")
	}
	msg.IndexMap, err = toIndexMap(protoMsg.IndexMap.IndexMap)
	if err != nil {
		return nil, err
	}
	msg.ChannelUpdateMsg, err = toChannelUpdate(protoMsg.ChannelUpdateMsg)
	return msg, err
}

func toVirtualChannelSettlementProposalMsg(protoEnvMsg *Envelope_VirtualChannelSettlementProposalMsg) (
	msg *client.VirtualChannelSettlementProposalMsg,
	err error,
) {
	protoMsg := protoEnvMsg.VirtualChannelSettlementProposalMsg

	msg = &client.VirtualChannelSettlementProposalMsg{}
	msg.Final, err = toSignedState(protoMsg.Final)
	if err != nil {
		return nil, errors.WithMessage(err, "final state")
	}
	msg.ChannelUpdateMsg, err = toChannelUpdate(protoMsg.ChannelUpdateMsg)
	return msg, err
}

func toChannelUpdateAccMsg(protoEnvMsg *Envelope_ChannelUpdateAccMsg) (msg *client.ChannelUpdateAccMsg) {
	protoMsg := protoEnvMsg.ChannelUpdateAccMsg

	msg = &client.ChannelUpdateAccMsg{}
	copy(msg.ChannelID[:], protoMsg.ChannelId)
	msg.Version = protoMsg.Version
	msg.Sig = make([]byte, len(protoMsg.Sig))
	copy(msg.Sig, protoMsg.Sig)
	return msg
}

func toChannelUpdateRejMsg(protoEnvMsg *Envelope_ChannelUpdateRejMsg) (msg *client.ChannelUpdateRejMsg) {
	protoMsg := protoEnvMsg.ChannelUpdateRejMsg

	msg = &client.ChannelUpdateRejMsg{}
	copy(msg.ChannelID[:], protoMsg.ChannelId)
	msg.Version = protoMsg.Version
	msg.Reason = protoMsg.Reason
	return msg
}

func toChannelUpdate(protoUpdate *ChannelUpdateMsg) (update client.ChannelUpdateMsg, err error) {
	if protoUpdate.ChannelUpdate.ActorIdx > math.MaxUint16 {
		return update, errors.New("actor index is invalid")
	}
	update.ActorIdx = channel.Index(protoUpdate.ChannelUpdate.ActorIdx)
	update.Sig = make([]byte, len(protoUpdate.Sig))
	copy(update.Sig, protoUpdate.Sig)
	update.State, err = toState(protoUpdate.ChannelUpdate.State)
	return update, err
}

func toSignedState(protoSignedState *SignedState) (signedState channel.SignedState, err error) {
	signedState.Params, err = toParams(protoSignedState.Params)
	if err != nil {
		return signedState, err
	}
	signedState.Sigs = make([][]byte, len(protoSignedState.Sigs))
	for i := range protoSignedState.Sigs {
		signedState.Sigs[i] = make([]byte, len(protoSignedState.Sigs[i]))
		copy(signedState.Sigs[i], protoSignedState.Sigs[i])
	}
	signedState.State, err = toState(protoSignedState.State)
	return signedState, err
}

func toParams(protoParams *Params) (*channel.Params, error) {
	app, err := toApp(protoParams.App)
	if err != nil {
		return nil, err
	}
	parts, err := toWalletAddrs(protoParams.Parts)
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

func toState(protoState *State) (state *channel.State, err error) {
	state = &channel.State{}
	copy(state.ID[:], protoState.Id)
	state.Version = protoState.Version
	state.IsFinal = protoState.IsFinal
	allocation, err := toAllocation(protoState.Allocation)
	if err != nil {
		return nil, errors.WithMessage(err, "allocation")
	}
	state.Allocation = *allocation
	state.App, state.Data, err = toAppAndData(protoState.App, protoState.Data)
	return state, err
}

func fromChannelUpdateMsg(msg *client.ChannelUpdateMsg) (*Envelope_ChannelUpdateMsg, error) {
	protoMsg, err := fromChannelUpdate(msg)
	return &Envelope_ChannelUpdateMsg{protoMsg}, err
}

func fromVirtualChannelFundingProposalMsg(msg *client.VirtualChannelFundingProposalMsg) (
	_ *Envelope_VirtualChannelFundingProposalMsg,
	err error,
) {
	protoMsg := &VirtualChannelFundingProposalMsg{}

	protoMsg.Initial, err = fromSignedState(&msg.Initial)
	if err != nil {
		return nil, errors.WithMessage(err, "initial state")
	}
	protoMsg.IndexMap = &IndexMap{IndexMap: fromIndexMap(msg.IndexMap)}
	protoMsg.ChannelUpdateMsg, err = fromChannelUpdate(&msg.ChannelUpdateMsg)
	return &Envelope_VirtualChannelFundingProposalMsg{protoMsg}, err
}

func fromVirtualChannelSettlementProposalMsg(msg *client.VirtualChannelSettlementProposalMsg) (
	_ *Envelope_VirtualChannelSettlementProposalMsg,
	err error,
) {
	protoMsg := &VirtualChannelSettlementProposalMsg{}

	protoMsg.ChannelUpdateMsg, err = fromChannelUpdate(&msg.ChannelUpdateMsg)
	if err != nil {
		return nil, err
	}
	protoMsg.Final, err = fromSignedState(&msg.Final)
	return &Envelope_VirtualChannelSettlementProposalMsg{protoMsg}, err
}

func fromChannelUpdateAccMsg(msg *client.ChannelUpdateAccMsg) *Envelope_ChannelUpdateAccMsg {
	protoMsg := &ChannelUpdateAccMsg{}

	protoMsg.ChannelId = make([]byte, len(msg.ChannelID))
	copy(protoMsg.ChannelId, msg.ChannelID[:])
	protoMsg.Sig = make([]byte, len(msg.Sig))
	copy(protoMsg.Sig, msg.Sig)
	protoMsg.Version = msg.Version
	return &Envelope_ChannelUpdateAccMsg{protoMsg}
}

func fromChannelUpdateRejMsg(msg *client.ChannelUpdateRejMsg) *Envelope_ChannelUpdateRejMsg {
	protoMsg := &ChannelUpdateRejMsg{}
	protoMsg.ChannelId = make([]byte, len(msg.ChannelID))
	copy(protoMsg.ChannelId, msg.ChannelID[:])
	protoMsg.Version = msg.Version
	protoMsg.Reason = msg.Reason
	return &Envelope_ChannelUpdateRejMsg{protoMsg}
}

func fromChannelUpdate(update *client.ChannelUpdateMsg) (protoUpdate *ChannelUpdateMsg, err error) {
	protoUpdate = &ChannelUpdateMsg{}
	protoUpdate.ChannelUpdate = &ChannelUpdate{}

	protoUpdate.ChannelUpdate.ActorIdx = uint32(update.ChannelUpdate.ActorIdx)
	protoUpdate.Sig = make([]byte, len(update.Sig))
	copy(protoUpdate.Sig, update.Sig)
	protoUpdate.ChannelUpdate.State, err = fromState(update.ChannelUpdate.State)
	return protoUpdate, err
}

func fromSignedState(signedState *channel.SignedState) (protoSignedState *SignedState, err error) {
	protoSignedState = &SignedState{}
	protoSignedState.Sigs = make([][]byte, len(signedState.Sigs))
	for i := range signedState.Sigs {
		protoSignedState.Sigs[i] = make([]byte, len(signedState.Sigs[i]))
		copy(protoSignedState.Sigs[i], signedState.Sigs[i])
	}
	protoSignedState.Params, err = fromParams(signedState.Params)
	if err != nil {
		return nil, err
	}
	protoSignedState.State, err = fromState(signedState.State)
	return protoSignedState, err
}

func fromParams(params *channel.Params) (protoParams *Params, err error) {
	protoParams = &Params{}

	protoParams.Nonce = params.Nonce.Bytes()
	protoParams.ChallengeDuration = params.ChallengeDuration
	protoParams.LedgerChannel = params.LedgerChannel
	protoParams.VirtualChannel = params.VirtualChannel
	protoParams.Parts, err = fromWalletAddrs(params.Parts)
	if err != nil {
		return nil, errors.WithMessage(err, "parts")
	}
	protoParams.App, err = fromApp(params.App)
	return protoParams, err
}

func fromState(state *channel.State) (protoState *State, err error) {
	protoState = &State{}

	protoState.Id = make([]byte, len(state.ID))
	copy(protoState.Id, state.ID[:])
	protoState.Version = state.Version
	protoState.IsFinal = state.IsFinal
	protoState.Allocation, err = fromAllocation(state.Allocation)
	if err != nil {
		return nil, errors.WithMessage(err, "allocation")
	}
	protoState.App, protoState.Data, err = fromAppAndData(state.App, state.Data)
	return protoState, err
}

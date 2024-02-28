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
	"perun.network/go-perun/channel"
	"perun.network/go-perun/client"
)

//nolint:forbidigo
func toChannelSyncMsg(protoEnvMsg *Envelope_ChannelSyncMsg) (*client.ChannelSyncMsg, error) {
	protoMsg := protoEnvMsg.ChannelSyncMsg

	var err error
	msg := &client.ChannelSyncMsg{}
	msg.Phase = channel.Phase(protoMsg.GetPhase())
	msg.CurrentTX.Sigs = make([][]byte, len(protoMsg.GetCurrentTx().GetSigs()))
	for i := range protoMsg.GetCurrentTx().GetSigs() {
		msg.CurrentTX.Sigs[i] = make([]byte, len(protoMsg.GetCurrentTx().GetSigs()[i]))
		copy(msg.CurrentTX.Sigs[i], protoMsg.GetCurrentTx().GetSigs()[i])
	}
	msg.CurrentTX.State, err = ToState(protoMsg.GetCurrentTx().GetState())
	return msg, err
}

//nolint:protogetter
func fromChannelSyncMsg(msg *client.ChannelSyncMsg) (*Envelope_ChannelSyncMsg, error) {
	var err error
	protoMsg := &ChannelSyncMsg{}
	protoMsg.CurrentTx = &Transaction{}

	protoMsg.Phase = uint32(msg.Phase)
	protoMsg.CurrentTx.Sigs = make([][]byte, len(msg.CurrentTX.Sigs))
	for i := range msg.CurrentTX.Sigs {
		protoMsg.CurrentTx.Sigs[i] = make([]byte, len(msg.CurrentTX.Sigs[i]))
		copy(protoMsg.CurrentTx.Sigs[i], msg.CurrentTX.Sigs[i])
	}
	protoMsg.CurrentTx.State, err = FromState(msg.CurrentTX.State)
	return &Envelope_ChannelSyncMsg{protoMsg}, err
}

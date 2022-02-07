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
	"github.com/pkg/errors"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/client"
)

func toChannelUpdate(in *ChannelUpdateMsg) (*client.ChannelUpdateMsg, error) {
	state, err := toState(in.GetChannelUpdate().GetState())
	if err != nil {
		return nil, errors.WithMessage(err, "encoding state")
	}
	out := &client.ChannelUpdateMsg{
		ChannelUpdate: client.ChannelUpdate{
			State:    state,
			ActorIdx: channel.Index(in.ChannelUpdate.ActorIdx), // write a test for safe conversion.
		},
		Sig: make([]byte, len(in.Sig)),
	}
	copy(out.Sig, in.Sig)
	return out, nil
}

func toState(in *State) (*channel.State, error) {
	allocation, err := toAllocation(in.Allocation)
	if err != nil {
		return nil, errors.WithMessage(err, "encoding allocation")
	}
	app, data, err := toAppAndData(in.App, in.Data)
	if err != nil {
		return nil, err
	}

	out := &channel.State{
		Version:    in.Version,
		App:        app,
		Allocation: allocation,
		Data:       data,
		IsFinal:    in.IsFinal,
	}
	copy(out.ID[:], in.Id)
	return out, nil
}

func toChannelUpdateAcc(in *ChannelUpdateAccMsg) (out *client.ChannelUpdateAccMsg) {
	out = &client.ChannelUpdateAccMsg{
		Version: in.Version,
		Sig:     make([]byte, len(in.Sig)),
	}
	copy(out.ChannelID[:], in.ChannelId)
	copy(out.Sig, in.Sig)
	return
}

func toChannelUpdateRej(in *ChannelUpdateRejMsg) (out *client.ChannelUpdateRejMsg) {
	out = &client.ChannelUpdateRejMsg{
		Version: in.Version,
		Reason:  in.Reason,
	}
	copy(out.ChannelID[:], in.ChannelId)
	return
}

func fromChannelUpdate(in *client.ChannelUpdateMsg) (*ChannelUpdateMsg, error) {
	state, err := fromState(in.ChannelUpdate.State)
	if err != nil {
		return nil, errors.WithMessage(err, "encoding state")
	}
	out := &ChannelUpdateMsg{
		ChannelUpdate: &ChannelUpdate{
			State:    state,
			ActorIdx: uint32(in.ChannelUpdate.ActorIdx),
		},
		Sig: make([]byte, len(in.Sig)),
	}
	copy(out.Sig, in.Sig)
	return out, nil
}

func fromState(in *channel.State) (*State, error) {
	allocation, err := fromAllocation(in.Allocation)
	if err != nil {
		return nil, errors.WithMessage(err, "encoding allocation")
	}
	app, data, err := fromAppAndData(in.App, in.Data)
	if err != nil {
		return nil, err
	}

	out := &State{
		Id:         make([]byte, len(in.ID)),
		Version:    in.Version,
		App:        app,
		Allocation: allocation,
		Data:       data,
		IsFinal:    in.IsFinal,
	}
	copy(out.Id, in.ID[:])
	return out, nil
}

func fromChannelUpdateAcc(in *client.ChannelUpdateAccMsg) (out *ChannelUpdateAccMsg) {
	out = &ChannelUpdateAccMsg{
		ChannelId: make([]byte, len(in.ChannelID)),
		Sig:       make([]byte, len(in.Sig)),
		Version:   in.Version,
	}
	copy(out.ChannelId, in.ChannelID[:])
	copy(out.Sig, in.Sig)
	return
}

func fromChannelUpdateRej(in *client.ChannelUpdateRejMsg) (out *ChannelUpdateRejMsg) {
	out = &ChannelUpdateRejMsg{
		ChannelId: make([]byte, len(in.ChannelID)),
		Version:   in.Version,
		Reason:    in.Reason,
	}
	copy(out.ChannelId, in.ChannelID[:])
	return
}

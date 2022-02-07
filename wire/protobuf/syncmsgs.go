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

func toChannelSync(in *ChannelSyncMsg) (*client.ChannelSyncMsg, error) {
	state, err := toState(in.CurrentTx.State)
	if err != nil {
		return nil, err
	}
	sigs := make([][]byte, len(in.CurrentTx.Sigs))
	for i := range in.CurrentTx.Sigs {
		sigs[i] = make([]byte, len(in.CurrentTx.Sigs[i]))
		copy(sigs[i], in.CurrentTx.Sigs[i])
	}
	out := &client.ChannelSyncMsg{
		Phase: channel.Phase(in.Phase),
		CurrentTX: channel.Transaction{
			State: state,
			Sigs:  sigs,
		},
	}
	return out, nil
}

func fromChannelSync(in *client.ChannelSyncMsg) (*ChannelSyncMsg, error) {
	state, err := fromState(in.CurrentTX.State)
	if err != nil {
		return nil, errors.WithMessage(err, "encoding state")
	}
	sigs := make([][]byte, len(in.CurrentTX.Sigs))
	for i := range in.CurrentTX.Sigs {
		sigs[i] = make([]byte, len(in.CurrentTX.Sigs[i]))
		copy(sigs[i], in.CurrentTX.Sigs[i])
	}
	out := &ChannelSyncMsg{
		Phase: uint32(in.Phase),
		CurrentTx: &Transaction{
			State: state,
			Sigs:  sigs,
		},
	}
	return out, nil
}

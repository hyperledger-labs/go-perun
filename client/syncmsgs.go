// Copyright 2020 - See NOTICE file for copyright holders.
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

package client

import (
	"io"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/wire"
	"perun.network/go-perun/wire/perunio"
)

func init() {
	wire.RegisterDecoder(wire.ChannelSync,
		func(r io.Reader) (wire.Msg, error) {
			var m msgChannelSync
			return &m, m.Decode(r)
		})
}

type msgChannelSync struct {
	Phase     channel.Phase       // Phase is the phase of the sender.
	CurrentTX channel.Transaction // CurrentTX is the sender's current transaction.
}

var _ ChannelMsg = (*msgChannelSync)(nil)

func newChannelSyncMsg(s channel.Source) *msgChannelSync {
	return &msgChannelSync{
		Phase:     s.Phase(),
		CurrentTX: s.CurrentTX(),
	}
}

// Encode implements perunio.Encode.
func (m *msgChannelSync) Encode(w io.Writer) error {
	return perunio.Encode(w,
		m.Phase,
		m.CurrentTX)
}

// Decode implements perunio.Decode.
func (m *msgChannelSync) Decode(r io.Reader) error {
	return perunio.Decode(r,
		&m.Phase,
		&m.CurrentTX)
}

// ID returns the channel's ID.
func (m *msgChannelSync) ID() channel.ID {
	return m.CurrentTX.ID
}

// Type implements wire.Type.
func (m *msgChannelSync) Type() wire.Type {
	return wire.ChannelSync
}

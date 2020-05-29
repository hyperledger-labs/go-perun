// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package client

import (
	"io"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/wire"
	"perun.network/go-perun/wire/msg"
)

func init() {
	msg.RegisterDecoder(msg.ChannelSyncReq,
		func(r io.Reader) (msg.Msg, error) {
			var m msgChannelSyncReq
			return &m, m.Decode(r)
		})
	msg.RegisterDecoder(msg.ChannelSyncRes,
		func(r io.Reader) (msg.Msg, error) {
			var m msgChannelSyncRes
			return &m, m.Decode(r)
		})
}

type msgChannelSync struct {
	ChannelID channel.ID          // ChannelID is the channel ID.
	Phase     channel.Phase       // Phase is the phase of the sender.
	CurrentTX channel.Transaction // CurrentTX is the sender's current transaction.
	StagingTX channel.Transaction // StagingTX is the sender's staging transaction.
}

var _ ChannelMsg = (*msgChannelSyncReq)(nil)
var _ ChannelMsg = (*msgChannelSyncRes)(nil)

// Encode implements msg.Encode.
func (m *msgChannelSync) Encode(w io.Writer) error {
	return wire.Encode(w,
		m.ChannelID,
		m.Phase,
		m.CurrentTX,
		m.StagingTX)
}

// Decode implements msg.Decode.
func (m *msgChannelSync) Decode(r io.Reader) error {
	return wire.Decode(r,
		&m.ChannelID,
		&m.Phase,
		&m.CurrentTX,
		&m.StagingTX)
}

// ID returns the channel's ID.
func (m *msgChannelSync) ID() channel.ID {
	return m.ChannelID
}

type msgChannelSyncReq struct{ msgChannelSync }
type msgChannelSyncRes struct{ msgChannelSync }

// Type implements msg.Type.
func (m *msgChannelSyncReq) Type() msg.Type {
	return msg.ChannelSyncReq
}

// Type implements msg.Type.
func (m *msgChannelSyncRes) Type() msg.Type {
	return msg.ChannelSyncRes
}

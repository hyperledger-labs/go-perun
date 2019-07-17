// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package msg

import (
	"io"

	"perun.network/go-perun/wire"
)

// PingMsg is a ping request.
// It contains the time at which it was sent, so that the recipient can also
// measure the time it took to transmit the ping request.
type PingMsg struct {
	controlMsg
	Created wire.Time
}

func (m *PingMsg) Type() ControlMsgType {
	return Ping
}

func (m *PingMsg) encode(writer io.Writer) error {
	return wire.Encode(writer, &m.Created)
}

func (m *PingMsg) decode(reader io.Reader) error {
	return wire.Decode(reader, &m.Created)
}

// PongMsg is the response to a ping message.
// It contains the time at which it was sent, so that the recipient knows how
// long the ping request took to be transmitted, and how quickly the response
// was sent.
type PongMsg struct {
	controlMsg
	Created wire.Time
}

func (m *PongMsg) Type() ControlMsgType {
	return Pong
}

func (m *PongMsg) encode(writer io.Writer) error {
	return wire.Encode(writer, &m.Created)
}

func (m *PongMsg) decode(reader io.Reader) error {
	return wire.Decode(reader, &m.Created)
}

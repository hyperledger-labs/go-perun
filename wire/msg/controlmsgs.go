// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package msg

import (
	"io"
	"time"

	"perun.network/go-perun/wire"
)

// Since ping and pong messages are essentially the same, this is a common
// implementation for both.
type pingPongMsg struct {
	controlMsg
	Created time.Time
}

func (m pingPongMsg) encode(writer io.Writer) error {
	return wire.Encode(writer, &m.Created)
}

func (m *pingPongMsg) decode(reader io.Reader) error {
	return wire.Decode(reader, &m.Created)
}

// PingMsg is a ping request.
// It contains the time at which it was sent, so that the recipient can also
// measure the time it took to transmit the ping request.
type PingMsg struct {
	pingPongMsg
}

func (m *PingMsg) Type() ControlMsgType {
	return Ping
}

// PongMsg is the response to a ping message.
// It contains the time at which it was sent, so that the recipient knows how
// long the ping request took to be transmitted, and how quickly the response
// was sent.
type PongMsg struct {
	pingPongMsg
}

func (m *PongMsg) Type() ControlMsgType {
	return Pong
}

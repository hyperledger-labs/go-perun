// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package msg

import (
	"io"

	"perun.network/go-perun/wire"
)

// DummyChannelMsg is a dummy message type used for testing.
type DummyChannelMsg struct {
	channelMsg
	dummy int64
}

func (DummyChannelMsg) Type() ChannelMsgType {
	return ChannelDummy
}

func (m DummyChannelMsg) encode(writer io.Writer) error {
	return wire.Encode(writer, m.dummy)
}

func (m *DummyChannelMsg) decode(reader io.Reader) error {
	return wire.Decode(reader, &m.dummy)
}

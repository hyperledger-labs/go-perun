// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package msg

import (
	"io"
	"testing"

	"perun.network/go-perun/pkg/io/test"
)

type serializableMsg struct {
	Msg Msg
}

func (msg *serializableMsg) Encode(writer io.Writer) error {
	return Encode(msg.Msg, writer)
}

func (msg *serializableMsg) Decode(reader io.Reader) error {
	var err error
	msg.Msg, err = Decode(reader)
	return err
}

func testMsg(t *testing.T, msg Msg) {
	test.GenericSerializableTest(t, &serializableMsg{msg})
}

func TestPingMsg(t *testing.T) {
	testMsg(t, NewPingMsg())
}

func TestPongMsg(t *testing.T) {
	testMsg(t, NewPongMsg())
}

func newChannelMsg() (m channelMsg) {
	// Set a non-0 channel ID, so that we can detect serializing errors.
	for i := range m.channelID {
		m.channelID[i] = byte(i)
	}
	return
}

func TestDummyChannelMsg(t *testing.T) {
	testMsg(t, &DummyChannelMsg{newChannelMsg(), int64(-0x7172635445362718)})
}

func TestDummyPeerMsg(t *testing.T) {
	testMsg(t, &DummyPeerMsg{peerMsg{}, int64(-0x7172635445362718)})
}

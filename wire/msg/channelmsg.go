// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package msg

import (
	"io"
	"strconv"

	"github.com/pkg/errors"
)

// ChannelMsg objects are channel-specific messages that are sent between
// perun nodes.
type ChannelMsg interface {
	Msg
	// Connection returns the channel message's associated connection's ID.
	Channel() *ChannelID
	// Type returns the message's implementing type.
	Type() ChannelMsgType
}

func decodeChannelMsg(reader io.Reader) (ChannelMsg, error) {
	var Type ChannelMsgType
	if err := Type.Decode(reader); err != nil {
		return nil, errors.Wrap(err, "failed to read the message type")
	}

	var conn ChannelID
	if err := conn.Decode(reader); err != nil {
		return nil, errors.Wrap(err, "failed to read the channel ID")
	}

	switch Type {
	default:
		panic("decodeChannelMsg(): Unhandled channel message type: " + Type.String())
	}
}

func encodeChannelMsg(msg ChannelMsg, writer io.Writer) error {
	if err := msg.Type().Encode(writer); err != nil {
		return errors.Wrap(err, "failed to write the message type")
	}

	if err := msg.Channel().Encode(writer); err != nil {
		return errors.Wrap(err, "failed to write the channel id")
	}

	if err := msg.encode(writer); err != nil {
		return errors.Wrap(err, "failed to write the message contents")
	}

	return nil
}

// channelMsg allows default-implementing the Category function in channel
// messages.
type channelMsg struct {
	channelID ChannelID
}

func (m *channelMsg) Channel() *ChannelID {
	return &m.channelID
}

func (*channelMsg) Category() Category {
	return Channel
}

// ChannelID uniquely identifies a virtual connection to a peer node.
type ChannelID [32]byte

func (id *ChannelID) Encode(writer io.Writer) error {
	if _, err := writer.Write(id[:]); err != nil {
		return errors.Wrap(err, "failed to write channel id")
	}
	return nil
}
func (id *ChannelID) Decode(reader io.Reader) error {
	if _, err := reader.Read(id[:]); err != nil {
		return errors.Wrap(err, "failed to read channel id")
	}
	return nil
}

// ChannelMsgType is an enumeration used for (de)serializing channel messages
// and identifying a channel message's type.
//
// When changing this type, also change encode() and decode().
type ChannelMsgType uint8

// Enumeration of channel message types.
const (
	// A dummy message, replace with real message types.
	Dummy ChannelMsgType = iota
	lastChannelMsgType
)

func (t ChannelMsgType) String() string {
	if !t.Valid() {
		return strconv.Itoa(int(t))
	}
	return []string{
		"DummyMsg",
	}[t]
}

// Valid checks whether a ChannelMsgType is a valid value.
func (t ChannelMsgType) Valid() bool {
	return t < lastChannelMsgType
}

func (t ChannelMsgType) Encode(writer io.Writer) error {
	if _, err := writer.Write([]byte{byte(t)}); err != nil {
		return errors.Wrap(err, "failed to write channel message type")
	}
	return nil
}

func (t *ChannelMsgType) Decode(reader io.Reader) error {
	buf := [1]byte{}
	if _, err := reader.Read(buf[:]); err != nil {
		return errors.Wrap(err, "failed to read channel message type")
	}
	*t = ChannelMsgType(buf[0])
	if !t.Valid() {
		return errors.New("invalid channel message type encoding: " + t.String())
	}
	return nil
}

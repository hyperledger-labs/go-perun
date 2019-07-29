// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package msg

import (
	"strconv"

	"perun.network/go-perun/log"
	"perun.network/go-perun/pkg/io"

	"github.com/pkg/errors"
)

// ChannelMsg objects are channel-specific messages that are sent between
// Perun nodes.
type ChannelMsg interface {
	Msg
	// Connection returns the channel message's associated channel's ID.
	ChannelID() ChannelID
	// Type returns the message's implementing type.
	Type() ChannelMsgType
}

func decodeChannelMsg(reader io.Reader) (msg ChannelMsg, err error) {
	var Type ChannelMsgType
	if err := Type.Decode(reader); err != nil {
		return nil, errors.WithMessage(err, "failed to read the message type")
	}

	var m channelMsg
	if err := m.channelID.Decode(reader); err != nil {
		return nil, errors.WithMessage(err, "failed to read the channel ID")
	}

	// Type is guaranteed to be valid at this point.
	// This switch handles all channel message types, but if any was forgotten,
	// the program panics.
	switch Type {
	case ChannelDummy:
		msg = &DummyChannelMsg{channelMsg: m}
	default:
		log.Panicf("decodeChannelMsg(): Unhandled channel message type: %v", Type)
	}

	if err := msg.decode(reader); err != nil {
		return nil, errors.WithMessagef(err, "failed to decode %v", Type)
	}
	return msg, nil
}

func encodeChannelMsg(msg ChannelMsg, writer io.Writer) error {
	if err := msg.Type().Encode(writer); err != nil {
		return errors.WithMessage(err, "failed to write the message type")
	}

	if err := msg.ChannelID().Encode(writer); err != nil {
		return errors.WithMessage(err, "failed to write the channel id")
	}

	if err := msg.encode(writer); err != nil {
		return errors.WithMessage(err, "failed to write the message contents")
	}

	return nil
}

// channelMsg allows default-implementing the Category(), Channel() functions
// in channel messages.
//
// Example:
// 	type SomeChannelMsg struct {
//  	channelMsg
//  }
type channelMsg struct {
	channelID ChannelID
}

func (m *channelMsg) ChannelID() ChannelID {
	return m.channelID
}

func (*channelMsg) Category() Category {
	return Channel
}

// ChannelID uniquely identifies a virtual connection to a peer node.
type ChannelID [32]byte

func (id ChannelID) Encode(writer io.Writer) error {
	if _, err := writer.Write(id[:]); err != nil {
		return errors.Wrap(err, "failed to write channel id")
	}
	return nil
}

func (id *ChannelID) Decode(reader io.Reader) error {
	if err := io.ReadAll(reader, id[:]); err != nil {
		return errors.WithMessage(err, "failed to read channel id")
	}
	return nil
}

// ChannelMsgType is an enumeration used for (de)serializing channel messages
// and identifying a channel message's type.
//
// When changing this type, also change Encode() and Decode().
type ChannelMsgType uint8

// Enumeration of channel message types.
const (
	// This is a dummy peer message. It is used for testing purposes until the
	// first actual channel message type is added.
	ChannelDummy ChannelMsgType = iota

	// This constant marks the first invalid enum value.
	channelMsgTypeEnd
)

// String returns the name of a channel message type, if it is valid, otherwise,
// returns its numerical representation for debugging purposes.
func (t ChannelMsgType) String() string {
	if !t.Valid() {
		return strconv.Itoa(int(t))
	}
	return [...]string{
		"DummyChannelMsg",
	}[t]
}

// Valid checks whether a ChannelMsgType is a valid value.
func (t ChannelMsgType) Valid() bool {
	return t < channelMsgTypeEnd
}

func (t ChannelMsgType) Encode(writer io.Writer) error {
	if _, err := writer.Write([]byte{byte(t)}); err != nil {
		return errors.Wrap(err, "failed to write channel message type")
	}
	return nil
}

func (t *ChannelMsgType) Decode(reader io.Reader) error {
	buf := make([]byte, 1)
	if err := io.ReadAll(reader, buf); err != nil {
		return errors.WithMessage(err, "failed to read channel message type")
	}
	*t = ChannelMsgType(buf[0])
	if !t.Valid() {
		return errors.New("invalid channel message type encoding: " + t.String())
	}
	return nil
}

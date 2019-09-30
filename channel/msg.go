// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel

import (
	"io"
	"strconv"

	"github.com/pkg/errors"

	"perun.network/go-perun/log"
	wire "perun.network/go-perun/wire/msg"
)

// Msg objects are channel-specific messages that are sent between
// Perun nodes.
type Msg interface {
	wire.Msg
	// Connection returns the channel message's associated channel's ID.
	ChannelID() ID
	// Type returns the message's implementing type.
	Type() MsgType

	// encode encodes the message's contents (without headers or type info).
	encode(io.Writer) error
	// decode decodes the message's contents (without headers or type info).
	decode(io.Reader) error
}

func decodeMsg(reader io.Reader) (wire.Msg, error) {
	var Type MsgType
	if err := Type.Decode(reader); err != nil {
		return nil, errors.WithMessage(err, "failed to read the message type")
	}

	var baseMsg msg
	if _, err := io.ReadFull(reader, baseMsg.channelID[:]); err != nil {
		return nil, errors.WithMessage(err, "failed to read the channel ID")
	}

	var message Msg
	// Type is guaranteed to be valid at this point.
	// This switch handles all channel message types, but if any was forgotten,
	// the program panics.
	switch Type {
	case ChannelDummy:
		message = &DummyChannelMsg{msg: baseMsg}
	default:
		log.Panicf("decodeMsg(): Unhandled channel message type: %v", Type)
	}

	if err := message.decode(reader); err != nil {
		return nil, errors.WithMessagef(err, "failed to decode %v", Type)
	}
	return message, nil
}

func encodeMsg(writer io.Writer, message wire.Msg) error {
	var cmsg = message.(Msg)
	if err := cmsg.Type().Encode(writer); err != nil {
		return errors.WithMessage(err, "failed to write the message type")
	}

	id := cmsg.ChannelID()
	if _, err := writer.Write(id[:]); err != nil {
		return errors.WithMessage(err, "failed to write the channel id")
	}

	if err := cmsg.encode(writer); err != nil {
		return errors.WithMessage(err, "failed to write the message contents")
	}

	return nil
}

// msg allows default-implementing the Category(), Channel() functions
// in channel messages.
//
// Example:
// 	type SomeChannelMsg struct {
//  	msg
//  }
type msg struct {
	channelID ID
}

func (m *msg) ChannelID() ID {
	return m.channelID
}

func (*msg) Category() wire.Category {
	return wire.Channel
}

// MsgType is an enumeration used for (de)serializing channel messages
// and identifying a channel message's type.
//
// When changing this type, also change Encode() and Decode().
type MsgType uint8

// Enumeration of channel message types.
const (
	// This is a dummy peer message. It is used for testing purposes until the
	// first actual channel message type is added.
	ChannelDummy MsgType = iota

	// This constant marks the first invalid enum value.
	msgTypeEnd
)

// String returns the name of a channel message type, if it is valid, otherwise,
// returns its numerical representation for debugging purposes.
func (t MsgType) String() string {
	if !t.Valid() {
		return strconv.Itoa(int(t))
	}
	return [...]string{
		"DummyChannelMsg",
	}[t]
}

// Valid checks whether a MsgType is a valid value.
func (t MsgType) Valid() bool {
	return t < msgTypeEnd
}

func (t MsgType) Encode(writer io.Writer) error {
	if _, err := writer.Write([]byte{byte(t)}); err != nil {
		return errors.Wrap(err, "failed to write channel message type")
	}
	return nil
}

func (t *MsgType) Decode(reader io.Reader) error {
	buf := make([]byte, 1)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return errors.WithMessage(err, "failed to read channel message type")
	}
	*t = MsgType(buf[0])
	if !t.Valid() {
		return errors.New("invalid channel message type encoding: " + t.String())
	}
	return nil
}

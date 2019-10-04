// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package peer

import (
	"io"
	"strconv"

	"github.com/pkg/errors"

	wire "perun.network/go-perun/wire/msg"
)

// Msg objects are messages that are sent between peers, but do not belong
// to a specific state channel, such as channel creation requests.
type Msg interface {
	wire.Msg
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
		return nil, errors.WithMessage(err, "failed to decode message type")
	}

	var message Msg
	switch Type {
	case PeerDummy:
		message = &DummyPeerMsg{}
	default:
		return nil, errors.Errorf("unknown peer message type: %v", Type)
	}

	return message, message.decode(reader)
}

func encodeMsg(writer io.Writer, message wire.Msg) error {
	var pmsg = message.(Msg)
	if err := pmsg.Type().Encode(writer); err != nil {
		return errors.WithMessage(err, "failed to encode message type")
	}

	if err := pmsg.encode(writer); err != nil {
		return errors.WithMessage(err, "failed to encode peer message content")
	}

	return nil
}

// msg allows default-implementing the Category function in peer messages.
type msg struct {
	// Currently empty, until we know what peer messages actually look like.
}

func (*msg) Category() wire.Category {
	return wire.Peer
}

// PeerMsgType is an enumeration used for (de)serializing channel messages
// and identifying a channel message's type.
//
// When changing this type, also change Encode() and Decode().
type MsgType uint8

// Enumeration of channel message types.
const (
	// This is a dummy peer message. It is used for testing purposes until the
	// first actual peer message type is added.
	PeerDummy MsgType = iota
	PeerProposal

	// This constant marks the first invalid enum value.
	msgTypeEnd
)

// String returns the name of a peer message type, if it is valid, otherwise,
// returns its numerical representation for debugging purposes.
func (t MsgType) String() string {
	if !t.Valid() {
		return strconv.Itoa(int(t))
	}
	return [...]string{
		"DummyPeerMsg",
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
	return nil
}

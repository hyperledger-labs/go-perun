// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package msg

import (
	"fmt"
	"io"
	"strconv"

	"github.com/pkg/errors"
)

func init() {
	RegisterEncoder(Control, encodeControlMsg)
	RegisterDecoder(Control, decodeControlMsg)
}

// ControlMsg objects are messages that are outside of the perun core protcol
// that directly control what happens with a peer connection.
type ControlMsg interface {
	Msg
	// Type returns the control message's type.
	Type() ControlMsgType

	// encode encodes the message's contents (without headers or type).
	encode(io.Writer) error
	// decode decodes the message's contents (without headers or type).
	decode(io.Reader) error
}

// encodeControlMsg is a helper function that encodes a control message header.
func encodeControlMsg(writer io.Writer, _msg Msg) error {
	msg := _msg.(ControlMsg) // should panic if wrong message type is passed
	Type := msg.Type()
	if !Type.Valid() {
		panic(fmt.Sprintf("invalid control message type: %v", Type))
	}
	if err := Type.Encode(writer); err != nil {
		return errors.WithMessage(err, "error encoding control message type")
	}

	return msg.encode(writer)
}

// decodeControlMsg decodes a ControlMsg from a reader.
func decodeControlMsg(reader io.Reader) (Msg, error) {
	var Type ControlMsgType
	if err := Type.Decode(reader); err != nil {
		return nil, errors.WithMessage(err, "error decoding message type")
	}

	var msg ControlMsg
	// Decode message payload.
	switch Type {
	case Ping:
		msg = &PingMsg{}
	case Pong:
		msg = &PongMsg{}
	default:
		return nil, errors.Errorf("invalid control message type %v", Type)
	}

	return msg, msg.decode(reader)
}

// controlMsg allows default-implementing the Category() function in control
// messages.
//
// Example:
// 	type SomeControlMsg struct {
//  	controlMsg
//  }
type controlMsg struct{}

func (m *controlMsg) Category() Category {
	return Control
}

// ControlMsgType is an enumeration used for (de)serializing control messages
// and identifying a control message's type.
type ControlMsgType uint8

// Enumeration of control message types.
const (
	Ping ControlMsgType = iota
	Pong

	// This constant marks the first invalid enum value.
	controlMsgTypeEnd
)

// String returns the name of a peer message type, if it is valid, otherwise,
// returns its numerical representation for debugging purposes.
func (t ControlMsgType) String() string {
	if !t.Valid() {
		return strconv.Itoa(int(t))
	}
	return [...]string{
		"PingMsg",
		"PongMsg",
	}[t]
}

// Valid checks whether a ControlMsgType is a valid value.
func (t ControlMsgType) Valid() bool {
	return t < controlMsgTypeEnd
}

func (t ControlMsgType) Encode(writer io.Writer) error {
	_, err := writer.Write([]byte{byte(t)})
	return errors.Wrap(err, "failed to write control message type")
}

func (t *ControlMsgType) Decode(reader io.Reader) error {
	buf := make([]byte, 1)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return errors.Wrap(err, "failed to read control message type")
	}
	*t = ControlMsgType(buf[0])
	return nil
}

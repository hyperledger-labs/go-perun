// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package msg

import (
	"strconv"

	"perun.network/go-perun/pkg/io"

	"github.com/pkg/errors"
)

// ControlMsg objects are messages that are outside of the perun core protcol
// that directly control what happens with a peer connection.
type ControlMsg interface {
	Msg
	// Type returns the control message's implementing type.
	Type() ControlMsgType
}

// encodeControlMsg is a helper function that encodes a control message header.
func encodeControlMsg(msg ControlMsg, writer io.Writer) error {
	if err := msg.Type().Encode(writer); err != nil {
		return err
	}

	if err := msg.encode(writer); err != nil {
		return errors.WithMessage(err, "failed to encode message payload.")
	}

	return nil
}

// decodeControlMsg decodes a ControlMsg from a reader.
func decodeControlMsg(reader io.Reader) (ControlMsg, error) {
	var Type ControlMsgType
	if err := Type.Decode(reader); err != nil {
		return nil, errors.WithMessage(err, "failed to read the message type")
	}

	var msg ControlMsg
	// Decode message payload.
	switch Type {
	case Ping:
		msg = &PingMsg{}
	case Pong:
		msg = &PongMsg{}
	default:
		panic("decodeControlMsg(): Unhandled control message type: " + Type.String())
	}

	if err := msg.decode(reader); err != nil {
		return nil, errors.WithMessagef(err, "failed to decode %v", Type)
	}
	return msg, nil
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
	controlMsgTypeEnd
)

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
	if _, err := writer.Write([]byte{byte(t)}); err != nil {
		return errors.Wrap(err, "failed to write control message type")
	}
	return nil
}

func (t *ControlMsgType) Decode(reader io.Reader) error {
	buf := [1]byte{}
	if _, err := reader.Read(buf[:]); err != nil {
		return errors.Wrap(err, "failed to read control message type")
	}
	*t = ControlMsgType(buf[0])
	if !t.Valid() {
		return errors.New("invalid control message type encoding: " + t.String())
	}
	return nil
}

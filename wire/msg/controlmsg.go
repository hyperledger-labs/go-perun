// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package msg

import (
	"io"
	"strconv"

	"github.com/pkg/errors"

	"perun.network/go-perun/log"
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

	if controlDecodeFuns[Type] == nil {
		log.Panicf("decodeControlMsg(): Unhandled control message type: %v", Type)
	}
	return controlDecodeFuns[Type](reader)
}

var controlDecodeFuns map[ControlMsgType]func(io.Reader) (ControlMsg, error)

// RegisterControlDecode register the function that will decode all messages of category ControlMsg
func RegisterControlDecode(t ControlMsgType, fun func(io.Reader) (ControlMsg, error)) {
	if controlDecodeFuns[t] != nil || fun == nil {
		log.Panic("RegisterControlDecode called twice or with invalid argument")
	}

	controlDecodeFuns[t] = fun
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
	if !t.Valid() {
		return errors.New("invalid control message type encoding: " + t.String())
	}
	return nil
}

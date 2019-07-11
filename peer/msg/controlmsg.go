// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package msg

/*
	ControlMsgType is an enumeration used for (de)serializing control messages
	and identifying a control message's type.
*/
type ControlMsgType uint8

// Enumeration of control message types.
const (
	Ping ControlMsgType = iota
	Pong
	lastControlMsgType
)

func (t ControlMsgType) String() string {
	if !t.Valid() {
		return "<Invalid ControlMsgType (" + uint8(t) + ")>"
	}
	return []string{
		"PingMsg",
		"PongMsg",
	}[t]
}

// Valid checks whether a ControlMsgType is a valid value.
func (t ControlMsgType) Valid() bool {
	return t < lastControlMsgType
}

/*
	ControlMsg objects are messages that are outside of the perun core protcol
	that directly control what happens with a peer connection.
*/
type ControlMsg interface {
	Msg
	// Type returns the control message's implementing type.
	Type() ControlMsgType
}

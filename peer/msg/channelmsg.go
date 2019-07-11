// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package msg

/*
	ConnectionID uniquely identifies a virtual connection to a peer node.
*/
type ConnectionID [20]byte

/*
	ChannelMsgType is an enumeration used for (de)serializing channel messages
	and identifying a channel message's type.
*/
type ChannelMsgType uint16

// Enumeration of channel message types.
const (
	// A dummy message, replace with real message types.
	Dummy ChannelMsgType = iota
	lastChannelMsgType
)

func (t ChannelMsgType) String() string {
	if !t.Valid() {
		return "<Invalid ChannelMsgType (" + uint8(t) + ")>"
	}
	return []string{
		"DummyMsg",
	}[t]
}

// Valid checks whether a ChannelMsgType is a valid value.
func (t ChannelMsgType) Valid() bool {
	return t < lastChannelMsgType
}

/*
	ChannelMsg objects are channel-specific messages that are sent between
	perun nodes.
*/
type ChannelMsg interface {
	Msg
	// Connection returns the channel message's associated connection's ID.
	Connection() ConnectionID
	// Type returns the message's implementing type.
	Type() ChannelMsgType
}

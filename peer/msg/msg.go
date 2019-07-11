// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

// Package msg contains all message types, as well as serialising and
// deserialising logic used in peer communications.
package msg // import "perun.network/go-perun/peer/msg"

import (
	"perun.network/go-perun/peer"
)

// Msg is the top-level abstraction for all messages sent between perun
// nodes.
type Msg interface {
	peer.Serializable
	// Category returns the message's subcategory.
	Category() Category
}

// MsgCategory is an enumeration used for (de)serializing messages and
// identifying a message's subcategory.
type Category uint8

// Enumeration of message categories.
const (
	Channel Category = iota
	Control
	categoryEnd
)

func (c Category) String() string {
	return []string{
		"ChannelMsg",
		"ControlMsg",
	}[c]
}

// Valid checks whether a ControlMsgType is a valid value.
func (c Category) Valid() bool {
	return c < categoryEnd
}

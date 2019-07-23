// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

// Package msg contains all message types, as well as serialising and
// deserialising logic used in peer communications.
package msg // import "perun.network/go-perun/wire/msg"

import (
	"strconv"

	"perun.network/go-perun/pkg/io"

	"github.com/pkg/errors"
)

// Msg is the top-level abstraction for all messages sent between perun
// nodes.
type Msg interface {
	// Category returns the message's subcategory.
	Category() Category

	// encode encodes the message's contents (without headers or type info).
	encode(io.Writer) error
	// decode decodes the message's contents (without headers or type info).
	decode(io.Reader) error
}

// Encode encodes a message into an io.Writer.
func Encode(msg Msg, writer io.Writer) (err error) {
	// Encode the message category, then encode the message.
	if err = msg.Category().Encode(writer); err == nil {
		switch msg.Category() {
		case Control:
			cmsg, _ := msg.(ControlMsg)
			err = encodeControlMsg(cmsg, writer)
		case Channel:
			cmsg, _ := msg.(ChannelMsg)
			err = encodeChannelMsg(cmsg, writer)
		case Peer:
			pmsg, _ := msg.(PeerMsg)
			err = encodePeerMsg(pmsg, writer)
		default:
			panic("Encode(): Unhandled message category: " + msg.Category().String())
		}
	}

	// Handle both error sources at once.
	if err != nil {
		err = errors.WithMessagef(err, "failed to write message with category %v", msg.Category())
	}
	return
}

// Decode decodes a message from an io.Reader.
func Decode(reader io.Reader) (Msg, error) {
	var cat Category
	if err := cat.Decode(reader); err != nil {
		return nil, errors.WithMessage(err, "failed to decode message category")
	}

	switch cat {
	case Control:
		return decodeControlMsg(reader)
	case Channel:
		return decodeChannelMsg(reader)
	case Peer:
		return decodePeerMsg(reader)
	default:
		panic("Decode(): Unhandled message category: " + cat.String())
	}
}

// Category is an enumeration used for (de)serializing messages and
// identifying a message's subcategory.
type Category uint8

// Enumeration of message categories.
const (
	Control Category = iota
	Peer
	Channel

	// This constant marks the first invalid enum value.
	categoryEnd
)

// String returns the name of a message category, if it is valid, otherwise,
// returns its numerical representation for debugging purposes.
func (c Category) String() string {
	if !c.Valid() {
		return strconv.Itoa(int(c))
	}
	return [...]string{
		"ControlMsg",
		"PeerMsg",
		"ChannelMsg",
	}[c]
}

// Valid checks whether a Category is a valid value.
func (c Category) Valid() bool {
	return c < categoryEnd
}

func (c Category) Encode(writer io.Writer) error {
	if _, err := writer.Write([]byte{byte(c)}); err != nil {
		return errors.Wrap(err, "failed to write category")
	}
	return nil
}

func (c *Category) Decode(reader io.Reader) error {
	buf := make([]byte, 1)
	if err := io.ReadAll(reader, buf); err != nil {
		return errors.WithMessage(err, "failed to write category")
	}

	*c = Category(buf[0])
	if !c.Valid() {
		return errors.New("invalid message category encoding: " + c.String())
	}
	return nil
}

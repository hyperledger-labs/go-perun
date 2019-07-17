// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

// Package msg contains all message types, as well as serialising and
// deserialising logic used in peer communications.
package msg // import "perun.network/go-perun/wire/msg"

import (
	"io"
	"strconv"

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
func Encode(msg Msg, writer io.Writer) error {
	err := msg.Category().Encode(writer)
	if err == nil {
		switch msg.Category() {
		case Control:
			cmsg, _ := msg.(ControlMsg)
			err = encodeControlMsg(cmsg, writer)
		case Channel:
			cmsg, _ := msg.(ChannelMsg)
			err = encodeChannelMsg(cmsg, writer)
		default:
			panic("Encode(): Unhandled message category: " + msg.Category().String())
		}
	}

	if err != nil {
		return errors.Wrapf(err, "failed to write message (%v)", msg.Category())
	}

	return nil
}

// Decode decodes a message from an io.Reader.
func Decode(reader io.Reader) (msg Msg, err error) {
	var cat Category
	if err = cat.Decode(reader); err != nil {
		return nil, errors.Wrap(err, "failed to decode message category")
	}

	switch cat {
	case Control:
		msg, err = decodeControlMsg(reader)
	case Channel:
		msg, err = decodeChannelMsg(reader)
	default:
		panic("Decode(): Unhandled message category: " + cat.String())
	}

	return
}

// Category is an enumeration used for (de)serializing messages and
// identifying a message's subcategory.
type Category uint8

// Enumeration of message categories.
const (
	Channel Category = iota
	Control
	Peer
	categoryEnd
)

func (c Category) String() string {
	if !c.Valid() {
		return strconv.Itoa(int(uint8(c)))
	}
	return [...]string{
		"ChannelMsg",
		"ControlMsg",
		"PeerMsg",
	}[c]
}

// Valid checks whether a ControlMsgType is a valid value.
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
	buf := [1]byte{}
	_, err := reader.Read(buf[:])
	if err != nil {
		return errors.Wrap(err, "failed to write category")
	}

	*c = Category(buf[0])
	if !c.Valid() {
		return errors.New("invalid message category encoding: " + c.String())
	}
	return nil
}

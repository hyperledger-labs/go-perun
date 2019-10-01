// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

// Package msg contains all message types, as well as serialising and
// deserialising logic used in peer communications.
package msg // import "perun.network/go-perun/wire/msg"

import (
	"fmt"
	"io"
	"strconv"

	"github.com/pkg/errors"

	"perun.network/go-perun/log"
)

// Msg is the top-level abstraction for all messages sent between perun
// nodes.
type Msg interface {
	// Category returns the message's subcategory.
	Category() Category
}

// Encode encodes a message into an io.Writer.
func Encode(msg Msg, writer io.Writer) (err error) {
	// Encode the message category, then encode the message.
	cat := msg.Category()
	encoder, ok := encoders[cat]
	// we don't use Category.Valid() here because it might happen that a decoder,
	// but no encoder is set, because Valid() tests whether a decoder is set.
	if !ok {
		log.Panicf("no encoder registered for message category %v", cat)
	}
	if err = cat.Encode(writer); err != nil {
		return err
	}

	return encoder(writer, msg)
}

var encoders = make(map[Category]func(io.Writer, Msg) error)

// RegisterEncoder sets the encoder of messages of category `cat`.
func RegisterEncoder(cat Category, encoder func(io.Writer, Msg) error) {
	if encoder == nil {
		// encoder registration happens during init(), so we don't use log.Panic
		panic("wire: encoder nil")
	}
	if encoders[cat] != nil {
		panic(fmt.Sprintf("wire: encoder for category %v already set", cat))
	}

	encoders[cat] = encoder
}

// Decode decodes a message from an io.Reader.
func Decode(reader io.Reader) (Msg, error) {
	var cat Category
	if err := cat.Decode(reader); err != nil {
		return nil, errors.WithMessage(err, "failed to decode message category")
	}

	if !cat.Valid() {
		return nil, errors.Errorf("wire: invalid message category in Decode(): %v", cat)
	}
	return decoders[cat](reader)
}

var decoders = make(map[Category]func(io.Reader) (Msg, error))

// RegisterDecoder sets the decoder of messages of category `cat`.
func RegisterDecoder(cat Category, decoder func(io.Reader) (Msg, error)) {
	if decoder == nil {
		// decoder registration happens during init(), so we don't use log.Panic
		panic("wire: decoder nil")
	}
	if decoders[cat] != nil {
		panic(fmt.Sprintf("wire: decoder for category %v already set", cat))
	}

	decoders[cat] = decoder
}

// Category is an enumeration used for (de)serializing messages and
// identifying a message's category.
type Category uint8

// Enumeration of message categories known to the Perun framework.
const (
	Control Category = iota
	Peer
	Channel
)

// String returns the name of a message category if it is valid and name known
// or otherwise its numerical representation.
func (c Category) String() string {
	// Channel is currently the last known category to the framework
	if c > Channel {
		return strconv.Itoa(int(c))
	}
	return [...]string{
		"ControlMsg",
		"PeerMsg",
		"ChannelMsg",
	}[c]
}

// Valid checks whether a Category is a valid value, i.e., if a decoder is set.
func (c Category) Valid() bool {
	_, ok := decoders[c]
	return ok
}

func (c Category) Encode(writer io.Writer) error {
	if _, err := writer.Write([]byte{byte(c)}); err != nil {
		return errors.Wrap(err, "failed to write category")
	}
	return nil
}

func (c *Category) Decode(reader io.Reader) error {
	buf := make([]byte, 1)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return errors.Wrap(err, "failed to read category")
	}

	*c = Category(buf[0])
	return nil
}

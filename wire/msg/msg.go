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

	perunio "perun.network/go-perun/pkg/io"
	"perun.network/go-perun/wire"
)

// Msg is the top-level abstraction for all messages sent between perun
// nodes.
type Msg interface {
	// Type returns the message's type.
	Type() Type
	// encoding of payload. Type byte should not be encoded.
	perunio.Encoder
}

// Encode encodes a message into an io.Writer.
func Encode(msg Msg, w io.Writer) (err error) {
	// Encode the message type and payload
	return wire.Encode(w, byte(msg.Type()), msg)
}

// Decode decodes a message from an io.Reader.
func Decode(r io.Reader) (Msg, error) {
	var t Type
	if err := wire.Decode(r, (*byte)(&t)); err != nil {
		return nil, errors.WithMessage(err, "failed to decode message Type")
	}

	if !t.Valid() {
		return nil, errors.Errorf("wire: invalid message Type in Decode(): %v", t)
	}
	return decoders[t](r)
}

var decoders = make(map[Type]func(io.Reader) (Msg, error))

// RegisterDecoder sets the decoder of messages of Type `t`.
func RegisterDecoder(t Type, decoder func(io.Reader) (Msg, error)) {
	if decoder == nil {
		// decoder registration happens during init(), so we don't use log.Panic
		panic("wire: decoder nil")
	}
	if decoders[t] != nil {
		panic(fmt.Sprintf("wire: decoder for Type %v already set", t))
	}

	decoders[t] = decoder
}

// Type is an enumeration used for (de)serializing messages and
// identifying a message's Type.
type Type uint8

// Enumeration of message categories known to the Perun framework.
const (
	Ping Type = iota
	Pong
	ChannelProposal
	ChannelProposalRes
	AuthResponse
	msgTypeEnd // upper bound on the message type byte
)

// String returns the name of a message type if it is valid and name known
// or otherwise its numerical representation.
func (t Type) String() string {
	if t > msgTypeEnd {
		return strconv.Itoa(int(t))
	}
	return [...]string{
		"Ping",
		"Pong",
		"ChannelProposal",
		"ChannelProposalRes",
		"AuthResponse",
	}[t]
}

// Valid checks whether the type is known.
func (t Type) Valid() bool {
	return t < msgTypeEnd
}

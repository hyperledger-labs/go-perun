// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

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
		return nil, errors.Errorf("wire: no decoder known for message Type): %v", t)
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

// RegisterExternalDecoder sets the decoder of messages of external type `t`.
// This is like RegisterDecoder but for message types not part of the Perun wire
// protocol and thus not known natively. This can be used by users of the
// framework to create additional message types and send them over the same
// peer connection. It also comes in handy to register types for testing.
func RegisterExternalDecoder(t Type, decoder func(io.Reader) (Msg, error), name string) {
	if t < LastType {
		panic("external decoders can only be registered for alien types")
	}
	RegisterDecoder(t, decoder)
	// above registration panics if already set, so we don't need to check the
	// next assignment.
	typeNames[t] = name
}

// Type is an enumeration used for (de)serializing messages and
// identifying a message's Type.
type Type uint8

// Enumeration of message categories known to the Perun framework.
const (
	Ping Type = iota
	Pong
	AuthResponse
	ChannelProposal
	ChannelProposalAcc
	ChannelProposalRej
	ChannelUpdate
	ChannelUpdateAcc
	ChannelUpdateRej
	LastType // upper bound on the message types of the Perun wire protocol
)

var typeNames = map[Type]string{
	Ping:               "Ping",
	Pong:               "Pong",
	AuthResponse:       "AuthResponse",
	ChannelProposal:    "ChannelProposal",
	ChannelProposalAcc: "ChannelProposalAcc",
	ChannelProposalRej: "ChannelProposalRej",
	ChannelUpdate:      "ChannelUpdate",
	ChannelUpdateAcc:   "ChannelUpdateAcc",
	ChannelUpdateRej:   "ChannelUpdateRej",
}

// String returns the name of a message type if it is valid and name known
// or otherwise its numerical representation.
func (t Type) String() string {
	if name, ok := typeNames[t]; !ok {
		return strconv.Itoa(int(t))
	} else {
		return name
	}
}

// Valid checks whether a decoder is known for the type.
func (t Type) Valid() bool {
	_, ok := decoders[t]
	return ok
}

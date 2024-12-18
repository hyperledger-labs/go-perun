// Copyright 2019 - See NOTICE file for copyright holders.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package wire

import (
	"fmt"
	"io"
	"strconv"

	"perun.network/go-perun/wallet"

	"github.com/pkg/errors"

	"perun.network/go-perun/wire/perunio"
)

type (
	// Msg is the top-level abstraction for all messages sent between Perun
	// nodes.
	Msg interface {
		// Type returns the message's type.
		Type() Type
	}

	// An Envelope encapsulates a message with routing information, that is, the
	// sender and intended recipient.
	Envelope struct {
		Sender    map[wallet.BackendID]Address // Sender of the message.
		Recipient map[wallet.BackendID]Address // Recipient of the message.
		// Msg contained in this Envelope. Not embedded so Envelope doesn't implement Msg.
		Msg Msg
	}

	// EnvelopeSerializer serializes/deserializes envelopes into/from streams.
	EnvelopeSerializer interface {
		Encode(w io.Writer, env *Envelope) error
		Decode(r io.Reader) (*Envelope, error)
	}
)

// EncodeMsg encodes a message into an io.Writer. It also encodes the message
// type whereas the Msg.Encode implementation is assumed not to write the type.
func EncodeMsg(msg Msg, w io.Writer) error {
	// Encode the message type and payload
	return perunio.Encode(w, byte(msg.Type()), msg)
}

// DecodeMsg decodes a message from an io.Reader.
func DecodeMsg(r io.Reader) (Msg, error) {
	var t Type
	if err := perunio.Decode(r, (*byte)(&t)); err != nil {
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
	Shutdown
	AuthResponse
	LedgerChannelProposal
	LedgerChannelProposalAcc
	SubChannelProposal
	SubChannelProposalAcc
	VirtualChannelProposal
	VirtualChannelProposalAcc
	ChannelProposalRej
	ChannelUpdate
	VirtualChannelFundingProposal
	VirtualChannelSettlementProposal
	ChannelUpdateAcc
	ChannelUpdateRej
	ChannelSync
	LastType // upper bound on the message types of the Perun wire protocol
)

var typeNames = map[Type]string{
	Ping:                             "Ping",
	Pong:                             "Pong",
	Shutdown:                         "Shutdown",
	AuthResponse:                     "AuthResponse",
	LedgerChannelProposal:            "LedgerChannelProposal",
	LedgerChannelProposalAcc:         "LedgerChannelProposalAcc",
	SubChannelProposal:               "SubChannelProposal",
	SubChannelProposalAcc:            "SubChannelProposalAcc",
	VirtualChannelProposal:           "VirtualChannelProposal",
	VirtualChannelProposalAcc:        "VirtualChannelProposalAcc",
	ChannelProposalRej:               "ChannelProposalRej",
	ChannelUpdate:                    "ChannelUpdate",
	VirtualChannelFundingProposal:    "VirtualChannelFundingProposal",
	VirtualChannelSettlementProposal: "VirtualChannelSettlementProposal",
	ChannelUpdateAcc:                 "ChannelUpdateAcc",
	ChannelUpdateRej:                 "ChannelUpdateRej",
	ChannelSync:                      "ChannelSync",
}

// String returns the name of a message type if it is valid and name known
// or otherwise its numerical representation.
func (t Type) String() string {
	name, ok := typeNames[t]
	if !ok {
		return strconv.Itoa(int(t))
	}
	return name
}

// Valid checks whether a decoder is known for the type.
func (t Type) Valid() bool {
	_, ok := decoders[t]
	return ok
}

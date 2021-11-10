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

package client

import (
	"io"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wire"
	perunio "polycry.pt/poly-go/io"
)

func init() {
	wire.RegisterDecoder(wire.ChannelUpdate,
		func(r io.Reader) (wire.Msg, error) {
			var m msgChannelUpdate
			return &m, m.Decode(r)
		})
	wire.RegisterDecoder(wire.ChannelUpdateAcc,
		func(r io.Reader) (wire.Msg, error) {
			var m msgChannelUpdateAcc
			return &m, m.Decode(r)
		})
	wire.RegisterDecoder(wire.ChannelUpdateRej,
		func(r io.Reader) (wire.Msg, error) {
			var m msgChannelUpdateRej
			return &m, m.Decode(r)
		})
	wire.RegisterDecoder(wire.VirtualChannelFundingProposal,
		func(r io.Reader) (wire.Msg, error) {
			var m virtualChannelFundingProposal
			return &m, m.Decode(r)
		})
	wire.RegisterDecoder(wire.VirtualChannelSettlementProposal,
		func(r io.Reader) (wire.Msg, error) {
			var m virtualChannelSettlementProposal
			return &m, m.Decode(r)
		})
}

type (
	// ChannelMsg are all messages that can be routed to a particular channel
	// controller.
	ChannelMsg interface {
		wire.Msg
		ID() channel.ID
	}

	channelUpdateResMsg interface {
		ChannelMsg
		Ver() uint64
	}

	// msgChannelUpdate is the wire message of a channel update proposal. It
	// additionally holds the signature on the proposed state.
	msgChannelUpdate struct {
		ChannelUpdate
		// Sig is the signature on the proposed state by the peer sending the
		// ChannelUpdate.
		Sig wallet.Sig
	}

	// ChannelUpdateProposal represents an abstract update proposal message.
	ChannelUpdateProposal interface {
		wire.Msg
		perunio.Decoder
		Base() *msgChannelUpdate
	}

	// msgChannelUpdateAcc is the wire message sent as a positive reply to a
	// ChannelUpdate.  It references the channel ID and version and contains the
	// signature on the accepted new state by the sender.
	msgChannelUpdateAcc struct {
		// ChannelID is the channel ID.
		ChannelID channel.ID
		// Version of the state that is accepted.
		Version uint64
		// Sig is the signature on the proposed new state by the sender.
		Sig wallet.Sig
	}

	// msgChannelUpdateRej is the wire message sent as a negative reply to a
	// ChannelUpdate.  It references the channel ID and version and states a
	// reason for the rejection.
	msgChannelUpdateRej struct {
		// ChannelID is the channel ID.
		ChannelID channel.ID
		// Version of the state that is accepted.
		Version uint64
		// Reason states why the sender rejectes the proposed new state.
		Reason string
	}
)

var (
	_ ChannelMsg          = (*msgChannelUpdate)(nil)
	_ channelUpdateResMsg = (*msgChannelUpdateAcc)(nil)
	_ channelUpdateResMsg = (*msgChannelUpdateRej)(nil)
)

// Type returns this message's type: ChannelUpdate.
func (*msgChannelUpdate) Type() wire.Type {
	return wire.ChannelUpdate
}

// Type returns this message's type: ChannelUpdateAcc.
func (*msgChannelUpdateAcc) Type() wire.Type {
	return wire.ChannelUpdateAcc
}

// Type returns this message's type: ChannelUpdateRej.
func (*msgChannelUpdateRej) Type() wire.Type {
	return wire.ChannelUpdateRej
}

// Base returns the core channel update message.
func (c *msgChannelUpdate) Base() *msgChannelUpdate {
	return c
}

func (c msgChannelUpdate) Encode(w io.Writer) error {
	return perunio.Encode(w, c.State, c.ActorIdx, c.Sig)
}

func (c *msgChannelUpdate) Decode(r io.Reader) (err error) {
	if c.State == nil {
		c.State = new(channel.State)
	}
	if err := perunio.Decode(r, c.State, &c.ActorIdx); err != nil {
		return err
	}
	c.Sig, err = wallet.DecodeSig(r)
	return err
}

func (c msgChannelUpdateAcc) Encode(w io.Writer) error {
	return perunio.Encode(w, c.ChannelID, c.Version, c.Sig)
}

func (c *msgChannelUpdateAcc) Decode(r io.Reader) (err error) {
	if err := perunio.Decode(r, &c.ChannelID, &c.Version); err != nil {
		return err
	}
	c.Sig, err = wallet.DecodeSig(r)
	return err
}

func (c msgChannelUpdateRej) Encode(w io.Writer) error {
	return perunio.Encode(w, c.ChannelID, c.Version, c.Reason)
}

func (c *msgChannelUpdateRej) Decode(r io.Reader) (err error) {
	return perunio.Decode(r, &c.ChannelID, &c.Version, &c.Reason)
}

// ID returns the id of the channel this update refers to.
func (c *msgChannelUpdate) ID() channel.ID {
	return c.State.ID
}

// ID returns the id of the channel this update acceptance refers to.
func (c *msgChannelUpdateAcc) ID() channel.ID {
	return c.ChannelID
}

// ID returns the id of the channel this update rejection refers to.
func (c *msgChannelUpdateRej) ID() channel.ID {
	return c.ChannelID
}

// Ver returns the version of the state this update acceptance refers to.
func (c *msgChannelUpdateAcc) Ver() uint64 {
	return c.Version
}

// Ver returns the version of the state this update rejection refers to.
func (c *msgChannelUpdateRej) Ver() uint64 {
	return c.Version
}

/*
Virtual channel
*/

type (
	// virtualChannelFundingProposal is a channel update that proposes the funding of a virtual channel.
	virtualChannelFundingProposal struct {
		msgChannelUpdate
		Initial  channel.SignedState
		IndexMap []channel.Index
	}

	// virtualChannelSettlementProposal is a channel update that proposes the settlement of a virtual channel.
	virtualChannelSettlementProposal struct {
		msgChannelUpdate
		Final channel.SignedState
	}
)

// Type returns the message type.
func (*virtualChannelFundingProposal) Type() wire.Type {
	return wire.VirtualChannelFundingProposal
}

func (m virtualChannelFundingProposal) Encode(w io.Writer) (err error) {
	err = perunio.Encode(w,
		m.msgChannelUpdate,
		m.Initial.Params,
		*m.Initial.State,
		indexMapWithLen(m.IndexMap),
	)
	if err != nil {
		return
	}

	return wallet.EncodeSparseSigs(w, m.Initial.Sigs)
}

func (m *virtualChannelFundingProposal) Decode(r io.Reader) (err error) {
	m.Initial = channel.SignedState{
		Params: &channel.Params{},
		State:  &channel.State{},
	}
	err = perunio.Decode(r,
		&m.msgChannelUpdate,
		m.Initial.Params,
		m.Initial.State,
		(*indexMapWithLen)(&m.IndexMap),
	)
	if err != nil {
		return
	}

	m.Initial.Sigs = make([]wallet.Sig, m.Initial.State.NumParts())
	return wallet.DecodeSparseSigs(r, &m.Initial.Sigs)
}

// Type returns the message type.
func (*virtualChannelSettlementProposal) Type() wire.Type {
	return wire.VirtualChannelSettlementProposal
}

func (m virtualChannelSettlementProposal) Encode(w io.Writer) (err error) {
	err = perunio.Encode(w,
		m.msgChannelUpdate,
		m.Final.Params,
		*m.Final.State,
	)
	if err != nil {
		return
	}

	return wallet.EncodeSparseSigs(w, m.Final.Sigs)
}

func (m *virtualChannelSettlementProposal) Decode(r io.Reader) (err error) {
	m.Final = channel.SignedState{
		Params: &channel.Params{},
		State:  &channel.State{},
	}
	err = perunio.Decode(r,
		&m.msgChannelUpdate,
		m.Final.Params,
		m.Final.State,
	)
	if err != nil {
		return
	}

	m.Final.Sigs = make([]wallet.Sig, m.Final.State.NumParts())
	return wallet.DecodeSparseSigs(r, &m.Final.Sigs)
}

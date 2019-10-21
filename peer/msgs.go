// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package peer

import (
	"io"
	"math"
	"math/big"

	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
	perunio "perun.network/go-perun/pkg/io"
	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wire"
	wiremsg "perun.network/go-perun/wire/msg"
)

// DummyPeerMsg is a dummy message type used for testing.
type DummyPeerMsg struct {
	msg
	dummy int64
}

func (DummyPeerMsg) Type() MsgType {
	return PeerDummy
}

func (DummyPeerMsg) Category() wiremsg.Category {
	return wiremsg.Peer
}

func (m DummyPeerMsg) encode(writer io.Writer) error {
	return wire.Encode(writer, m.dummy)
}

func (m *DummyPeerMsg) decode(reader io.Reader) error {
	return wire.Decode(reader, &m.dummy)
}

// ChannelProposal contains all data necessary to propose a new
// channel to a given set of peers.
//
// The type implements the channel proposal messages from the Multi-Party
// Channel Proposal Protocol (MPCPP).
type ChannelProposal struct {
	ChallengeDuration uint64
	Nonce             *big.Int
	ParticipantAddr   Address
	AppDef            Address
	InitData          channel.Data
	InitBals          *channel.Allocation
	Parts             []Address
}

func (ChannelProposal) Category() wiremsg.Category {
	return wiremsg.Peer
}

func (*ChannelProposal) Type() MsgType {
	return PeerChannelProposal
}

func (c ChannelProposal) encode(w io.Writer) error {
	if err := wire.Encode(w, c.ChallengeDuration, c.Nonce); err != nil {
		return err
	}

	if err := perunio.Encode(w, c.ParticipantAddr, c.AppDef, c.InitData, c.InitBals); err != nil {
		return err
	}

	if len(c.Parts) > math.MaxInt32 {
		return errors.Errorf(
			"expected maximum number of participants %d, got %d",
			math.MaxInt32, len(c.Parts))
	}

	numParts := int32(len(c.Parts))
	if err := wire.Encode(w, numParts); err != nil {
		return err
	}
	for i := range c.Parts {
		if err := c.Parts[i].Encode(w); err != nil {
			return errors.Errorf(
				"error encoding participant %d", i)
		}
	}

	return nil
}

func (c *ChannelProposal) decode(r io.Reader) (err error) {
	if err := wire.Decode(r, &c.ChallengeDuration, &c.Nonce); err != nil {
		return err
	}

	// read c.ParticipantAddr, c.AppDef
	if c.ParticipantAddr, err = wallet.DecodeAddress(r); err != nil {
		return err
	}
	if c.AppDef, err = wallet.DecodeAddress(r); err != nil {
		return err
	}

	c.InitData = &channel.DummyData{}
	c.InitBals = &channel.Allocation{}
	if err := perunio.Decode(r, c.InitData, c.InitBals); err != nil {
		return err
	}

	var numParts int32
	if err := wire.Decode(r, &numParts); err != nil {
		return err
	}
	if numParts < 2 {
		return errors.Errorf(
			"expected at least 2 participants, got %d", numParts)
	}

	c.Parts = make([]wallet.Address, numParts)
	for i := 0; i < len(c.Parts); i++ {
		if c.Parts[i], err = wallet.DecodeAddress(r); err != nil {
			return err
		}
	}

	return nil
}

// SessionID is a unique identifier generated for every instantiantiation of
// a channel.
type SessionID = wire.Byte32

// ChannelProposalRes contains all data for a response to a channel proposal
// message. The SessID must be computed from the channel proposal messages one
// wishes to respond to. ParticipantAddr should be an ephemeral address just
// for this channel instantiation.
//
// The type implements the channel proposal response messages from the
// Multi-Party Channel Proposal Protocol (MPCPP).
type ChannelProposalRes struct {
	SessID          SessionID
	ParticipantAddr wallet.Address
}

func (*ChannelProposalRes) Category() wiremsg.Category {
	return wiremsg.Peer
}

func (*ChannelProposalRes) Type() MsgType {
	return PeerChannelProposalRes
}

func (r *ChannelProposalRes) encode(w io.Writer) error {
	if err := r.SessID.Encode(w); err != nil {
		return errors.WithMessagef(err, "response SID encoding")
	}

	if err := r.ParticipantAddr.Encode(w); err != nil {
		return errors.WithMessagef(err, "response ephemeral address encoding")
	}

	return nil
}

func (response *ChannelProposalRes) decode(r io.Reader) error {
	if err := response.SessID.Decode(r); err != nil {
		return errors.WithMessagef(err, "response SID decoding")
	}

	if ephemeralAddr, err := wallet.DecodeAddress(r); err != nil {
		return errors.WithMessagef(err, "app address decoding")
	} else {
		response.ParticipantAddr = ephemeralAddr
	}

	return nil
}

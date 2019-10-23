// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel

import (
	"io"
	"math"
	"math/big"

	"github.com/pkg/errors"

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

func (DummyPeerMsg) Type() PeerMsgType {
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
	ParticipantAddr   wallet.Address
	AppDef            wallet.Address
	InitData          Data
	InitBals          *Allocation
	Parts             []wallet.Address
}

func (ChannelProposal) Category() wiremsg.Category {
	return wiremsg.Peer
}

func (ChannelProposal) Type() PeerMsgType {
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
			return errors.Errorf("error encoding participant %d", i)
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

	c.InitData = &DummyData{}
	c.InitBals = &Allocation{}
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
type SessionID = [32]byte

// ChannelProposalRes contains all data for a response to a channel proposal
// message. The SessID must be computed from the channel proposal messages one
// wishes to respond to. ParticipantAddr should be a participant address just
// for this channel instantiation.
//
// The type implements the channel proposal response messages from the
// Multi-Party Channel Proposal Protocol (MPCPP).
type ChannelProposalRes struct {
	SessID          SessionID
	ParticipantAddr wallet.Address
}

func (ChannelProposalRes) Category() wiremsg.Category {
	return wiremsg.Peer
}

func (ChannelProposalRes) Type() PeerMsgType {
	return PeerChannelProposalRes
}

func (res ChannelProposalRes) encode(w io.Writer) error {
	if err := wire.Encode(w, res.SessID); err != nil {
		return errors.WithMessage(err, "response SID encoding")
	}

	if err := res.ParticipantAddr.Encode(w); err != nil {
		return errors.WithMessage(err, "response participant address encoding")
	}

	return nil
}

func (res *ChannelProposalRes) decode(r io.Reader) (err error) {
	if err = wire.Decode(r, &res.SessID); err != nil {
		return errors.WithMessage(err, "response SID decoding")
	}

	if res.ParticipantAddr, err = wallet.DecodeAddress(r); err != nil {
		return errors.WithMessage(err, "app address decoding")
	}

	return nil
}

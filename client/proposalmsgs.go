// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package client

import (
	"golang.org/x/crypto/sha3"
	"io"
	"log"
	"math"
	"math/big"

	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
	perunio "perun.network/go-perun/pkg/io"
	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wire"
	"perun.network/go-perun/wire/msg"
)

func init() {
	msg.RegisterDecoder(msg.ChannelProposal,
		func(r io.Reader) (msg.Msg, error) {
			var m ChannelProposal
			return &m, m.Decode(r)
		})
	msg.RegisterDecoder(msg.ChannelProposalAcc,
		func(r io.Reader) (msg.Msg, error) {
			var m ChannelProposalAcc
			return &m, m.Decode(r)
		})
	msg.RegisterDecoder(msg.ChannelProposalRej,
		func(r io.Reader) (msg.Msg, error) {
			var m ChannelProposalRej
			return &m, m.Decode(r)
		})
}

// SessionID is a unique identifier generated for every instantiantiation of
// a channel.
type SessionID = [32]byte

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
	InitData          channel.Data
	InitBals          *channel.Allocation
	Parts             []wallet.Address
}

func (ChannelProposal) Type() msg.Type {
	return msg.ChannelProposal
}

func (c ChannelProposal) Encode(w io.Writer) error {
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

func (c *ChannelProposal) Decode(r io.Reader) (err error) {
	if err := wire.Decode(r, &c.ChallengeDuration, &c.Nonce); err != nil {
		return err
	}

	if c.ParticipantAddr, err = wallet.DecodeAddress(r); err != nil {
		return err
	}
	if c.AppDef, err = wallet.DecodeAddress(r); err != nil {
		return err
	}
	var app channel.App
	if app, err = channel.AppFromDefinition(c.AppDef); err != nil {
		return err
	}

	if c.InitData, err = app.DecodeData(r); err != nil {
		return err
	}

	if c.InitBals == nil {
		c.InitBals = new(channel.Allocation)
	}
	if err := perunio.Decode(r, c.InitBals); err != nil {
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

func (c ChannelProposal) SessID() (sid SessionID) {
	hasher := sha3.New256()
	if err := wire.Encode(hasher, c.Nonce); err != nil {
		log.Panicf("session ID nonce encoding: %v", err)
	}

	for _, p := range c.Parts {
		if err := wire.Encode(hasher, p); err != nil {
			log.Panicf("session ID participant encoding: %v", err)
		}
	}

	if err := wire.Encode(
		hasher,
		c.ChallengeDuration,
		c.InitData,
		c.InitBals,
		c.AppDef,
	); err != nil {
		log.Panicf("session ID data encoding error: %v", err)
	}

	copy(sid[:], hasher.Sum(nil))
	return
}

// Valid checks that the channel proposal is valid:
// * ParticipantAddr, InitBals and Nonce must not be nil
// * 2 <= len(Parts) <= channel.MaxNumParts
// * InitBals match the dimension of Parts (TODO: check is valid)
// * No locked sub-allocations
// * non-zero ChallengeDuration
func (c ChannelProposal) Valid() error {
	if c.ParticipantAddr == nil || c.InitBals == nil || c.Nonce == nil {
		return errors.New("invalid nil fields")
	} else if c.ChallengeDuration == 0 {
		return errors.New("challenge duration must not be zero")
	} else if len(c.Parts) < 2 || len(c.Parts) > channel.MaxNumParts {
		return errors.New("invalid number of participants")
	} else if err := c.InitBals.Valid(); err != nil {
		return err
	} else if len(c.InitBals.Locked) > 0 {
		return errors.New("initial allocation cannot have locked funds")
	} else if len(c.InitBals.OfParts) != len(c.Parts) {
		return errors.New("wrong dimension of initial balances")
	}
	return nil
}

// ChannelProposalAcc contains all data for a response to a channel proposal
// message. The SessID must be computed from the channel proposal messages one
// wishes to respond to. ParticipantAddr should be a participant address just
// for this channel instantiation.
//
// The type implements the channel proposal response messages from the
// Multi-Party Channel Proposal Protocol (MPCPP).
type ChannelProposalAcc struct {
	SessID          SessionID
	ParticipantAddr wallet.Address
}

func (ChannelProposalAcc) Type() msg.Type {
	return msg.ChannelProposalAcc
}

func (acc ChannelProposalAcc) Encode(w io.Writer) error {
	if err := wire.Encode(w, acc.SessID); err != nil {
		return errors.WithMessage(err, "SID encoding")
	}

	if err := acc.ParticipantAddr.Encode(w); err != nil {
		return errors.WithMessage(err, "participant address encoding")
	}

	return nil
}

func (acc *ChannelProposalAcc) Decode(r io.Reader) (err error) {
	if err = wire.Decode(r, &acc.SessID); err != nil {
		return errors.WithMessage(err, "SID decoding")
	}

	acc.ParticipantAddr, err = wallet.DecodeAddress(r)
	return errors.WithMessage(err, "participant address decoding")
}

type ChannelProposalRej struct {
	SessID SessionID
	Reason string
}

func (ChannelProposalRej) Type() msg.Type {
	return msg.ChannelProposalRej
}

func (rej ChannelProposalRej) Encode(w io.Writer) error {
	return wire.Encode(w, rej.SessID, rej.Reason)
}

func (rej *ChannelProposalRej) Decode(r io.Reader) error {
	return wire.Decode(r, &rej.SessID, &rej.Reason)
}

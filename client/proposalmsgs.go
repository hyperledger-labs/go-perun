// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package client

import (
	"golang.org/x/crypto/sha3"
	"io"
	"log"
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
			var m ChannelProposalReq
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

// ChannelProposalReq is the wire message that is derived from the
// ChannelProposal.
//
// ChannelProposalReq implements the channel proposal messages from the
// Multi-Party Channel Proposal Protocol (MPCPP).
type ChannelProposalReq struct {
	ChallengeDuration uint64
	Nonce             *big.Int
	ParticipantAddr   wallet.Address
	AppDef            wallet.Address
	InitData          channel.Data
	InitBals          *channel.Allocation
	PeerAddrs         []wallet.Address
}

// AsReq returns a shallow copy of the ChannelProposal as a ChannelProposalReq,
// i.e., as a wire message.
func (c *ChannelProposal) AsReq() *ChannelProposalReq {
	return &ChannelProposalReq{
		ChallengeDuration: c.ChallengeDuration,
		Nonce:             c.Nonce,
		ParticipantAddr:   c.Account.Address(),
		AppDef:            c.AppDef,
		InitData:          c.InitData,
		InitBals:          c.InitBals,
		PeerAddrs:         c.PeerAddrs,
	}
}

// AsProp returns a shallow copy of the ChannelProposalReq as a ChannelProposal.
func (c *ChannelProposalReq) AsProp(acc wallet.Account) *ChannelProposal {
	return &ChannelProposal{
		ChallengeDuration: c.ChallengeDuration,
		Nonce:             c.Nonce,
		Account:           acc,
		AppDef:            c.AppDef,
		InitData:          c.InitData,
		InitBals:          c.InitBals,
		PeerAddrs:         c.PeerAddrs,
	}
}

// Type returns msg.ChannelProposal.
func (ChannelProposalReq) Type() msg.Type {
	return msg.ChannelProposal
}

// Encode encodes the ChannelProposalReq into an io.writer.
func (c ChannelProposalReq) Encode(w io.Writer) error {
	if w == nil {
		return errors.New("writer must not be nil")
	}

	if err := wire.Encode(w, c.ChallengeDuration, c.Nonce); err != nil {
		return err
	}

	if err := perunio.Encode(w, c.ParticipantAddr, c.AppDef, c.InitData, c.InitBals); err != nil {
		return err
	}

	if len(c.PeerAddrs) > channel.MaxNumParts {
		return errors.Errorf(
			"expected maximum number of participants %d, got %d",
			channel.MaxNumParts, len(c.PeerAddrs))
	}

	numParts := int32(len(c.PeerAddrs))
	if err := wire.Encode(w, numParts); err != nil {
		return err
	}
	for i := range c.PeerAddrs {
		if err := c.PeerAddrs[i].Encode(w); err != nil {
			return errors.Errorf("error encoding participant %d", i)
		}
	}

	return nil
}

// Decode decodes a ChannelProposalRequest from an io.Reader.
func (c *ChannelProposalReq) Decode(r io.Reader) (err error) {
	if r == nil {
		return errors.New("reader must not be nil")
	}

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
	if numParts > channel.MaxNumParts {
		return errors.Errorf(
			"expected at most %d participants, got %d",
			channel.MaxNumParts, numParts)
	}

	c.PeerAddrs = make([]wallet.Address, numParts)
	for i := 0; i < len(c.PeerAddrs); i++ {
		if c.PeerAddrs[i], err = wallet.DecodeAddress(r); err != nil {
			return err
		}
	}

	return nil
}

// SessID calculates the SessionID of a ChannelProposalReq.
func (c ChannelProposalReq) SessID() (sid SessionID) {
	hasher := sha3.New256()
	if err := wire.Encode(hasher, c.Nonce); err != nil {
		log.Panicf("session ID nonce encoding: %v", err)
	}

	for _, p := range c.PeerAddrs {
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
// * ParticipantAddr, InitBals must not be nil
// * ValidateParameters returns nil
// * InitBals are valid
// * No locked sub-allocations
// * InitBals match the dimension of Parts
// * non-zero ChallengeDuration
func (c ChannelProposalReq) Valid() error {
	if c.InitBals == nil || c.ParticipantAddr == nil {
		return errors.New("invalid nil fields")
	} else if err := channel.ValidateParameters(
		c.ChallengeDuration, len(c.PeerAddrs), c.AppDef, c.Nonce); err != nil {
		return errors.WithMessage(err, "invalid channel parameters")
	} else if err := c.InitBals.Valid(); err != nil {
		return err
	} else if len(c.InitBals.Locked) != 0 {
		return errors.New("initial allocation cannot have locked funds")
	} else if len(c.InitBals.OfParts) != len(c.PeerAddrs) {
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

// Type returns msg.ChannelProposalAcc.
func (ChannelProposalAcc) Type() msg.Type {
	return msg.ChannelProposalAcc
}

// Encode encodes the ChannelProposalAcc into an io.Writer.
func (acc ChannelProposalAcc) Encode(w io.Writer) error {
	if err := wire.Encode(w, acc.SessID); err != nil {
		return errors.WithMessage(err, "SID encoding")
	}

	if err := acc.ParticipantAddr.Encode(w); err != nil {
		return errors.WithMessage(err, "participant address encoding")
	}

	return nil
}

// Decode decodes a ChannelProposalAcc from an io.Reader.
func (acc *ChannelProposalAcc) Decode(r io.Reader) (err error) {
	if err = wire.Decode(r, &acc.SessID); err != nil {
		return errors.WithMessage(err, "SID decoding")
	}

	acc.ParticipantAddr, err = wallet.DecodeAddress(r)
	return errors.WithMessage(err, "participant address decoding")
}

// ChannelProposalRej is used to reject a ChannelProposalReq.
// An optional reason for the rejection can be set.
//
// The message is one of two possible responses in the
// Multi-Party Channel Proposal Protocol (MPCPP).
type ChannelProposalRej struct {
	SessID SessionID
	Reason string
}

// Type returns msg.ChannelProposalRej.
func (ChannelProposalRej) Type() msg.Type {
	return msg.ChannelProposalRej
}

// Encode encodes a ChannelProposalRej into an io.Writer.
func (rej ChannelProposalRej) Encode(w io.Writer) error {
	return wire.Encode(w, rej.SessID, rej.Reason)
}

// Decode decodes a ChannelProposalRej from an io.Reader.
func (rej *ChannelProposalRej) Decode(r io.Reader) error {
	return wire.Decode(r, &rej.SessID, &rej.Reason)
}

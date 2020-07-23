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
	"log"
	"math/big"

	"github.com/pkg/errors"
	"golang.org/x/crypto/sha3"

	"perun.network/go-perun/channel"
	perunio "perun.network/go-perun/pkg/io"
	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wire"
)

func init() {
	wire.RegisterDecoder(wire.ChannelProposal,
		func(r io.Reader) (wire.Msg, error) {
			var m ChannelProposal
			return &m, m.Decode(r)
		})
	wire.RegisterDecoder(wire.ChannelProposalAcc,
		func(r io.Reader) (wire.Msg, error) {
			var m ChannelProposalAcc
			return &m, m.Decode(r)
		})
	wire.RegisterDecoder(wire.ChannelProposalRej,
		func(r io.Reader) (wire.Msg, error) {
			var m ChannelProposalRej
			return &m, m.Decode(r)
		})
}

// SessionID is a unique identifier generated for every instantiantiation of
// a channel.
type SessionID = [32]byte

// ChannelProposal contains all data necessary to propose a new
// channel to a given set of peers. It is also sent over the wire.
//
// ChannelProposal implements the channel proposal messages from the
// Multi-Party Channel Proposal Protocol (MPCPP).
type ChannelProposal struct {
	ChallengeDuration uint64
	Nonce             *big.Int
	ParticipantAddr   wallet.Address
	AppDef            wallet.Address
	InitData          channel.Data
	InitBals          *channel.Allocation
	PeerAddrs         []wire.Address
}

// Type returns wire.ChannelProposal.
func (ChannelProposal) Type() wire.Type {
	return wire.ChannelProposal
}

// Encode encodes the ChannelProposalReq into an io.writer.
func (c ChannelProposal) Encode(w io.Writer) error {
	if w == nil {
		return errors.New("writer must not be nil")
	}

	if err := perunio.Encode(w, c.ChallengeDuration, c.Nonce); err != nil {
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
	if err := perunio.Encode(w, numParts); err != nil {
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
func (c *ChannelProposal) Decode(r io.Reader) (err error) {
	if r == nil {
		return errors.New("reader must not be nil")
	}

	if err := perunio.Decode(r, &c.ChallengeDuration, &c.Nonce); err != nil {
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
	if err := perunio.Decode(r, &numParts); err != nil {
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
	for i := range c.PeerAddrs {
		if c.PeerAddrs[i], err = wallet.DecodeAddress(r); err != nil {
			return err
		}
	}

	return nil
}

// SessID calculates the SessionID of a ChannelProposalReq.
func (c ChannelProposal) SessID() (sid SessionID) {
	hasher := sha3.New256()
	if err := perunio.Encode(hasher, c.Nonce); err != nil {
		log.Panicf("session ID nonce encoding: %v", err)
	}

	for _, p := range c.PeerAddrs {
		if err := perunio.Encode(hasher, p); err != nil {
			log.Panicf("session ID participant encoding: %v", err)
		}
	}

	if err := perunio.Encode(
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
func (c ChannelProposal) Valid() error {
	if c.InitBals == nil || c.ParticipantAddr == nil {
		return errors.New("invalid nil fields")
	} else if err := channel.ValidateParameters(
		c.ChallengeDuration, len(c.PeerAddrs), c.AppDef, c.Nonce); err != nil {
		return errors.WithMessage(err, "invalid channel parameters")
	} else if err := c.InitBals.Valid(); err != nil {
		return err
	} else if len(c.InitBals.Locked) != 0 {
		return errors.New("initial allocation cannot have locked funds")
	} else if len(c.InitBals.Balances[0]) != len(c.PeerAddrs) {
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

// Type returns wire.ChannelProposalAcc.
func (ChannelProposalAcc) Type() wire.Type {
	return wire.ChannelProposalAcc
}

// Encode encodes the ChannelProposalAcc into an io.Writer.
func (acc ChannelProposalAcc) Encode(w io.Writer) error {
	if err := perunio.Encode(w, acc.SessID); err != nil {
		return errors.WithMessage(err, "SID encoding")
	}

	if err := acc.ParticipantAddr.Encode(w); err != nil {
		return errors.WithMessage(err, "participant address encoding")
	}

	return nil
}

// Decode decodes a ChannelProposalAcc from an io.Reader.
func (acc *ChannelProposalAcc) Decode(r io.Reader) (err error) {
	if err = perunio.Decode(r, &acc.SessID); err != nil {
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

// Type returns wire.ChannelProposalRej.
func (ChannelProposalRej) Type() wire.Type {
	return wire.ChannelProposalRej
}

// Encode encodes a ChannelProposalRej into an io.Writer.
func (rej ChannelProposalRej) Encode(w io.Writer) error {
	return perunio.Encode(w, rej.SessID, rej.Reason)
}

// Decode decodes a ChannelProposalRej from an io.Reader.
func (rej *ChannelProposalRej) Decode(r io.Reader) error {
	return perunio.Decode(r, &rej.SessID, &rej.Reason)
}

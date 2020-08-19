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
	"hash"
	"io"
	"log"

	"golang.org/x/crypto/sha3"

	"github.com/pkg/errors"

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

func newHasher() hash.Hash { return sha3.New256() }

// ProposalID uniquely identifies the channel proposal as
// specified by the Channel Proposal Protocol (CPP).
type ProposalID = [32]byte

// NonceShare is used to cooperatively calculate a channel's nonce.
type NonceShare = [32]byte

// ChannelProposal contains all data necessary to propose a new
// channel to a given set of peers. It is also sent over the wire.
//
// ChannelProposal implements the channel proposal messages from the
// Multi-Party Channel Proposal Protocol (MPCPP).
type ChannelProposal struct {
	ChallengeDuration uint64
	NonceShare        NonceShare
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

	if err := perunio.Encode(w, c.ChallengeDuration, c.NonceShare); err != nil {
		return err
	}

	if err := perunio.Encode(w, c.ParticipantAddr, OptAppDefAndDataEnc{c.AppDef, c.InitData}, c.InitBals); err != nil {
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
	return wallet.Addresses(c.PeerAddrs).Encode(w)
}

// OptAppDefAndDataEnc makes an optional pair of App definition and Data encodable.
type OptAppDefAndDataEnc struct {
	wallet.Address
	channel.Data
}

// Encode encodes an optional pair of App definition and Data.
func (o OptAppDefAndDataEnc) Encode(w io.Writer) error {
	if o.Address == nil {
		return perunio.Encode(w, false)
	}
	return perunio.Encode(w, true, o.Address, o.Data)
}

// OptAppDefAndDataDec makes an optional pair of App definition and Data decodable.
type OptAppDefAndDataDec struct {
	Address *wallet.Address
	Data    *channel.Data
}

// Decode decodes an optional pair of App definition and Data.
func (o OptAppDefAndDataDec) Decode(r io.Reader) (err error) {
	*o.Data = nil
	*o.Address = nil
	var app channel.App
	if err = perunio.Decode(r, channel.OptAppDec{App: &app}); err != nil {
		return err
	}

	if app == nil {
		return nil
	}

	*o.Address = app.Def()
	*o.Data, err = app.DecodeData(r)
	return err
}

// Decode decodes a ChannelProposalRequest from an io.Reader.
func (c *ChannelProposal) Decode(r io.Reader) (err error) {
	if r == nil {
		return errors.New("reader must not be nil")
	}

	if err := perunio.Decode(r, &c.ChallengeDuration, &c.NonceShare); err != nil {
		return err
	}

	if c.InitBals == nil {
		c.InitBals = new(channel.Allocation)
	}

	if err := perunio.Decode(r,
		wallet.AddressDec{Addr: &c.ParticipantAddr},
		OptAppDefAndDataDec{&c.AppDef, &c.InitData},
		c.InitBals); err != nil {
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
	return wallet.Addresses(c.PeerAddrs).Decode(r)
}

// ProposalID returns the identifier of this channel proposal request as
// specified by the Channel Proposal Protocol (CPP).
func (c ChannelProposal) ProposalID() (propID ProposalID) {
	hasher := newHasher()
	if err := perunio.Encode(hasher, c.NonceShare); err != nil {
		log.Panicf("proposal ID nonce encoding: %v", err)
	}

	for _, p := range c.PeerAddrs {
		if err := perunio.Encode(hasher, p); err != nil {
			log.Panicf("proposal ID participant encoding: %v", err)
		}
	}

	if err := perunio.Encode(
		hasher,
		c.ChallengeDuration,
		c.InitData,
		c.InitBals,
		c.AppDef,
	); err != nil {
		log.Panicf("proposal ID data encoding error: %v", err)
	}

	copy(propID[:], hasher.Sum(nil))
	return
}

// Valid checks that the channel proposal is valid:
// * ParticipantAddr, InitBals must not be nil
// * ValidateProposalParameters returns nil
// * InitBals are valid
// * No locked sub-allocations
// * InitBals match the dimension of Parts
// * non-zero ChallengeDuration.
func (c ChannelProposal) Valid() error {
	// nolint: gocritic
	if c.InitBals == nil || c.ParticipantAddr == nil {
		return errors.New("invalid nil fields")
	} else if err := channel.ValidateProposalParameters(
		c.ChallengeDuration, len(c.PeerAddrs), c.AppDef); err != nil {
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
// message. The ProposalID must correspond to the channel proposal request one
// wishes to respond to. ParticipantAddr should be a participant address just
// for this channel instantiation.
//
// The type implements the channel proposal response messages from the
// Multi-Party Channel Proposal Protocol (MPCPP).
type ChannelProposalAcc struct {
	ProposalID      ProposalID
	NonceShare      NonceShare
	ParticipantAddr wallet.Address
}

// Type returns wire.ChannelProposalAcc.
func (ChannelProposalAcc) Type() wire.Type {
	return wire.ChannelProposalAcc
}

// Encode encodes the ChannelProposalAcc into an io.Writer.
func (acc ChannelProposalAcc) Encode(w io.Writer) error {
	return perunio.Encode(w,
		acc.ProposalID,
		acc.NonceShare,
		acc.ParticipantAddr)
}

// Decode decodes a ChannelProposalAcc from an io.Reader.
func (acc *ChannelProposalAcc) Decode(r io.Reader) (err error) {
	return perunio.Decode(r,
		&acc.ProposalID,
		&acc.NonceShare,
		wallet.AddressDec{Addr: &acc.ParticipantAddr})
}

// ChannelProposalRej is used to reject a ChannelProposalReq.
// An optional reason for the rejection can be set.
//
// The message is one of two possible responses in the
// Multi-Party Channel Proposal Protocol (MPCPP).
type ChannelProposalRej struct {
	ProposalID ProposalID
	Reason     string
}

// Type returns wire.ChannelProposalRej.
func (ChannelProposalRej) Type() wire.Type {
	return wire.ChannelProposalRej
}

// Encode encodes a ChannelProposalRej into an io.Writer.
func (rej ChannelProposalRej) Encode(w io.Writer) error {
	return perunio.Encode(w, rej.ProposalID, rej.Reason)
}

// Decode decodes a ChannelProposalRej from an io.Reader.
func (rej *ChannelProposalRej) Decode(r io.Reader) error {
	return perunio.Decode(r, &rej.ProposalID, &rej.Reason)
}

// CalcNonce calculates a nonce from its shares. The order of the shares must
// correspond to the participant indices.
func CalcNonce(nonceShares []NonceShare) channel.Nonce {
	hasher := newHasher()
	for i, share := range nonceShares {
		if err := perunio.Encode(hasher, share); err != nil {
			log.Panicf("Failed to encode nonce share %d for hashing", i)
		}
	}
	return channel.NonceFromBytes(hasher.Sum(nil))
}

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

	"golang.org/x/crypto/sha3"

	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/log"
	perunio "perun.network/go-perun/pkg/io"
	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wire"
)

func init() {
	wire.RegisterDecoder(wire.LedgerChannelProposal,
		func(r io.Reader) (wire.Msg, error) {
			m := LedgerChannelProposal{}
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

type (
	// ChannelProposal is the interface that describes all channel proposal message
	// types.
	ChannelProposal interface {
		wire.Msg
		perunio.Decoder

		// Proposal returns the channel proposal's common values.
		Proposal() *BaseChannelProposal
	}

	// BaseChannelProposal contains all data necessary to propose a new
	// channel to a given set of peers. It is also sent over the wire.
	//
	// ChannelProposal implements the channel proposal messages from the
	// Multi-Party Channel Proposal Protocol (MPCPP).
	BaseChannelProposal struct {
		ChallengeDuration uint64              // Dispute challenge duration.
		NonceShare        NonceShare          // Proposer's channel nonce share.
		ParticipantAddr   wallet.Address      // Proposer's address in the channel.
		AppDef            wallet.Address      // App definition, or nil.
		InitData          channel.Data        // Initial App data, or nil (if App nil).
		InitBals          *channel.Allocation // Initial balances.
		PeerAddrs         []wire.Address      // Participants' wire addresses.
	}

	// LedgerChannelProposal is a channel proposal for ledger channels and has no
	// additional values beyond a BaseChannelProposal.
	LedgerChannelProposal struct {
		BaseChannelProposal
	}
)

// makeBaseChannelProposal creates a BaseChannelProposal and applies the supplied
// options. For more information, see ProposalOpts.
func makeBaseChannelProposal(
	challengeDuration uint64,
	participantAddr wallet.Address,
	initBals *channel.Allocation,
	peerAddrs []wire.Address,
	opts ...ProposalOpts,
) BaseChannelProposal {
	opt := union(opts...)

	return BaseChannelProposal{
		ChallengeDuration: challengeDuration,
		NonceShare:        opt.nonce(),
		ParticipantAddr:   participantAddr,
		AppDef:            opt.appDef(),
		InitData:          opt.appData(),
		InitBals:          initBals,
		PeerAddrs:         peerAddrs,
	}
}

// Proposal returns the channel proposal's common values.
func (p *BaseChannelProposal) Proposal() *BaseChannelProposal {
	return p
}

// Encode encodes the ChannelProposalReq into an io.writer.
func (p *BaseChannelProposal) Encode(w io.Writer) error {
	if w == nil {
		return errors.New("writer must not be nil")
	}

	if err := perunio.Encode(w, p.ChallengeDuration, p.NonceShare); err != nil {
		return err
	}

	if err := perunio.Encode(w, p.ParticipantAddr, OptAppDefAndDataEnc{p.AppDef, p.InitData}, p.InitBals); err != nil {
		return err
	}

	if len(p.PeerAddrs) > channel.MaxNumParts {
		return errors.Errorf(
			"expected maximum number of participants %d, got %d",
			channel.MaxNumParts, len(p.PeerAddrs))
	}

	numParts := int32(len(p.PeerAddrs))
	if err := perunio.Encode(w, numParts); err != nil {
		return err
	}
	return wallet.Addresses(p.PeerAddrs).Encode(w)
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
	Def  *wallet.Address
	Data *channel.Data
}

// Decode decodes an optional pair of App definition and Data.
func (o OptAppDefAndDataDec) Decode(r io.Reader) (err error) {
	*o.Data = nil
	*o.Def = nil
	var app channel.App
	if err = perunio.Decode(r, channel.OptAppDec{App: &app}); err != nil {
		return err
	}

	if app == nil {
		return nil
	}

	*o.Def = app.Def()
	*o.Data, err = app.DecodeData(r)
	return err
}

// Decode decodes a ChannelProposalRequest from an io.Reader.
func (p *BaseChannelProposal) Decode(r io.Reader) (err error) {
	if r == nil {
		return errors.New("reader must not be nil")
	}

	if err := perunio.Decode(r, &p.ChallengeDuration, &p.NonceShare); err != nil {
		return err
	}

	if p.InitBals == nil {
		p.InitBals = new(channel.Allocation)
	}

	if err := perunio.Decode(r,
		wallet.AddressDec{Addr: &p.ParticipantAddr},
		OptAppDefAndDataDec{&p.AppDef, &p.InitData},
		p.InitBals); err != nil {
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

	p.PeerAddrs = make([]wallet.Address, numParts)
	return wallet.Addresses(p.PeerAddrs).Decode(r)
}

// ProposalID returns the identifier of this channel proposal request as
// specified by the Channel Proposal Protocol (CPP).
func (p *BaseChannelProposal) ProposalID() (propID ProposalID) {
	hasher := newHasher()
	if err := perunio.Encode(hasher, p.NonceShare); err != nil {
		log.Panicf("proposal ID nonce encoding: %v", err)
	}

	for _, p := range p.PeerAddrs {
		if err := perunio.Encode(hasher, p); err != nil {
			log.Panicf("proposal ID participant encoding: %v", err)
		}
	}

	if err := perunio.Encode(
		hasher,
		p.ChallengeDuration,
		p.InitData,
		p.InitBals,
		p.AppDef,
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
func (p *BaseChannelProposal) Valid() error {
	// nolint: gocritic
	if p.InitBals == nil || p.ParticipantAddr == nil {
		return errors.New("invalid nil fields")
	} else if err := channel.ValidateProposalParameters(
		p.ChallengeDuration, len(p.PeerAddrs), p.AppDef); err != nil {
		return errors.WithMessage(err, "invalid channel parameters")
	} else if err := p.InitBals.Valid(); err != nil {
		return err
	} else if len(p.InitBals.Locked) != 0 {
		return errors.New("initial allocation cannot have locked funds")
	} else if len(p.InitBals.Balances[0]) != len(p.PeerAddrs) {
		return errors.New("wrong dimension of initial balances")
	}
	return nil
}

// NewChannelProposalAcc constructs an accept message that belongs to a proposal
// message. It should be used instead of manually constructing an accept
// message.
func (p *BaseChannelProposal) NewChannelProposalAcc(
	participantAddr wallet.Address,
	nonceShare ProposalOpts,
) *ChannelProposalAcc {
	if !nonceShare.isNonce() {
		log.WithField("proposal", p.ProposalID()).
			Panic("NewChannelProposalAcc: nonceShare has no configured nonce")
	}
	return &ChannelProposalAcc{
		ProposalID:      p.ProposalID(),
		NonceShare:      nonceShare.nonce(),
		ParticipantAddr: participantAddr,
	}
}

// NewLedgerChannelProposal creates a ledger channel proposal and applies the
// supplied options. For more information, see ProposalOpts.
func NewLedgerChannelProposal(
	challengeDuration uint64,
	participantAddr wallet.Address,
	initBals *channel.Allocation,
	peerAddrs []wire.Address,
	opts ...ProposalOpts,
) *LedgerChannelProposal {
	return &LedgerChannelProposal{
		makeBaseChannelProposal(
			challengeDuration,
			participantAddr,
			initBals,
			peerAddrs,
			opts...)}
}

// Type returns wire.ChannelProposal.
func (LedgerChannelProposal) Type() wire.Type {
	return wire.LedgerChannelProposal
}

// ChannelProposalAcc contains all data for a response to a channel proposal
// message. The ProposalID must correspond to the channel proposal request one
// wishes to respond to. ParticipantAddr should be a participant address just
// for this channel instantiation.
//
// The type implements the channel proposal response messages from the
// Multi-Party Channel Proposal Protocol (MPCPP).
type ChannelProposalAcc struct {
	ProposalID      ProposalID     // Proposal session ID we're answering.
	NonceShare      NonceShare     // Responder's channel nonce share.
	ParticipantAddr wallet.Address // Responder's participant address.
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
	ProposalID ProposalID // The channel proposal to reject.
	Reason     string     // The rejection reason.
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

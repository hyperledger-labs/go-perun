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
	"crypto/rand"
	"hash"
	"io"

	"golang.org/x/crypto/sha3"

	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/log"
	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wire"
	"perun.network/go-perun/wire/perunio"
	perunbig "polycry.pt/poly-go/math/big"
)

func init() {
	wire.RegisterDecoder(wire.LedgerChannelProposal,
		func(r io.Reader) (wire.Msg, error) {
			m := LedgerChannelProposalMsg{}
			return &m, m.Decode(r)
		})
	wire.RegisterDecoder(wire.LedgerChannelProposalAcc,
		func(r io.Reader) (wire.Msg, error) {
			var m LedgerChannelProposalAccMsg
			return &m, m.Decode(r)
		})
	wire.RegisterDecoder(wire.SubChannelProposal,
		func(r io.Reader) (wire.Msg, error) {
			m := SubChannelProposalMsg{}
			return &m, m.Decode(r)
		})
	wire.RegisterDecoder(wire.SubChannelProposalAcc,
		func(r io.Reader) (wire.Msg, error) {
			var m SubChannelProposalAccMsg
			return &m, m.Decode(r)
		})
	wire.RegisterDecoder(wire.ChannelProposalRej,
		func(r io.Reader) (wire.Msg, error) {
			var m ChannelProposalRejMsg
			return &m, m.Decode(r)
		})
	wire.RegisterDecoder(wire.VirtualChannelProposal,
		func(r io.Reader) (wire.Msg, error) {
			m := VirtualChannelProposalMsg{}
			return &m, m.Decode(r)
		})
	wire.RegisterDecoder(wire.VirtualChannelProposalAcc,
		func(r io.Reader) (wire.Msg, error) {
			m := VirtualChannelProposalAccMsg{}
			return &m, m.Decode(r)
		})
}

func newHasher() hash.Hash { return sha3.New256() }

// ProposalID uniquely identifies the channel proposal as  specified by the
// Channel Proposal Protocol (CPP).
type ProposalID = [32]byte

// NonceShare is used to cooperatively calculate a channel's nonce.
type NonceShare = [32]byte

type (
	// ChannelProposal is the interface that describes all channel proposal
	// message types.
	ChannelProposal interface {
		wire.Msg
		perunio.Decoder

		// Base returns the channel proposal's common values.
		Base() *BaseChannelProposal

		// Matches checks whether an accept message is of the correct type. This
		// does not check any contents of the accept message, only its type.
		Matches(ChannelProposalAccept) bool

		// Valid checks whether a channel proposal is valid.
		Valid() error
	}

	// BaseChannelProposal contains all data necessary to propose a new
	// channel to a given set of peers. It is also sent over the wire.
	//
	// BaseChannelProposal implements the channel proposal messages from the
	// Multi-Party Channel Proposal Protocol (MPCPP).
	BaseChannelProposal struct {
		ProposalID        ProposalID          // Unique ID for the proposal.
		ChallengeDuration uint64              // Dispute challenge duration.
		NonceShare        NonceShare          // Proposer's channel nonce share.
		App               channel.App         // App definition, or nil.
		InitData          channel.Data        // Initial App data.
		InitBals          *channel.Allocation // Initial balances.
		FundingAgreement  channel.Balances    // Possibly different funding agreement from initial state's balances.
	}

	// LedgerChannelProposalMsg is a channel proposal for ledger channels.
	LedgerChannelProposalMsg struct {
		BaseChannelProposal
		Participant map[wallet.BackendID]wallet.Address // Proposer's address in the channel.
		Peers       []map[wallet.BackendID]wire.Address // Participants' wire addresses.
	}

	// SubChannelProposalMsg is a channel proposal for subchannels.
	SubChannelProposalMsg struct {
		BaseChannelProposal
		Parent map[wallet.BackendID]channel.ID
	}
)

// proposalPeers returns the wire addresses of a proposed channel's
// participants.
func (c *Client) proposalPeers(p ChannelProposal) (peers []map[wallet.BackendID]wire.Address) {
	switch prop := p.(type) {
	case *LedgerChannelProposalMsg:
		peers = prop.Peers
	case *SubChannelProposalMsg:
		ch, ok := c.channels.Channel(prop.Parent)
		if !ok {
			c.log.Panic("ProposalPeers: invalid parent channel")
		}
		peers = ch.Peers()
	case *VirtualChannelProposalMsg:
		peers = prop.Peers
	default:
		c.log.Panicf("ProposalPeers: unhandled proposal type %T")
	}
	return
}

// makeBaseChannelProposal creates a BaseChannelProposal and applies the supplied
// options. For more information, see ProposalOpts.
func makeBaseChannelProposal(
	challengeDuration uint64,
	initBals *channel.Allocation,
	opts ...ProposalOpts,
) (BaseChannelProposal, error) {
	opt := union(opts...)

	fundingAgreement := initBals.Balances
	if opt.isFundingAgreement() {
		fundingAgreement = opt.fundingAgreement()
		if equal, err := perunbig.EqualSum(initBals.Balances, fundingAgreement); err != nil {
			return BaseChannelProposal{}, errors.WithMessage(err, "comparing FundingAgreement and initial balances sum")
		} else if !equal {
			return BaseChannelProposal{}, errors.New("FundingAgreement and initial balances differ")
		}
	}

	var proposalID ProposalID
	if _, err := io.ReadFull(rand.Reader, proposalID[:]); err != nil {
		return BaseChannelProposal{}, errors.Wrap(err, "generating proposal ID")
	}

	return BaseChannelProposal{
		ChallengeDuration: challengeDuration,
		ProposalID:        proposalID,
		NonceShare:        opt.nonce(),
		App:               opt.App(),
		InitData:          opt.AppData(),
		InitBals:          initBals,
		FundingAgreement:  fundingAgreement,
	}, nil
}

// Base returns the channel proposal's common values.
func (p *BaseChannelProposal) Base() *BaseChannelProposal {
	return p
}

// NumPeers returns the number of peers in a channel.
func (p BaseChannelProposal) NumPeers() int {
	return len(p.InitBals.Balances[0])
}

// Encode encodes the BaseChannelProposal into an io.Writer.
func (p BaseChannelProposal) Encode(w io.Writer) error {
	optAppAndDataEnc := channel.OptAppAndDataEnc{App: p.App, Data: p.InitData}
	return perunio.Encode(w, p.ProposalID, p.ChallengeDuration, p.NonceShare,
		optAppAndDataEnc, p.InitBals, p.FundingAgreement)
}

// Decode decodes a BaseChannelProposal from an io.Reader.
func (p *BaseChannelProposal) Decode(r io.Reader) (err error) {
	if p.InitBals == nil {
		p.InitBals = new(channel.Allocation)
	}
	optAppAndDataDec := channel.OptAppAndDataDec{App: &p.App, Data: &p.InitData}
	return perunio.Decode(r, &p.ProposalID, &p.ChallengeDuration, &p.NonceShare,
		optAppAndDataDec, p.InitBals, &p.FundingAgreement)
}

// Valid checks that the channel proposal is valid:
// * InitBals must not be nil
// * ValidateProposalParameters returns nil
// * InitBals are valid
// * No locked sub-allocations
// * non-zero ChallengeDuration.
func (p *BaseChannelProposal) Valid() error {
	if p.InitBals == nil {
		return errors.New("invalid nil fields")
	} else if err := channel.ValidateProposalParameters(
		p.ChallengeDuration, p.NumPeers(), p.App); err != nil {
		return errors.WithMessage(err, "invalid channel parameters")
	} else if err := p.InitBals.Valid(); err != nil {
		return err
	} else if len(p.InitBals.Locked) != 0 {
		return errors.New("initial allocation cannot have locked funds")
	}
	return nil
}

// Accept constructs an accept message that belongs to a proposal message. It
// should be used instead of manually constructing an accept message.
func (p LedgerChannelProposalMsg) Accept(
	participant map[wallet.BackendID]wallet.Address,
	nonceShare ProposalOpts,
) *LedgerChannelProposalAccMsg {
	log.Println("LedgerChannelProposalMsg.Accept")
	if !nonceShare.isNonce() {
		log.WithField("proposal", p.ProposalID).
			Panic("LedgerChannelProposal.Accept: nonceShare has no configured nonce")
	}
	return &LedgerChannelProposalAccMsg{
		BaseChannelProposalAcc: makeBaseChannelProposalAcc(
			p.ProposalID, nonceShare.nonce()),
		Participant: participant,
	}
}

// Matches requires that the accept message is a LedgerChannelAcc message.
func (LedgerChannelProposalMsg) Matches(acc ChannelProposalAccept) bool {
	_, ok := acc.(*LedgerChannelProposalAccMsg)
	return ok
}

// NewLedgerChannelProposal creates a ledger channel proposal and applies the
// supplied options.
// challengeDuration is the on-chain challenge duration in seconds.
// participant is our wallet address.
// initBals are the initial balances.
// peers are the wire addresses of the channel participants.
// For more information, see ProposalOpts.
func NewLedgerChannelProposal(
	challengeDuration uint64,
	participant map[wallet.BackendID]wallet.Address,
	initBals *channel.Allocation,
	peers []map[wallet.BackendID]wire.Address,
	opts ...ProposalOpts,
) (prop *LedgerChannelProposalMsg, err error) {
	prop = &LedgerChannelProposalMsg{
		Participant: participant,
		Peers:       peers,
	}
	prop.BaseChannelProposal, err = makeBaseChannelProposal(
		challengeDuration,
		initBals,
		opts...)
	return
}

// Type returns wire.LedgerChannelProposal.
func (LedgerChannelProposalMsg) Type() wire.Type {
	return wire.LedgerChannelProposal
}

// Encode encodes a ledger channel proposal.
func (p LedgerChannelProposalMsg) Encode(w io.Writer) error {
	if err := p.assertValidNumParts(); err != nil {
		return err
	}
	return perunio.Encode(w,
		p.BaseChannelProposal,
		wallet.AddressDecMap(p.Participant),
		wire.AddressMapArray(p.Peers))
}

// Decode decodes a ledger channel proposal.
func (p *LedgerChannelProposalMsg) Decode(r io.Reader) error {
	err := perunio.Decode(r,
		&p.BaseChannelProposal,
		(*wallet.AddressDecMap)(&p.Participant),
		(*wire.AddressMapArray)(&p.Peers))
	if err != nil {
		return err
	}

	return p.assertValidNumParts()
}

func (p LedgerChannelProposalMsg) assertValidNumParts() error {
	if len(p.Peers) < 2 || len(p.Peers) > channel.MaxNumParts {
		return errors.Errorf("expected 2-%d participants, got %d",
			channel.MaxNumParts, len(p.Peers))
	}
	return nil
}

// Valid checks whether the participant address is nil.
func (p LedgerChannelProposalMsg) Valid() error {
	if err := p.BaseChannelProposal.Valid(); err != nil {
		return err
	}
	if p.Participant == nil {
		return errors.New("invalid nil participant")
	}
	return nil
}

// NewSubChannelProposal creates a subchannel proposal and applies the
// supplied options. For more information, see ProposalOpts.
func NewSubChannelProposal(
	parent map[wallet.BackendID]channel.ID,
	challengeDuration uint64,
	initBals *channel.Allocation,
	opts ...ProposalOpts,
) (prop *SubChannelProposalMsg, err error) {
	if union(opts...).isFundingAgreement() {
		return nil, errors.New("Sub-Channels currently do not support funding agreements")
	}
	prop = &SubChannelProposalMsg{Parent: parent}
	prop.BaseChannelProposal, err = makeBaseChannelProposal(
		challengeDuration,
		initBals,
		opts...)
	return
}

// Encode encodes the SubChannelProposal into an io.Writer.
func (p SubChannelProposalMsg) Encode(w io.Writer) error {
	return perunio.Encode(w, p.BaseChannelProposal, channel.IDMap(p.Parent))
}

// Decode decodes a SubChannelProposal from an io.Reader.
func (p *SubChannelProposalMsg) Decode(r io.Reader) error {
	return perunio.Decode(r, &p.BaseChannelProposal, (*channel.IDMap)(&p.Parent))
}

// Type returns wire.SubChannelProposal.
func (SubChannelProposalMsg) Type() wire.Type {
	return wire.SubChannelProposal
}

// Accept constructs an accept message that belongs to a proposal message. It
// should be used instead of manually constructing an accept message.
func (p SubChannelProposalMsg) Accept(
	nonceShare ProposalOpts,
) *SubChannelProposalAccMsg {
	if !nonceShare.isNonce() {
		log.WithField("proposal", p.ProposalID).
			Panic("SubChannelProposal.Accept: nonceShare has no configured nonce")
	}
	return &SubChannelProposalAccMsg{
		BaseChannelProposalAcc: makeBaseChannelProposalAcc(
			p.ProposalID, nonceShare.nonce()),
	}
}

// Matches requires that the accept message is a sub channel proposal accept
// message.
func (SubChannelProposalMsg) Matches(acc ChannelProposalAccept) bool {
	_, ok := acc.(*SubChannelProposalAccMsg)
	return ok
}

type (
	// ChannelProposalAccept is the generic interface for channel proposal
	// accept messages.
	ChannelProposalAccept interface {
		wire.Msg
		Base() *BaseChannelProposalAcc
	}

	// BaseChannelProposalAcc contains all data for a response to a channel proposal
	// message. The ProposalID must correspond to the channel proposal request one
	// wishes to respond to. Participant should be a participant address just
	// for this channel instantiation.
	//
	// The type implements the channel proposal response messages from the
	// Multi-Party Channel Proposal Protocol (MPCPP).
	BaseChannelProposalAcc struct {
		ProposalID ProposalID // Proposal session ID we're answering.
		NonceShare NonceShare // Responder's channel nonce share.
	}

	// LedgerChannelProposalAccMsg is the accept message type corresponding to
	// ledger channel proposals. ParticipantAdd is recommended to be unique for
	// each channel instantiation.
	LedgerChannelProposalAccMsg struct {
		BaseChannelProposalAcc
		Participant map[wallet.BackendID]wallet.Address // Responder's participant address.
	}

	// SubChannelProposalAccMsg is the accept message type corresponding to sub
	// channel proposals.
	SubChannelProposalAccMsg struct {
		BaseChannelProposalAcc
	}
)

func makeBaseChannelProposalAcc(
	proposalID ProposalID,
	nonceShare NonceShare,
) BaseChannelProposalAcc {
	return BaseChannelProposalAcc{
		ProposalID: proposalID,
		NonceShare: nonceShare,
	}
}

// Encode encodes a BaseChannelProposalAcc.
func (acc BaseChannelProposalAcc) Encode(w io.Writer) error {
	return perunio.Encode(w,
		acc.ProposalID,
		acc.NonceShare)
}

// Decode decodes a BaseChannelProposalAcc.
func (acc *BaseChannelProposalAcc) Decode(r io.Reader) error {
	return perunio.Decode(r,
		&acc.ProposalID,
		&acc.NonceShare)
}

// Type returns wire.ChannelProposalAcc.
func (LedgerChannelProposalAccMsg) Type() wire.Type {
	return wire.LedgerChannelProposalAcc
}

// Base returns the common proposal accept values.
func (acc *LedgerChannelProposalAccMsg) Base() *BaseChannelProposalAcc {
	return &acc.BaseChannelProposalAcc
}

// Encode encodes the LedgerChannelProposalAcc into an io.Writer.
func (acc LedgerChannelProposalAccMsg) Encode(w io.Writer) error {
	return perunio.Encode(w,
		acc.BaseChannelProposalAcc,
		wallet.AddressDecMap(acc.Participant))
}

// Decode decodes a LedgerChannelProposalAcc from an io.Reader.
func (acc *LedgerChannelProposalAccMsg) Decode(r io.Reader) error {
	return perunio.Decode(r,
		&acc.BaseChannelProposalAcc,
		(*wallet.AddressDecMap)(&acc.Participant))
}

// Type returns wire.SubChannelProposalAcc.
func (SubChannelProposalAccMsg) Type() wire.Type {
	return wire.SubChannelProposalAcc
}

// Base returns the common proposal accept values.
func (acc *SubChannelProposalAccMsg) Base() *BaseChannelProposalAcc {
	return &acc.BaseChannelProposalAcc
}

// Encode encodes the SubChannelProposalAcc into an io.Writer.
func (acc SubChannelProposalAccMsg) Encode(w io.Writer) error {
	return perunio.Encode(w, acc.BaseChannelProposalAcc)
}

// Decode decodes a SubChannelProposalAcc from an io.Reader.
func (acc *SubChannelProposalAccMsg) Decode(r io.Reader) error {
	return perunio.Decode(r, &acc.BaseChannelProposalAcc)
}

// ChannelProposalRejMsg is used to reject a ChannelProposalReq.
// An optional reason for the rejection can be set.
//
// The message is one of two possible responses in the
// Multi-Party Channel Proposal Protocol (MPCPP).
//
// Reason should be a UTF-8 encodable string.
type ChannelProposalRejMsg struct {
	ProposalID ProposalID // The channel proposal to reject.
	Reason     string     // The rejection reason.
}

// Type returns wire.ChannelProposalRej.
func (ChannelProposalRejMsg) Type() wire.Type {
	return wire.ChannelProposalRej
}

// Encode encodes a ChannelProposalRej into an io.Writer.
func (rej ChannelProposalRejMsg) Encode(w io.Writer) error {
	return perunio.Encode(w, rej.ProposalID, rej.Reason)
}

// Decode decodes a ChannelProposalRej from an io.Reader.
func (rej *ChannelProposalRejMsg) Decode(r io.Reader) error {
	return perunio.Decode(r, &rej.ProposalID, &rej.Reason)
}

/*
Virtual channels
*/

type (
	// VirtualChannelProposalMsg is a channel proposal for virtual channels.
	VirtualChannelProposalMsg struct {
		BaseChannelProposal
		Proposer  map[wallet.BackendID]wallet.Address // Proposer's address in the channel.
		Peers     []map[wallet.BackendID]wire.Address // Participants' wire addresses.
		Parents   []map[wallet.BackendID]channel.ID   // Parent channels for each participant.
		IndexMaps [][]channel.Index                   // Index mapping for each participant in relation to the root channel.
	}

	// VirtualChannelProposalAccMsg is the accept message type corresponding to
	// virtual channel proposals.
	VirtualChannelProposalAccMsg struct {
		BaseChannelProposalAcc
		Responder map[wallet.BackendID]wallet.Address // Responder's participant address.
	}
)

// NewVirtualChannelProposal creates a virtual channel proposal.
func NewVirtualChannelProposal(
	challengeDuration uint64,
	participant map[wallet.BackendID]wallet.Address,
	initBals *channel.Allocation,
	peers []map[wallet.BackendID]wire.Address,
	parents []map[wallet.BackendID]channel.ID,
	indexMaps [][]channel.Index,
	opts ...ProposalOpts,
) (prop *VirtualChannelProposalMsg, err error) {
	base, err := makeBaseChannelProposal(
		challengeDuration,
		initBals,
		opts...,
	)
	if err != nil {
		return
	}
	prop = &VirtualChannelProposalMsg{
		BaseChannelProposal: base,
		Proposer:            participant,
		Peers:               peers,
		Parents:             parents,
		IndexMaps:           indexMaps,
	}
	return
}

// Encode encodes the proposal into an io.Writer.
func (p VirtualChannelProposalMsg) Encode(w io.Writer) error {
	return perunio.Encode(
		w,
		p.BaseChannelProposal,
		wallet.AddressDecMap(p.Proposer),
		wire.AddressMapArray(p.Peers),
		channelIDsWithLen(p.Parents),
		indexMapsWithLen(p.IndexMaps),
	)
}

// Decode decodes a proposal from an io.Reader.
func (p *VirtualChannelProposalMsg) Decode(r io.Reader) error {
	return perunio.Decode(
		r,
		&p.BaseChannelProposal,
		(*wallet.AddressDecMap)(&p.Proposer),
		(*wire.AddressMapArray)(&p.Peers),
		(*channelIDsWithLen)(&p.Parents),
		(*indexMapsWithLen)(&p.IndexMaps),
	)
}

// Type returns the message type.
func (VirtualChannelProposalMsg) Type() wire.Type {
	return wire.VirtualChannelProposal
}

// Accept constructs an accept message that belongs to a proposal message.
func (p VirtualChannelProposalMsg) Accept(
	responder map[wallet.BackendID]wallet.Address,
	opts ...ProposalOpts,
) *VirtualChannelProposalAccMsg {
	_opts := union(opts...)
	return &VirtualChannelProposalAccMsg{
		BaseChannelProposalAcc: makeBaseChannelProposalAcc(p.ProposalID, _opts.nonce()),
		Responder:              responder,
	}
}

// Matches requires that the accept message has the correct type.
func (VirtualChannelProposalMsg) Matches(acc ChannelProposalAccept) bool {
	_, ok := acc.(*VirtualChannelProposalAccMsg)
	return ok
}

// Type returns the message type.
func (VirtualChannelProposalAccMsg) Type() wire.Type {
	return wire.VirtualChannelProposalAcc
}

// Base returns the common proposal accept values.
func (acc *VirtualChannelProposalAccMsg) Base() *BaseChannelProposalAcc {
	return &acc.BaseChannelProposalAcc
}

// Encode encodes the message into an io.Writer.
func (acc VirtualChannelProposalAccMsg) Encode(w io.Writer) error {
	return perunio.Encode(w, acc.BaseChannelProposalAcc, wallet.AddressDecMap(acc.Responder))
}

// Decode decodes a message from an io.Reader.
func (acc *VirtualChannelProposalAccMsg) Decode(r io.Reader) error {
	return perunio.Decode(r, &acc.BaseChannelProposalAcc, (*wallet.AddressDecMap)(&acc.Responder))
}

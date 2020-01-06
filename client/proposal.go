// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package client

import (
	"context"
	"math/big"

	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/log"
	"perun.network/go-perun/peer"
	"perun.network/go-perun/pkg/sync/atomic"
	"perun.network/go-perun/wallet"
	wire "perun.network/go-perun/wire/msg"
)

type (
	// ChannelProposal contains all data necessary to propose a new
	// channel to a given set of peers.
	//
	// This is the same as ChannelProposalMsg but with an account instead of only
	// the address of the proposer. ChannelProposal is not sent over the wire.
	ChannelProposal struct {
		ChallengeDuration uint64
		Nonce             *big.Int
		Account           wallet.Account // local account to use when creating this channel
		AppDef            wallet.Address
		InitData          channel.Data
		InitBals          *channel.Allocation
		PeerAddrs         []wallet.Address // Perun addresses of all peers, including the proposer's
	}

	// A ProposalHandler decides how to handle incoming channel proposals from
	// other channel network peers.
	ProposalHandler interface {
		// Handle is the user callback called by the Client on an incoming channel
		// proposal.
		Handle(*ChannelProposalReq, *ProposalResponder)
	}

	// ProposalResponder lets the user respond to a channel proposal. If the user
	// wants to accept the proposal, they should call Accept(), otherwise Reject().
	// Only a single function must be called and every further call causes a
	// panic.
	ProposalResponder struct {
		client *Client
		peer   *peer.Peer
		req    *ChannelProposalReq
		called atomic.Bool
	}

	// ProposalAcc is the proposal acceptance struct that the user passes to
	// ProposalResponder.Accept() when they want to accept an incoming channel
	// proposal.
	ProposalAcc struct {
		Participant wallet.Account
	}
)

// Accept lets the user signal that they want to accept the channel proposal.
// Returns whether the acceptance message was successfully sent. Panics if the
// proposal was already accepted or rejected.
func (r *ProposalResponder) Accept(ctx context.Context, acc ProposalAcc) (*Channel, error) {
	if !r.called.TrySet() {
		log.Panic("multiple calls on proposal responder")
	}
	if ctx == nil {
		log.Panic("nil context")
	}

	return r.client.handleChannelProposalAcc(ctx, r.peer, r.req, acc)
}

// Reject lets the user signal that they reject the channel proposal.
// Returns whether the rejection message was successfully sent. Panics if the
// proposal was already accepted or rejected.
func (r *ProposalResponder) Reject(ctx context.Context, reason string) error {
	if !r.called.TrySet() {
		log.Panic("multiple calls on proposal responder")
	}
	if ctx == nil {
		log.Panic("nil context")
	}

	return r.client.handleChannelProposalRej(ctx, r.peer, r.req, reason)
}

// ProposeChannel attempts to open a channel with the parameters and peers from
// ChannelProposal prop:
// - the proposal is sent to the peers and if all peers accept,
// - the channel is funded. If successful,
// - the channel controller is returned.
// The user is required to start the update handler with
// Channel.ListenUpdates(UpdateHandler)
func (c *Client) ProposeChannel(ctx context.Context, prop *ChannelProposal) (*Channel, error) {
	if ctx == nil || prop == nil {
		c.log.Panic("invalid nil argument")
	}

	// 1. check valid proposal
	req := prop.AsReq()
	if err := c.validTwoPartyProposal(req, 0, req.PeerAddrs[1]); err != nil {
		return nil, errors.WithMessage(err, "invalid channel proposal")
	}

	// 2. send proposal and wait for response
	parts, err := c.exchangeTwoPartyProposal(ctx, req)
	if err != nil {
		return nil, errors.WithMessage(err, "sending proposal")
	}

	// 3. create params, channel machine from gathered participant addresses
	// 4. fund channel
	// 5. return controller on successful funding
	return c.setupChannel(ctx, prop, parts)
}

// This function is called during the setup of new peers by the registry. The
// passed peer is not yet receiving any messages, thus, subscription is
// race-free. After the function returns, the peer starts receiving messages.
func (c *Client) subChannelProposals(p *peer.Peer) {
	proposalReceiver := peer.NewReceiver()
	if err := p.Subscribe(proposalReceiver,
		func(m wire.Msg) bool { return m.Type() == wire.ChannelProposal },
	); err != nil {
		c.logPeer(p).Errorf("failed to subscribe to channel proposals on new peer: %v", err)
		proposalReceiver.Close()
		return
	}

	// Aborts the proposal handler loop when the Peer is closed.
	p.OnCloseAlways(func() {
		if err := proposalReceiver.Close(); err != nil {
			c.logPeer(p).Errorf("failed to close proposal receiver: %v", err)
		}
	})

	// proposal handler loop.
	go func() {
		for {
			_p, m := proposalReceiver.Next(context.Background())
			if _p == nil {
				c.logPeer(p).Debug("proposal subscription closed")
				return
			}
			proposal := m.(*ChannelProposalReq) // safe because that's the predicate
			go c.handleChannelProposal(p, proposal)
		}
	}()
}

// handleChannelProposal implements the receiving side of the (currently)
// two-party channel proposal protocol.
// The proposer is expected to be the first peer in the participant list.
func (c *Client) handleChannelProposal(p *peer.Peer, req *ChannelProposalReq) {
	if err := c.validTwoPartyProposal(req, 1, p.PerunAddress); err != nil {
		c.logPeer(p).Debugf("received invalid channel proposal: %v", err)
		return
	}

	c.logPeer(p).Trace("calling proposal handler")
	responder := &ProposalResponder{client: c, peer: p, req: req}
	c.propHandler.Handle(req, responder)
}

func (c *Client) handleChannelProposalAcc(
	ctx context.Context, p *peer.Peer,
	req *ChannelProposalReq, acc ProposalAcc,
) (*Channel, error) {
	if acc.Participant == nil {
		c.logPeer(p).Error("user returned nil Participant in ProposalAcc")
		return nil, errors.New("nil Participant in ProposalAcc")
	}

	// enables caching of incoming version 0 signatures before sending any message
	// that might trigger a fast peer to send those. We don't know the channel id
	// yet so the cache predicate is coarser than the later subscription.
	enableVer0Cache(ctx, p)

	msgAccept := &ChannelProposalAcc{
		SessID:          req.SessID(),
		ParticipantAddr: acc.Participant.Address(),
	}
	if err := p.Send(ctx, msgAccept); err != nil {
		c.logPeer(p).Errorf("error sending proposal acceptance: %v", err)
		return nil, errors.WithMessage(err, "sending proposal acceptance")
	}

	parts := []wallet.Address{req.ParticipantAddr, acc.Participant.Address()}
	return c.setupChannel(ctx, req.AsProp(acc.Participant), parts)
}

func (c *Client) handleChannelProposalRej(
	ctx context.Context, p *peer.Peer,
	req *ChannelProposalReq, reason string,
) error {
	msgReject := &ChannelProposalRej{
		SessID: req.SessID(),
		Reason: reason,
	}
	if err := p.Send(ctx, msgReject); err != nil {
		c.logPeer(p).Warn("error sending proposal rejection")
		return err
	}
	return nil
}

// exchangeTwoPartyProposal implements the multi-party channel proposal
// protocol for the two-party case.
func (c *Client) exchangeTwoPartyProposal(
	ctx context.Context,
	proposal *ChannelProposalReq,
) ([]wallet.Address, error) {
	p, err := c.peers.Get(ctx, proposal.PeerAddrs[1])
	if err != nil {
		return nil, errors.WithMessage(err, "failed to Get() participant[1]")
	}

	// enables caching of incoming version 0 signatures before sending any message
	// that might trigger a fast peer to send those. We don't know the channel id
	// yet so the cache predicate is coarser than the later subscription.
	enableVer0Cache(ctx, p)

	sessID := proposal.SessID()
	isResponse := func(m wire.Msg) bool {
		return (m.Type() == wire.ChannelProposalAcc &&
			m.(*ChannelProposalAcc).SessID == sessID) ||
			(m.Type() == wire.ChannelProposalRej &&
				m.(*ChannelProposalRej).SessID == sessID)
	}
	receiver := peer.NewReceiver()
	defer receiver.Close()

	if err := p.Subscribe(receiver, isResponse); err != nil {
		return nil, errors.WithMessagef(err, "subscribing peer %v", p)
	}

	if err := p.Send(ctx, proposal); err != nil {
		return nil, errors.WithMessage(err, "channel proposal broadcast")
	}

	_, rawResponse := receiver.Next(ctx)
	if rawResponse == nil {
		return nil, errors.New("timeout when waiting for proposal response")
	}
	if rej, ok := rawResponse.(*ChannelProposalRej); ok {
		return nil, errors.Errorf("channel proposal rejected: %v", rej.Reason)
	}

	acc := rawResponse.(*ChannelProposalAcc) // this is safe because of predicate isResponse
	return []wallet.Address{proposal.ParticipantAddr, acc.ParticipantAddr}, nil
}

// validTwoPartyProposal checks that the proposal is valid in the two-party
// setting, where the proposer is expected to have index 0 in the peer list and
// the receiver to have index 1. The generic validity of the proposal is also
// checked.
func (c *Client) validTwoPartyProposal(
	proposal *ChannelProposalReq,
	ourIdx int,
	peerAddr wallet.Address,
) error {
	if err := proposal.Valid(); err != nil {
		return err
	}

	if len(proposal.PeerAddrs) != 2 {
		return errors.Errorf("exptected 2 peers, got %d", len(proposal.PeerAddrs))
	}

	peerIdx := ourIdx ^ 1
	// In the 2PCPP, the proposer is expected to have index 0
	if !proposal.PeerAddrs[peerIdx].Equals(peerAddr) {
		return errors.Errorf("remote peer doesn't have peer index %d", peerIdx)
	}

	// In the 2PCPP, the receiver is expected to have index 1
	if !proposal.PeerAddrs[ourIdx].Equals(c.id.Address()) {
		return errors.Errorf("we don't have peer index %d", ourIdx)
	}

	return nil
}

// setupChannel sets up a new channel controller for the given proposal and
// participant addresses, using the account for our participant. The parameters
// are assembled and the channel controller is started. The channel will be
// funded and if successful, the *Channel is returned. It does not perform a
// validity check on the proposal, so make sure to only paste valid proposals.
func (c *Client) setupChannel(
	ctx context.Context,
	prop *ChannelProposal,
	parts []wallet.Address, // result of the MPCPP on prop
) (*Channel, error) {
	params := channel.NewParamsUnsafe(prop.ChallengeDuration, parts, prop.AppDef, prop.Nonce)

	peers, err := c.getPeers(ctx, prop.PeerAddrs)
	if err != nil {
		return nil, errors.WithMessage(err, "getting peers from the registry")
	}

	ch, err := newChannel(prop.Account, peers, *params, c.settler)
	if err != nil {
		return nil, err
	}
	ch.setLogger(c.logChan(params.ID()))
	if err := ch.init(prop.InitBals, prop.InitData); err != nil {
		return ch, errors.WithMessage(err, "setting initial bals and data")
	}
	if err := ch.initExchangeSigsAndEnable(ctx); err != nil {
		return ch, errors.WithMessage(err, "exchanging initial sigs and enabling state")
	}

	if err = c.funder.Fund(ctx,
		channel.FundingReq{
			Params:     params,
			Allocation: prop.InitBals,
			Idx:        ch.machine.Idx(),
		}); channel.IsPeerTimedOutFundingError(err) {
		// TODO: initiate dispute and withdrawal
		ch.log.Warnf("error while funding channel: %v", err)
		return ch, errors.WithMessage(err, "error while funding channel")
	} else if err != nil { // other runtime error
		ch.log.Warnf("error while funding channel: %v", err)
		return ch, errors.WithMessage(err, "error while funding channel")
	}

	return ch, ch.machine.SetFunded()
}

// enableVer0Cache enables caching of incoming version 0 signatures
func enableVer0Cache(ctx context.Context, c wire.Cacher) {
	c.Cache(ctx, func(m wire.Msg) bool {
		return m.Type() == wire.ChannelUpdateAcc &&
			m.(*msgChannelUpdateAcc).Version == 0
	})
}

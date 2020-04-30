// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package client

import (
	"context"

	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/log"
	"perun.network/go-perun/peer"
	"perun.network/go-perun/pkg/sync/atomic"
	"perun.network/go-perun/wallet"
	wire "perun.network/go-perun/wire/msg"
)

type (
	// A ProposalHandler decides how to handle incoming channel proposals from
	// other channel network peers.
	ProposalHandler interface {
		// Handle is the user callback called by the Client on an incoming channel
		// proposal.
		Handle(*ChannelProposal, *ProposalResponder)
	}

	// ProposalResponder lets the user respond to a channel proposal. If the user
	// wants to accept the proposal, they should call Accept(), otherwise Reject().
	// Only a single function must be called and every further call causes a
	// panic.
	ProposalResponder struct {
		client *Client
		peer   *peer.Peer
		req    *ChannelProposal
		called atomic.Bool
	}

	// ProposalAcc is the proposal acceptance struct that the user passes to
	// ProposalResponder.Accept() when they want to accept an incoming channel
	// proposal.
	ProposalAcc struct {
		Participant wallet.Address
	}
)

// Accept lets the user signal that they want to accept the channel proposal.
// Returns whether the acceptance message was successfully sent. Panics if the
// proposal was already accepted or rejected.
func (r *ProposalResponder) Accept(ctx context.Context, acc ProposalAcc) (*Channel, error) {
	if ctx == nil {
		return nil, errors.New("context must not be nil")
	}

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
func (c *Client) ProposeChannel(ctx context.Context, req *ChannelProposal) (*Channel, error) {
	if ctx == nil || req == nil {
		c.log.Panic("invalid nil argument")
	}

	// 1. check valid proposal
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
	return c.setupChannel(ctx, req, parts)
}

// This function is called during the setup of new peers by the registry. The
// passed peer is not yet receiving any messages, thus, subscription is
// race-free. After the function returns, the peer starts receiving messages.
func (c *Client) subChannelProposals(p *peer.Peer) {
	if err := p.Subscribe(c.propRecv,
		func(m wire.Msg) bool { return m.Type() == wire.ChannelProposal },
	); err != nil {
		c.logPeer(p).Errorf("failed to subscribe to channel proposals on new peer: %v", err)
		return
	}

}

// HandleChannelProposals is the incoming channel proposal handler routine. It
// must only be started at most once by the user. Incoming channel proposals are
// handled using the passed handler.
func (c *Client) HandleChannelProposals(handler ProposalHandler) {
	for {
		p, m := c.propRecv.Next(context.Background())
		if p == nil {
			c.log.Debug("proposal receiver closed")
			return
		}
		req := m.(*ChannelProposal) // safe because that's the predicate
		go c.handleChannelProposal(handler, p, req)
	}
}

// handleChannelProposal implements the receiving side of the (currently)
// two-party channel proposal protocol.
// The proposer is expected to be the first peer in the participant list.
func (c *Client) handleChannelProposal(
	handler ProposalHandler, p *peer.Peer, req *ChannelProposal) {
	if err := c.validTwoPartyProposal(req, 1, p.PerunAddress); err != nil {
		c.logPeer(p).Debugf("received invalid channel proposal: %v", err)
		return
	}

	c.logPeer(p).Trace("calling proposal handler")
	responder := &ProposalResponder{client: c, peer: p, req: req}
	handler.Handle(req, responder)
	// control flow continues in responder.Accept/Reject
}

func (c *Client) handleChannelProposalAcc(
	ctx context.Context, p *peer.Peer,
	req *ChannelProposal, acc ProposalAcc,
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
		ParticipantAddr: acc.Participant,
	}
	if err := p.Send(ctx, msgAccept); err != nil {
		c.logPeer(p).Errorf("error sending proposal acceptance: %v", err)
		return nil, errors.WithMessage(err, "sending proposal acceptance")
	}

	// In the 2-party case, we hardcode the proposer to index 0 and responder to 1.
	parts := []wallet.Address{req.ParticipantAddr, acc.Participant}
	// Change ParticipantAddr to own address because setupChannel reads own
	// address from this field. The ChannelProposal is consumed by setupChannel so
	// there's no harm in changing it.
	req.ParticipantAddr = acc.Participant
	return c.setupChannel(ctx, req, parts)
}

func (c *Client) handleChannelProposalRej(
	ctx context.Context, p *peer.Peer,
	req *ChannelProposal, reason string,
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
	proposal *ChannelProposal,
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
	proposal *ChannelProposal,
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
// participant addresses, using the wallet to unlock the account for our
// participant. The own participant is chosen as prop.ParticipantAddr, so make
// sure that this field has been set correctly and not to the initial proposer.
//
// The parameters are assembled and the initial state with signatures is
// exchanged. The channel will be funded and if successful, the channel
// controller is returned.
//
// It does not perform a validity check on the proposal, so make sure to only
// pass valid proposals.
func (c *Client) setupChannel(
	ctx context.Context,
	prop *ChannelProposal,
	parts []wallet.Address, // result of the MPCPP on prop
) (*Channel, error) {
	params := channel.NewParamsUnsafe(prop.ChallengeDuration, parts, prop.AppDef, prop.Nonce)
	if c.channels.Has(params.ID()) {
		return nil, errors.New("channel already exists")
	}

	peers, err := c.getPeers(ctx, prop.PeerAddrs)
	if err != nil {
		return nil, errors.WithMessage(err, "getting peers from the registry")
	}
	acc, err := c.wallet.Unlock(prop.ParticipantAddr)
	if err != nil {
		return nil, errors.WithMessage(err, "unlocking account")
	}

	ch, err := c.newChannel(acc, peers, *params)
	if err != nil {
		return nil, err
	}

	if err := c.pr.ChannelCreated(ctx, ch.machine, prop.PeerAddrs); err != nil {
		return ch, errors.WithMessage(err, "persisting new channel")
	}

	if err := ch.init(ctx, prop.InitBals, prop.InitData); err != nil {
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
		}); channel.IsFundingTimeoutError(err) {
		ch.log.Warnf("Peers timed out funding channel(%v); settling...", err)
		serr := ch.Settle(ctx)
		return ch, errors.WithMessagef(err,
			"peers timed out funding (subsequent settlement error: %v)", serr)
	} else if err != nil { // other runtime error
		ch.log.Warnf("error while funding channel: %v", err)
		return ch, errors.WithMessage(err, "error while funding channel")
	}

	if err := ch.machine.SetFunded(ctx); err != nil {
		return ch, errors.WithMessage(err, "error in SetFunded()")
	}
	if !c.channels.Put(params.ID(), ch) {
		return ch, errors.New("channel already exists")
	}
	c.wallet.IncrementUsage(acc.Address())

	return ch, nil
}

// enableVer0Cache enables caching of incoming version 0 signatures
func enableVer0Cache(ctx context.Context, c wire.Cacher) {
	c.Cache(ctx, func(m wire.Msg) bool {
		return m.Type() == wire.ChannelUpdateAcc &&
			m.(*msgChannelUpdateAcc).Version == 0
	})
}

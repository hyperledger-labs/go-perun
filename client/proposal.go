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

	ProposalHandler interface {
		Handle(*ChannelProposalReq, *ProposalResponder)
	}

	// ProposalResponder lets the user respond to a channel proposal. If the user
	// wants to accept the proposal, they should call Accept(), otherwise Reject().
	// Only a single function must be called and every further call causes a
	// panic.
	ProposalResponder struct {
		accept    chan ctxProposalAcc
		acceptRet chan acceptRet // Accept returns (*Channel, error)
		reject    chan ctxProposalRej
		rejectRet chan error // Reject returns error
		called    atomic.Bool
	}

	ProposalAcc struct {
		Participant wallet.Account
		// TODO add UpdateHandler
	}

	// The following type is only needed to bundle the ctx and res of
	// ProposalResponder.Accept() into a single struct so that they can be sent
	// over a go channel
	ctxProposalAcc struct {
		ctx context.Context
		ProposalAcc
	}

	// acceptRet is needed to bundle the return variable of
	// ProposalResponder.Accept so they can be sent over a go channel
	acceptRet struct {
		ch  *Channel
		err error
	}

	// The following type is only needed to bundle the ctx and reason of
	// ProposalResponder.Reject() into a single struct so that they can be sent
	// over a go channel
	ctxProposalRej struct {
		ctx    context.Context
		reason string
	}
)

func newProposalResponder() *ProposalResponder {
	return &ProposalResponder{
		accept:    make(chan ctxProposalAcc),
		acceptRet: make(chan acceptRet, 1),
		reject:    make(chan ctxProposalRej),
		rejectRet: make(chan error, 1),
	}
}

// Accept lets the user signal that they want to accept the channel proposal.
// Returns whether the acceptance message was successfully sent. Panics if the
// proposal was already accepted or rejected.
func (r *ProposalResponder) Accept(ctx context.Context, res ProposalAcc) (*Channel, error) {
	if !r.called.TrySet() {
		log.Panic("multiple calls on proposal responder")
	}
	r.accept <- ctxProposalAcc{ctx, res}
	ret := <-r.acceptRet
	return ret.ch, ret.err
}

// Reject lets the user signal that they reject the channel proposal.
// Returns whether the rejection message was successfully sent. Panics if the
// proposal was already accepted or rejected.
func (r *ProposalResponder) Reject(ctx context.Context, reason string) error {
	if !r.called.TrySet() {
		log.Panic("multiple calls on proposal responder")
	}
	r.reject <- ctxProposalRej{ctx, reason}
	return <-r.rejectRet
}

// ProposeChannel attempts to open a channel witht the parameters and peers from
// ChannelProposal prop:
// - the proposal is sent to the peers and if all peers accept,
// - the channel is funded. If successful,
// - the channel controller is returned.
// The user is required to start the update handler with
// Channel.ListenUpdates(UpdateHandler)
func (c *Client) ProposeChannel(ctx context.Context, prop *ChannelProposal) (*Channel, error) {
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
		c.logPeer(p).Errorf("failed to subscribe to channel proposals on new peer")
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
				c.logPeer(p).Debugf("proposal subscription closed")
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

	responder := newProposalResponder()
	go c.propHandler.Handle(req, responder)

	// wait for user response
	select {
	case acc := <-responder.accept:
		if acc.Participant == nil {
			c.logPeer(p).Error("user returned nil Participant in ProposalAcc")
			responder.acceptRet <- acceptRet{nil, errors.New("nil Participant in ProposalAcc")}
			return
		}

		msgAccept := &ChannelProposalAcc{
			SessID:          req.SessID(),
			ParticipantAddr: acc.Participant.Address(),
		}
		if err := p.Send(acc.ctx, msgAccept); err != nil {
			c.logPeer(p).Errorf("error sending proposal acceptance: %v", err)
			responder.acceptRet <- acceptRet{nil, errors.WithMessage(err, "sending proposal acceptance")}
			return
		}

		parts := []wallet.Address{req.ParticipantAddr, acc.Participant.Address()}
		ch, err := c.setupChannel(acc.ctx, req.AsProp(acc.Participant), parts)
		if err != nil {
			c.logPeer(p).Errorf("error setting up channel controller: %v", err)
		}
		responder.acceptRet <- acceptRet{ch, err}
		return

	case rej := <-responder.reject:
		msgReject := &ChannelProposalRej{
			SessID: req.SessID(),
			Reason: rej.reason,
		}
		if err := p.Send(rej.ctx, msgReject); err != nil {
			c.logPeer(p).Warn("error sending proposal rejection")
			responder.rejectRet <- err
			return
		}
		responder.rejectRet <- nil
		return
	}
}

func (c *Client) exchangeTwoPartyProposal(
	ctx context.Context,
	proposal *ChannelProposalReq,
) ([]wallet.Address, error) {
	p, err := c.peers.Get(ctx, proposal.PeerAddrs[1])
	if err != nil {
		return nil, errors.WithMessage(err, "failed to Get() participant[1]")
	}

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

	ch, err := newChannel(prop.Account, peers, *params)
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

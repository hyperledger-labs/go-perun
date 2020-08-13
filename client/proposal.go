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
	"context"

	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/log"
	"perun.network/go-perun/pkg/sync/atomic"
	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wire"
)

const proposerIdx, proposeeIdx = 0, 1

type (
	// A ProposalHandler decides how to handle incoming channel proposals from
	// other channel network peers.
	ProposalHandler interface {
		// HandleProposal is the user callback called by the Client on an incoming channel
		// proposal.
		HandleProposal(*ChannelProposal, *ProposalResponder)
	}

	// ProposalHandlerFunc is an adapter type to allow the use of functions as
	// proposal handlers. ProposalHandlerFunc(f) is a ProposalHandler that calls
	// f when HandleProposal is called.
	ProposalHandlerFunc func(*ChannelProposal, *ProposalResponder)

	// ProposalResponder lets the user respond to a channel proposal. If the user
	// wants to accept the proposal, they should call Accept(), otherwise Reject().
	// Only a single function must be called and every further call causes a
	// panic.
	ProposalResponder struct {
		client *Client
		peer   wire.Address
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

// HandleProposal calls the proposal handler function.
func (f ProposalHandlerFunc) HandleProposal(p *ChannelProposal, r *ProposalResponder) { f(p, r) }

// Accept lets the user signal that they want to accept the channel proposal.
// Returns the newly created channel controller if the channel was successfully
// created and funded. Panics if the proposal was already accepted or rejected.
//
// After the channel got successfully created, the user is required to start the
// update handler with Channel.ListenUpdates(UpdateHandler) and to start the
// channel watcher with Channel.Watch() on the returned channel controller.
//
// It is important that the passed context does not cancel before twice the
// ChallengeDuration has passed (at least for real blockchain backends with wall
// time), or the channel cannot be settled if a peer times out funding.
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
//
// After the channel got successfully created, the user is required to start the
// update handler with Channel.ListenUpdates(UpdateHandler) and to start the
// channel watcher with Channel.Watch(context.Context) on the returned channel
// controller.
//
// It is important that the passed context does not cancel before twice the
// ChallengeDuration has passed (at least for real blockchain backends with wall
// time), or the channel cannot be settled if a peer times out funding.
func (c *Client) ProposeChannel(ctx context.Context, req *ChannelProposal) (*Channel, error) {
	if ctx == nil || req == nil {
		c.log.Panic("invalid nil argument")
	}

	// 1. check valid proposal
	if err := c.validTwoPartyProposal(req, proposerIdx, req.PeerAddrs[proposeeIdx]); err != nil {
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
	return c.setupChannel(ctx, req, parts, proposerIdx)
}

// handleChannelProposal implements the receiving side of the (currently)
// two-party channel proposal protocol.
// The proposer is expected to be the first peer in the participant list.
//
// This handler is dispatched from the Client.Handle routine.
func (c *Client) handleChannelProposal(
	handler ProposalHandler, p wire.Address, req *ChannelProposal) {
	if err := c.validTwoPartyProposal(req, proposeeIdx, p); err != nil {
		c.logPeer(p).Debugf("received invalid channel proposal: %v", err)
		return
	}

	c.logPeer(p).Trace("calling proposal handler")
	responder := &ProposalResponder{client: c, peer: p, req: req}
	handler.HandleProposal(req, responder)
	// control flow continues in responder.Accept/Reject
}

func (c *Client) handleChannelProposalAcc(
	ctx context.Context, p wire.Address,
	req *ChannelProposal, acc ProposalAcc,
) (*Channel, error) {
	if acc.Participant == nil {
		c.logPeer(p).Error("user returned nil Participant in ProposalAcc")
		return nil, errors.New("nil Participant in ProposalAcc")
	}

	// enables caching of incoming version 0 signatures before sending any message
	// that might trigger a fast peer to send those. We don't know the channel id
	// yet so the cache predicate is coarser than the later subscription.
	enableVer0Cache(ctx, c.conn)

	msgAccept := &ChannelProposalAcc{
		ProposalID:      req.ProposalID(),
		ParticipantAddr: acc.Participant,
	}
	if err := c.conn.pubMsg(ctx, msgAccept, p); err != nil {
		c.logPeer(p).Errorf("error sending proposal acceptance: %v", err)
		return nil, errors.WithMessage(err, "sending proposal acceptance")
	}

	parts := participants(req.ParticipantAddr, acc.Participant)
	return c.setupChannel(ctx, req, parts, proposeeIdx)
}

func (c *Client) handleChannelProposalRej(
	ctx context.Context, p wire.Address,
	req *ChannelProposal, reason string,
) error {
	msgReject := &ChannelProposalRej{
		ProposalID: req.ProposalID(),
		Reason:     reason,
	}
	if err := c.conn.pubMsg(ctx, msgReject, p); err != nil {
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
	peer := proposal.PeerAddrs[proposeeIdx]

	// enables caching of incoming version 0 signatures before sending any message
	// that might trigger a fast peer to send those. We don't know the channel id
	// yet so the cache predicate is coarser than the later subscription.
	enableVer0Cache(ctx, c.conn)

	proposalID := proposal.ProposalID()
	isResponse := func(e *wire.Envelope) bool {
		return (e.Msg.Type() == wire.ChannelProposalAcc &&
			e.Msg.(*ChannelProposalAcc).ProposalID == proposalID) ||
			(e.Msg.Type() == wire.ChannelProposalRej &&
				e.Msg.(*ChannelProposalRej).ProposalID == proposalID)
	}
	receiver := wire.NewReceiver()
	// nolint:errcheck
	defer receiver.Close()

	if err := c.conn.Subscribe(receiver, isResponse); err != nil {
		return nil, errors.WithMessage(err, "subscribing proposal response recv")
	}

	if err := c.conn.pubMsg(ctx, proposal, peer); err != nil {
		return nil, errors.WithMessage(err, "publishing channel proposal")
	}

	env, err := receiver.Next(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "receiving proposal response")
	}
	if rej, ok := env.Msg.(*ChannelProposalRej); ok {
		return nil, errors.Errorf("channel proposal rejected: %v", rej.Reason)
	}

	acc := env.Msg.(*ChannelProposalAcc) // this is safe because of predicate isResponse
	return participants(proposal.ParticipantAddr, acc.ParticipantAddr), nil
}

func participants(proposer, proposee wallet.Address) []wallet.Address {
	parts := make([]wallet.Address, 2)
	parts[proposerIdx] = proposer
	parts[proposeeIdx] = proposee
	return parts
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
	if !proposal.PeerAddrs[ourIdx].Equals(c.address) {
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
//
// It is important that the passed context does not cancel before twice the
// ChallengeDuration has passed (at least for real blockchain backends with wall
// time), or the channel cannot be settled if a peer times out funding.
func (c *Client) setupChannel(
	ctx context.Context,
	prop *ChannelProposal,
	parts []wallet.Address, // result of the MPCPP on prop
	idx channel.Index, // our index
) (*Channel, error) {
	params := channel.NewParamsUnsafe(prop.ChallengeDuration, parts, prop.AppDef, prop.Nonce)
	if c.channels.Has(params.ID()) {
		return nil, errors.New("channel already exists")
	}

	acc, err := c.wallet.Unlock(parts[idx])
	if err != nil {
		return nil, errors.WithMessage(err, "unlocking account")
	}

	ch, err := c.newChannel(acc, prop.PeerAddrs, *params)
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
			Params: params,
			State:  ch.machine.State(), // initial state
			Idx:    ch.machine.Idx(),
		}); channel.IsFundingTimeoutError(err) {
		ch.Log().Warnf("Peers timed out funding channel(%v); settling...", err)
		serr := ch.Settle(ctx)
		return ch, errors.WithMessagef(err,
			"peers timed out funding (subsequent settlement error: %v)", serr)
	} else if err != nil { // other runtime error
		ch.Log().Warnf("error while funding channel: %v", err)
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

// enableVer0Cache enables caching of incoming version 0 signatures.
func enableVer0Cache(ctx context.Context, c wire.Cacher) {
	c.Cache(ctx, func(m *wire.Envelope) bool {
		return m.Msg.Type() == wire.ChannelUpdateAcc &&
			m.Msg.(*msgChannelUpdateAcc).Version == 0
	})
}

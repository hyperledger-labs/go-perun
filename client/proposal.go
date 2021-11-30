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
	"bytes"
	"context"
	"fmt"
	"sync"

	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/log"
	"perun.network/go-perun/pkg/io"
	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wire"
	pcontext "polycry.pt/poly-go/context"
	"polycry.pt/poly-go/sync/atomic"
)

const proposerIdx, proposeeIdx = 0, 1

// number of participants that is used unless specified otherwise.
const proposalNumParts = 2

type (
	// A ProposalHandler decides how to handle incoming channel proposals from
	// other channel network peers.
	ProposalHandler interface {
		// HandleProposal is the user callback called by the Client on an incoming channel
		// proposal.
		// The response on the proposal responder must be called within the same go routine.
		HandleProposal(ChannelProposal, *ProposalResponder)
	}

	// ProposalHandlerFunc is an adapter type to allow the use of functions as
	// proposal handlers. ProposalHandlerFunc(f) is a ProposalHandler that calls
	// f when HandleProposal is called.
	ProposalHandlerFunc func(ChannelProposal, *ProposalResponder)

	// ProposalResponder lets the user respond to a channel proposal. If the user
	// wants to accept the proposal, they should call Accept(), otherwise Reject().
	// Only a single function must be called and every further call causes a
	// panic.
	ProposalResponder struct {
		client *Client
		peer   wire.Address
		req    ChannelProposal
		called atomic.Bool
	}

	// PeerRejectedError indicates the channel proposal or channel update was
	// rejected by the peer.
	PeerRejectedError struct {
		ItemType string // ItemType indicates the type of item rejected (channel proposal or channel update).
		Reason   string // Reason sent by the peer for the rejection.
	}
)

// HandleProposal calls the proposal handler function.
func (f ProposalHandlerFunc) HandleProposal(p ChannelProposal, r *ProposalResponder) { f(p, r) }

// Accept lets the user signal that they want to accept the channel proposal.
// The ChannelProposalAcc message has to be created using
// ChannelProposal.Proposal().NewChannelProposalAcc on the proposal that was
// passed to the handler.
//
// Accept returns the newly created channel controller if the channel was
// successfully created and funded. Panics if the proposal was already accepted
// or rejected.
//
// After the channel controller got successfully set up, it is passed to the
// callback registered with Client.OnNewChannel. Accept returns after this
// callback has run.
//
// It is important that the passed context does not cancel before twice the
// ChallengeDuration has passed (at least for real blockchain backends with wall
// time), or the channel cannot be settled if a peer times out funding.
//
// After the channel got successfully created, the user is required to start the
// channel watcher with Channel.Watch() on the returned channel controller.
//
// Returns TxTimedoutError when the program times out waiting for a transaction
// to be mined.
// Returns ChainNotReachableError if the connection to the blockchain network
// fails when sending a transaction to / reading from the blockchain.
func (r *ProposalResponder) Accept(ctx context.Context, acc ChannelProposalAccept) (*Channel, error) {
	if ctx == nil {
		return nil, errors.New("context must not be nil")
	}

	if !r.called.TrySet() {
		log.Panic("multiple calls on proposal responder")
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
	return r.client.handleChannelProposalRej(ctx, r.peer, r.req, reason)
}

// ProposeChannel attempts to open a channel with the parameters and peers from
// ChannelProposal prop:
// - the proposal is sent to the peers and if all peers accept,
// - the channel is funded. If successful,
// - the channel controller is returned.
//
// After the channel controller got successfully set up, it is passed to the
// callback registered with Client.OnNewChannel. Accept returns after this
// callback has run.
//
// It is important that the passed context does not cancel before twice the
// ChallengeDuration has passed (at least for real blockchain backends with wall
// time), or the channel cannot be settled if a peer times out funding.
//
// After the channel got successfully created, the user is required to start the
// channel watcher with Channel.Watch() on the returned channel
// controller.
//
// Returns PeerRejectedProposalError if the channel is rejected by the peer.
// Returns RequestTimedOutError if the peer did not respond before the context
// expires or is cancelled.
// Returns FundingTimeoutError if any of the participants do not fund the
// channel in time.
// Returns TxTimedoutError when the program times out waiting for a transaction
// to be mined.
// Returns ChainNotReachableError if the connection to the blockchain network
// fails when sending a transaction to / reading from the blockchain.
func (c *Client) ProposeChannel(ctx context.Context, prop ChannelProposal) (*Channel, error) {
	if ctx == nil {
		c.log.Panic("invalid nil argument")
	}

	// Prepare and cleanup, e.g., for locking and unlocking parent channel.
	err := c.prepareChannelOpening(ctx, prop, proposerIdx)
	if err != nil {
		return nil, errors.WithMessage(err, "preparing channel opening")
	}
	defer c.cleanupChannelOpening(prop, proposerIdx)

	// 1. validate input
	peer := c.proposalPeers(prop)[proposeeIdx]
	if err := c.validTwoPartyProposal(prop, proposerIdx, peer); err != nil {
		return nil, errors.WithMessage(err, "invalid channel proposal")
	}

	// 2. send proposal, wait for response, create channel object
	c.enableVer1Cache()        // cache version 1 updates until channel is opened
	defer c.releaseVer1Cache() // replay cached version 1 updates
	ch, err := c.proposeTwoPartyChannel(ctx, prop)
	if err != nil {
		return nil, errors.WithMessage(err, "channel proposal")
	}

	// 3. fund
	fundingErr := c.fundChannel(ctx, ch, prop)
	return ch, fundingErr
}

func (c *Client) prepareChannelOpening(ctx context.Context, prop ChannelProposal, ourIdx channel.Index) (err error) {
	_, parentCh, err := c.proposalParent(prop, ourIdx)
	if err != nil {
		return
	}
	if parentCh != nil {
		if !parentCh.machMtx.TryLockCtx(ctx) {
			return ctx.Err()
		}
	}
	return
}

func (c *Client) cleanupChannelOpening(prop ChannelProposal, ourIdx channel.Index) {
	_, parentCh, err := c.proposalParent(prop, ourIdx)
	if err != nil {
		c.log.Warn("getting proposal parent:", err)
		return
	}
	if parentCh != nil {
		parentCh.machMtx.Unlock()
	}
}

// handleChannelProposal implements the receiving side of the (currently)
// two-party channel proposal protocol.
// The proposer is expected to be the first peer in the participant list.
//
// This handler is dispatched from the Client.Handle routine.
func (c *Client) handleChannelProposal(handler ProposalHandler, p wire.Address, req ChannelProposal) {
	ourIdx := channel.Index(proposeeIdx)

	// Prepare and cleanup, e.g., for locking and unlocking parent channel.
	err := c.prepareChannelOpening(c.Ctx(), req, ourIdx)
	if err != nil {
		c.log.Warn("preparing channel opening:", err)
		return
	}
	defer c.cleanupChannelOpening(req, ourIdx)

	if err := c.validTwoPartyProposal(req, ourIdx, p); err != nil {
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
	prop ChannelProposal, acc ChannelProposalAccept,
) (ch *Channel, err error) {
	if err := c.validChannelProposalAcc(prop, acc); err != nil {
		return ch, errors.WithMessage(err, "validating channel proposal acceptance")
	}

	c.enableVer1Cache()        // cache version 1 updates
	defer c.releaseVer1Cache() // replay cached version 1 updates

	if ch, err = c.acceptChannelProposal(ctx, prop, p, acc); err != nil {
		return ch, errors.WithMessage(err, "accept channel proposal")
	}

	fundingErr := c.fundChannel(ctx, ch, prop)
	return ch, fundingErr
}

func (c *Client) acceptChannelProposal(
	ctx context.Context,
	prop ChannelProposal,
	p wire.Address,
	acc ChannelProposalAccept,
) (*Channel, error) {
	if acc == nil {
		c.logPeer(p).Error("user passed nil ChannelProposalAcc")
		return nil, errors.New("nil ChannelProposalAcc")
	}

	// enables caching of incoming version 0 signatures before sending any message
	// that might trigger a fast peer to send those. We don't know the channel id
	// yet so the cache predicate is coarser than the later subscription.
	enableVer0Cache(ctx, c.conn)

	if err := c.conn.pubMsg(ctx, acc, p); err != nil {
		c.logPeer(p).Errorf("error sending proposal acceptance: %v", err)
		return nil, errors.WithMessage(err, "sending proposal acceptance")
	}

	return c.completeCPP(ctx, prop, acc, proposeeIdx)
}

func (c *Client) handleChannelProposalRej(
	ctx context.Context, p wire.Address,
	req ChannelProposal, reason string,
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

// proposeTwoPartyChannel implements the multi-party channel proposal
// protocol for the two-party case. It returns the agreed upon channel
// parameters.
func (c *Client) proposeTwoPartyChannel(
	ctx context.Context,
	proposal ChannelProposal,
) (*Channel, error) {
	peer := c.proposalPeers(proposal)[proposeeIdx]

	// enables caching of incoming version 0 signatures before sending any message
	// that might trigger a fast peer to send those. We don't know the channel id
	// yet so the cache predicate is coarser than the later subscription.
	enableVer0Cache(ctx, c.conn)

	proposalID := proposal.ProposalID()
	isResponse := func(e *wire.Envelope) bool {
		acc, isAcc := e.Msg.(ChannelProposalAccept)
		return (isAcc && acc.Base().ProposalID == proposalID) ||
			(e.Msg.Type() == wire.ChannelProposalRej &&
				e.Msg.(*ChannelProposalRej).ProposalID == proposalID)
	}
	receiver := wire.NewReceiver()
	defer receiver.Close()

	if err := c.conn.Subscribe(receiver, isResponse); err != nil {
		return nil, errors.WithMessage(err, "subscribing proposal response recv")
	}

	if err := c.conn.pubMsg(ctx, proposal, peer); err != nil {
		return nil, errors.WithMessage(err, "publishing channel proposal")
	}

	env, err := receiver.Next(ctx)
	if err != nil {
		if pcontext.IsContextError(err) {
			return nil, newRequestTimedOutError("channel proposal", err.Error())
		}
		return nil, errors.WithMessage(err, "receiving proposal response")
	}
	if rej, ok := env.Msg.(*ChannelProposalRej); ok {
		return nil, newPeerRejectedError("channel proposal", rej.Reason)
	}

	acc, ok := env.Msg.(ChannelProposalAccept) // this is safe because of predicate isResponse
	if !ok {
		log.Panic("internal error: wrong message type")
	}

	if err := c.validChannelProposalAcc(proposal, acc); err != nil {
		return nil, errors.WithMessage(err, "validating channel proposal acceptance")
	}

	return c.completeCPP(ctx, proposal, acc, proposerIdx)
}

// validTwoPartyProposal checks that the proposal is valid in the two-party
// setting, where the proposer is expected to have index 0 in the peer list and
// the receiver to have index 1. The generic validity of the proposal is also
// checked.
func (c *Client) validTwoPartyProposal(
	proposal ChannelProposal,
	ourIdx channel.Index,
	peerAddr wallet.Address,
) error {
	if err := proposal.Valid(); err != nil {
		return err
	}

	peers := c.proposalPeers(proposal)
	if proposal.Base().NumPeers() != len(peers) {
		return errors.Errorf("participants (%d) and peers (%d) dimension mismatch",
			proposal.Base().NumPeers(), len(peers))
	}
	if len(peers) != proposalNumParts {
		return errors.Errorf("expected 2 peers, got %d", len(peers))
	}

	if !(ourIdx == proposerIdx || ourIdx == proposeeIdx) {
		return errors.Errorf("invalid index: %d", ourIdx)
	}

	peerIdx := ourIdx ^ 1
	// In the 2PCPP, the proposer is expected to have index 0
	if !peers[peerIdx].Equal(peerAddr) {
		return errors.Errorf("remote peer doesn't have peer index %d", peerIdx)
	}

	// In the 2PCPP, the receiver is expected to have index 1
	if !peers[ourIdx].Equal(c.address) {
		return errors.Errorf("we don't have peer index %d", ourIdx)
	}

	switch prop := proposal.(type) {
	case *SubChannelProposal:
		if err := c.validSubChannelProposal(prop); err != nil {
			return errors.WithMessage(err, "validate subchannel proposal")
		}
	case *VirtualChannelProposal:
		if err := c.validVirtualChannelProposal(prop, ourIdx); err != nil {
			return errors.WithMessage(err, "validate subchannel proposal")
		}
	}

	return nil
}

func (c *Client) validSubChannelProposal(proposal *SubChannelProposal) error {
	parent, ok := c.channels.Channel(proposal.Parent)
	if !ok {
		return errors.New("parent channel does not exist")
	}

	base := proposal.Base()
	parentState := parent.state() // We assume that the channel is locked.

	if err := channel.AssetsAssertEqual(parentState.Assets, base.InitBals.Assets); err != nil {
		return errors.WithMessage(err, "parent channel and sub-channel assets do not match")
	}

	if err := parentState.Balances.AssertGreaterOrEqual(base.InitBals.Balances); err != nil {
		return errors.WithMessage(err, "insufficient funds")
	}

	return nil
}

func (c *Client) validVirtualChannelProposal(prop *VirtualChannelProposal, ourIdx channel.Index) error {
	numParents := len(prop.Parents)
	numPeers := prop.NumPeers()
	if numParents != numPeers {
		return errors.Errorf("expected %d parent channels, got %d", numPeers, numParents)
	}

	parent, err := c.Channel(prop.Parents[ourIdx])
	if err != nil {
		return errors.New("parent channel not found")
	}

	parentState := parent.state() // We assume that the channel is locked.

	if err := channel.AssetsAssertEqual(parentState.Assets, prop.InitBals.Assets); err != nil {
		return errors.WithMessage(err, "unequal assets")
	}

	if !prop.InitBals.Balances.Equal(prop.FundingAgreement) {
		return errors.WithMessage(err, "unequal funding agreement")
	}

	numIndexMaps := len(prop.IndexMaps)
	if numIndexMaps != numPeers {
		return errors.Errorf("expected %d index maps, got %d", numPeers, numIndexMaps)
	}

	// Check index map entries.
	indexMap := prop.IndexMaps[ourIdx]
	for i, p := range indexMap {
		if int(p) >= numPeers {
			return errors.Errorf("invalid index map entry %d: %d", i, p)
		}
	}

	virtualBals := transformBalances(prop.InitBals.Balances, parentState.NumParts(), indexMap)
	if err := parentState.Balances.AssertGreaterOrEqual(virtualBals); err != nil {
		return errors.WithMessage(err, "insufficient funds")
	}
	return nil
}

func (c *Client) validChannelProposalAcc(
	proposal ChannelProposal,
	response ChannelProposalAccept,
) error {
	if !proposal.Matches(response) {
		return errors.Errorf("Received invalid accept message %T to proposal %T", response, proposal)
	}

	propID := proposal.ProposalID()
	accID := response.Base().ProposalID
	if !bytes.Equal(propID[:], accID[:]) {
		return errors.Errorf("mismatched proposal ID %b and accept ID %b", propID, accID)
	}

	return nil
}

func participants(proposer, proposee wallet.Address) []wallet.Address {
	parts := make([]wallet.Address, proposalNumParts)
	parts[proposerIdx] = proposer
	parts[proposeeIdx] = proposee
	return parts
}

func nonceShares(proposer, proposee NonceShare) []NonceShare {
	shares := make([]NonceShare, proposalNumParts)
	shares[proposerIdx] = proposer
	shares[proposeeIdx] = proposee
	return shares
}

// calcNonce calculates a nonce from its shares. The order of the shares must
// correspond to the participant indices.
func calcNonce(nonceShares []NonceShare) channel.Nonce {
	hasher := newHasher()
	for i, share := range nonceShares {
		if err := io.Encode(hasher, share); err != nil {
			log.Panicf("Failed to encode nonce share %d for hashing", i)
		}
	}
	return channel.NonceFromBytes(hasher.Sum(nil))
}

// completeCPP completes the channel proposal protocol and sets up a new channel
// controller. The initial state with signatures is exchanged using the wallet
// to unlock the account for our participant.
//
// It does not perform a validity check on the proposal, so make sure to only
// pass valid proposals.
//
// It is important that the passed context does not cancel before twice the
// ChallengeDuration has passed (at least for real blockchain backends with wall
// time), or the channel cannot be settled if a peer times out funding.
func (c *Client) completeCPP(
	ctx context.Context,
	prop ChannelProposal,
	acc ChannelProposalAccept,
	partIdx channel.Index,
) (*Channel, error) {
	propBase := prop.Base()
	params := channel.NewParamsUnsafe(
		propBase.ChallengeDuration,
		c.mpcppParts(prop, acc),
		propBase.App,
		calcNonce(nonceShares(propBase.NonceShare, acc.Base().NonceShare)),
		prop.Type() == wire.LedgerChannelProposal,
		prop.Type() == wire.VirtualChannelProposal,
	)

	if c.channels.Has(params.ID()) {
		return nil, errors.New("channel already exists")
	}

	account, err := c.wallet.Unlock(params.Parts[partIdx])
	if err != nil {
		return nil, errors.WithMessage(err, "unlocking account")
	}

	parentChannelID, parent, err := c.proposalParent(prop, partIdx)
	if err != nil {
		return nil, err
	}

	peers := c.proposalPeers(prop)
	ch, err := c.newChannel(account, parent, peers, *params)
	if err != nil {
		return nil, err
	}

	// If subchannel proposal receiver, setup register funding update.
	if prop.Type() == wire.SubChannelProposal && partIdx == proposeeIdx {
		parent.registerSubChannelFunding(ch.ID(), propBase.InitBals.Sum())
	}

	if err := c.pr.ChannelCreated(ctx, ch.machine, peers, parentChannelID); err != nil {
		return ch, errors.WithMessage(err, "persisting new channel")
	}

	if err := ch.init(ctx, propBase.InitBals, propBase.InitData); err != nil {
		return ch, errors.WithMessage(err, "setting initial bals and data")
	}
	if err := ch.initExchangeSigsAndEnable(ctx); err != nil {
		return ch, errors.WithMessage(err, "exchanging initial sigs and enabling state")
	}

	return ch, nil
}

func (c *Client) proposalParent(prop ChannelProposal, partIdx channel.Index) (parentChannelID *channel.ID, parent *Channel, err error) {
	switch prop := prop.(type) {
	case *SubChannelProposal:
		parentChannelID = &prop.Parent
	case *VirtualChannelProposal:
		parentChannelID = &prop.Parents[partIdx]
	}

	if parentChannelID != nil {
		var ok bool
		if parent, ok = c.channels.Channel(*parentChannelID); !ok {
			err = errors.New("referenced parent channel not found")
			return
		}
	}
	return
}

// mpcppParts returns a proposed channel's participant addresses.
func (c *Client) mpcppParts(
	prop ChannelProposal,
	acc ChannelProposalAccept,
) (parts []wallet.Address) {
	switch p := prop.(type) {
	case *LedgerChannelProposal:
		parts = participants(
			p.Participant,
			acc.(*LedgerChannelProposalAcc).Participant)
	case *SubChannelProposal:
		ch, ok := c.channels.Channel(p.Parent)
		if !ok {
			c.log.Panic("unknown parent channel ID")
		}
		parts = ch.Params().Parts
	case *VirtualChannelProposal:
		parts = participants(
			p.Proposer,
			acc.(*VirtualChannelProposalAcc).Responder,
		)
	default:
		c.log.Panicf("unhandled %T", p)
	}
	return
}

func (c *Client) fundChannel(ctx context.Context, ch *Channel, prop ChannelProposal) error {
	switch prop := prop.(type) {
	case *LedgerChannelProposal:
		err := c.fundLedgerChannel(ctx, ch, prop.Base().FundingAgreement)
		return errors.WithMessage(err, "funding ledger channel")
	case *SubChannelProposal:
		err := c.fundSubchannel(ctx, prop, ch)
		return errors.WithMessage(err, "funding subchannel")
	case *VirtualChannelProposal:
		err := c.fundVirtualChannel(ctx, ch, prop)
		return errors.WithMessage(err, "funding virtual channel")
	}
	c.log.Panicf("invalid channel proposal type %T", prop)
	return nil
}

func (c *Client) completeFunding(ctx context.Context, ch *Channel) error {
	params := ch.Params()
	if err := ch.machine.SetFunded(ctx); err != nil {
		return errors.WithMessage(err, "error in SetFunded()")
	}
	if !c.channels.Put(params.ID(), ch) {
		return errors.New("channel already exists")
	}
	c.wallet.IncrementUsage(params.Parts[ch.machine.Idx()])
	return nil
}

func (c *Client) fundLedgerChannel(ctx context.Context, ch *Channel, agreement channel.Balances) (err error) {
	if err = c.funder.Fund(ctx,
		*channel.NewFundingReq(
			ch.Params(),
			ch.machine.State(), // initial state
			ch.machine.Idx(),
			agreement,
		)); channel.IsFundingTimeoutError(err) {
		return errors.WithMessage(err, "waiting for peer funding")
	} else if err != nil { // other runtime error
		ch.Log().Warnf("error while funding channel: %v", err)
		return errors.WithMessage(err, "error while funding channel")
	}

	return c.completeFunding(ctx, ch)
}

func (c *Client) fundSubchannel(ctx context.Context, prop *SubChannelProposal, subChannel *Channel) (err error) {
	parentChannel, ok := c.channels.Channel(prop.Parent)
	if !ok {
		return errors.New("referenced parent channel not found")
	}

	switch subChannel.Idx() {
	case proposerIdx:
		if err := parentChannel.fundSubChannel(ctx, subChannel.ID(), prop.InitBals); err != nil {
			return errors.WithMessage(err, "parent channel update failed")
		}

	case proposeeIdx:
		if err := parentChannel.awaitSubChannelFunding(ctx, subChannel.ID()); err != nil {
			return errors.WithMessage(err, "await subchannel funding update")
		}
	default:
		return errors.New("invalid participant index")
	}

	return c.completeFunding(ctx, subChannel)
}

// enableVer0Cache enables caching of incoming version 0 signatures.
func enableVer0Cache(ctx context.Context, c wire.Cacher) {
	c.Cache(ctx, func(m *wire.Envelope) bool {
		return m.Msg.Type() == wire.ChannelUpdateAcc &&
			m.Msg.(*msgChannelUpdateAcc).Version == 0
	})
}

func (c *Client) enableVer1Cache() {
	c.log.Trace("Enabling version 1 cache")

	c.version1Cache.mu.Lock()
	defer c.version1Cache.mu.Unlock()

	c.version1Cache.enabled++
}

func (c *Client) releaseVer1Cache() {
	c.log.Trace("Releasing version 1 cache")

	c.version1Cache.mu.Lock()
	defer c.version1Cache.mu.Unlock()

	c.version1Cache.enabled--
	for _, u := range c.version1Cache.cache {
		go c.handleChannelUpdate(u.uh, u.p, u.m)
	}
	c.version1Cache.cache = nil
}

type version1Cache struct {
	mu      sync.Mutex
	enabled uint // counter to support concurrent channel openings
	cache   []cachedUpdate
}

type cachedUpdate struct {
	uh UpdateHandler
	p  wire.Address
	m  ChannelUpdateProposal
}

// Error implements the error interface.
func (e PeerRejectedError) Error() string {
	return fmt.Sprintf("%s rejected by peer: %s", e.ItemType, e.Reason)
}

func newPeerRejectedError(rejectedItemType, reason string) error {
	return errors.WithStack(PeerRejectedError{rejectedItemType, reason})
}

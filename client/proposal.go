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
	"perun.network/go-perun/channel/multi"
	"perun.network/go-perun/log"
	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wire"
	pcontext "polycry.pt/poly-go/context"
	"polycry.pt/poly-go/sync/atomic"
)

const (
	// ProposerIdx is the index of the channel proposer.
	ProposerIdx = 0
	// ProposeeIdx is the index of the channel proposal receiver.
	ProposeeIdx = 1
)

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
		peer   map[wallet.BackendID]wire.Address
		req    ChannelProposal
		called atomic.Bool
	}

	// PeerRejectedError indicates the channel proposal or channel update was
	// rejected by the peer.
	//
	// Reason should be a UTF-8 encodable string.
	PeerRejectedError struct {
		ItemType string // ItemType indicates the type of item rejected (channel proposal or channel update).
		Reason   string // Reason sent by the peer for the rejection.
	}

	// ChannelFundingError indicates an error during channel funding.
	ChannelFundingError struct {
		Err error
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
// It is important that the passed context does not cancel before the
// ChallengeDuration has passed. Otherwise funding may not complete.
//
// If funding fails, ChannelFundingError is thrown and an unfunded channel
// object is returned, which can be used for withdrawing the funds.
//
// After the channel got successfully created, the user is required to start the
// channel watcher with Channel.Watch() on the returned channel controller.
//
// Returns ChannelFundingError if an error happened during funding. The internal
// error gives more information.
// - Contains FundingTimeoutError if any of the participants do not fund the
// channel in time.
// - Contains TxTimedoutError when the program times out waiting for a
// transaction to be mined.
// - Contains ChainNotReachableError if the connection to the blockchain network
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

// SetEgoisticChain sets the egoistic chain flag for a given ledger.
func (r *ProposalResponder) SetEgoisticChain(egoistic multi.AssetID, id int) {
	mf, ok := r.client.funder.(*multi.Funder)
	if !ok {
		log.Panic("unexpected type for funder")
	}
	mf.SetEgoisticChain(egoistic, id, true)
}

// RemoveEgoisticChain removes the egoistic chain flag for a given ledger.
func (r *ProposalResponder) RemoveEgoisticChain(egoistic multi.AssetID, id int) {
	mf, ok := r.client.funder.(*multi.Funder)
	if !ok {
		log.Panic("unexpected type for funder")
	}
	mf.SetEgoisticChain(egoistic, id, false)
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
// It is important that the passed context does not cancel before the
// ChallengeDuration has passed. Otherwise funding may not complete.
//
// If funding fails, ChannelFundingError is thrown and an unfunded channel
// object is returned, which can be used for withdrawing the funds.
//
// After the channel got successfully created, the user is required to start the
// channel watcher with Channel.Watch() on the returned channel
// controller.
//
// Returns PeerRejectedProposalError if the channel is rejected by the peer.
// Returns RequestTimedOutError if the peer did not respond before the context
// expires or is cancelled.
// Returns ChannelFundingError if an error happened during funding. The internal
// error gives more information.
// - Contains FundingTimeoutError if any of the participants do not fund the
// channel in time.
// - Contains TxTimedoutError when the program times out waiting for a
// transaction to be mined.
// - Contains ChainNotReachableError if the connection to the blockchain network
// fails when sending a transaction to / reading from the blockchain.
func (c *Client) ProposeChannel(ctx context.Context, prop ChannelProposal) (*Channel, error) {
	if ctx == nil {
		c.log.Panic("invalid nil argument")
	}

	// Prepare and cleanup, e.g., for locking and unlocking parent channel.
	err := c.prepareChannelOpening(ctx, prop, ProposerIdx)
	if err != nil {
		return nil, errors.WithMessage(err, "preparing channel opening")
	}
	defer c.cleanupChannelOpening(prop, ProposerIdx)

	// 1. validate input
	peer := c.proposalPeers(prop)[ProposeeIdx]
	if err := c.validTwoPartyProposal(prop, ProposerIdx, peer); err != nil {
		return nil, errors.WithMessage(err, "invalid channel proposal")
	}

	// 2. send proposal, wait for response, create channel object
	// cache version 1 updates until channel is opened
	c.enableVer1Cache()
	// replay cached version 1 updates
	defer c.releaseVer1Cache() //nolint:contextcheck
	ch, err := c.proposeTwoPartyChannel(ctx, prop)
	if err != nil {
		return nil, errors.WithMessage(err, "channel proposal")
	}

	// 3. fund
	err = c.fundChannel(ctx, ch, prop)
	if err != nil {
		return ch, newChannelFundingError(err)
	}

	return ch, nil
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
func (c *Client) handleChannelProposal(handler ProposalHandler, p map[wallet.BackendID]wire.Address, req ChannelProposal) {
	ourIdx := channel.Index(ProposeeIdx)

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
	ctx context.Context, p map[wallet.BackendID]wire.Address,
	prop ChannelProposal, acc ChannelProposalAccept,
) (ch *Channel, err error) {
	if err := c.validChannelProposalAcc(prop, acc); err != nil {
		return ch, errors.WithMessage(err, "validating channel proposal acceptance")
	}

	// cache version 1 updates
	c.enableVer1Cache()
	// replay cached version 1 updates
	defer c.releaseVer1Cache() //nolint:contextcheck

	if ch, err = c.acceptChannelProposal(ctx, prop, p, acc); err != nil {
		return ch, errors.WithMessage(err, "accept channel proposal")
	}

	err = c.fundChannel(ctx, ch, prop)
	if err != nil {
		return ch, newChannelFundingError(err)
	}
	return ch, nil
}

func (c *Client) acceptChannelProposal(
	ctx context.Context,
	prop ChannelProposal,
	p map[wallet.BackendID]wire.Address,
	acc ChannelProposalAccept,
) (*Channel, error) {
	if acc == nil {
		c.logPeer(p).Error("user passed nil ChannelProposalAcc")
		return nil, errors.New("nil ChannelProposalAcc")
	}

	// enables caching of incoming version 0 signatures before sending any message
	// that might trigger a fast peer to send those. We don't know the channel id
	// yet so the cache predicate is coarser than the later subscription.
	pred := enableVer0Cache(c.conn)
	defer c.conn.ReleaseCache(pred)

	if err := c.conn.pubMsg(ctx, acc, p); err != nil {
		c.logPeer(p).Errorf("error sending proposal acceptance: %v", err)
		return nil, errors.WithMessage(err, "sending proposal acceptance")
	}

	return c.completeCPP(ctx, prop, acc, ProposeeIdx)
}

func (c *Client) handleChannelProposalRej(
	ctx context.Context, p map[wallet.BackendID]wire.Address,
	req ChannelProposal, reason string,
) error {
	msgReject := &ChannelProposalRejMsg{
		ProposalID: req.Base().ProposalID,
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
	peer := c.proposalPeers(proposal)[ProposeeIdx]

	// enables caching of incoming version 0 signatures before sending any message
	// that might trigger a fast peer to send those. We don't know the channel id
	// yet so the cache predicate is coarser than the later subscription.
	pred := enableVer0Cache(c.conn)
	defer c.conn.ReleaseCache(pred)

	proposalID := proposal.Base().ProposalID
	isResponse := func(e *wire.Envelope) bool {
		switch msg := e.Msg.(type) {
		case ChannelProposalAccept:
			return msg.Base().ProposalID == proposalID
		case *ChannelProposalRejMsg:
			return msg.ProposalID == proposalID
		default:
			return false
		}
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
	if rej, ok := env.Msg.(*ChannelProposalRejMsg); ok {
		return nil, newPeerRejectedError("channel proposal", rej.Reason)
	}

	acc, ok := env.Msg.(ChannelProposalAccept) // this is safe because of predicate isResponse
	if !ok {
		log.Panic("internal error: wrong message type")
	}

	if err := c.validChannelProposalAcc(proposal, acc); err != nil {
		return nil, errors.WithMessage(err, "validating channel proposal acceptance")
	}

	return c.completeCPP(ctx, proposal, acc, ProposerIdx)
}

// validTwoPartyProposal checks that the proposal is valid in the two-party
// setting, where the proposer is expected to have index 0 in the peer list and
// the receiver to have index 1. The generic validity of the proposal is also
// checked.
func (c *Client) validTwoPartyProposal(
	proposal ChannelProposal,
	ourIdx channel.Index,
	peerAddr map[wallet.BackendID]wire.Address,
) error {
	if err := proposal.Valid(); err != nil {
		return err
	}

	multiLedger := multi.IsMultiLedgerAssets(proposal.Base().InitBals.Assets)
	appChannel := !channel.IsNoApp(proposal.Base().App)
	if multiLedger && appChannel {
		return errors.New("multi-ledger app channel not supported")
	}

	peers := c.proposalPeers(proposal)
	if proposal.Base().NumPeers() != len(peers) {
		return errors.Errorf("participants (%d) and peers (%d) dimension mismatch",
			proposal.Base().NumPeers(), len(peers))
	}
	if len(peers) != proposalNumParts {
		return errors.Errorf("expected 2 peers, got %d", len(peers))
	}

	if !(ourIdx == ProposerIdx || ourIdx == ProposeeIdx) {
		return errors.Errorf("invalid index: %d", ourIdx)
	}

	peerIdx := ourIdx ^ 1
	// In the 2PCPP, the proposer is expected to have index 0
	if !channel.EqualWireMaps(peers[peerIdx], peerAddr) {
		return errors.Errorf("remote peer doesn't have peer index %d", peerIdx)
	}

	// In the 2PCPP, the receiver is expected to have index 1
	if !channel.EqualWireMaps(peers[ourIdx], c.address) {
		return errors.Errorf("we don't have peer index %d", ourIdx)
	}

	switch prop := proposal.(type) {
	case *SubChannelProposalMsg:
		if err := c.validSubChannelProposal(prop); err != nil {
			return errors.WithMessage(err, "validate subchannel proposal")
		}
	case *VirtualChannelProposalMsg:
		if err := c.validVirtualChannelProposal(prop, ourIdx); err != nil {
			return errors.WithMessage(err, "validate subchannel proposal")
		}
	}

	return nil
}

func (c *Client) validSubChannelProposal(proposal *SubChannelProposalMsg) error {
	parent, ok := c.channels.Channel(proposal.Parent)
	if !ok {
		return errors.New("parent channel does not exist")
	}

	base := proposal.Base()
	parentState := parent.state() // We assume that the channel is locked.

	if err := channel.AssertAssetsEqual(parentState.Assets, base.InitBals.Assets); err != nil {
		return errors.WithMessage(err, "parent channel and sub-channel assets do not match")
	}

	if err := parentState.Balances.AssertGreaterOrEqual(base.InitBals.Balances); err != nil {
		return errors.WithMessage(err, "insufficient funds")
	}

	return nil
}

func (c *Client) validVirtualChannelProposal(prop *VirtualChannelProposalMsg, ourIdx channel.Index) error {
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

	if err := channel.AssertAssetsEqual(parentState.Assets, prop.InitBals.Assets); err != nil {
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

	propID := proposal.Base().ProposalID
	accID := response.Base().ProposalID
	if !bytes.Equal(propID[:], accID[:]) {
		return errors.Errorf("mismatched proposal ID %b and accept ID %b", propID, accID)
	}

	return nil
}

func participants(proposer, proposee map[wallet.BackendID]wallet.Address) []map[wallet.BackendID]wallet.Address {
	parts := make([]map[wallet.BackendID]wallet.Address, proposalNumParts)
	parts[ProposerIdx] = proposer
	parts[ProposeeIdx] = proposee
	return parts
}

func nonceShares(proposer, proposee NonceShare) []NonceShare {
	shares := make([]NonceShare, proposalNumParts)
	shares[ProposerIdx] = proposer
	shares[ProposeeIdx] = proposee
	return shares
}

// calcNonce calculates a nonce from its shares. The order of the shares must
// correspond to the participant indices.
func calcNonce(nonceShares []NonceShare) channel.Nonce {
	hasher := newHasher()
	for i, share := range nonceShares {
		if _, err := hasher.Write(share[:]); err != nil {
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

	accounts := make(map[wallet.BackendID]wallet.Account)
	var err error
	for i, wall := range c.wallet {
		accounts[i], err = wall.Unlock(params.Parts[partIdx][i])
		if err != nil {
			return nil, errors.WithMessage(err, "unlocking account")
		}
	}

	parentChannelID, parent, err := c.proposalParent(prop, partIdx)
	if err != nil {
		return nil, err
	}

	peers := c.proposalPeers(prop)
	ch, err := c.newChannel(accounts, parent, peers, *params)
	if err != nil {
		return nil, err
	}

	// If subchannel proposal receiver, setup register funding update.
	if prop.Type() == wire.SubChannelProposal && partIdx == ProposeeIdx {
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

	for i, wall := range c.wallet {
		wall.IncrementUsage(params.Parts[partIdx][i])
	}
	return ch, nil
}

func (c *Client) proposalParent(prop ChannelProposal, partIdx channel.Index) (parentChannelID *map[wallet.BackendID]channel.ID, parent *Channel, err error) {
	switch prop := prop.(type) {
	case *SubChannelProposalMsg:
		parentChannelID = &prop.Parent
	case *VirtualChannelProposalMsg:
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
) (parts []map[wallet.BackendID]wallet.Address) {
	switch p := prop.(type) {
	case *LedgerChannelProposalMsg:
		ledgerAcc, ok := acc.(*LedgerChannelProposalAccMsg)
		if !ok {
			c.log.Panicf("unexpected message type: expected *LedgerChannelProposalAccMsg, got %T", acc)
		}
		parts = participants(p.Participant, ledgerAcc.Participant)
	case *SubChannelProposalMsg:
		ch, ok := c.channels.Channel(p.Parent)
		if !ok {
			c.log.Panic("unknown parent channel ID")
		}
		parts = ch.Params().Parts
	case *VirtualChannelProposalMsg:
		virtualAcc, ok := acc.(*VirtualChannelProposalAccMsg)
		if !ok {
			c.log.Panicf("unexpected message type: expected *VirtualChannelProposalAccMsg, got %T", acc)
		}
		parts = participants(p.Proposer, virtualAcc.Responder)
	default:
		c.log.Panicf("unhandled %T", p)
	}
	return
}

func (c *Client) fundChannel(ctx context.Context, ch *Channel, prop ChannelProposal) error {
	switch prop := prop.(type) {
	case *LedgerChannelProposalMsg:
		err := c.fundLedgerChannel(ctx, ch, prop.Base().FundingAgreement)
		return errors.WithMessage(err, "funding ledger channel")
	case *SubChannelProposalMsg:
		err := c.fundSubchannel(ctx, prop, ch)
		return errors.WithMessage(err, "funding subchannel")
	case *VirtualChannelProposalMsg:
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
	for i, wall := range c.wallet {
		wall.IncrementUsage(params.Parts[ch.machine.Idx()][i])
	}
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

func (c *Client) fundSubchannel(ctx context.Context, prop *SubChannelProposalMsg, subChannel *Channel) (err error) {
	parentChannel, ok := c.channels.Channel(prop.Parent)
	if !ok {
		return errors.New("referenced parent channel not found")
	}

	switch subChannel.Idx() {
	case ProposerIdx:
		if err := parentChannel.fundSubChannel(ctx, subChannel.ID(), prop.InitBals); err != nil {
			return errors.WithMessage(err, "parent channel update failed")
		}

	case ProposeeIdx:
		if err := parentChannel.awaitSubChannelFunding(ctx, subChannel.ID()); err != nil {
			return errors.WithMessage(err, "await subchannel funding update")
		}
	default:
		return errors.New("invalid participant index")
	}

	return c.completeFunding(ctx, subChannel)
}

// enableVer0Cache enables caching of incoming version 0 signatures.
func enableVer0Cache(c wire.Cacher) *wire.Predicate {
	p := func(m *wire.Envelope) bool {
		msg, ok := m.Msg.(*ChannelUpdateAccMsg)
		return ok && msg.Version == 0
	}
	c.Cache(&p)
	return &p
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
		go c.handleChannelUpdate(u.uh, u.p, u.m) //nolint:contextcheck
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
	p  map[wallet.BackendID]wire.Address
	m  ChannelUpdateProposal
}

// Error implements the error interface.
func (e PeerRejectedError) Error() string {
	return fmt.Sprintf("%s rejected by peer: %s", e.ItemType, e.Reason)
}

func newPeerRejectedError(rejectedItemType, reason string) error {
	return errors.WithStack(PeerRejectedError{rejectedItemType, reason})
}

func newChannelFundingError(err error) *ChannelFundingError {
	return &ChannelFundingError{err}
}

func (e ChannelFundingError) Error() string {
	return fmt.Sprintf("channel funding failed: %v", e.Err.Error())
}

func GetPeerMapWire(peers map[wallet.BackendID][]wire.Address, index int) map[wallet.BackendID]wire.Address {
	address := make(map[wallet.BackendID]wire.Address)
	for i, p := range peers {
		address[i] = p[index]
	}
	return address
}

// Copyright 2021 - See NOTICE file for copyright holders.
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
	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wire"
	pcontext "polycry.pt/poly-go/context"
	"polycry.pt/poly-go/sync/atomic"
)

// handleChannelUpdate forwards incoming channel update requests to the
// respective channel's update handler (Channel.handleUpdateReq). If the channel
// is unknown, an error is logged.
//
// This handler is dispatched from the Client.Handle routine.
func (c *Client) handleChannelUpdate(uh UpdateHandler, p wire.Address, m ChannelUpdateProposal) {
	ch, ok := c.channels.Channel(m.Base().ID())
	if !ok {
		if !c.cacheVersion1Update(uh, p, m) {
			c.logChan(m.Base().ID()).WithField("peer", p).Error("received update for unknown channel")
		}
		return
	}
	pidx := ch.Idx() ^ 1
	ch.handleUpdateReq(pidx, m, uh) //nolint:contextcheck
}

func (c *Client) cacheVersion1Update(uh UpdateHandler, p wire.Address, m ChannelUpdateProposal) bool {
	c.version1Cache.mu.Lock()
	defer c.version1Cache.mu.Unlock()

	if !(m.Base().State.Version == 1 && c.version1Cache.enabled > 0) {
		return false
	}

	c.version1Cache.cache = append(c.version1Cache.cache, cachedUpdate{
		uh: uh,
		p:  p,
		m:  m,
	})
	return true
}

type (
	// ChannelUpdate is a channel update proposal.
	ChannelUpdate struct {
		// State is the proposed new state.
		State *channel.State
		// ActorIdx is the actor causing the new state. It does not need to
		// coincide with the sender of the request.
		ActorIdx channel.Index
	}

	// An UpdateHandler decides how to handle incoming channel update requests
	// from other channel participants.
	UpdateHandler interface {
		// HandleUpdate is the user callback called by the channel controller on an
		// incoming update request. The first argument contains the current state
		// of the channel before the update is applied. Clone it if you want to
		// modify it.
		HandleUpdate(*channel.State, ChannelUpdate, *UpdateResponder)
	}

	// UpdateHandlerFunc is an adapter type to allow the use of functions as
	// update handlers. UpdateHandlerFunc(f) is an UpdateHandler that calls
	// f when HandleUpdate is called.
	UpdateHandlerFunc func(*channel.State, ChannelUpdate, *UpdateResponder)

	// The UpdateResponder allows the user to react to the incoming channel update
	// request. If the user wants to accept the update, Accept() should be called,
	// otherwise Reject(), possibly giving a reason for the rejection.
	// Only a single function must be called and every further call causes a
	// panic.
	UpdateResponder struct {
		channel *Channel
		pidx    channel.Index
		req     ChannelUpdateProposal
		called  atomic.Bool
	}

	// RequestTimedOutError indicates that a peer has not responded within the
	// expected time period.
	RequestTimedOutError string
)

// HandleUpdate calls the update handler function.
func (f UpdateHandlerFunc) HandleUpdate(s *channel.State, u ChannelUpdate, r *UpdateResponder) {
	f(s, u, r)
}

// Accept lets the user signal that they want to accept the channel update.
func (r *UpdateResponder) Accept(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context must not be nil")
	}
	if !r.called.TrySet() {
		log.Panic("multiple calls on channel update responder")
	}

	return r.channel.handleUpdateAcc(ctx, r.pidx, r.req)
}

// Reject lets the user signal that they reject the channel update.
func (r *UpdateResponder) Reject(ctx context.Context, reason string) error {
	if ctx == nil {
		return errors.New("context must not be nil")
	}
	if !r.called.TrySet() {
		log.Panic("multiple calls on channel update responder")
	}

	return r.channel.handleUpdateRej(ctx, r.pidx, r.req, reason)
}

// Update updates the channel state using the update function and proposes the
// new state to all other channel participants. The update function must not
// update the version counter.
//
// Returns nil if all peers accept the update. Returns RequestTimedOutError if
// any peer did not respond before the context expires or is cancelled. Returns
// an error if any runtime error occurs or any peer rejects the update.
func (c *Channel) Update(ctx context.Context, update func(*channel.State) error) (err error) {
	if ctx == nil {
		return errors.New("context must not be nil")
	}

	// Lock machine while update is in progress.
	if !c.machMtx.TryLockCtx(ctx) {
		return errors.Errorf("locking machine mutex in time: %v", ctx.Err())
	}
	defer c.machMtx.Unlock()

	return c.update(ctx,
		func(state *channel.State) error {
			// apply update
			if err := update(state); err != nil {
				return err
			}

			// validate
			return c.validTwoPartyUpdateState(state)
		},
	)
}

// updateGeneric proposes the `next` state to all channel participants.
// `prepareMsg` allows to control which message type is being used.
// `next` should not be modified while this function runs.
//
// It assumes that the channel is locked and the update is validated.
//
// Returns nil if all peers accept the update. Returns RequestTimedOutError if
// any peer did not respond before the context expires or is cancelled. Returns
// an error if any runtime error occurs or any peer rejects the update.
func (c *Channel) updateGeneric(
	ctx context.Context,
	next *channel.State,
	prepareMsg func(*msgChannelUpdate) wire.Msg,
) (err error) {
	up := makeChannelUpdate(next, c.machine.Idx())
	if err = c.machine.Update(ctx, up.State, up.ActorIdx); err != nil {
		return errors.WithMessage(err, "updating machine")
	}
	// if anything goes wrong from now on, we discard the update.
	defer func() { c.checkUpdateError(ctx, err) }()

	sig, err := c.machine.Sig(ctx)
	if err != nil {
		return errors.WithMessage(err, "signing update")
	}

	resRecv, err := c.conn.NewUpdateResRecv(up.State.Version)
	if err != nil {
		return errors.WithMessage(err, "creating update response receiver")
	}
	defer resRecv.Close()

	msgUpdate := &msgChannelUpdate{
		ChannelUpdate: up,
		Sig:           sig,
	}
	msg := prepareMsg(msgUpdate)
	if err = c.conn.Send(ctx, msg); err != nil {
		return errors.WithMessage(err, "sending update")
	}

	pidx, res, err := resRecv.Next(ctx)
	if err != nil {
		if pcontext.IsContextError(err) {
			err = newRequestTimedOutError("channel update", err.Error())
			return err
		}
		return errors.WithMessage(err, "receiving update response")
	}
	c.Log().Tracef("Received update response (%T): %v", res, res)

	if rej, ok := res.(*msgChannelUpdateRej); ok {
		return newPeerRejectedError("channel update", rej.Reason)
	}

	acc, ok := res.(*msgChannelUpdateAcc) // safe by predicate of the updateResRecv
	if !ok {
		log.Panic("wrong message type")
	}
	if err := c.machine.AddSig(ctx, pidx, acc.Sig); err != nil {
		return errors.WithMessage(err, "adding peer signature")
	}

	return c.enableNotifyUpdate(ctx)
}

// checkUpdateError is a helper function that checks whether an error occurred
// and in this case attempts to discard the update.
func (c *Channel) checkUpdateError(ctx context.Context, updateErr error) {
	if updateErr != nil {
		if derr := c.machine.DiscardUpdate(ctx); derr != nil {
			// discarding update should never fail
			c.Log().Warn("discarding update failed:", derr)
		}
	}
}

func (c *Channel) update(ctx context.Context, update func(*channel.State) error) (err error) {
	state := c.machine.State().Clone()
	if err := update(state); err != nil {
		return err
	}
	state.Version++

	return c.updateGeneric(ctx, state, func(mcu *msgChannelUpdate) wire.Msg { return mcu })
}

// handleUpdateReq is called by the controller on incoming channel update
// requests.
func (c *Channel) handleUpdateReq(
	pidx channel.Index,
	req ChannelUpdateProposal,
	uh UpdateHandler,
) {
	c.machMtx.Lock() // Lock machine while update is in progress.
	defer c.machMtx.Unlock()

	if err := c.machine.CheckUpdate(req.Base().State, req.Base().ActorIdx, req.Base().Sig, pidx); err != nil {
		c.logPeer(pidx).Warnf("invalid update received: %v", err)
		return
	}

	responder := &UpdateResponder{channel: c, pidx: pidx, req: req}
	client := c.client

	if prop, ok := req.(*virtualChannelFundingProposal); ok {
		client.handleVirtualChannelFundingProposal(c, prop, responder) //nolint:contextcheck
		return
	}

	if prop, ok := req.(*virtualChannelSettlementProposal); ok {
		client.handleVirtualChannelSettlementProposal(c, prop, responder) //nolint:contextcheck
		return
	}

	if ui, ok := c.subChannelFundings.Filter(req.Base().ChannelUpdate); ok {
		ui.HandleUpdate(req.Base().ChannelUpdate, responder)
		return
	}

	if ui, ok := c.subChannelWithdrawals.Filter(req.Base().ChannelUpdate); ok {
		ui.HandleUpdate(req.Base().ChannelUpdate, responder)
		return
	}

	if err := c.validTwoPartyUpdate(req.Base().ChannelUpdate, pidx); err != nil {
		c.logPeer(pidx).Warnf("invalid update received: %v", err)
		return
	}

	uh.HandleUpdate(c.machine.State(), req.Base().ChannelUpdate, responder)
}

func (c *Channel) handleUpdateAcc(
	ctx context.Context,
	pidx channel.Index,
	req ChannelUpdateProposal,
) (err error) {
	defer func() {
		if err != nil {
			c.logPeer(pidx).Errorf("error accepting state: %v", err)
		}
	}()

	// machine.Update and AddSig should never fail after CheckUpdate...
	if err = c.machine.Update(ctx, req.Base().State, req.Base().ActorIdx); err != nil {
		return errors.WithMessage(err, "updating machine")
	}
	// if anything goes wrong from now on, we discard the update.
	defer func() {
		if err != nil {
			// we discard the update if anything went wrong
			if derr := c.machine.DiscardUpdate(ctx); derr != nil {
				// discarding update should never fail at this point
				err = errors.WithMessagef(derr,
					"sending accept message failed: %v, then discarding update failed", err)
			}
		}
	}()

	if err = c.machine.AddSig(ctx, pidx, req.Base().Sig); err != nil {
		return errors.WithMessage(err, "adding peer signature")
	}
	var sig wallet.Sig
	sig, err = c.machine.Sig(ctx)
	if err != nil {
		return errors.WithMessage(err, "signing updated state")
	}

	// If subchannel is final, register settlement update at parent channel.
	if c.IsSubChannel() && req.Base().State.IsFinal {
		c.Parent().registerSubChannelSettlement(c.ID(), req.Base().State.Balances)
	}

	msgUpAcc := &msgChannelUpdateAcc{
		ChannelID: c.ID(),
		Version:   req.Base().State.Version,
		Sig:       sig,
	}
	if err := c.conn.Send(ctx, msgUpAcc); err != nil {
		return errors.WithMessage(err, "sending accept message")
	}

	return c.enableNotifyUpdate(ctx)
}

func (c *Channel) handleUpdateRej(
	ctx context.Context,
	pidx channel.Index,
	req ChannelUpdateProposal,
	reason string,
) (err error) {
	defer func() {
		if err != nil {
			c.logPeer(pidx).Errorf("error rejecting state: %v", err)
		}
	}()

	msgUpRej := &msgChannelUpdateRej{
		ChannelID: c.ID(),
		Version:   req.Base().State.Version,
		Reason:    reason,
	}
	return errors.WithMessage(c.conn.Send(ctx, msgUpRej), "sending reject message")
}

// enableNotifyUpdate enables the current staging state of the machine. If the
// state is final, machine.EnableFinal is called. Finally, if there is a
// notification on channel updates, the enabled state is sent on it.
func (c *Channel) enableNotifyUpdate(ctx context.Context) error {
	var err error
	from := c.machine.State()
	to := c.machine.StagingState()
	if to.IsFinal {
		err = c.machine.EnableFinal(ctx)
	} else {
		err = c.machine.EnableUpdate(ctx)
	}

	if err != nil {
		return errors.WithMessage(err, "enabling update")
	}

	if c.onUpdate != nil {
		c.onUpdate(from, to)
	}

	if err = c.statesPub.Publish(ctx, c.machine.CurrentTX()); err != nil {
		c.Log().WithField("Version", c.state().Version).Errorf("Error publishe state to watcher: %v", err)
	}
	return nil
}

// OnUpdate sets up a callback to state updates for the channel.
// The subscription cannot be canceled, but it can be replaced.
// The States that are passed to the callback are not clones but pointers to the
// State in the channel machine, so they must not be modified. If you need to
// modify the State, .Clone() them first.
func (c *Channel) OnUpdate(cb func(from, to *channel.State)) {
	c.onUpdate = cb
}

// validTwoPartyUpdate performs additional protocol-dependent checks on the
// proposed update that go beyond the machine's checks:
// * Actor and signer must be the same.
// * Sub-allocations do not change.
func (c *Channel) validTwoPartyUpdate(up ChannelUpdate, sigIdx channel.Index) error {
	if up.ActorIdx != sigIdx {
		return errors.Errorf(
			"Currently, only update proposals with the proposing peer as actor are allowed.")
	}
	if err := channel.SubAllocsAssertEqual(c.machine.State().Locked, up.State.Locked); err != nil {
		return errors.WithMessage(err, "sub-allocation changed")
	}
	return nil
}

func (c *Channel) validTwoPartyUpdateState(next *channel.State) error {
	up := makeChannelUpdate(next, c.machine.Idx())
	return c.validTwoPartyUpdate(up, c.machine.Idx())
}

func makeChannelUpdate(next *channel.State, actor channel.Index) ChannelUpdate {
	return ChannelUpdate{
		State:    next,
		ActorIdx: actor,
	}
}

// Error implements error interface for RequestTimedOutError.
func (e RequestTimedOutError) Error() string {
	return string(e)
}

func newRequestTimedOutError(requestType, msg string) error {
	return errors.Wrap(RequestTimedOutError("peer did not respond to the "+requestType), msg)
}

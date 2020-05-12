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
	"perun.network/go-perun/pkg/sync/atomic"
	"perun.network/go-perun/wallet"
)

type (
	// ChannelUpdate is a channel update proposal.
	ChannelUpdate struct {
		// State is the proposed new state.
		State *channel.State
		// ActorIdx is the actor causing the new state.  It does not need to
		// coincide with the sender of the request.
		ActorIdx uint16
	}

	// An UpdateHandler decides how to handle incoming channel update requests
	// from other channel participants.
	UpdateHandler interface {
		// Handle is the user callback called by the channel controller on an
		// incoming update request.
		Handle(ChannelUpdate, *UpdateResponder)
	}

	// The UpdateResponder allows the user to react to the incoming channel update
	// request. If the user wants to accept the update, Accept() should be called,
	// otherwise Reject(), possibly giving a reason for the rejection.
	// Only a single function must be called and every further call causes a
	// panic.
	UpdateResponder struct {
		channel *Channel
		pidx    channel.Index
		req     *msgChannelUpdate
		called  atomic.Bool
	}
)

// Accept lets the user signal that they want to accept the channel update.
func (r *UpdateResponder) Accept(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context must not be nil")
	}
	if !r.called.TrySet() {
		log.Panic("multiple calls on channel update responder")
	}
	if ctx == nil {
		log.Panic("nil context")
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
	if ctx == nil {
		log.Panic("nil context")
	}

	return r.channel.handleUpdateRej(ctx, r.pidx, r.req, reason)
}

// Update proposes the given channel update to all channel participants.
//
// It returns nil if all peers accept the update. If any runtime error occurs or
// any peer rejects the update, an error is returned.
func (c *Channel) Update(ctx context.Context, up ChannelUpdate) (err error) {
	if ctx == nil {
		return errors.New("context must not be nil")
	}
	if err := c.validTwoPartyUpdate(up, c.machine.Idx()); err != nil {
		return err
	}

	c.machMtx.Lock() // lock machine while update is in progress
	defer c.machMtx.Unlock()

	if err = c.machine.Update(ctx, up.State, up.ActorIdx); err != nil {
		return errors.WithMessage(err, "updating machine")
	}
	// if anything goes wrong from now on, we discard the update.
	// TODO: this is insecure after we sent our signature.
	defer func() {
		if err != nil {
			if derr := c.machine.DiscardUpdate(); derr != nil {
				// discarding update should never fail
				err = errors.WithMessagef(derr,
					"progressing update failed: %v, then discarding update failed", err)
			}
		}
	}()

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
	if err = c.conn.Send(ctx, msgUpdate); err != nil {
		return errors.WithMessage(err, "sending update")
	}

	pidx, res := resRecv.Next(ctx)
	c.log.Tracef("Received update response (%T): %v", res, res)
	if res == nil {
		return errors.WithMessage(err, "receiving update response")
	}

	if rej, ok := res.(*msgChannelUpdateRej); ok {
		return errors.Errorf("update rejected: %s", rej.Reason)
	}

	acc := res.(*msgChannelUpdateAcc) // safe by predicate of the updateResRecv
	if err := c.machine.AddSig(ctx, pidx, acc.Sig); err != nil {
		return errors.WithMessage(err, "adding peer signature")
	}

	return c.enableNotifyUpdate(ctx)
}

// UpdateBy updates the channel state using the update function and proposes the new state
// to all other channel participants.
//
// It returns nil if all peers accept the update. If any runtime error occurs or
// any peer rejects the update, an error is returned.
func (c *Channel) UpdateBy(ctx context.Context, update func(*channel.State)) (err error) {
	state := c.State().Clone()
	update(state)
	state.Version++

	return c.Update(ctx, ChannelUpdate{
		State:    state,
		ActorIdx: c.Idx(),
	})
}

// ListenUpdates starts the handling of incoming channel update requests. It
// should immediately be started by the user after they receive the channel
// controller.
func (c *Channel) ListenUpdates(uh UpdateHandler) {
	if uh == nil {
		c.log.Panicf("update handler must not be nil")
	}

	for {
		pidx, req := c.conn.NextUpdateReq(context.Background())
		if req == nil {
			c.log.Debug("update request receiver closed")
			return
		}
		c.handleUpdateReq(pidx, req, uh)
	}
}

// handleUpdateReq is called by the controller on incoming channel update
// requests.
func (c *Channel) handleUpdateReq(
	pidx channel.Index,
	req *msgChannelUpdate,
	uh UpdateHandler) {
	if err := c.validTwoPartyUpdate(req.ChannelUpdate, pidx); err != nil {
		// TODO: how to handle invalid updates? Just drop and ignore them?
		c.logPeer(pidx).Warnf("invalid update received: %v", err)
		return
	}

	c.machMtx.Lock() // lock machine while update is in progress
	defer c.machMtx.Unlock()

	if err := c.machine.CheckUpdate(req.State, req.ActorIdx, req.Sig, pidx); err != nil {
		// TODO: how to handle invalid updates? Just drop and ignore them?
		c.logPeer(pidx).Warnf("invalid update received: %v", err)
		return
	}

	responder := &UpdateResponder{channel: c, pidx: pidx, req: req}
	uh.Handle(req.ChannelUpdate, responder)
}

func (c *Channel) handleUpdateAcc(
	ctx context.Context,
	pidx channel.Index,
	req *msgChannelUpdate,
) (err error) {
	defer func() {
		if err != nil {
			c.logPeer(pidx).Errorf("error accepting state: %v", err)
		}
	}()

	// machine.Update and AddSig should never fail after CheckUpdate...
	if err = c.machine.Update(ctx, req.State, req.ActorIdx); err != nil {
		return errors.WithMessage(err, "updating machine")
	}
	// if anything goes wrong from now on, we discard the update.
	// TODO: this is insecure after we sent our signature.
	defer func() {
		if err != nil {
			// we discard the update if anything went wrong
			if derr := c.machine.DiscardUpdate(); derr != nil {
				// discarding update should never fail at this point
				err = errors.WithMessagef(derr,
					"sending accept message failed: %v, then discarding update failed", err)
			}
		}
	}()

	if err = c.machine.AddSig(ctx, pidx, req.Sig); err != nil {
		return errors.WithMessage(err, "adding peer signature")
	}
	var sig wallet.Sig
	sig, err = c.machine.Sig(ctx)
	if err != nil {
		return errors.WithMessage(err, "signing updated state")
	}

	msgUpAcc := &msgChannelUpdateAcc{
		ChannelID: c.ID(),
		Version:   req.State.Version,
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
	req *msgChannelUpdate,
	reason string,
) (err error) {
	defer func() {
		if err != nil {
			c.logPeer(pidx).Errorf("error rejecting state: %v", err)
		}
	}()

	msgUpRej := &msgChannelUpdateRej{
		ChannelID: c.ID(),
		Version:   req.State.Version,
		Reason:    reason,
	}
	return errors.WithMessage(c.conn.Send(ctx, msgUpRej), "sending reject message")
}

// enableNotifyUpdate enables the current staging state of the machine. If the
// state is final, machine.EnableFinal is called. Finally, if there is a
// notification on channel updates, the enabled state is sent on it.
func (c *Channel) enableNotifyUpdate(ctx context.Context) error {
	var err error
	if c.machine.StagingState().IsFinal {
		err = c.machine.EnableFinal(ctx)
	} else {
		err = c.machine.EnableUpdate(ctx)
	}

	if err != nil {
		return errors.WithMessage(err, "enabling update")
	}

	if c.updateSub != nil {
		c.updateSub <- c.machine.State()
	}
	return nil
}

// SubUpdates sets up a subscription to state updates on the provided go channel.
// The subscription cannot be canceled, but it can be replaced.
// The provided go channel is not closed if the Channel is closed. It must not
// be closed while the Channel is not closed.
// The States that are sent on the channel are not clones but pointers to the
// State in the channel machine, so they must not be modified. If you need to
// modify the State, .Clone() it first.
func (c *Channel) SubUpdates(updateSub chan<- *channel.State) {
	c.updateSub = updateSub
}

// validTwoPartyUpdate performs additional protocol-dependent checks on the
// proposed update that go beyond the machine's checks:
// * actor and signer must be the same
// * no locked sub-allocations
func (c *Channel) validTwoPartyUpdate(up ChannelUpdate, sigIdx channel.Index) error {
	if up.ActorIdx != sigIdx {
		return errors.Errorf(
			"Currently, only update proposals with the proposing peer as actor are allowed.")
	}
	if len(up.State.Locked) > 0 {
		return errors.New("no locked sub-allocations allowed")
	}
	return nil
}

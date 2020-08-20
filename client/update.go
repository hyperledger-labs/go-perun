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

// handleChannelUpdate forwards incoming channel update requests to the
// respective channel's update handler (Channel.handleUpdateReq). If the channel
// is unknown, an error is logged.
//
// This handler is dispatched from the Client.Handle routine.
func (c *Client) handleChannelUpdate(uh UpdateHandler, p wire.Address, m *msgChannelUpdate) {
	ch, ok := c.channels.Get(m.ID())
	if !ok {
		c.logChan(m.ID()).WithField("peer", p).Error("received update for unknown channel")
		return
	}
	pidx := ch.Idx() ^ 1
	ch.handleUpdateReq(pidx, m, uh)
}

type (
	// ChannelUpdate is a channel update proposal.
	ChannelUpdate struct {
		// State is the proposed new state.
		State *channel.State
		// ActorIdx is the actor causing the new state. It does not need to
		// coincide with the sender of the request.
		ActorIdx uint16
	}

	// An UpdateHandler decides how to handle incoming channel update requests
	// from other channel participants.
	UpdateHandler interface {
		// HandleUpdate is the user callback called by the channel controller on an
		// incoming update request.
		//
		// A channel peer might request an update from you by setting your index
		// as the ActorIdx, so don't forget to check the ActorIdx according to
		// your application logic.
		HandleUpdate(ChannelUpdate, *UpdateResponder)
	}

	// UpdateHandlerFunc is an adapter type to allow the use of functions as
	// update handlers. UpdateHandlerFunc(f) is an UpdateHandler that calls
	// f when HandleUpdate is called.
	UpdateHandlerFunc func(ChannelUpdate, *UpdateResponder)

	// The UpdateResponder allows the user to react to the incoming channel update
	// request. If the user wants to accept the update, Accept() should be called,
	// otherwise Reject(), possibly giving a reason for the rejection.
	// Only a single function must be called and every further call causes a
	// panic.
	//
	// The user has to check the ActorIdx of the request since it is possible
	// to request updates by using the other peers actor index. They should only
	// accept updates which benefit him.
	UpdateResponder struct {
		channel *Channel
		pidx    channel.Index
		req     *msgChannelUpdate
		called  atomic.Bool
	}
)

// HandleUpdate calls the update handler function.
func (f UpdateHandlerFunc) HandleUpdate(u ChannelUpdate, r *UpdateResponder) { f(u, r) }

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
// It is possible to request updates from other peers by using their actor index.
//
// Returns nil if all peers accept the update. If any runtime error occurs or
// any peer rejects the update, an error is returned.
// nolint: funlen
func (c *Channel) Update(ctx context.Context, up ChannelUpdate) (err error) {
	if ctx == nil {
		return errors.New("context must not be nil")
	}
	if err := c.validTwoPartyUpdate(up); err != nil {
		return err
	}
	// Lock machine while update is in progress.
	if !c.machMtx.TryLockCtx(ctx) {
		return errors.Errorf("locking machine mutex in time: %v", ctx.Err())
	}
	defer c.machMtx.Unlock()

	if err = c.machine.Update(ctx, up.State, up.ActorIdx); err != nil {
		return errors.WithMessage(err, "updating machine")
	}
	// if anything goes wrong from now on, we discard the update.
	// TODO: this is insecure after we sent our signature.
	defer func() {
		if err != nil {
			if derr := c.machine.DiscardUpdate(ctx); derr != nil {
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
	// nolint:errcheck
	defer resRecv.Close()

	msgUpdate := &msgChannelUpdate{
		ChannelUpdate: up,
		Sig:           sig,
	}
	if err = c.conn.Send(ctx, msgUpdate); err != nil {
		return errors.WithMessage(err, "sending update")
	}

	pidx, res, err := resRecv.Next(ctx)
	if err != nil {
		return errors.WithMessage(err, "receiving update response")
	}
	c.Log().Tracef("Received update response (%T): %v", res, res)

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
// to all other channel participants. If the actor is different from the sender of the
// update, it is called a request.
//
// Returns nil if all peers accept the update. If any runtime error occurs or
// any peer rejects the update, an error is returned.
func (c *Channel) UpdateBy(ctx context.Context, actor channel.Index, update func(*channel.State)) (err error) {
	state := c.State().Clone()
	update(state)
	state.Version++

	return c.Update(ctx, ChannelUpdate{
		State:    state,
		ActorIdx: actor,
	})
}

// handleUpdateReq is called by the controller on incoming channel update
// requests.
func (c *Channel) handleUpdateReq(
	pidx channel.Index,
	req *msgChannelUpdate,
	uh UpdateHandler) {
	if err := c.validTwoPartyUpdate(req.ChannelUpdate); err != nil {
		// TODO: how to handle invalid updates? Just drop and ignore them?
		c.logPeer(pidx).Warnf("invalid update received: %v", err)
		return
	}

	c.machMtx.Lock() // Lock machine while update is in progress.
	defer c.machMtx.Unlock()

	if err := c.machine.CheckUpdate(req.State, req.ActorIdx, req.Sig, pidx); err != nil {
		// TODO: how to handle invalid updates? Just drop and ignore them?
		c.logPeer(pidx).Warnf("invalid update received: %v", err)
		return
	}

	responder := &UpdateResponder{channel: c, pidx: pidx, req: req}
	uh.HandleUpdate(req.ChannelUpdate, responder)
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
			if derr := c.machine.DiscardUpdate(ctx); derr != nil {
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
// * no locked sub-allocations.
func (c *Channel) validTwoPartyUpdate(up ChannelUpdate) error {
	if len(up.State.Locked) > 0 {
		return errors.New("no locked sub-allocations allowed")
	}
	return nil
}

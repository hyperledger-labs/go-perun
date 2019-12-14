// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

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

	UpdateHandler interface {
		Handle(ChannelUpdate, *UpdateResponder)
	}

	UpdateResponder struct {
		accept chan context.Context
		reject chan ctxUpdateRej
		err    chan error // return error
		called atomic.Bool
	}

	// The following type is only needed to bundle the ctx and channel update
	// rejection of UpdateResponder.Reject() into a single struct so that they can
	// be sent over a channel
	ctxUpdateRej struct {
		ctx    context.Context
		reason string
	}
)

func newUpdateResponder() *UpdateResponder {
	return &UpdateResponder{
		accept: make(chan context.Context),
		reject: make(chan ctxUpdateRej),
		err:    make(chan error, 1),
	}
}

// Accept lets the user signal that they want to accept the channel update.
func (r *UpdateResponder) Accept(ctx context.Context) error {
	if !r.called.TrySet() {
		log.Panic("multiple calls on channel update responder")
	}
	r.accept <- ctx
	return <-r.err
}

// Reject lets the user signal that they reject the channel update.
func (r *UpdateResponder) Reject(ctx context.Context, reason string) error {
	if !r.called.TrySet() {
		log.Panic("multiple calls on channel update responder")
	}
	r.reject <- ctxUpdateRej{ctx, reason}
	return <-r.err
}

func (c *Channel) Update(ctx context.Context, up ChannelUpdate) (err error) {
	if err := c.validTwoPartyUpdate(up, c.machine.Idx()); err != nil {
		return err
	}

	c.machMtx.Lock() // lock machine while update is in progress
	defer c.machMtx.Unlock()

	if err = c.machine.Update(up.State, up.ActorIdx); err != nil {
		return errors.WithMessage(err, "updating machine")
	}

	discardUpdate := true
	defer func() {
		if discardUpdate && err != nil {
			if derr := c.machine.DiscardUpdate(); derr != nil {
				// discarding update should never fail
				err = errors.WithMessagef(derr,
					"progressing update failed: %v, then discarding update failed", err)
			}
		}
	}()

	sig, err := c.machine.Sig()
	if err != nil {
		return errors.WithMessage(err, "signing update")
	}
	// from now on, we don't discard the update on errors
	discardUpdate = false

	resRecv, err := c.conn.NewUpdateResRecv(up.State.Version)
	if err != nil {
		return errors.WithMessage(err, "creating update response receiver")
	}
	defer resRecv.Close()

	msgUpdate := &msgChannelUpdate{
		ChannelUpdate: ChannelUpdate{
			State:    up.State,
			ActorIdx: c.machine.Idx(),
		},
		Sig: sig,
	}
	if err = c.conn.Send(ctx, msgUpdate); err != nil {
		return errors.WithMessage(err, "sending update acceptance")
	}

	pidx, res := resRecv.Next(ctx)
	if res == nil {
		return errors.WithMessage(err, "receiving update response")
	}

	if rej, ok := res.(*msgChannelUpdateRej); ok {
		// on reject, we discard the update again... TODO: alternative state
		discardUpdate = true
		return errors.Errorf("update rejected: %s", rej.Reason)
	}

	acc := res.(*msgChannelUpdateAcc) // safe by predicate of the updateResRecv
	if err := c.machine.AddSig(pidx, acc.Sig); err != nil {
		return errors.WithMessage(err, "adding peer signature")
	}

	return c.enableNotifyUpdate()
}

func (c *Channel) ListenUpdates(uh UpdateHandler) {
	for {
		pidx, req := c.conn.NextUpdateReq(context.Background())
		if req == nil {
			c.log.Debug("update request receiver closed")
			return
		}
		go c.handleUpdateReq(pidx, req, uh)
	}
}

func (c *Channel) handleUpdateReq(
	pidx channel.Index,
	req *msgChannelUpdate,
	uh UpdateHandler) {
	if err := c.validTwoPartyUpdate(req.ChannelUpdate, pidx); err != nil {
		// TODO: how to handle invalid updates? Just drop and ignore em?
		c.logPeer(pidx).Warnf("invalid update received: %v", err)
		return
	}

	c.machMtx.Lock() // lock machine while update is in progress
	defer c.machMtx.Unlock()

	if err := c.machine.CheckUpdate(req.State, req.ActorIdx, req.Sig, pidx); err != nil {
		// TODO: how to handle invalid updates? Just drop and ignore em?
		c.logPeer(pidx).Warnf("invalid update received: %v", err)
		return
	}

	res := newUpdateResponder()
	go uh.Handle(req.ChannelUpdate, res)

	// wait for user response
	select {
	case accCtx := <-res.accept:
		err := func(ctx context.Context) (err error) {
			// machine.Update and AddSig should never fail after CheckUpdate...
			if err = c.machine.Update(req.State, req.ActorIdx); err != nil {
				return errors.WithMessage(err, "updating machine")
			}
			// if anything below goes wrong, we discard the update
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

			if err = c.machine.AddSig(pidx, req.Sig); err != nil {
				return errors.WithMessage(err, "adding peer signature")
			}
			var sig wallet.Sig
			sig, err = c.machine.Sig()
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

			return c.enableNotifyUpdate()
		}(accCtx)

		if err != nil {
			c.logPeer(pidx).Errorf("error accepting state: %v", err)
		}
		res.err <- err

	case rej := <-res.reject:
		err := func(ctx context.Context) error {
			msgUpRej := &msgChannelUpdateRej{
				ChannelID: c.ID(),
				Version:   req.State.Version,
				Reason:    rej.reason,
			}
			return c.conn.Send(ctx, msgUpRej)
		}(rej.ctx)

		if err != nil {
			c.logPeer(pidx).Errorf("error rejecting state: %v", err)
		}
		res.err <- err
	}
}

// enableNotifyUpdate enables the current staging state of the machine. If the
// state is final, machine.EnableFinal is called. Finally, if there is a
// notification on channel updates, the enabled state is sent on it.
func (c *Channel) enableNotifyUpdate() error {
	var updater func() error
	if c.machine.StagingState().IsFinal {
		updater = c.machine.EnableFinal
	} else {
		updater = c.machine.EnableUpdate
	}

	if err := updater(); err != nil {
		return errors.WithMessage(c.machine.EnableUpdate(), "enabling update")
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

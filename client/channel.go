// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package client

import (
	"context"
	"sync"

	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/persistence"
	"perun.network/go-perun/log"
	"perun.network/go-perun/peer"
	perunsync "perun.network/go-perun/pkg/sync"
	"perun.network/go-perun/wallet"
)

// Channel is the channel controller, progressing the channel state machine and
// executing the channel update and dispute protocols.
//
// Currently, only the two-party protocol is fully implemented.
type Channel struct {
	perunsync.OnCloser
	log log.Logger

	conn        *channelConn
	machine     persistence.StateMachine
	machMtx     sync.RWMutex
	updateSub   chan<- *channel.State
	adjudicator channel.Adjudicator
	wallet      wallet.Wallet
}

// newChannel is internally used by the Client to create a new channel
// controller after the channel proposal protocol ran successfully.
func (c *Client) newChannel(
	acc wallet.Account,
	peers []*peer.Peer,
	params channel.Params,
) (*Channel, error) {
	machine, err := channel.NewStateMachine(acc, params)
	if err != nil {
		return nil, errors.WithMessage(err, "creating state machine")
	}

	pmachine := persistence.FromStateMachine(machine, c.pr)

	// bundle peers into channel connection
	conn, err := newChannelConn(params.ID(), peers, machine.Idx())
	if err != nil {
		return nil, errors.WithMessagef(err, "setting up channel connection")
	}

	logger := c.logChan(params.ID())
	conn.SetLogger(logger)
	return &Channel{
		OnCloser:    conn,
		log:         logger,
		conn:        conn,
		machine:     pmachine,
		adjudicator: c.adjudicator,
		wallet:      c.wallet,
	}, nil
}

// Close closes the channel and all associated peer subscriptions.
func (c *Channel) Close() error {
	return c.conn.Close()
}

// IsClosed returns whether the channel is closed.
func (c *Channel) IsClosed() bool {
	return c.conn.r.IsClosed()
}

// Ctx returns a context that is active for the channel's lifetime.
func (c *Channel) Ctx() context.Context {
	return c.conn.r.Ctx()
}

func (c *Channel) setLogger(l log.Logger) {
	c.conn.SetLogger(l)
	c.log = l
}

func (c *Channel) logPeer(idx channel.Index) log.Logger {
	return c.log.WithField("peerIdx", idx)
}

// ID returns the channel ID.
func (c *Channel) ID() channel.ID {
	return c.machine.ID()
}

// Idx returns our index in the channel.
func (c *Channel) Idx() channel.Index {
	return c.machine.Idx()
}

// Params returns the channel parameters.
func (c *Channel) Params() *channel.Params {
	return c.machine.Params()
}

// State returns the current state.
// Clone it if you want to modify it.
func (c *Channel) State() *channel.State {
	c.machMtx.RLock()
	defer c.machMtx.RUnlock()

	return c.machine.State()
}

// Phase returns the current phase of the channel state machine.
func (c *Channel) Phase() channel.Phase {
	c.machMtx.RLock()
	defer c.machMtx.RUnlock()

	return c.machine.Phase()
}

// init brings the state machine into the InitSigning phase. It is not callable
// by the user since the Client initializes the channel controller.
// The state machine is not locked as this function is expected to be called
// during the initialization phase of the channel controller.
func (c *Channel) init(ctx context.Context, initBals *channel.Allocation, initData channel.Data) error {
	return c.machine.Init(ctx, *initBals, initData)
}

// initExchangeSigsAndEnable exchanges signatures on the initial state.
// The state machine is not locked as this function is expected to be called
// during the initialization phase of the channel controller.
func (c *Channel) initExchangeSigsAndEnable(ctx context.Context) error {
	sig, err := c.machine.Sig(ctx)
	if err != nil {
		return err
	}

	resRecv, err := c.conn.NewUpdateResRecv(0)
	if err != nil {
		return errors.WithMessage(err, "creating update response receiver")
	}
	defer resRecv.Close()

	send := make(chan error)
	go func() {
		send <- c.conn.Send(ctx, &msgChannelUpdateAcc{
			ChannelID: c.ID(),
			Version:   0,
			Sig:       sig,
		})
	}()

	pidx, cm := resRecv.Next(ctx)
	acc, ok := cm.(*msgChannelUpdateAcc)
	if !ok {
		return errors.Errorf(
			"received unexpected message of type (%T) from peer[%d]: %v",
			cm, pidx, cm)
	}

	if err := c.machine.AddSig(ctx, pidx, acc.Sig); err != nil {
		return err
	}
	if err := c.machine.EnableInit(ctx); err != nil {
		return err
	}

	return errors.WithMessage(<-send, "sending initial signature")
}

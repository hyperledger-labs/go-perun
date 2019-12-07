// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package client

import (
	"context"
	"sync"

	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/log"
	"perun.network/go-perun/peer"
	perunsync "perun.network/go-perun/pkg/sync"
	"perun.network/go-perun/wallet"
	wire "perun.network/go-perun/wire/msg"
)

// Channel is the channel controller, progressing the channel state machine and
// executing the channel update and dispute protocols.
type Channel struct {
	perunsync.Closer
	log log.Logger

	conn    *channelConn
	machine channel.StateMachine
	machMtx sync.RWMutex
}

func newChannel(acc wallet.Account, peers []*peer.Peer, params channel.Params) (*Channel, error) {
	machine, err := channel.NewStateMachine(acc, params)
	if err != nil {
		return nil, errors.WithMessage(err, "creating state machine")
	}

	// bundle peers into channel connection
	conn, err := newChannelConn(params.ID(), peers, machine.Idx())
	if err != nil {
		return nil, errors.WithMessagef(err, "setting up channel connection")
	}

	return &Channel{
		log:     log.WithField("channel", params.ID), // default to global field logger
		conn:    conn,
		machine: *machine,
	}, nil
}

func (c *Channel) setLogger(l log.Logger) {
	c.log = l
}

func (c *Channel) logPeer(idx channel.Index) log.Logger {
	return c.log.WithField("peerIdx", idx)
}

func (c *Channel) ID() channel.ID {
	return c.machine.ID()
}

// init brings the state machine into the InitSigning phase. It is not callable
// by the user since the Client initializes the channel controller.
// The state machine is not locked as this function is expected to be called
// during the initialization phase of the channel controller.
func (c *Channel) init(initBals *channel.Allocation, initData channel.Data) error {
	return c.machine.Init(*initBals, initData)
}

// A channelConn bundles the message sending and receiving infrastructure for a
// channel. It is an abstraction over a set of peers. Peers are translated into
// their index in the channel.
type channelConn struct {
	b         *peer.Broadcaster
	r         *peer.Relay
	upReqRecv *peer.Receiver
	upResRecv *peer.Receiver
	peerIdx   map[*peer.Peer]channel.Index

	log log.Logger
}

// initExchangeSigsAndEnable exchanges signatures on the initial state.
// The state machine is not locked as this function is expected to be called
// during the initialization phase of the channel controller.
func (c *Channel) initExchangeSigsAndEnable(ctx context.Context) error {
	sig, err := c.machine.Sig()
	if err != nil {
		return err
	}

	send := make(chan error)
	go func() {
		send <- c.conn.send(ctx, &msgChannelUpdateAcc{
			ChannelID: c.ID(),
			Version:   0,
			Sig:       sig,
		})
	}()

	pidx, cm := c.conn.nextUpdateRes(ctx)
	acc, ok := cm.(*msgChannelUpdateAcc)
	if !ok {
		return errors.Errorf(
			"received unexpected message of type (%T) from peer: %v",
			cm, cm)
	}
	if acc.Version != 0 {
		return errors.Errorf(
			"received signature on unexpected version %d from peer",
			acc.Version)
	}

	if err := c.machine.AddSig(pidx, acc.Sig); err != nil {
		return err
	}
	if err := c.machine.EnableInit(); err != nil {
		return err
	}

	return errors.WithMessage(<-send, "sending initial signature")
}

// newChannelConn creates a new channel connection for the given channel ID. It
// subscribes on all peers to all messages regarding this channel. The order of
// the peers is important: it must match their position in the channel
// participant slice, or one less if their index is above our index, since we
// are not part of the peer slice.
func newChannelConn(id channel.ID, peers []*peer.Peer, idx channel.Index) (*channelConn, error) {
	forThisChannel := func(m wire.Msg) bool {
		cm, ok := m.(ChannelMsg)
		if !ok {
			return false
		}
		return cm.ID() == id
	}

	// setup receiving infrastructure:
	// 1. one relay to combine all channel messages from all peers
	// 2. two receivers for update requests and update responses
	relay := peer.NewRelay()
	peerIdx := make(map[*peer.Peer]channel.Index)
	for i, peer := range peers {
		i := channel.Index(i)
		peerIdx[peer] = i
		// We are not in the peer list, so the peer index is increased after our index
		if i >= idx {
			peerIdx[peer]++
		}
		if err := peer.Subscribe(relay, forThisChannel); err != nil {
			return nil, errors.WithMessagef(err,
				"subscribing relay to peer[%d] (%v)", i, peer)
		}
	}

	upReqRecv, upResRecv := peer.NewReceiver(), peer.NewReceiver()
	if err := relay.Subscribe(upReqRecv, func(m wire.Msg) bool {
		return m.Type() == wire.ChannelUpdate
	}); err != nil {
		return nil, errors.WithMessagef(err, "subscribing update request receiver")
	}
	if err := relay.Subscribe(upResRecv, func(m wire.Msg) bool {
		return (m.Type() == wire.ChannelUpdateAcc) ||
			(m.Type() == wire.ChannelUpdateRej)
	}); err != nil {
		return nil, errors.WithMessagef(err, "subscribing update response receiver")
	}

	return &channelConn{
		b:         peer.NewBroadcaster(peers),
		r:         relay,
		upReqRecv: upReqRecv,
		upResRecv: upResRecv,
		peerIdx:   peerIdx,
		log:       log.WithField("channel", id),
	}, nil
}

// send broadcasts the message to all channel participants
func (c *channelConn) send(ctx context.Context, msg wire.Msg) error {
	return c.b.Send(ctx, msg)
}

func (c *channelConn) nextUpdateReq(ctx context.Context) (channel.Index, *msgChannelUpdate) {
	peer, msg := c.upReqRecv.Next(ctx)
	idx, ok := c.peerIdx[peer]
	if !ok {
		c.log.Panicf("channel connection received message from unknown peer %v", peer)
	}
	return idx, msg.(*msgChannelUpdate) // safe by the predicate
}

func (c *channelConn) nextUpdateRes(ctx context.Context) (channel.Index, ChannelMsg) {
	peer, msg := c.upResRecv.Next(ctx)
	idx, ok := c.peerIdx[peer]
	if !ok {
		c.log.Panicf("channel connection received message from unknown peer %v", peer)
	}
	return idx, msg.(ChannelMsg) // safe by the predicate
}

func (c *channelConn) close() error {
	err := c.r.Close()
	if rerr := c.upReqRecv.Close(); err != nil && rerr != nil {
		err = rerr
	}
	if rerr := c.upResRecv.Close(); err != nil && rerr != nil {
		err = rerr
	}
	return err
}

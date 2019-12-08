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
		send <- c.conn.Send(ctx, &msgChannelUpdateAcc{
			ChannelID: c.ID(),
			Version:   0,
			Sig:       sig,
		})
	}()

	resRecv, err := c.conn.NewUpdateResRecv(0)
	if err != nil {
		return errors.WithMessage(err, "creating update response receiver")
	}
	defer resRecv.Close()

	pidx, cm := resRecv.Next(ctx)
	acc, ok := cm.(*msgChannelUpdateAcc)
	if !ok {
		return errors.Errorf(
			"received unexpected message of type (%T) from peer: %v",
			cm, cm)
	}

	if err := c.machine.AddSig(pidx, acc.Sig); err != nil {
		return err
	}
	if err := c.machine.EnableInit(); err != nil {
		return err
	}

	return errors.WithMessage(<-send, "sending initial signature")
}

// A channelConn bundles the message sending and receiving infrastructure for a
// channel. It is an abstraction over a set of peers. Peers are translated into
// their index in the channel.
type channelConn struct {
	b         *peer.Broadcaster
	r         *peer.Relay
	upReqRecv *peer.Receiver
	peerIdx   map[*peer.Peer]channel.Index

	log log.Logger
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

	upReqRecv := peer.NewReceiver()
	if err := relay.Subscribe(upReqRecv, func(m wire.Msg) bool {
		return m.Type() == wire.ChannelUpdate
	}); err != nil {
		return nil, errors.WithMessagef(err, "subscribing update request receiver")
	}

	return &channelConn{
		b:         peer.NewBroadcaster(peers),
		r:         relay,
		upReqRecv: upReqRecv,
		peerIdx:   peerIdx,
		log:       log.WithField("channel", id),
	}, nil
}

func (c *channelConn) Close() error {
	err := c.r.Close()
	if rerr := c.upReqRecv.Close(); err != nil && rerr != nil {
		err = rerr
	}
	return err
}

// send broadcasts the message to all channel participants
func (c *channelConn) Send(ctx context.Context, msg wire.Msg) error {
	return c.b.Send(ctx, msg)
}

func (c *channelConn) NextUpdateReq(ctx context.Context) (channel.Index, *msgChannelUpdate) {
	peer, msg := c.upReqRecv.Next(ctx)
	idx, ok := c.peerIdx[peer]
	if !ok {
		c.log.Panicf("channel connection received message from unknown peer %v", peer)
	}
	return idx, msg.(*msgChannelUpdate) // safe by the predicate
}

// newUpdateResRecv creates a new update response receiver for the given version.
// The receiver should be closed after all exptected responses are received.
// The receiver is also closed when the channel connection is closed.
func (c *channelConn) NewUpdateResRecv(version uint64) (*updateResRecv, error) {
	recv := peer.NewReceiver()
	if err := c.r.Subscribe(recv, func(m wire.Msg) bool {
		resMsg, ok := m.(channelUpdateResMsg)
		if !ok {
			return false
		}
		return resMsg.Ver() == version
	}); err != nil {
		return nil, errors.WithMessagef(err, "subscribing update response receiver")
	}

	resRecv := &updateResRecv{
		r:       recv,
		peerIdx: c.peerIdx,
		log:     c.log.WithField("version", version),
	}
	c.r.OnClose(func() { resRecv.Close() })
	return resRecv, nil
}

type updateResRecv struct {
	r       *peer.Receiver
	peerIdx map[*peer.Peer]channel.Index
	log     log.Logger
}

func (r *updateResRecv) Next(ctx context.Context) (channel.Index, channelUpdateResMsg) {
	peer, msg := r.r.Next(ctx)
	idx, ok := r.peerIdx[peer]
	if !ok {
		r.log.Panicf("channel connection received message from unknown peer %v", peer)
	}
	return idx, msg.(channelUpdateResMsg) // safe by the predicate
}

func (r *updateResRecv) Close() error {
	return r.r.Close()
}

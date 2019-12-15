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

	conn      *channelConn
	machine   channel.StateMachine
	machMtx   sync.RWMutex
	updateSub chan<- *channel.State
	settler   channel.Settler
}

func newChannel(
	acc wallet.Account,
	peers []*peer.Peer,
	params channel.Params,
	settler channel.Settler,
) (*Channel, error) {
	machine, err := channel.NewStateMachine(acc, params)
	if err != nil {
		return nil, errors.WithMessage(err, "creating state machine")
	}

	// bundle peers into channel connection
	conn, err := newChannelConn(params.ID(), peers, machine.Idx())
	if err != nil {
		return nil, errors.WithMessagef(err, "setting up channel connection")
	}

	logger := log.WithFields(log.Fields{"channel": params.ID, "id": acc.Address()})
	conn.SetLogger(logger)
	return &Channel{
		log:     logger,
		conn:    conn,
		machine: *machine,
		settler: settler,
	}, nil
}

// Close closes the channel and all associated peer subscriptions.
func (c *Channel) Close() error {
	if err := c.Closer.Close(); err != nil {
		return err
	}
	return c.conn.Close()
}

func (c *Channel) setLogger(l log.Logger) {
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

// State returns the current state.
// Clone it if you want to modify it.
func (c *Channel) State() *channel.State {
	return c.machine.State()
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

// Settle settles the channel using the Settler. The channel must be in a
// final state.
func (c *Channel) Settle(ctx context.Context) error {
	c.machMtx.Lock()
	defer c.machMtx.Unlock()

	// check final state
	if c.machine.Phase() != channel.Final || !c.machine.State().IsFinal {
		return errors.New("currently, only channels in a final state can be settled")
	}

	if err := c.settler.Settle(ctx, c.machine.SettleReq(), c.machine.Account()); err != nil {
		return errors.WithMessage(err, "calling settler")
	}

	return c.machine.SetSettled()
}

// A channelConn bundles the message sending and receiving infrastructure for a
// channel. It is an abstraction over a set of peers. Peers are translated into
// their index in the channel.
type channelConn struct {
	b         *peer.Broadcaster
	r         *peer.Relay
	upReqRecv *channelMsgRecv
	peerIdx   map[*peer.Peer]channel.Index

	log log.Logger
}

// newChannelConn creates a new channel connection for the given channel ID. It
// subscribes on all peers to all messages regarding this channel. The order of
// the peers is important: it must match their position in the channel
// participant slice, or one less if their index is above our index, since we
// are not part of the peer slice.
func newChannelConn(id channel.ID, peers []*peer.Peer, idx channel.Index) (_ *channelConn, err error) {
	// setup receiving infrastructure:
	// 1. one relay to combine all channel messages from all peers
	// 2. two receivers for update requests and update responses
	relay := peer.NewRelay()
	// we cache all channel messsages for the lifetime of the relay
	relay.Cache(context.Background(), func(wire.Msg) bool { return true })
	// Close the relay if anything goes wrong in the following.
	// We could have a leaky subscription otherwise.
	defer func() {
		if err != nil {
			if cerr := relay.Close(); cerr != nil {
				err = errors.WithMessagef(err,
					"error closing relay: %v, caused by error", cerr)
			}
		}
	}()

	forThisChannel := func(m wire.Msg) bool {
		cm, ok := m.(ChannelMsg)
		if !ok {
			return false
		}
		return cm.ID() == id
	}
	peerIdx := make(map[*peer.Peer]channel.Index)
	for i, peer := range peers {
		i := channel.Index(i)
		peerIdx[peer] = i
		// We are not in the peer list, so the peer index is increased after our index
		if i >= idx {
			peerIdx[peer]++
		}
		if err = peer.Subscribe(relay, forThisChannel); err != nil {
			return nil, errors.WithMessagef(err,
				"subscribing relay to peer[%d] (%v)", i, peer)
		}
	}

	logger := log.WithField("channel", id)
	upReqRecv := &channelMsgRecv{
		Receiver: peer.NewReceiver(),
		peerIdx:  peerIdx,
		log:      logger,
	}
	if err = relay.Subscribe(upReqRecv, func(m wire.Msg) bool {
		return m.Type() == wire.ChannelUpdate
	}); err != nil {
		return nil, errors.WithMessagef(err, "subscribing update request receiver")
	}

	return &channelConn{
		b:         peer.NewBroadcaster(peers),
		r:         relay,
		upReqRecv: upReqRecv,
		peerIdx:   peerIdx,
		log:       logger,
	}, nil
}

// SetLogger sets the logger of the channel connection. It is assumed to be
// called once before usage of the connection, so it isn't thread-safe.
func (c *channelConn) SetLogger(l log.Logger) {
	c.upReqRecv.log = l
	c.log = l
}

// Close closes the broadcaster and
func (c *channelConn) Close() error {
	err := c.r.Close()
	if rerr := c.upReqRecv.Close(); err == nil && rerr != nil {
		err = rerr
	}
	return err
}

// send broadcasts the message to all channel participants
func (c *channelConn) Send(ctx context.Context, msg wire.Msg) error {
	return c.b.Send(ctx, msg)
}

// NextUpdateReq returns the next channel update request that the channel
// connection receives. The update request receiver
func (c *channelConn) NextUpdateReq(ctx context.Context) (channel.Index, *msgChannelUpdate) {
	idx, m := c.upReqRecv.Next(ctx)
	if m == nil {
		return idx, nil // nil conversion doesn't work...
	}
	return idx, m.(*msgChannelUpdate) // safe by the predicate
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
		channelMsgRecv{
			Receiver: recv,
			peerIdx:  c.peerIdx,
			log:      c.log.WithField("version", version),
		},
	}
	return resRecv, nil
}

type (
	// A channelMsgRecv is a receiver of channel messages. Messages are received
	// with Next(), which returns the peer's channel index and the message.
	channelMsgRecv struct {
		*peer.Receiver
		peerIdx map[*peer.Peer]channel.Index
		log     log.Logger
	}

	// An updateResRecv is just a wrapper around a channelMsgRecv whose Next()
	// method returns channelUpdateResMsg's.
	updateResRecv struct {
		channelMsgRecv
	}
)

// Next returns the next message. If the receiver is closed or the context is
// done, (0, nil) is returned.
func (r *channelMsgRecv) Next(ctx context.Context) (channel.Index, ChannelMsg) {
	peer, msg := r.Receiver.Next(ctx)
	if peer == nil || msg == nil {
		return 0, nil // receiver was closed or context is done
	}
	idx, ok := r.peerIdx[peer]
	if !ok {
		r.log.Panicf("channel connection received message from unknown peer %v", peer)
	}
	return idx, msg.(ChannelMsg) // predicate must guarantee that this is safe
}

func (r *updateResRecv) Next(ctx context.Context) (channel.Index, channelUpdateResMsg) {
	idx, m := r.channelMsgRecv.Next(ctx)
	if m == nil {
		return idx, nil // nil conversion doesn't work...
	}
	return idx, m.(channelUpdateResMsg) // predicate must guarantee that this is safe
}

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
	"perun.network/go-perun/peer"
	wire "perun.network/go-perun/wire/msg"
)

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
		return ok && cm.ID() == id
	}
	peerIdx := make(map[*peer.Peer]channel.Index)
	for i, peer := range peers {
		i := channel.Index(i)
		peerIdx[peer] = i
		// We are not in the peer list, so the peer index is increased after our index.
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

// Close closes the broadcaster and update request receiver.
func (c *channelConn) Close() error {
	err := c.r.Close()
	if rerr := c.upReqRecv.Close(); err == nil && rerr != nil {
		err = rerr
	}
	return err
}

// send broadcasts the message to all channel participants.
func (c *channelConn) Send(ctx context.Context, msg wire.Msg) error {
	return c.b.Send(ctx, msg)
}

// NextUpdateReq returns the next channel update request that the channel
// connection receives.
func (c *channelConn) NextUpdateReq(ctx context.Context) (channel.Index, *msgChannelUpdate) {
	idx, m := c.upReqRecv.Next(ctx)
	if m == nil {
		return idx, nil // nil conversion doesn't work...
	}
	return idx, m.(*msgChannelUpdate) // safe by the predicate
}

// newUpdateResRecv creates a new update response receiver for the given version.
// The receiver should be closed after all expected responses are received.
// The receiver is also closed when the channel connection is closed.
func (c *channelConn) NewUpdateResRecv(version uint64) (*channelMsgRecv, error) {
	recv := peer.NewReceiver()
	if err := c.r.Subscribe(recv, func(m wire.Msg) bool {
		resMsg, ok := m.(channelVerMsg)
		return ok && resMsg.Ver() == version
	}); err != nil {
		return nil, errors.WithMessagef(err, "subscribing update response receiver")
	}

	return &channelMsgRecv{
		Receiver: recv,
		peerIdx:  c.peerIdx,
		log:      c.log.WithField("version", version),
	}, nil
}

type (
	// A channelMsgRecv is a receiver of channel messages. Messages are received
	// with Next(), which returns the peer's channel index and the message.
	channelMsgRecv struct {
		*peer.Receiver
		peerIdx map[*peer.Peer]channel.Index
		log     log.Logger
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

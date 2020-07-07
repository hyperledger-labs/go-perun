// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package client

import (
	"context"

	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/log"
	"perun.network/go-perun/pkg/sync"
	"perun.network/go-perun/wire"
)

// A channelConn bundles the message sending and receiving infrastructure for a
// channel. It is an abstraction over a set of peers. Peers are translated into
// their index in the channel.
type channelConn struct {
	sync.OnCloser

	b       *wire.Broadcaster
	r       *wire.Relay // update response relay
	peerIdx map[*wire.Endpoint]channel.Index

	log log.Logger
}

// newChannelConn creates a new channel connection for the given channel ID. It
// subscribes on all peers to all messages regarding this channel. The order of
// the peers is important: it must match their position in the channel
// participant slice, or one less if their index is above our index, since we
// are not part of the peer slice.
func newChannelConn(id channel.ID, peers []*wire.Endpoint, idx channel.Index) (_ *channelConn, err error) {
	// relay to receive all update responses
	relay := wire.NewRelay()
	// we cache all responses for the lifetime of the relay
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

	isUpdateRes := func(m wire.Msg) bool {
		ok := m.Type() == wire.ChannelUpdateAcc || m.Type() == wire.ChannelUpdateRej
		return ok && m.(ChannelMsg).ID() == id
	}

	peerIdx := make(map[*wire.Endpoint]channel.Index)
	for i, peer := range peers {
		i := channel.Index(i)
		peerIdx[peer] = i
		// We are not in the peer list, so the peer index is increased after our index.
		if i >= idx {
			peerIdx[peer]++
		}
		if err = peer.Subscribe(relay, isUpdateRes); err != nil {
			return nil, errors.WithMessagef(err,
				"subscribing relay to peer[%d] (%v)", i, peer)
		}
	}

	ch := &channelConn{
		OnCloser: relay,
		b:        wire.NewBroadcaster(peers),
		r:        relay,
		peerIdx:  peerIdx,
		log:      log.WithField("channel", id),
	}
	for _, peer := range peers {
		peer.OnCloseAlways(func() { ch.Close() })
	}
	return ch, nil
}

// SetLogger sets the logger of the channel connection. It is assumed to be
// called once before usage of the connection, so it isn't thread-safe.
func (c *channelConn) SetLogger(l log.Logger) {
	c.log = l
}

// Close closes the broadcaster and update request receiver.
func (c *channelConn) Close() error {
	return c.r.Close()
}

// Send broadcasts the message to all channel participants.
func (c *channelConn) Send(ctx context.Context, msg wire.Msg) error {
	return c.b.Send(ctx, msg)
}

// Peers returns the ordered list of peer addresses. Note that the length is
// the number of channel participants minus one, since the own peer is excluded.
func (c *channelConn) Peers() []wire.Address {
	ps := make([]wire.Address, len(c.peerIdx)+1) // +1 for own nil entry
	for p, i := range c.peerIdx {
		ps[i] = p.PerunAddress
	}

	// clean possible nil entry for own peer
	for i, a := range ps {
		if a == nil {
			ps = append(ps[:i], ps[i+1:]...)
			return ps // there's only at most one nil entry
		}
	}
	return ps
}

// newUpdateResRecv creates a new update response receiver for the given version.
// The receiver should be closed after all expected responses are received.
// The receiver is also closed when the channel connection is closed.
func (c *channelConn) NewUpdateResRecv(version uint64) (*channelMsgRecv, error) {
	recv := wire.NewReceiver()
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
		*wire.Receiver
		peerIdx map[*wire.Endpoint]channel.Index
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

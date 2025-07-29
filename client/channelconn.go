// Copyright 2025 - See NOTICE file for copyright holders.
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
	"math"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"perun.network/go-perun/wallet"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/log"
	"perun.network/go-perun/wire"
	"polycry.pt/poly-go/sync"
)

// A channelConn bundles the message sending and receiving infrastructure for a
// channel. It is an abstraction over a set of peers. Peers are translated into
// their index in the channel.
type channelConn struct {
	sync.OnCloser

	pub   wire.Publisher // outgoing message publisher
	r     *wire.Relay    // update response relay/incoming messages
	peers []map[wallet.BackendID]wire.Address
	idx   channel.Index // our index

	log log.Logger
}

// Send broadcasts the message to all channel participants.
func (c *channelConn) Send(ctx context.Context, msg wire.Msg) error {
	var eg errgroup.Group
	for i, peer := range c.peers {
		if i < 0 || i > math.MaxUint16 {
			return errors.Errorf("channel connection peer index out of bounds: %d", i)
		}
		if channel.Index(i) == c.idx {
			continue // skip own peer
		}
		c.log.WithField("peer", peer).Debugf("channelConn: publishing message: %v: %+v", msg.Type(), msg)
		env := &wire.Envelope{
			Sender:    c.sender(),
			Recipient: peer,
			Msg:       msg,
		}
		eg.Go(func() error { return c.pub.Publish(ctx, env) })
	}
	return errors.WithMessage(eg.Wait(), "publishing message")
}

// Peers returns the ordered list of peer addresses. Note that the own peer is
// included in the list.
func (c *channelConn) Peers() []map[wallet.BackendID]wire.Address {
	return c.peers
}

// newUpdateResRecv creates a new update response receiver for the given version.
// The receiver should be closed after all expected responses are received.
// The receiver is also closed when the channel connection is closed.
func (c *channelConn) NewUpdateResRecv(version uint64) (*channelMsgRecv, error) {
	recv := wire.NewReceiver()
	if err := c.r.Subscribe(recv, func(e *wire.Envelope) bool {
		resMsg, ok := e.Msg.(channelUpdateResMsg)
		return ok && resMsg.Ver() == version
	}); err != nil {
		return nil, errors.WithMessagef(err, "subscribing update response receiver")
	}

	return &channelMsgRecv{
		Receiver: recv,
		peers:    c.peers,
		log:      c.log.WithField("version", version),
	}, nil
}

// Close closes the broadcaster and update request receiver.
func (c *channelConn) Close() error {
	return c.r.Close()
}

// newChannelConn creates a new channel connection for the given channel ID. It
// subscribes on the subscriber to all messages regarding this channel.
func newChannelConn(id channel.ID, peers []map[wallet.BackendID]wire.Address, idx channel.Index, sub wire.Subscriber, pub wire.Publisher) (_ *channelConn, err error) {
	// relay to receive all update responses
	relay := wire.NewRelay()
	// we cache all responses for the lifetime of the relay
	cacheAll := func(*wire.Envelope) bool { return true }
	relay.Cache(&cacheAll)
	// Close the relay if anything goes wrong in the following.
	// We could have a leaky subscription otherwise.
	defer func() {
		if err != nil {
			if cerr := relay.Close(); cerr != nil {
				err = errors.WithMessagef(err,
					"(error closing relay: %v)", cerr)
			}
		}
	}()

	isUpdateRes := func(e *wire.Envelope) bool {
		switch msg := e.Msg.(type) {
		case *ChannelUpdateAccMsg:
			return msg.ID() == id
		case *ChannelUpdateRejMsg:
			return msg.ID() == id
		default:
			return false
		}
	}

	if err = sub.Subscribe(relay, isUpdateRes); err != nil {
		return nil, errors.WithMessagef(err, "subscribing relay")
	}

	return &channelConn{
		OnCloser: relay,
		r:        relay,
		pub:      pub,
		peers:    peers,
		idx:      idx,
		log:      log.WithField("channel", id),
	}, nil
}

// SetLog sets the logger of the channel connection. It is assumed to be
// called once before usage of the connection, so it isn't thread-safe.
func (c *channelConn) SetLog(l log.Logger) {
	c.log = l
}

func (c *channelConn) sender() map[wallet.BackendID]wire.Address {
	return c.peers[c.idx]
}

type (
	// A channelMsgRecv is a receiver of channel messages. Messages are received
	// with Next(), which returns the peer's channel index and the message.
	channelMsgRecv struct {
		*wire.Receiver

		peers []map[wallet.BackendID]wire.Address
		log   log.Logger
	}
)

// Next returns the next message. If the receiver is closed or the context is
// done, (0, nil) is returned.
func (r *channelMsgRecv) Next(ctx context.Context) (channel.Index, ChannelMsg, error) {
	env, err := r.Receiver.Next(ctx)
	if err != nil {
		return 0, nil, err
	}
	idx := wire.IndexOfAddrs(r.peers, env.Sender)
	if idx == -1 {
		return 0, nil, errors.Errorf("channel connection received message from unexpected peer %v", env.Sender)
	}
	msg, ok := env.Msg.(ChannelMsg)
	if !ok {
		return 0, nil, errors.Errorf("unexpected message type: expected ChannelMsg, got %T", env.Msg)
	}
	index, err := channel.FromInt(idx)
	if err != nil {
		return 0, nil, errors.WithMessagef(err, "converting index %d to channel.Index", idx)
	}

	return index, msg, nil
}

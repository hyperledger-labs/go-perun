// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package client // import "perun.network/go-perun/client"

import (
	"context"

	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/log"
	"perun.network/go-perun/peer"
	"perun.network/go-perun/pkg/sync"
	"perun.network/go-perun/wallet"
	wire "perun.network/go-perun/wire/msg"
)

// Client is a state channel client. It is the central controller to interact
// with a state channel network. It can be used to propose channels to other
// channel network peers.
//
// Currently, only the two-party protocol is fully implemented.
type Client struct {
	id          peer.Identity
	peers       *peer.Registry
	propHandler ProposalHandler
	funder      channel.Funder
	settler     channel.Settler
	log         log.Logger // structured logger for this client

	sync.Closer
}

// New creates a new State Channel Client.
//
// id is the channel network identity. It is the persistend identifier in the
// network and not necessarily related to any on-chain identity or channel
// identity.
//
// dialer is used to dial new peers when a peer connection is not yet
// established, e.g. when proposing a channel.
//
// proposalHandler is the user callback that is called by the Client when a peer
// proposes a valid channel to this Client.
//
// funder and settler are used to fund and settle a ledger channel, respectively.
//
// If any argument is nil, New panics.
func New(
	id peer.Identity,
	dialer peer.Dialer,
	propHandler ProposalHandler,
	funder channel.Funder,
	settler channel.Settler,
) *Client {
	if id == nil || dialer == nil || propHandler == nil || funder == nil || settler == nil {
		log.Panic("invalid nil argument")
	}

	c := &Client{
		id:          id,
		propHandler: propHandler,
		funder:      funder,
		settler:     settler,
		log:         log.WithField("id", id.Address()),
	}
	c.peers = peer.NewRegistry(id, c.subscribePeer, dialer)
	return c
}

// Close closes this state channel client.
// It also closes the peer registry.
func (c *Client) Close() error {
	if err := c.Closer.Close(); err != nil {
		return err
	}

	return errors.WithMessage(c.peers.Close(), "closing registry")
}

// Listen starts listening for incoming connections on the provided listener and
// currently just automatically accepts them after successful authentication.
// This function does not start go routines but instead should
// be started by the user as `go client.Listen()`. The client takes ownership of
// the listener and will close it when the client is closed.
func (c *Client) Listen(listener peer.Listener) {
	c.peers.Listen(listener)
}

func (c *Client) subscribePeer(p *peer.Peer) {
	c.logPeer(p).Debugf("setting up default subscriptions")

	// handle incoming channel proposals
	c.subChannelProposals(p)

	log := c.logPeer(p)
	p.SetDefaultMsgHandler(func(m wire.Msg) {
		log.Debugf("Received %T message without subscription: %v", m, m)
	})
}

// Log is the getter for the client's field logger.
func (c *Client) Log() log.Logger {
	return c.log
}

func (c *Client) logPeer(p *peer.Peer) log.Logger {
	return c.log.WithField("peer", p.PerunAddress)
}

func (c *Client) logChan(id channel.ID) log.Logger {
	return c.log.WithField("channel", id)
}

// getPeers gets all peers from the registry for the provided addresses,
// skipping the own peer, if present in the list.
func (c *Client) getPeers(
	ctx context.Context,
	addrs []peer.Address,
) (peers []*peer.Peer, err error) {
	idx := wallet.IndexOfAddr(addrs, c.id.Address())
	l := len(addrs)
	if idx != -1 {
		l--
	}

	peers = make([]*peer.Peer, l)
	for i, a := range addrs {
		if idx == -1 || i < idx {
			peers[i], err = c.peers.Get(ctx, a)
		} else if i > idx {
			peers[i-1], err = c.peers.Get(ctx, a)
		}
		if err != nil {
			return
		}
	}

	return
}

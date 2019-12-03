// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package client

import (
	"github.com/pkg/errors"

	"perun.network/go-perun/log"
	"perun.network/go-perun/peer"
	"perun.network/go-perun/pkg/sync"

	wire "perun.network/go-perun/wire/msg"
)

type Client struct {
	id          peer.Identity
	peers       *peer.Registry
	propHandler ProposalHandler
	log         log.Logger // structured logger for this client

	sync.Closer
}

func New(id peer.Identity, dialer peer.Dialer, propHandler ProposalHandler) *Client {
	c := &Client{
		id:          id,
		propHandler: propHandler,
		log:         log.WithField("client", id.Address),
	}
	c.peers = peer.NewRegistry(id, c.subscribePeer, dialer)
	return c
}

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

func (c *Client) logPeer(p *peer.Peer) log.Logger {
	return c.log.WithField("peer", p.PerunAddress)
}

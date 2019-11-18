// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package client

import (
	"github.com/pkg/errors"

	"perun.network/go-perun/log"
	"perun.network/go-perun/peer"
	"perun.network/go-perun/pkg/sync/atomic"

	wire "perun.network/go-perun/wire/msg"
)

type Client struct {
	id          peer.Identity
	peers       *peer.Registry
	propHandler ProposalHandler
	log         log.Logger // structured logger for this client
	quit        chan struct{}
	closed      atomic.Bool
}

func New(id peer.Identity, dialer peer.Dialer, propHandler ProposalHandler) *Client {
	c := &Client{
		id:          id,
		propHandler: propHandler,
		quit:        make(chan struct{}),
		log:         log.WithField("client", id.Address),
	}
	c.peers = peer.NewRegistry(c.subscribePeer, dialer)
	return c
}

func (c *Client) Close() error {
	if !c.closed.TrySet() {
		return errors.New("client already closed")
	}
	close(c.quit)

	return errors.WithMessage(c.peers.Close(), "closing registry")
}

// Listen starts listening for incoming connections on the provided listener and
// currently just automatically accepts them after successful authentication.
// This function does not start go routines but instead should
// be started by the user as `go client.Listen()`.
func (c *Client) Listen(listener peer.Listener) {
	go func() {
		<-c.quit
		if err := listener.Close(); err != nil {
			c.log.Debugf("Closing listener while closing client failed: %v", err)
		}
	}()
	// start listener and accept all incoming peer connections, writing them to the registry
	for {
		conn, err := listener.Accept()
		if err != nil {
			c.log.Debugf("peer listener closed: %v", err)
			return
		}

		// setup connection in a serparate routine so that new incoming connections
		// can immediately be handled.
		go c.setupConn(conn)
	}
}

func (c *Client) setupConn(conn peer.Conn) {
	if peerAddr, err := peer.ExchangeAddrs(c.id, conn); err != nil {
		c.log.Warnf("could not authenticate peer: %v", err)
	} else {
		// the peer registry is thread safe
		c.peers.Register(peerAddr, conn)
	}
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

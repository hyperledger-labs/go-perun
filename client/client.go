// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package client

import (
	"perun.network/go-perun/log"
	"perun.network/go-perun/peer"
)

type Client struct {
	id      peer.Identity
	peerReg *peer.Registry
	quit        chan struct{}
	log         log.Logger // structured logger for this client
}

func New(id peer.Identity, dialer peer.Dialer) *Client {
	c := &Client{
		id: id,
		quit:        make(chan struct{}),
		log:         log.WithField("client", id.Address),
	}
	c.peerReg = peer.NewRegistry(c.subscribePeer, dialer)
	return c
}

func (c *Client) Close() {
	close(c.quit)
}

// Listen starts listening for incoming connections on the provided listener and
// currently just automatically accepts them after successful authentication.
// This function does not start go routines but instead should
// be started by the user as `go client.Listen()`.
func (c *Client) Listen(listener peer.Listener) {
	// start listener and accept all incoming peer connections, writing them to the registry
	for {
		conn, err := listener.Accept()
		if err != nil {
			c.log.Debugf("peer listener closed: %v", err)
			return
		}

		if peerAddr, err := peer.ExchangeAddrs(c.id, conn); err != nil {
			c.log.Warnf("could not authenticate peer: %v", err)
		} else {
			// the peer registry is thread safe
			c.peerReg.Register(peerAddr, conn)
		}
	}
}

func (c *Client) subscribePeer(p *peer.Peer) {
	log.Debugf("Client: subscribing peer: %v", p.PerunAddress)
	// TODO actual subscriptions
}

// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package client

import (
	"context"

	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/persistence"
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
	channels    chanRegistry
	propRecv    *peer.Receiver
	funder      channel.Funder
	adjudicator channel.Adjudicator
	wallet      wallet.Wallet
	pr          persistence.PersistRestorer
	log         log.Logger // structured logger for this client

	sync.Closer
}

// New creates a new State Channel Client.
//
// id is the channel network identity. It is the persistend identifier in the
// network and not necessarily related to any on-chain identity or channel
// identity.
//
// The dialer is used to dial new peers when a peer connection is not yet
// established, e.g. when proposing a channel.
//
// The funder and adjudicator are used to fund and dispute or settle a ledger
// channel, respectively.
//
// The wallet is used to resolve addresses to accounts when creating or
// restoring channels.
//
// If any argument is nil, New panics.
func New(
	id peer.Identity,
	dialer peer.Dialer,
	funder channel.Funder,
	adjudicator channel.Adjudicator,
	wallet wallet.Wallet,
) *Client {
	if id == nil {
		log.Panic("identity must not be nil")
	}
	log := log.WithField("id", id.Address())
	if dialer == nil {
		log.Panic("dialer must not be nil")
	} else if funder == nil {
		log.Panic("funder must not be nil")
	} else if adjudicator == nil {
		log.Panic("adjudicator must not be nil")
	} else if wallet == nil {
		log.Panic("wallet must not be nil")
	}

	c := &Client{
		id:          id,
		channels:    makeChanRegistry(),
		propRecv:    peer.NewReceiver(),
		funder:      funder,
		adjudicator: adjudicator,
		wallet:      wallet,
		pr:          persistence.NonPersistRestorer,
		log:         log,
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

	err := errors.WithMessage(c.channels.CloseAll(), "closing channels")
	if cerr := c.peers.Close(); err == nil {
		err = errors.WithMessage(cerr, "closing registry")
	}
	if rerr := c.propRecv.Close(); err == nil {
		err = errors.WithMessage(rerr, "closing proposal receiver")
	}
	return err
}

// EnablePersistence sets the PersistRestorer that the client is going to use for channel
// persistence. This methods is expected to be called once during the setup of
// the client and is hence not thread-safe.
func (c *Client) EnablePersistence(pr persistence.PersistRestorer) {
	c.pr = pr
}

// Channel queries a channel by its ID.
func (c *Client) Channel(id channel.ID) (*Channel, error) {
	if ch, ok := c.channels.Get(id); ok {
		return ch, nil
	}
	return nil, errors.New("unknown channel ID")
}

// Listen starts listening for incoming connections on the provided listener and
// currently just automatically accepts them after successful authentication.
// This function does not start go routines but instead should
// be started by the user as `go client.Listen()`. The client takes ownership of
// the listener and will close it when the client is closed.
func (c *Client) Listen(listener peer.Listener) {
	if listener == nil {
		c.log.Panic("listener must not be nil")
	}

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

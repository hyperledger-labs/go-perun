// Copyright 2019 - See NOTICE file for copyright holders.
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

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/persistence"
	"perun.network/go-perun/log"
	"perun.network/go-perun/pkg/sync"
	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wire"
)

// Client is a state channel client. It is the central controller to interact
// with a state channel network. It can be used to propose channels to other
// channel network peers.
//
// Currently, only the two-party protocol is fully implemented.
type Client struct {
	address     wire.Address
	conn        clientConn
	channels    chanRegistry
	funder      channel.Funder
	adjudicator channel.Adjudicator
	wallet      wallet.Wallet
	pr          persistence.PersistRestorer
	log         log.Logger // structured logger for this client

	sync.Closer
}

// New creates a new State Channel Client.
//
// address is the channel network address of this client. It is the persistend
// identifier in the network and not necessarily related to any on-chain
// identity or channel participant address.
//
// bus is the wire protocol bus over which messages to other network clients are
// sent and received from.
//
// The funder and adjudicator are used to fund and dispute or settle a ledger
// channel, respectively.
//
// The wallet is used to resolve addresses to accounts when creating or
// restoring channels.
//
// If any argument is nil, New panics.
func New(
	address wire.Address,
	bus wire.Bus,
	funder channel.Funder,
	adjudicator channel.Adjudicator,
	wallet wallet.Wallet,
) (c *Client, err error) {
	if address == nil {
		log.Panic("address must not be nil")
	}
	log := log.WithField("id", address)
	// nolint: gocritic
	if bus == nil {
		log.Panic("bus must not be nil")
	} else if funder == nil {
		log.Panic("funder must not be nil")
	} else if adjudicator == nil {
		log.Panic("adjudicator must not be nil")
	} else if wallet == nil {
		log.Panic("wallet must not be nil")
	}

	conn, err := makeClientConn(address, bus)
	if err != nil {
		return nil, errors.WithMessage(err, "setting up client connection")
	}

	return &Client{
		address:     address,
		conn:        conn,
		channels:    makeChanRegistry(),
		funder:      funder,
		adjudicator: adjudicator,
		wallet:      wallet,
		pr:          persistence.NonPersistRestorer,
		log:         log,
	}, nil
}

// Close closes this state channel client.
// It also closes the peer registry.
func (c *Client) Close() error {
	if err := c.Closer.Close(); err != nil {
		return err
	}

	err := errors.WithMessage(c.channels.CloseAll(), "closing channels")
	if cerr := c.conn.Close(); err == nil {
		err = errors.WithMessage(cerr, "closing channel connection")
	}
	return err
}

// OnNewChannel sets a callback to be called whenever a new channel is created
// or restored. Only one such handler can be set at a time, and repeated calls
// to this function will overwrite the currently existing handler. This function
// may be safely called at any time.
func (c *Client) OnNewChannel(handler func(*Channel)) {
	c.channels.OnNewChannel(handler)
}

// EnablePersistence sets the PersistRestorer that the client is going to use for channel
// persistence. This methods is expected to be called once during the setup of
// the client and is hence not thread-safe.
//
// The PersistRestorer is not closed when the Client is closed.
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

// Handle is the incoming request handler routine. It handles channel proposals
// and channel update requests. It must be started exactly once by the user,
// during the setup of the Client. Incoming requests are handled by the passed
// respecive handlers.
func (c *Client) Handle(ph ProposalHandler, uh UpdateHandler) {
	if ph == nil || uh == nil {
		c.log.Panic("handlers must not be nil")
	}

	for {
		env, err := c.conn.nextReq(c.Ctx())
		if err != nil {
			c.log.Debug("request receiver closed: ", err)
			return
		}
		msg := env.Msg

		switch msg.Type() {
		case wire.ChannelProposal:
			go c.handleChannelProposal(ph, env.Sender, msg.(*ChannelProposal))
		case wire.ChannelUpdate:
			go c.handleChannelUpdate(uh, env.Sender, msg.(*msgChannelUpdate))
		case wire.ChannelSync:
			go c.handleSyncMsg(env.Sender, msg.(*msgChannelSync))
		default:
			c.log.Error("Unexpected %T message received in request loop")
		}
	}
}

// Log returns the logger used by the client. It is not thread-safe.
func (c *Client) Log() log.Logger {
	return c.log
}

// SetLog sets the logger used by the client. It is not thread-safe.
func (c *Client) SetLog(l log.Logger) {
	c.log = l
}

func (c *Client) logPeer(p wire.Address) log.Logger {
	return c.log.WithField("peer", p)
}

func (c *Client) logChan(id channel.ID) log.Logger {
	return c.log.WithField("channel", id)
}

// Restore restores all channels from persistence. Channels are restored in
// parallel. Newly restored channels. should be acquired through the
// OnNewChannel callback.
func (c *Client) Restore(ctx context.Context) error {
	ps, err := c.pr.ActivePeers(ctx)
	if err != nil {
		return errors.WithMessage(err, "restoring active peers")
	}

	var eg errgroup.Group
	for _, p := range ps {
		if p.Equals(c.address) {
			continue // skip own peer
		}
		p := p
		eg.Go(func() error { return c.restorePeerChannels(ctx, p) })
	}

	return eg.Wait()
}

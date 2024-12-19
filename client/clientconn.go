// Copyright 2020 - See NOTICE file for copyright holders.
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

	"perun.network/go-perun/wallet"

	"github.com/pkg/errors"

	"perun.network/go-perun/log"
	"perun.network/go-perun/wire"
)

// A clientConn bundles all the messaging infrastructure for a Client.
type clientConn struct {
	*wire.Relay // Client relay, subscribed to the bus. Embedded for methods Subscribe and Cache.
	bus         wire.Bus
	reqRecv     *wire.Receiver // subscription to incoming requests
	sender      map[wallet.BackendID]wire.Address
	log.Embedding
}

func makeClientConn(address map[wallet.BackendID]wire.Address, bus wire.Bus) (c clientConn, err error) {
	c.Embedding = log.MakeEmbedding(log.WithField("id", address))
	c.sender = address
	c.bus = bus
	c.Relay = wire.NewRelay()
	defer func() {
		if err != nil {
			if cerr := c.Relay.Close(); cerr != nil {
				err = errors.WithMessagef(err, "(error closing bus: %v)", cerr)
			}
		}
	}()

	c.Relay.SetDefaultMsgHandler(func(m *wire.Envelope) {
		log.Debugf("Received %T message without subscription: %v", m.Msg, m)
	})
	if err := bus.SubscribeClient(c, c.sender); err != nil {
		return c, errors.WithMessage(err, "subscribing client on bus")
	}

	c.reqRecv = wire.NewReceiver()
	if err := c.Subscribe(c.reqRecv, isReqMsg); err != nil {
		return c, errors.WithMessage(err, "subscribing request receiver")
	}

	return c, nil
}

func isReqMsg(m *wire.Envelope) bool {
	return m.Msg.Type() == wire.LedgerChannelProposal ||
		m.Msg.Type() == wire.SubChannelProposal ||
		m.Msg.Type() == wire.VirtualChannelProposal ||
		m.Msg.Type() == wire.VirtualChannelFundingProposal ||
		m.Msg.Type() == wire.VirtualChannelSettlementProposal ||
		m.Msg.Type() == wire.ChannelUpdate ||
		m.Msg.Type() == wire.ChannelSync
}

func (c clientConn) nextReq(ctx context.Context) (*wire.Envelope, error) {
	return c.reqRecv.Next(ctx)
}

// pubMsg publishes the given message on the wire bus, setting the own client as
// the sender.
func (c *clientConn) pubMsg(ctx context.Context, msg wire.Msg, rec map[wallet.BackendID]wire.Address) error {
	c.Log().WithField("peer", rec).Debugf("Publishing message: %v: %+v", msg.Type(), msg)
	return c.bus.Publish(ctx, &wire.Envelope{
		Sender:    c.sender,
		Recipient: rec,
		Msg:       msg,
	})
}

// Publish publishes the message on the bus. Makes clientConn implement the
// wire.Publisher interface.
func (c *clientConn) Publish(ctx context.Context, env *wire.Envelope) error {
	return c.bus.Publish(ctx, env)
}

func (c *clientConn) Close() error {
	err := c.Relay.Close()
	if rerr := c.reqRecv.Close(); err == nil {
		err = errors.WithMessage(rerr, "closing proposal receiver")
	}
	return err
}

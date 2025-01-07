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

	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/persistence"
	"perun.network/go-perun/wire"
)

type channelFromSourceSig = func(*Client, *persistence.Channel, *Channel, ...map[wallet.BackendID]wire.Address) (*Channel, error)

// clientChannelFromSource is the production behaviour of reconstructChannel.
// During testing, it is replaced by a simpler function that needs much less
// setup code.
func clientChannelFromSource(
	c *Client,
	ch *persistence.Channel,
	parent *Channel,
	peers ...map[wallet.BackendID]wire.Address,
) (*Channel, error) {
	return c.channelFromSource(ch, parent, peers)
}

func (c *Client) reconstructChannel(
	channelFromSource channelFromSourceSig,
	pch *persistence.Channel,
	db map[channel.ID]*persistence.Channel,
	chans map[channel.ID]*Channel,
) *Channel {
	if ch, ok := chans[pch.ID()]; ok {
		return ch
	}

	var parent *Channel
	if pch.Parent != nil {
		parent = c.reconstructChannel(
			channelFromSource,
			db[*pch.Parent],
			db,
			chans)
	}

	ch, err := channelFromSource(c, pch, parent, pch.PeersV...)
	if err != nil {
		c.logChan(pch.ID()).Panicf("Reconstruct channel: %v", err)
	}

	chans[pch.ID()] = ch
	return ch
}

func (c *Client) restorePeerChannels(ctx context.Context, p map[wallet.BackendID]wire.Address) (err error) {
	it, err := c.pr.RestorePeer(p)
	if err != nil {
		return errors.WithMessagef(err, "restoring channels for peer: %v", err)
	}
	defer func() {
		if cerr := it.Close(); cerr != nil {
			err = errors.WithMessagef(err, "(error closing iterator: %v)", cerr)
		}
	}()

	db := make(map[channel.ID]*persistence.Channel)

	// Serially restore channels. We might change this to parallel restoring once
	// we initiate the sync protocol from here again.
	for it.Next(ctx) {
		chdata := it.Channel()
		db[chdata.ID()] = chdata
	}

	if err := it.Close(); err != nil {
		return err
	}

	c.restoreChannelCollection(db, clientChannelFromSource)
	return nil
}

func (c *Client) restoreChannelCollection(
	db map[channel.ID]*persistence.Channel,
	channelFromSource channelFromSourceSig,
) {
	chs := make(map[channel.ID]*Channel)
	for _, pch := range db {
		ch := c.reconstructChannel(channelFromSource, pch, db, chs)
		log := c.logChan(ch.ID())
		log.Debug("Restoring channel...")

		// Putting the channel into the channel registry will call the
		// OnNewChannel callback so that the user can deal with the restored
		// channel.
		if !c.channels.Put(ch.ID(), ch) {
			log.Warn("Channel already present, closing restored channel.")
			// If the channel already existed, close this one.
			ch.Close()
		}
		log.Info("Channel restored.")
	}
}

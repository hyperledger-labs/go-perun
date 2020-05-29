// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package client

import (
	"perun.network/go-perun/peer"
)

func (c *Client) restorePeerChannels(p *peer.Peer) {
	log := c.logPeer(p)
	it, err := c.pr.RestorePeer(p.PerunAddress)
	if err != nil {
		log.Errorf("Failed to restore channels for peer: %v", err)
		p.Close()
	}

	for it.Next(c.Ctx()) {
		chdata := it.Channel()
		go func() {
			if c.channels.Has(chdata.ID()) {
				return
			}

			channel, err := c.channelFromSource(chdata, p)
			if err != nil {
				c.logChan(chdata.ID()).Errorf("Failed to restore channel: %v", err)
				return
			}
			c.channels.Put(chdata.ID(), channel)
		}()
	}

	if err := it.Close(); err != nil {
		log.Errorf("Error while restoring a channel: %v", err)
	}
}

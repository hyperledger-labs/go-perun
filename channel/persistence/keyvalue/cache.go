// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package keyvalue

import (
	"log"
	stdsync "sync"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/wire"
)

func newChannelCache() *channelCache {
	return &channelCache{
		peers:        make(map[channel.ID][]wire.Address),
		peerChannels: make(map[string]map[channel.ID]struct{}),
	}
}

// channelCache contains all channels
type channelCache struct {
	mutex        stdsync.RWMutex
	peers        map[channel.ID][]wire.Address      // Used when closing a channel.
	peerChannels map[string]map[channel.ID]struct{} // Address -> Set<chID>
}

func (c *channelCache) addPeerChannel(addr wire.Address, chID channel.ID) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	peers := c.peers[chID]
	c.peers[chID] = append(peers, addr)

	addrKey := string(addr.Bytes())
	if chans, ok := c.peerChannels[addrKey]; ok {
		chans[chID] = struct{}{}
	} else {
		c.peerChannels[addrKey] = map[channel.ID]struct{}{chID: {}}
	}
}

func (c *channelCache) deleteChannel(id channel.ID) []wire.Address {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	peers, ok := c.peers[id]
	if !ok {
		log.Panic("unknown channel ID")
	}
	delete(c.peers, id)

	// Delete the channel from all its peers.
	for _, addr := range peers {
		delete(c.peerChannels[string(addr.Bytes())], id)
	}

	return peers
}

func (c *channelCache) clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.peers = nil
	c.peerChannels = nil
}

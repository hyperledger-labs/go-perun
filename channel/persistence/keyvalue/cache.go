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

package keyvalue

import (
	"log"
	stdsync "sync"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/wire"
)

//nolint:deadcode
func newChannelCache() *channelCache {
	return &channelCache{
		peers:        make(map[channel.ID][]wire.Address),
		peerChannels: make(map[string]map[channel.ID]struct{}),
	}
}

// channelCache contains all channels.
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

	addrBytes, err := addr.MarshalBinary()
	if err != nil {
		panic("error marshaling address: " + err.Error())
	}
	addrKey := string(addrBytes)
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
		addrBytes, err := addr.MarshalBinary()
		if err != nil {
			panic("error marshaling address: " + err.Error())
		}
		delete(c.peerChannels[string(addrBytes)], id)
	}

	return peers
}

func (c *channelCache) clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.peers = nil
	c.peerChannels = nil
}

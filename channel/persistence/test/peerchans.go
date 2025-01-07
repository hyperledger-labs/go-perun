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

package test

import (
	"bytes"
	"fmt"

	"perun.network/go-perun/wallet"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/wire"
	"perun.network/go-perun/wire/perunio"
)

type peerChans map[string][]channel.ID

func (pc peerChans) ID(p map[wallet.BackendID]wire.Address) []channel.ID {
	ids, ok := pc[peerKey(p)]
	if !ok {
		return nil
	}
	return ids
}

func (pc peerChans) Peers() []map[wallet.BackendID]wire.Address {
	ps := make([]map[wallet.BackendID]wire.Address, 0, len(pc))
	for k := range pc {
		pk, _ := peerFromKey(k)
		ps = append(ps, pk)
	}
	return ps
}

// Add adds the given channel id to each peer's id list.
func (pc peerChans) Add(id channel.ID, ps ...map[wallet.BackendID]wire.Address) {
	for _, p := range ps {
		pc.add(id, p)
	}
}

// Don't use add, use Add.
func (pc peerChans) add(id channel.ID, p map[wallet.BackendID]wire.Address) {
	pk := peerKey(p)
	ids := pc[pk] // nil ok, since we append
	pc[pk] = append(ids, id)
}

func (pc peerChans) Delete(id channel.ID) {
	for pk, ids := range pc {
		for i, pid := range ids {
			if id == pid {
				// ch found, unsorted delete
				lim := len(ids) - 1
				if lim == 0 {
					// last channel, remove peer
					delete(pc, pk)
					break
				}

				ids[i] = ids[lim]
				pc[pk] = ids[:lim]
				break // next peer, no double channel ids
			}
		}
	}
}

func peerKey(a map[wallet.BackendID]wire.Address) string {
	key := new(bytes.Buffer)
	err := perunio.Encode(key, wire.AddressDecMap(a))
	if err != nil {
		panic("error encoding peer key: " + err.Error())
	}
	return key.String()
}

func peerFromKey(s string) (map[wallet.BackendID]wire.Address, error) {
	p := make(map[wallet.BackendID]wire.Address)
	decMap := wire.AddressDecMap(p)
	err := perunio.Decode(bytes.NewBuffer([]byte(s)), &decMap)
	if err != nil {
		return nil, fmt.Errorf("error decoding peer key: %w", err)
	}
	return decMap, nil
}

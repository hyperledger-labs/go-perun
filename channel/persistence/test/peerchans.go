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

	"perun.network/go-perun/channel"
	"perun.network/go-perun/pkg/io"
	"perun.network/go-perun/wire"
)

type peerChans map[string][]channel.ID

func (pc peerChans) ID(p wire.Address) []channel.ID {
	ids, ok := pc[peerKey(p)]
	if !ok {
		return nil
	}
	return ids
}

func (pc peerChans) Peers() []wire.Address {
	ps := make([]wire.Address, 0, len(pc))
	for k := range pc {
		ps = append(ps, peerFromKey(k))
	}
	return ps
}

// Add adds the given channel id to each peer's id list.
func (pc peerChans) Add(id channel.ID, ps ...wire.Address) {
	for _, p := range ps {
		pc.add(id, p)
	}
}

// Don't use add, use Add.
func (pc peerChans) add(id channel.ID, p wire.Address) {
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

func peerKey(a wire.Address) string {
	var key = new(bytes.Buffer)
	err := io.Encode(key, a)
	if err != nil {
		panic("error encoding peer key: " + err.Error())
	}
	return key.String()
}

func peerFromKey(s string) wire.Address {
	p := wire.NewAddress()
	err := io.Decode(bytes.NewBuffer([]byte(s)), p)
	if err != nil {
		panic("error decoding peer key: " + err.Error())
	}
	return p
}

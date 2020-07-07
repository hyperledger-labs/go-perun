// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package test

import (
	"strings"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/wire"
)

type peerChans map[string][]channel.ID

func (pc peerChans) Get(p wire.Address) []channel.ID {
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
	ids, _ := pc[pk] // nil ok, since we append
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

func peerKey(a wire.Address) string { return string(a.Bytes()) }

func peerFromKey(s string) wire.Address {
	p, err := wire.DecodeAddress(strings.NewReader(s))
	if err != nil {
		panic("error decoding peer key: " + err.Error())
	}
	return p
}

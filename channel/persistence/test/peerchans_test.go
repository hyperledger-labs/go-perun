// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/channel"
	ctest "perun.network/go-perun/channel/test"
	wtest "perun.network/go-perun/wallet/test"
	"perun.network/go-perun/wire"
)

func TestEndpointChans(t *testing.T) {
	assert := assert.New(t)
	rng := rand.New(rand.NewSource(20200525))
	id := []channel.ID{ctest.NewRandomChannelID(rng), ctest.NewRandomChannelID(rng)}
	ps := wtest.NewRandomAddresses(rng, 3)

	pc := make(peerChans)
	pc.Add(id[0], ps...)
	pc.Add(id[1], ps[0], ps[2]) // omit 2nd peer
	assert.ElementsMatch(id, pc.Get(ps[0]))
	assert.ElementsMatch(id, pc.Get(ps[2]))
	assert.ElementsMatch(id[:1], pc.Get(ps[1]))
	assert.ElementsMatch(ps, pc.Peers())

	pc.Delete(id[0]) // p[1] should be deleted as id[0] was their only channel
	assert.ElementsMatch(id[1:], pc.Get(ps[0]))
	assert.ElementsMatch(id[1:], pc.Get(ps[2]))
	assert.ElementsMatch([]wire.Address{ps[0], ps[2]}, pc.Peers())
	assert.Nil(pc.Get(ps[1]))

	pc.Delete(id[1]) // now all peers should have been deleted
	assert.Empty(pc.Peers())
}

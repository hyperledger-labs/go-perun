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
	"testing"

	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/channel"
	ctest "perun.network/go-perun/channel/test"
	"perun.network/go-perun/wire"
	wiretest "perun.network/go-perun/wire/test"
	pkgtest "polycry.pt/poly-go/test"
)

func TestEndpointChans(t *testing.T) {
	assert := assert.New(t)
	rng := pkgtest.Prng(t)
	id := []channel.ID{ctest.NewRandomChannelID(rng), ctest.NewRandomChannelID(rng)}
	ps := wiretest.NewRandomAddresses(rng, 3)

	pc := make(peerChans)
	pc.Add(id[0], ps...)
	pc.Add(id[1], ps[0], ps[2]) // omit 2nd peer
	assert.ElementsMatch(id, pc.ID(ps[0]))
	assert.ElementsMatch(id, pc.ID(ps[2]))
	assert.ElementsMatch(id[:1], pc.ID(ps[1]))
	assert.ElementsMatch(ps, pc.Peers())

	pc.Delete(id[0]) // p[1] should be deleted as id[0] was their only channel
	assert.ElementsMatch(id[1:], pc.ID(ps[0]))
	assert.ElementsMatch(id[1:], pc.ID(ps[2]))
	assert.ElementsMatch([]wire.Address{ps[0], ps[2]}, pc.Peers())
	assert.Nil(pc.ID(ps[1]))

	pc.Delete(id[1]) // now all peers should have been deleted
	assert.Empty(pc.Peers())
}

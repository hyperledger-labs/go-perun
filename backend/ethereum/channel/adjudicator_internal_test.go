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

package channel

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"perun.network/go-perun/backend/ethereum/bindings/adjudicator"
	"perun.network/go-perun/channel"
	channeltest "perun.network/go-perun/channel/test"
	pkgtest "perun.network/go-perun/pkg/test"
)

func Test_toEthSubStates(t *testing.T) {
	var (
		rng    = pkgtest.Prng(t)
		assert = assert.New(t)
	)

	tests := []struct {
		title string
		setup func() (state *channel.State, subStates channel.StateMap, expected []adjudicator.ChannelState)
	}{
		{
			title: "nil map gives nil slice",
			setup: func() (state *channel.State, subStates channel.StateMap, expected []adjudicator.ChannelState) {
				return channeltest.NewRandomState(rng), nil, nil
			},
		},
		{
			title: "fresh map gives nil slice",
			setup: func() (state *channel.State, subStates channel.StateMap, expected []adjudicator.ChannelState) {
				return channeltest.NewRandomState(rng), nil, nil
			},
		},
		{
			title: "1 layer of sub-channels",
			setup: func() (state *channel.State, subStates channel.StateMap, expected []adjudicator.ChannelState) {
				// ch[0]( ch[1], ch[2], ch[3] )
				ch := genStates(rng, 4)
				ch[0].AddSubAlloc(*ch[1].ToSubAlloc())
				ch[0].AddSubAlloc(*ch[2].ToSubAlloc())
				ch[0].AddSubAlloc(*ch[3].ToSubAlloc())
				return ch[0], toStateMap(ch[1:]...), toEthStates(ch[1:]...)
			},
		},
		{
			title: "2 layers of sub-channels",
			setup: func() (state *channel.State, subStates channel.StateMap, expected []adjudicator.ChannelState) {
				// ch[0]( ch[1]( ch[2], ch[3] ), ch[4], ch[5] (ch[6] ) )
				ch := genStates(rng, 7)
				ch[0].AddSubAlloc(*ch[1].ToSubAlloc())
				ch[0].AddSubAlloc(*ch[4].ToSubAlloc())
				ch[0].AddSubAlloc(*ch[5].ToSubAlloc())
				ch[1].AddSubAlloc(*ch[2].ToSubAlloc())
				ch[1].AddSubAlloc(*ch[3].ToSubAlloc())
				ch[5].AddSubAlloc(*ch[6].ToSubAlloc())
				return ch[0], toStateMap(ch[1:]...), toEthStates(ch[1:]...)
			},
		},
	}

	for _, tc := range tests {
		state, subStates, expected := tc.setup()
		got := toEthSubStates(state, subStates)
		assert.Equal(expected, got, tc.title)
	}
}

func genStates(rng *rand.Rand, n int) (states []*channel.State) {
	states = make([]*channel.State, n)
	for i := range states {
		states[i] = channeltest.NewRandomState(rng)
	}
	return
}

func toStateMap(states ...*channel.State) (_states channel.StateMap) {
	_states = channel.MakeStateMap()
	_states.Add(states...)
	return
}

func toEthStates(states ...*channel.State) (_states []adjudicator.ChannelState) {
	_states = make([]adjudicator.ChannelState, len(states))
	for i, s := range states {
		_states[i] = ToEthState(s)
	}
	return
}

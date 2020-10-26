// Copyright 2019 - See NOTICE file for copyright holders.
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
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	channeltest "perun.network/go-perun/channel/test"
	pkgtest "perun.network/go-perun/pkg/test"
	"perun.network/go-perun/wallet"
	wallettest "perun.network/go-perun/wallet/test"
	wiretest "perun.network/go-perun/wire/test"
)

func TestClient_validTwoPartyProposal(t *testing.T) {
	rng := pkgtest.Prng(t)

	// dummy client that only has an id
	c := &Client{
		address: wallettest.NewRandomAddress(rng),
	}
	validProp := NewRandomLedgerChannelProposal(rng, channeltest.WithNumParts(2))
	validProp.Base().PeerAddrs[0] = c.address // set us as the proposer
	peerAddr := validProp.Base().PeerAddrs[1] // peer at 1 as receiver
	require.False(t, peerAddr.Equals(c.address))
	require.Len(t, validProp.Base().PeerAddrs, 2)

	validProp3Peers := NewRandomLedgerChannelProposal(rng, channeltest.WithNumParts(3))
	invalidProp := &LedgerChannelProposal{}
	*invalidProp = *validProp                // shallow copy
	invalidProp.Base().ChallengeDuration = 0 // invalidate

	tests := []struct {
		prop     *LedgerChannelProposal
		ourIdx   int
		peerAddr wallet.Address
		valid    bool
	}{
		{
			validProp,
			0, peerAddr, true,
		},
		// test all three invalid combinations of peer address, index
		{
			validProp,
			1, peerAddr, false, // wrong ourIdx
		},
		{
			validProp,
			0, c.address, false, // wrong peerAddr (ours)
		},
		{
			validProp,
			1, c.address, false, // wrong index, wrong peer address
		},
		{
			validProp3Peers, // valid proposal but three peers
			0, peerAddr, false,
		},
		{
			invalidProp, // invalid proposal, correct other params
			0, peerAddr, false,
		},
	}

	for i, tt := range tests {
		valid := c.validTwoPartyProposal(tt.prop, tt.ourIdx, tt.peerAddr)
		if tt.valid && valid != nil {
			t.Errorf("[%d] Exptected proposal to be valid but got: %v", i, valid)
		} else if !tt.valid && valid == nil {
			t.Errorf("[%d] Exptected proposal to be invalid", i)
		}
	}
}

func NewRandomBaseChannelProposal(rng *rand.Rand, opts ...channeltest.RandomOpt) BaseChannelProposal {
	var opt channeltest.RandomOpt
	if len(opts) != 0 {
		opt = opts[0].Append(opts[1:]...)
	} else {
		opt = make(channeltest.RandomOpt)
	}
	alloc := channeltest.NewRandomAllocation(rng, channeltest.WithNumParts(opt.NumParts(rng)))
	peers := wiretest.NewRandomAddresses(rng, opt.NumParts(rng))

	app, data := channeltest.NewRandomAppAndData(rng)
	return makeBaseChannelProposal(
		rng.Uint64(),
		alloc,
		peers,
		WithNonceFrom(rng),
		WithApp(app, data))
}

func NewRandomLedgerChannelProposal(rng *rand.Rand, opts ...channeltest.RandomOpt) *LedgerChannelProposal {
	opt := make(channeltest.RandomOpt).Append(opts...)
	base := NewRandomBaseChannelProposal(rng, opt)
	return &LedgerChannelProposal{
		BaseChannelProposal: base,
		Participant:         wallettest.NewRandomAddress(rng),
	}
}

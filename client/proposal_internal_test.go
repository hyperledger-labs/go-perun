// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package client

import (
	"math/big"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	channeltest "perun.network/go-perun/channel/test"
	"perun.network/go-perun/peer"
	"perun.network/go-perun/wallet"
	wallettest "perun.network/go-perun/wallet/test"
)

func TestClient_validTwoPartyProposal(t *testing.T) {
	rng := rand.New(rand.NewSource(0xdeadbeef))

	// dummy client that only has an id
	c := &Client{
		id: wallettest.NewRandomAccount(rng),
	}
	validProp := *newRandomValidChannelProposalReq(rng, 2)
	validProp.PeerAddrs[0] = c.id.Address() // set us as the proposer
	peerAddr := validProp.PeerAddrs[1]      // peer at 1 as receiver
	require.False(t, peerAddr.Equals(c.id.Address()))
	require.Len(t, validProp.PeerAddrs, 2)

	validProp3Peers := *newRandomValidChannelProposalReq(rng, 3)
	invalidProp := validProp          // shallow copy
	invalidProp.ChallengeDuration = 0 // invalidate

	tests := []struct {
		prop     *ChannelProposalReq
		ourIdx   int
		peerAddr wallet.Address
		valid    bool
	}{
		{
			&validProp,
			0, peerAddr, true,
		},
		// test all three invalid combinations of peer address, index
		{
			&validProp,
			1, peerAddr, false, // wrong ourIdx
		},
		{
			&validProp,
			0, c.id.Address(), false, // wrong peerAddr (ours)
		},
		{
			&validProp,
			1, c.id.Address(), false, // wrong index, wrong peer address
		},
		{
			&validProp3Peers, // valid proposal but three peers
			0, peerAddr, false,
		},
		{
			&invalidProp, // invalid proposal, correct other params
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

func newRandomValidChannelProposalReq(rng *rand.Rand, numPeers int) *ChannelProposalReq {
	peerAddrs := make([]peer.Address, numPeers)
	for i := 0; i < numPeers; i++ {
		peerAddrs[i] = wallettest.NewRandomAddress(rng)
	}
	data := channeltest.NewRandomData(rng)
	alloc := channeltest.NewRandomAllocation(rng, numPeers)
	alloc.Locked = nil // make valid InitBals
	participantAddr := wallettest.NewRandomAddress(rng)
	return &ChannelProposalReq{
		ChallengeDuration: rng.Uint64(),
		Nonce:             big.NewInt(rng.Int63()),
		ParticipantAddr:   participantAddr,
		AppDef:            channeltest.NewRandomApp(rng).Def(),
		InitData:          data,
		InitBals:          alloc,
		PeerAddrs:         peerAddrs,
	}
}

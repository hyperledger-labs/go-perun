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

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/channel"
	channeltest "perun.network/go-perun/channel/test"
	wallettest "perun.network/go-perun/wallet/test"
	"perun.network/go-perun/wire"
	wiretest "perun.network/go-perun/wire/test"
	pkgtest "polycry.pt/poly-go/test"
)

func TestClient_validTwoPartyProposal(t *testing.T) {
	rng := pkgtest.Prng(t)

	// dummy client that only has an id
	c := &Client{
		address: wiretest.NewRandomAddress(rng),
	}
	validProp := NewRandomLedgerChannelProposal(rng, channeltest.WithNumParts(2))
	validProp.Peers[0] = c.address // set us as the proposer
	peerAddr := validProp.Peers[1] // peer at 1 as receiver
	require.False(t, peerAddr.Equal(c.address))
	require.Len(t, validProp.Peers, 2)

	validProp3Peers := NewRandomLedgerChannelProposal(rng, channeltest.WithNumParts(3))
	invalidProp := &LedgerChannelProposalMsg{}
	*invalidProp = *validProp                // shallow copy
	invalidProp.Base().ChallengeDuration = 0 // invalidate

	tests := []struct {
		prop     *LedgerChannelProposalMsg
		ourIdx   channel.Index
		peerAddr wire.Address
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

func TestChannelProposal_assertValidNumParts(t *testing.T) {
	require := require.New(t)

	rng := pkgtest.Prng(t)
	c := NewRandomLedgerChannelProposal(rng)
	require.NoError(c.assertValidNumParts())
	c.Peers = make([]wire.Address, channel.MaxNumParts+1)
	require.Error(c.assertValidNumParts())
}

func TestProposalResponder_Accept_Nil(t *testing.T) {
	p := new(ProposalResponder)
	_, err := p.Accept(nil, new(LedgerChannelProposalAccMsg)) //nolint:staticcheck
	assert.Error(t, err, "context")
}

func TestPeerRejectedProposalError(t *testing.T) {
	reason := "some-random-reason"
	err := newPeerRejectedError("update", reason)
	t.Run("direct_error", func(t *testing.T) {
		peerRejectedProposalError := PeerRejectedError{}
		gotPeerRejectedError := errors.As(err, &peerRejectedProposalError)
		require.True(t, gotPeerRejectedError)
		assert.Equal(t, reason, peerRejectedProposalError.Reason)
		assert.Contains(t, err.Error(), reason)
	})

	t.Run("wrapped_error", func(t *testing.T) {
		wrappedError := errors.WithMessage(err, "some higher level error")
		peerRejectedError := PeerRejectedError{}
		gotPeerRejectedError := errors.As(wrappedError, &peerRejectedError)
		require.True(t, gotPeerRejectedError)
		assert.Equal(t, reason, peerRejectedError.Reason)
		assert.Contains(t, err.Error(), reason)
	})
}

func NewRandomBaseChannelProposal(rng *rand.Rand, opts ...channeltest.RandomOpt) BaseChannelProposal {
	var opt channeltest.RandomOpt
	if len(opts) != 0 {
		opt = opts[0].Append(opts[1:]...)
	} else {
		opt = make(channeltest.RandomOpt)
	}
	alloc := channeltest.NewRandomAllocation(rng, channeltest.WithNumParts(opt.NumParts(rng)))
	app, data := channeltest.NewRandomAppAndData(rng)
	prop, err := makeBaseChannelProposal(
		rng.Uint64(),
		alloc,
		WithNonceFrom(rng),
		WithApp(app, data))
	if err != nil {
		panic("Error generating random channel proposal: " + err.Error())
	}
	return prop
}

func NewRandomLedgerChannelProposal(rng *rand.Rand, opts ...channeltest.RandomOpt) *LedgerChannelProposalMsg {
	opt := make(channeltest.RandomOpt).Append(opts...)
	base := NewRandomBaseChannelProposal(rng, opt)
	peers := wiretest.NewRandomAddresses(rng, base.NumPeers())
	return &LedgerChannelProposalMsg{
		BaseChannelProposal: base,
		Participant:         wallettest.NewRandomAddress(rng),
		Peers:               peers,
	}
}

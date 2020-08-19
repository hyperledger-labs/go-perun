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

package client_test

import (
	"bytes"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/test"
	"perun.network/go-perun/client"
	"perun.network/go-perun/pkg/io"
	pkgtest "perun.network/go-perun/pkg/test"
	wallettest "perun.network/go-perun/wallet/test"
	"perun.network/go-perun/wire"
)

func TestChannelProposalReq_NilArgs(t *testing.T) {
	rng := pkgtest.Prng(t)
	c := client.NewChannelProposal(
		1,
		wallettest.NewRandomAddress(rng),
		test.NewRandomAllocation(rng, test.WithNumParts(2)),
		wallettest.NewRandomAddresses(rng, 2),
		client.WithNonceFrom(rng),
		client.WithApp(test.NewRandomApp(rng).Def(), test.NewRandomData(rng)))

	err := c.Encode(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "writer")

	err = c.Decode(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reader")
}

func TestChannelProposalReqSerialization(t *testing.T) {
	rng := pkgtest.Prng(t)
	for i := 0; i < 8; i++ {
		var app client.ProposalOpts
		if i&1 == 0 {
			app = client.WithApp(test.NewRandomApp(rng).Def(), test.NewRandomData(rng))
		}

		m := client.NewChannelProposal(
			1,
			wallettest.NewRandomAddress(rng),
			test.NewRandomAllocation(rng, test.WithNumParts(2)),
			wallettest.NewRandomAddresses(rng, 2),
			client.WithNonceFrom(rng),
			app)
		wire.TestMsg(t, m)
	}
}

func TestChannelProposalReqDecode_CheckMaxNumParts(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	rng := pkgtest.Prng(t)
	c := client.NewRandomChannelProposalReq(rng)
	buffer := new(bytes.Buffer)

	// reimplementation of ChannelProposalReq.Encode modified to create the
	// maximum number of participants possible with the encoding
	require.NoError(io.Encode(buffer, c.ChallengeDuration, c.NonceShare))
	require.NoError(
		io.Encode(buffer, c.ParticipantAddr, client.OptAppDefAndDataEnc{c.AppDef, c.InitData}, c.InitBals))

	numParts := int32(channel.MaxNumParts + 1)
	require.NoError(io.Encode(buffer, numParts))

	for i := 0; i < int(numParts); i++ {
		require.NoError(wallettest.NewRandomAddress(rng).Encode(buffer))
	}
	// end of ChannelProposalReq.Encode clone

	var d client.ChannelProposal
	err := d.Decode(buffer)
	require.Error(err)
	assert.Contains(err.Error(), "participants")
}

func TestChannelProposalReqProposalID(t *testing.T) {
	rng := pkgtest.Prng(t)
	original := *client.NewRandomChannelProposalReq(rng)
	s := original.ProposalID()
	fake := client.NewRandomChannelProposalReq(rng)

	assert.NotEqual(t, original.ChallengeDuration, fake.ChallengeDuration)
	assert.NotEqual(t, original.NonceShare, fake.NonceShare)
	assert.NotEqual(t, original.ParticipantAddr, fake.ParticipantAddr)
	// TODO: while using the payment app in channel tests, they all have the same
	// address. Fixed in #266
	// assert.NotEqual(t, original.AppDef, fake.AppDef)

	c0 := original
	c0.ChallengeDuration = fake.ChallengeDuration
	assert.NotEqual(t, s, c0.ProposalID())

	c1 := original
	c1.NonceShare = fake.NonceShare
	assert.NotEqual(t, s, c1.ProposalID())

	c2 := original
	c2.ParticipantAddr = fake.ParticipantAddr
	assert.Equal(t, s, c2.ProposalID())

	// TODO: #266
	//c3 := original
	//c3.AppDef = fake.AppDef
	//assert.NotEqual(t, s, c3.ProposalID())

	//c4 := original
	//c4.InitData = fake.InitData
	//assert.NotEqual(t, s, c4.ProposalID())

	c5 := original
	c5.InitBals = fake.InitBals
	assert.NotEqual(t, s, c5.ProposalID())

	c6 := original
	c6.PeerAddrs = fake.PeerAddrs
	assert.NotEqual(t, s, c6.ProposalID())
}

func TestChannelProposalAccSerialization(t *testing.T) {
	rng := pkgtest.Prng(t)
	for i := 0; i < 16; i++ {
		m := &client.ChannelProposalAcc{
			ProposalID:      newRandomProposalID(rng),
			ParticipantAddr: wallettest.NewRandomAddress(rng),
		}
		wire.TestMsg(t, m)
	}
}

func TestChannelProposalRejSerialization(t *testing.T) {
	rng := pkgtest.Prng(t)
	for i := 0; i < 16; i++ {
		m := &client.ChannelProposalRej{
			ProposalID: newRandomProposalID(rng),
			Reason:     newRandomString(rng, 16, 16),
		}
		wire.TestMsg(t, m)
	}
}

func newRandomProposalID(rng *rand.Rand) (id client.ProposalID) {
	rng.Read(id[:])
	return
}

// newRandomstring returns a random string of length between minLen and
// minLen+maxLenDiff.
func newRandomString(rng *rand.Rand, minLen, maxLenDiff int) string {
	r := make([]byte, minLen+rng.Intn(maxLenDiff))
	rng.Read(r)
	return string(r)
}

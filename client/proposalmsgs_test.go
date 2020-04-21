// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package client_test

import (
	"bytes"
	"math/big"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/test"
	"perun.network/go-perun/client"
	"perun.network/go-perun/pkg/io"
	"perun.network/go-perun/wallet"
	wallettest "perun.network/go-perun/wallet/test"
	"perun.network/go-perun/wire"
	"perun.network/go-perun/wire/msg"
)

func TestChannelProposalReq_NilArgs(t *testing.T) {
	rng := rand.New(rand.NewSource(2020 - 01 - 0x8))
	c := &client.ChannelProposalReq{
		ChallengeDuration: 1,
		Nonce:             big.NewInt(2),
		ParticipantAddr:   wallettest.NewRandomAddress(rng),
		AppDef:            test.NewRandomApp(rng).Def(),
		InitData:          test.NewRandomData(rng),
		InitBals:          test.NewRandomAllocation(rng, test.WithNumParts(2)),
		PeerAddrs: []wallet.Address{
			wallettest.NewRandomAddress(rng),
			wallettest.NewRandomAddress(rng),
		},
	}

	err := c.Encode(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "writer")

	err = c.Decode(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reader")
}

func TestChannelProposalReqSerialization(t *testing.T) {
	rng := rand.New(rand.NewSource(0xdeadbeef))
	for i := 0; i < 4; i++ {
		m := &client.ChannelProposalReq{
			ChallengeDuration: 0,
			Nonce:             big.NewInt(rng.Int63()),
			ParticipantAddr:   wallettest.NewRandomAddress(rng),
			AppDef:            test.NewRandomApp(rng).Def(),
			InitData:          test.NewRandomData(rng),
			InitBals:          test.NewRandomAllocation(rng, test.WithNumParts(2)),
			PeerAddrs: []wallet.Address{
				wallettest.NewRandomAddress(rng),
				wallettest.NewRandomAddress(rng),
			},
		}
		msg.TestMsg(t, m)
	}
}

func TestChannelProposalReqDecode_CheckMaxNumParts(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	rng := rand.New(rand.NewSource(20191223))
	c := client.NewRandomChannelProposalReq(rng)
	buffer := new(bytes.Buffer)

	// reimplementation of ChannelProposalReq.Encode modified to create the
	// maximum number of participants possible with the encoding
	require.NoError(wire.Encode(buffer, c.ChallengeDuration, c.Nonce))
	require.NoError(
		io.Encode(buffer, c.ParticipantAddr, c.AppDef, c.InitData, c.InitBals))

	numParts := int32(channel.MaxNumParts + 1)
	require.NoError(wire.Encode(buffer, numParts))

	for i := 0; i < int(numParts); i++ {
		require.NoError(wallettest.NewRandomAddress(rng).Encode(buffer))
	}
	// end of ChannelProposalReq.Encode clone

	var d client.ChannelProposalReq
	err := d.Decode(buffer)
	require.Error(err)
	assert.Contains(err.Error(), "participants")
}

func TestChannelProposalReqSessID(t *testing.T) {
	original := *client.NewRandomChannelProposalReq(rand.New(rand.NewSource(0xc0ffee)))
	s := original.SessID()
	fake := client.NewRandomChannelProposalReq(rand.New(rand.NewSource(0xeeff0c)))

	assert.NotEqual(t, original.ChallengeDuration, fake.ChallengeDuration)
	assert.NotEqual(t, original.Nonce, fake.Nonce)
	assert.NotEqual(t, original.ParticipantAddr, fake.ParticipantAddr)
	// TODO: while using the payment app in channel tests, they all have the same
	// address. Fixed in #266
	// assert.NotEqual(t, original.AppDef, fake.AppDef)

	c0 := original
	c0.ChallengeDuration = fake.ChallengeDuration
	assert.NotEqual(t, s, c0.SessID())

	c1 := original
	c1.Nonce = fake.Nonce
	assert.NotEqual(t, s, c1.SessID())

	c2 := original
	c2.ParticipantAddr = fake.ParticipantAddr
	assert.Equal(t, s, c2.SessID())

	// TODO: #266
	//c3 := original
	//c3.AppDef = fake.AppDef
	//assert.NotEqual(t, s, c3.SessID())

	//c4 := original
	//c4.InitData = fake.InitData
	//assert.NotEqual(t, s, c4.SessID())

	c5 := original
	c5.InitBals = fake.InitBals
	assert.NotEqual(t, s, c5.SessID())

	c6 := original
	c6.PeerAddrs = fake.PeerAddrs
	assert.NotEqual(t, s, c6.SessID())
}

func TestChannelProposal_AsReqAsProp(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	acc := wallettest.NewRandomAccount(rng)
	prop := client.NewRandomChannelProposalReq(rng).AsProp(acc)
	req := prop.AsReq()
	assert.True(t, req.ParticipantAddr.Equals(acc.Address()))
	prop2 := req.AsProp(acc)
	assert.Equal(t, prop2, prop)
}

func TestChannelProposalAccSerialization(t *testing.T) {
	rng := rand.New(rand.NewSource(0xcafecafe))
	for i := 0; i < 16; i++ {
		m := &client.ChannelProposalAcc{
			SessID:          newRandomSessID(rng),
			ParticipantAddr: wallettest.NewRandomAddress(rng),
		}
		msg.TestMsg(t, m)
	}
}

func TestChannelProposalRejSerialization(t *testing.T) {
	rng := rand.New(rand.NewSource(0xcafecafe))
	for i := 0; i < 16; i++ {
		m := &client.ChannelProposalRej{
			SessID: newRandomSessID(rng),
			Reason: newRandomString(rng, 16, 16),
		}
		msg.TestMsg(t, m)
	}
}

func newRandomSessID(rng *rand.Rand) (id client.SessionID) {
	rng.Read(id[:])
	return
}

// newRandomstring returns a random string of length between minLen and
// minLen+maxLenDiff
func newRandomString(rng *rand.Rand, minLen, maxLenDiff int) string {
	r := make([]byte, minLen+rng.Intn(maxLenDiff))
	rng.Read(r)
	return string(r)
}

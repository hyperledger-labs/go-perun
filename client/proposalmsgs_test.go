// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package client_test

import (
	"math/big"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	// import for channel app back-end initialization
	_ "perun.network/go-perun/backend/sim/channel"
	"perun.network/go-perun/channel/test"
	"perun.network/go-perun/client"
	"perun.network/go-perun/wallet"
	wallettest "perun.network/go-perun/wallet/test"
	"perun.network/go-perun/wire/msg"
)

func init() {
	test.SetBackend(new(test.TestBackend))
	wallet.SetBackend(new(wallettest.DefaultWalletBackend))
	wallettest.SetBackend(new(wallettest.DefaultBackend))
}

func TestChannelProposalSerialization(t *testing.T) {
	rng := rand.New(rand.NewSource(0xdeadbeef))
	for i := 0; i < 4; i++ {
		m := &client.ChannelProposal{
			ChallengeDuration: 0,
			Nonce:             big.NewInt(rng.Int63()),
			ParticipantAddr:   wallettest.NewRandomAddress(rng),
			AppDef:            wallettest.NewRandomAddress(rng),
			InitData:          test.NewRandomData(rng),
			InitBals:          test.NewRandomAllocation(rng, 2),
			Parts: []wallet.Address{
				wallettest.NewRandomAddress(rng),
				wallettest.NewRandomAddress(rng),
			},
		}
		msg.TestMsg(t, m)
	}
}

func TestChannelProposalSessID(t *testing.T) {
	original := *newRandomChannelProposal(rand.New(rand.NewSource(0xc0ffee)))
	s := original.SessID()
	fake := newRandomChannelProposal(rand.New(rand.NewSource(0xeeff0c)))

	assert.NotEqual(t, original.ChallengeDuration, fake.ChallengeDuration)
	assert.NotEqual(t, original.Nonce, fake.Nonce)
	assert.NotEqual(t, original.ParticipantAddr, fake.ParticipantAddr)
	assert.NotEqual(t, original.AppDef, fake.AppDef)

	c0 := original
	c0.ChallengeDuration = fake.ChallengeDuration
	assert.NotEqual(t, s, c0.SessID())

	c1 := original
	c1.Nonce = fake.Nonce
	assert.NotEqual(t, s, c1.SessID())

	c2 := original
	c2.ParticipantAddr = fake.ParticipantAddr
	assert.Equal(t, s, c2.SessID())

	c3 := original
	c3.AppDef = fake.AppDef
	assert.NotEqual(t, s, c3.SessID())

	c4 := original
	c4.InitData = fake.InitData
	assert.NotEqual(t, s, c4.SessID())

	c5 := original
	c5.InitBals = fake.InitBals
	assert.NotEqual(t, s, c5.SessID())

	c6 := original
	c6.Parts = fake.Parts
	assert.NotEqual(t, s, c6.SessID())
}

func newRandomChannelProposal(rng *rand.Rand) *client.ChannelProposal {
	app := test.NewRandomApp(rng)
	params := test.NewRandomParams(rng, app)
	data := test.NewRandomData(rng)
	numParts := 2 + rng.Intn(8)
	alloc := test.NewRandomAllocation(rng, numParts)
	participantAddr := wallettest.NewRandomAddress(rng)
	return &client.ChannelProposal{
		params.ChallengeDuration,
		params.Nonce,
		participantAddr,
		params.App.Def(),
		data,
		alloc,
		params.Parts,
	}
}

func TestChannelProposalAccSerialization(t *testing.T) {
	rng := rand.New(rand.NewSource(0xcafecafe))
	for i := 0; i < 16; i++ {
		m := &client.ChannelProposalAcc{
			SessID:          NewRandomSessID(rng),
			ParticipantAddr: wallettest.NewRandomAddress(rng),
		}
		msg.TestMsg(t, m)
	}
}

func TestChannelProposalRejSerialization(t *testing.T) {
	rng := rand.New(rand.NewSource(0xcafecafe))
	for i := 0; i < 16; i++ {
		r := make([]byte, 16+rng.Intn(16)) // random string of length 16..32
		rng.Read(r)
		m := &client.ChannelProposalRej{
			SessID: NewRandomSessID(rng),
			Reason: string(r),
		}
		msg.TestMsg(t, m)
	}
}

func NewRandomSessID(rng *rand.Rand) (id client.SessionID) {
	rng.Read(id[:])
	return
}

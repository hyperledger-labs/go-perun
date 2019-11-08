// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package client_test

import (
	"math/big"
	"math/rand"
	"testing"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/test"
	"perun.network/go-perun/client"
	"perun.network/go-perun/wallet"
	wallettest "perun.network/go-perun/wallet/test"
	"perun.network/go-perun/wire/msg"
)

func init() {
	channel.SetAppBackend(new(test.NoAppBackend))
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

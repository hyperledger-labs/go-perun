// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package test

import (
	"math/big"
	"math/rand"
	"testing"

	"perun.network/go-perun/channel"
	simulatedWallet "perun.network/go-perun/backend/sim/wallet"
	"perun.network/go-perun/wallet"
	wire "perun.network/go-perun/wire/msg"
)

func newAddress(seed int64) wallet.Address {
	return simulatedWallet.NewRandomAddress(rand.New(rand.NewSource(seed)))
}

func TestChannelProposalSerialization(t *testing.T) {
	inputs := []channel.ChannelProposal{
		channel.ChannelProposal{
			ChallengeDuration: 0,
			Nonce:             big.NewInt(1),
			ParticipantAddr:   newAddress(2),
			AppDef:            newAddress(3),
			InitData:          &channel.DummyData{X: 6},
			InitBals: &channel.Allocation{
				Assets: []channel.Asset{&Asset{ID: 7}},
				OfParts: [][]channel.Bal{
					[]channel.Bal{big.NewInt(8)},
					[]channel.Bal{big.NewInt(9)}},
				Locked: []channel.SubAlloc{},
			},
			Parts: []wallet.Address{newAddress(4), newAddress(5)},
		},
		channel.ChannelProposal{
			ChallengeDuration: 99,
			Nonce:             big.NewInt(100),
			ParticipantAddr:   newAddress(101),
			AppDef:            newAddress(102),
			InitData:          &channel.DummyData{X: 103},
			InitBals: &channel.Allocation{
				Assets: []channel.Asset{&Asset{ID: 8}, &Asset{ID: 255}},
				OfParts: [][]channel.Bal{
					[]channel.Bal{big.NewInt(9), big.NewInt(131)},
					[]channel.Bal{big.NewInt(1), big.NewInt(1024)}},
				Locked: []channel.SubAlloc{
					channel.SubAlloc{
						ID:   channel.ID{0xCA, 0xFE},
						Bals: []channel.Bal{big.NewInt(11), big.NewInt(12)}}},
			},
			Parts: []wallet.Address{newAddress(101), newAddress(110)},
		},
	}

	for i := range inputs {
		wire.TestMsg(t, &inputs[i])
	}
}

func TestChannelProposalResSerialization(t *testing.T) {
	inputs := []channel.ChannelProposalRes{
		channel.ChannelProposalRes{
			SessID:          channel.SessionID{0, 1, 2},
			ParticipantAddr: newAddress(4),
		},
		channel.ChannelProposalRes{
			SessID:          channel.SessionID{0x0E, 0xA7, 0xBE, 0xEF},
			ParticipantAddr: newAddress(123),
		},
	}

	for i := range inputs {
		wire.TestMsg(t, &inputs[i])
	}
}

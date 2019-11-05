// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel_test

import (
	"math/big"
	"math/rand"
	"testing"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/test"
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

func newAddress(seed int64) wallet.Address {
	return wallettest.NewRandomAddress(rand.New(rand.NewSource(seed)))
}

func TestChannelProposalSerialization(t *testing.T) {
	inputs := []*channel.ChannelProposal{
		&channel.ChannelProposal{
			ChallengeDuration: 0,
			Nonce:             big.NewInt(1),
			ParticipantAddr:   newAddress(2),
			AppDef:            newAddress(3),
			InitData:          &test.NoAppData{Value: 6},
			InitBals: &channel.Allocation{
				Assets: []channel.Asset{&test.Asset{ID: 7}},
				OfParts: [][]channel.Bal{
					[]channel.Bal{big.NewInt(8)},
					[]channel.Bal{big.NewInt(9)}},
				Locked: []channel.SubAlloc{},
			},
			Parts: []wallet.Address{newAddress(4), newAddress(5)},
		},
		&channel.ChannelProposal{
			ChallengeDuration: 99,
			Nonce:             big.NewInt(100),
			ParticipantAddr:   newAddress(101),
			AppDef:            newAddress(102),
			InitData:          &test.NoAppData{Value: 103},
			InitBals: &channel.Allocation{
				Assets: []channel.Asset{&test.Asset{ID: 8}, &test.Asset{ID: 255}},
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

	for _, m := range inputs {
		msg.TestMsg(t, m)
	}
}

func TestChannelProposalResSerialization(t *testing.T) {
	inputs := []*channel.ChannelProposalRes{
		&channel.ChannelProposalRes{
			SessID:          channel.SessionID{0, 1, 2},
			ParticipantAddr: newAddress(4),
		},
		&channel.ChannelProposalRes{
			SessID:          channel.SessionID{0x0E, 0xA7, 0xBE, 0xEF},
			ParticipantAddr: newAddress(123),
		},
	}

	for _, m := range inputs {
		msg.TestMsg(t, m)
	}
}

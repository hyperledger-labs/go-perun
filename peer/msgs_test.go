// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package peer

import (
	"math/big"
	"testing"

	simulatedWallet "perun.network/go-perun/backend/sim/wallet"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"
	wire "perun.network/go-perun/wire/msg"
)

type SequentialGenerator struct {
	state byte
}

func (s *SequentialGenerator) Read(buf []byte) (int, error) {
	for i := 0; i < len(buf); i++ {
		buf[i] = s.state
		s.state++
	}

	return len(buf), nil
}

func newAddress(init byte) wallet.Address {
	return simulatedWallet.NewRandomAddress(&SequentialGenerator{init})
}

func TestProposalSerialization(t *testing.T) {
	inputs := []Proposal{
		Proposal{
			ChallengeDuration: 0,
			Nonce:             big.NewInt(1),
			EphemeralAddr:     newAddress(2),
			AppDef:            newAddress(3),
			InitData:          &channel.DummyData{X: 6},
			InitBals: channel.Allocation{
				Assets: []channel.Asset{&channel.DummyAsset{Value: 7}},
				OfParts: [][]channel.Bal{
					[]channel.Bal{big.NewInt(8)},
					[]channel.Bal{big.NewInt(9)}},
				Locked: []channel.SubAlloc{},
			},
			PerunParts: []wallet.Address{newAddress(4), newAddress(5)},
		},
		Proposal{
			ChallengeDuration: 99,
			Nonce:             big.NewInt(100),
			EphemeralAddr:     newAddress(101),
			AppDef:            newAddress(102),
			InitData:          &channel.DummyData{X: 103},
			InitBals: channel.Allocation{
				Assets: []channel.Asset{
					&channel.DummyAsset{Value: 8}, &channel.DummyAsset{Value: 255}},
				OfParts: [][]channel.Bal{
					[]channel.Bal{big.NewInt(9), big.NewInt(131)},
					[]channel.Bal{big.NewInt(1), big.NewInt(1024)}},
				Locked: []channel.SubAlloc{
					channel.SubAlloc{
						ID:   channel.ID{0xCA, 0xFE},
						Bals: []channel.Bal{big.NewInt(11), big.NewInt(12)}}},
			},
			PerunParts: []wallet.Address{newAddress(101), newAddress(110)},
		},
	}

	for i := range inputs {
		wire.TestMsg(t, &inputs[i])
	}
}

func TestResponseSerialization(t *testing.T) {
	inputs := []Response{
		Response{
			SessID:        SessionID{0, 1, 2},
			EphemeralAddr: newAddress(4),
		},
		Response{
			SessID:        SessionID{0x0E, 0xA7, 0xBE, 0xEF},
			EphemeralAddr: newAddress(123),
		},
	}

	for i := range inputs {
		wire.TestMsg(t, &inputs[i])
	}
}

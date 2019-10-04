// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package peer

import (
	"math/big"
	"testing"

	simulatedWallet "perun.network/go-perun/backend/sim/wallet"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/pkg/io/test"
	"perun.network/go-perun/wallet"
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
			Commit:            []byte{1, 2, 3, 4},
			EphemeralAddr:     newAddress(5),
			AppDef:            newAddress(6),
			InitData:          &channel.DummyData{X: 7},
			InitBals: channel.Allocation{
				Assets:  []channel.Asset{&channel.DummyAsset{Value: 8}},
				OfParts: [][]channel.Bal{[]channel.Bal{big.NewInt(9)}},
				Locked:  []channel.SubAlloc{},
			},
		},
		Proposal{
			ChallengeDuration: 100,
			Commit:            []byte{1, 2, 3, 4, 255, 127, 0, 128},
			EphemeralAddr:     newAddress(101),
			AppDef:            newAddress(102),
			InitData:          &channel.DummyData{X: 128},
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
		},
	}

	for i := range inputs {
		test.GenericSerializableTest(t, &inputs[i])
	}
}

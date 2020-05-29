// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package client_test

import (
	"math/big"
	"math/rand"
	"testing"

	chtest "perun.network/go-perun/channel/test"
	ctest "perun.network/go-perun/client/test"
	"perun.network/go-perun/peer"
)

func TestHappyAliceBob(t *testing.T) {
	rng := rand.New(rand.NewSource(0x1337))
	setups, _ := NewSetups(rng, []string{"Alice", "Bob"})
	roles := [2]ctest.Executer{
		ctest.NewAlice(setups[0], t),
		ctest.NewBob(setups[1], t),
	}

	cfg := ctest.ExecConfig{
		PeerAddrs:  [2]peer.Address{setups[0].Identity.Address(), setups[1].Identity.Address()},
		Asset:      chtest.NewRandomAsset(rng),
		InitBals:   [2]*big.Int{big.NewInt(100), big.NewInt(100)},
		NumUpdates: [2]int{2, 2},
		TxAmounts:  [2]*big.Int{big.NewInt(5), big.NewInt(3)},
	}

	executeTwoPartyTest(t, roles, cfg)
}

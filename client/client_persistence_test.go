// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package client_test

import (
	"math/big"
	"math/rand"
	"testing"

	chprtest "perun.network/go-perun/channel/persistence/test"
	chtest "perun.network/go-perun/channel/test"
	ctest "perun.network/go-perun/client/test"
	"perun.network/go-perun/wire"
)

func TestPersistencePetraRobert(t *testing.T) {
	rng := rand.New(rand.NewSource(0x70707))
	setups := NewSetupsPersistence(t, rng, []string{"Petra", "Robert"})
	roles := [2]ctest.Executer{
		ctest.NewPetra(setups[0], t),
		ctest.NewRobert(setups[1], t),
	}

	cfg := ctest.ExecConfig{
		PeerAddrs:  [2]wire.Address{setups[0].Identity.Address(), setups[1].Identity.Address()},
		Asset:      chtest.NewRandomAsset(rng),
		InitBals:   [2]*big.Int{big.NewInt(100), big.NewInt(100)},
		NumUpdates: [2]int{2, 2},
		TxAmounts:  [2]*big.Int{big.NewInt(5), big.NewInt(3)},
	}

	executeTwoPartyTest(t, roles, cfg)
}

func NewSetupsPersistence(t *testing.T, rng *rand.Rand, names []string) []ctest.RoleSetup {
	setups := NewSetups(rng, names)
	for i := range names {
		setups[i].PR = chprtest.NewPersistRestorer(t)
	}
	return setups
}

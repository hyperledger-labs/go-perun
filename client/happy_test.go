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
	"math/big"
	"testing"

	chtest "perun.network/go-perun/channel/test"
	ctest "perun.network/go-perun/client/test"
	"perun.network/go-perun/pkg/test"
	"perun.network/go-perun/wire"
)

func TestHappyAliceBob(t *testing.T) {
	rng := test.Prng(t)
	setups := NewSetups(rng, []string{"Alice", "Bob"})
	roles := [2]ctest.Executer{
		ctest.NewAlice(setups[0], t),
		ctest.NewBob(setups[1], t),
	}

	cfg := ctest.ExecConfig{
		PeerAddrs:  [2]wire.Address{setups[0].Identity.Address(), setups[1].Identity.Address()},
		Asset:      chtest.NewRandomAsset(rng),
		InitBals:   [2]*big.Int{big.NewInt(100), big.NewInt(100)},
		NumUpdates: [2]int{2, 2},
		TxAmounts:  [2]*big.Int{big.NewInt(5), big.NewInt(3)},
	}

	executeTwoPartyTest(roles, cfg)
}

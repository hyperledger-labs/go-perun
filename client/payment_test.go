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
	"context"
	"math/big"
	"math/rand"
	"testing"

	chtest "perun.network/go-perun/channel/test"
	"perun.network/go-perun/client"
	ctest "perun.network/go-perun/client/test"
	"perun.network/go-perun/wire"
	pkgtest "polycry.pt/poly-go/test"
)

func TestPaymentHappy(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), twoPartyTestTimeout)
	defer cancel()

	runAliceBobTest(ctx, t, func(rng *rand.Rand) ([]ctest.RoleSetup, [2]ctest.Executer) {
		setups := NewSetups(rng, []string{"Alice", "Bob"})
		roles := [2]ctest.Executer{
			ctest.NewAlice(t, setups[0]),
			ctest.NewBob(t, setups[1]),
		}
		return setups, roles
	})
}

func TestPaymentDispute(t *testing.T) {
	rng := pkgtest.Prng(t)
	ctx, cancel := context.WithTimeout(context.Background(), twoPartyTestTimeout)
	defer cancel()

	const mallory, carol = 0, 1 // Indices of Mallory and Carol
	setups := NewSetups(rng, []string{"Mallory", "Carol"})
	roles := [2]ctest.Executer{
		ctest.NewMallory(t, setups[0]),
		ctest.NewCarol(t, setups[1]),
	}

	cfg := &ctest.MalloryCarolExecConfig{
		BaseExecConfig: ctest.MakeBaseExecConfig(
			[2]wire.Address{setups[mallory].Identity.Address(), setups[carol].Identity.Address()},
			chtest.NewRandomAsset(rng),
			[2]*big.Int{big.NewInt(100), big.NewInt(1)},
			client.WithoutApp(),
		),
		NumPayments: [2]int{5, 0},
		TxAmounts:   [2]*big.Int{big.NewInt(20), big.NewInt(0)},
	}
	ctest.ExecuteTwoPartyTest(ctx, t, roles, cfg)
}

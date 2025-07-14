// Copyright 2025 - See NOTICE file for copyright holders.
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
	"time"

	"perun.network/go-perun/channel"

	"perun.network/go-perun/wallet"

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
		setups := NewSetups(rng, []string{"Alice", "Bob"}, channel.TestBackendID)
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
	setups := NewSetups(rng, []string{"Mallory", "Carol"}, channel.TestBackendID)
	roles := [2]ctest.Executer{
		ctest.NewMallory(t, setups[0]),
		ctest.NewCarol(t, setups[1]),
	}

	cfg := &ctest.MalloryCarolExecConfig{
		BaseExecConfig: ctest.MakeBaseExecConfig(
			[2]map[wallet.BackendID]wire.Address{wire.AddressMapfromAccountMap(setups[mallory].Identity), wire.AddressMapfromAccountMap(setups[carol].Identity)},
			chtest.NewRandomAsset(rng, channel.TestBackendID),
			channel.TestBackendID,
			[2]*big.Int{big.NewInt(100), big.NewInt(1)},
			client.WithoutApp(),
		),
		NumPayments: [2]int{5, 0},
		TxAmounts:   [2]*big.Int{big.NewInt(20), big.NewInt(0)},
	}
	ctest.ExecuteTwoPartyTest(ctx, t, roles, cfg)
}

func TestPaymentChannelsOptimistic(t *testing.T) {
	rng := pkgtest.Prng(t)

	ctx, cancel := context.WithTimeout(context.Background(), testDuration)
	defer cancel()

	setup := makePaymentChannelSetup(rng)
	ctest.TestPaymentChannelOptimistic(ctx, t, setup)
}

func TestPaymentChannelsDispute(t *testing.T) {
	rng := pkgtest.Prng(t)

	ctx, cancel := context.WithTimeout(context.Background(), testDuration)
	defer cancel()

	setup := makePaymentChannelSetup(rng)
	ctest.TestPaymentChannelDispute(ctx, t, setup)
}

func makePaymentChannelSetup(rng *rand.Rand) ctest.PaymentChannelSetup {
	return ctest.PaymentChannelSetup{
		Clients:           createPaymentChannelClients(rng),
		ChallengeDuration: challengeDuration,
		Asset:             chtest.NewRandomAsset(rng, channel.TestBackendID),
		Balances: ctest.PaymentChannelBalances{
			InitBalsAliceBob: []*big.Int{big.NewInt(5), big.NewInt(5)},
			BalsUpdated:      []*big.Int{big.NewInt(2), big.NewInt(8)},
			FinalBals:        []*big.Int{big.NewInt(2), big.NewInt(8)},
		},
		BalanceDelta:       big.NewInt(0),
		Rng:                rng,
		WaitWatcherTimeout: 100 * time.Millisecond,
		IsUTXO:             true,
	}
}

func createPaymentChannelClients(rng *rand.Rand) [2]ctest.RoleSetup {
	var setupsArray [2]ctest.RoleSetup
	setups := NewSetups(rng, []string{"Alice", "Bob"}, channel.TestBackendID)
	copy(setupsArray[:], setups)
	return setupsArray
}

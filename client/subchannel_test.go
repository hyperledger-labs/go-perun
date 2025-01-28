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
	"testing"

	"perun.network/go-perun/channel"

	"perun.network/go-perun/wallet"

	"perun.network/go-perun/apps/payment"
	chtest "perun.network/go-perun/channel/test"
	"perun.network/go-perun/client"
	ctest "perun.network/go-perun/client/test"
	"perun.network/go-perun/wire"
	"polycry.pt/poly-go/test"
)

func TestSubChannelHappy(t *testing.T) {
	rng := test.Prng(t)

	setups := NewSetups(rng, []string{"Susie", "Tim"}, channel.TestBackendID)
	roles := [2]ctest.Executer{
		ctest.NewSusie(t, setups[0]),
		ctest.NewTim(t, setups[1]),
	}

	cfg := ctest.NewSusieTimExecConfig(
		ctest.MakeBaseExecConfig(
			[2]map[wallet.BackendID]wire.Address{wire.AddressMapfromAccountMap(setups[0].Identity), wire.AddressMapfromAccountMap(setups[1].Identity)},
			chtest.NewRandomAsset(rng, channel.TestBackendID),
			channel.TestBackendID,
			[2]*big.Int{big.NewInt(100), big.NewInt(100)},
			client.WithoutApp(),
		),
		2,
		3,
		[][2]*big.Int{
			{big.NewInt(10), big.NewInt(10)},
			{big.NewInt(5), big.NewInt(5)},
		},
		[][2]*big.Int{
			{big.NewInt(3), big.NewInt(3)},
			{big.NewInt(2), big.NewInt(2)},
			{big.NewInt(1), big.NewInt(1)},
		},
		client.WithApp(
			chtest.NewRandomAppAndData(rng, chtest.WithAppRandomizer(new(payment.Randomizer))),
		),
		big.NewInt(1),
	)

	ctx, cancel := context.WithTimeout(context.Background(), twoPartyTestTimeout)
	defer cancel()
	ctest.ExecuteTwoPartyTest(ctx, t, roles, cfg)
}

func TestSubChannelDispute(t *testing.T) {
	rng := test.Prng(t)

	setups := NewSetups(rng, []string{"DisputeSusie", "DisputeTim"}, channel.TestBackendID)
	roles := [2]ctest.Executer{
		ctest.NewDisputeSusie(t, setups[0]),
		ctest.NewDisputeTim(t, setups[1]),
	}

	baseCfg := ctest.MakeBaseExecConfig(
		[2]map[wallet.BackendID]wire.Address{wire.AddressMapfromAccountMap(setups[0].Identity), wire.AddressMapfromAccountMap(setups[1].Identity)},
		chtest.NewRandomAsset(rng, channel.TestBackendID),
		channel.TestBackendID,
		[2]*big.Int{big.NewInt(100), big.NewInt(100)},
		client.WithoutApp(),
	)
	cfg := &ctest.DisputeSusieTimExecConfig{
		BaseExecConfig:  baseCfg,
		SubChannelFunds: [2]*big.Int{big.NewInt(10), big.NewInt(10)},
		TxAmount:        big.NewInt(1),
	}

	ctx, cancel := context.WithTimeout(context.Background(), twoPartyTestTimeout)
	defer cancel()
	ctest.ExecuteTwoPartyTest(ctx, t, roles, cfg)
}

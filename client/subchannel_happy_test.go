// Copyright 2020 - See NOTICE file for copyright holders.
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

	"github.com/stretchr/testify/assert"
	"perun.network/go-perun/apps/payment"
	chtest "perun.network/go-perun/channel/test"
	"perun.network/go-perun/client"
	ctest "perun.network/go-perun/client/test"
	"perun.network/go-perun/pkg/test"
	"perun.network/go-perun/wire"
)

func TestSubChannelHappy(t *testing.T) {
	rng := test.Prng(t)

	setups := NewSetups(rng, []string{"Susie", "Tim"})
	roles := [2]ctest.Executer{
		ctest.NewSusie(setups[0], t),
		ctest.NewTim(setups[1], t),
	}

	cfg := ctest.NewSusieTimExecConfig(
		ctest.MakeBaseExecConfig(
			[2]wire.Address{setups[0].Identity.Address(), setups[1].Identity.Address()},
			chtest.NewRandomAsset(rng),
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
	assert.NoError(t, ctest.ExecuteTwoPartyTest(ctx, roles, cfg))
}

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
	"time"

	"github.com/stretchr/testify/assert"
	"perun.network/go-perun/apps/payment"
	chtest "perun.network/go-perun/channel/test"
	"perun.network/go-perun/client"
	ctest "perun.network/go-perun/client/test"
	"perun.network/go-perun/pkg/test"
	"perun.network/go-perun/wire"
)

const twoPartyTestTimeout = 10 * time.Second

func TestHappyAliceBob(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), twoPartyTestTimeout)
	defer cancel()

	runTwoPartyTest(ctx, t, func(rng *rand.Rand) (setups []ctest.RoleSetup, roles [2]ctest.Executer) {
		setups = NewSetups(rng, []string{"Alice", "Bob"})
		roles = [2]ctest.Executer{
			ctest.NewAlice(setups[0], t),
			ctest.NewBob(setups[1], t),
		}
		return
	})
}

func runTwoPartyTest(ctx context.Context, t *testing.T, setup func(*rand.Rand) ([]ctest.RoleSetup, [2]ctest.Executer)) {
	rng := test.Prng(t)
	for i := 0; i < 2; i++ {
		setups, roles := setup(rng)
		app := client.WithoutApp()
		if i == 1 {
			app = client.WithApp(
				chtest.NewRandomAppAndData(rng, chtest.WithAppRandomizer(new(payment.Randomizer))),
			)
		}

		cfg := &ctest.AliceBobExecConfig{
			BaseExecConfig: ctest.MakeBaseExecConfig(
				[2]wire.Address{setups[0].Identity.Address(), setups[1].Identity.Address()},
				chtest.NewRandomAsset(rng),
				[2]*big.Int{big.NewInt(100), big.NewInt(100)},
				app,
			),
			NumPayments: [2]int{2, 2},
			TxAmounts:   [2]*big.Int{big.NewInt(5), big.NewInt(3)},
		}

		err := ctest.ExecuteTwoPartyTest(ctx, roles, cfg)
		assert.NoError(t, err)
	}
}

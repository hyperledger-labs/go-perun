// Copyright 2024 - See NOTICE file for copyright holders.
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

	"perun.network/go-perun/channel"
	chtest "perun.network/go-perun/channel/test"
	ctest "perun.network/go-perun/client/test"
	"polycry.pt/poly-go/test"
)

func TestFailingFunding(t *testing.T) {
	rng := test.Prng(t)

	ctx, cancel := context.WithTimeout(context.Background(), twoPartyTestTimeout)
	defer cancel()

	ctest.TestFundRecovery(
		ctx,
		t,
		ctest.FundSetup{
			ChallengeDuration: 1,
			FridaInitBal:      big.NewInt(100),
			FredInitBal:       big.NewInt(50),
			BalanceDelta:      big.NewInt(0),
		},
		func(r *rand.Rand) ([2]ctest.RoleSetup, channel.Asset) {
			roles := NewSetups(rng, []string{"Frida", "Fred"}, channel.TestBackendID)
			asset := chtest.NewRandomAsset(rng, channel.TestBackendID)
			return [2]ctest.RoleSetup{roles[0], roles[1]}, asset
		},
	)
}

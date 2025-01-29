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

	chtest "perun.network/go-perun/channel/test"
	ctest "perun.network/go-perun/client/test"
	"polycry.pt/poly-go/test"
)

const (
	challengeDuration = 10
	testDuration      = 10 * time.Second
)

func TestVirtualChannelsOptimistic(t *testing.T) {
	rng := test.Prng(t)
	ctx, cancel := context.WithTimeout(context.Background(), testDuration)
	defer cancel()

	setup := makeVirtualChannelSetup(rng)
	ctest.TestVirtualChannelOptimistic(ctx, t, setup)
}

func TestVirtualChannelsDispute(t *testing.T) {
	rng := test.Prng(t)
	ctx, cancel := context.WithTimeout(context.Background(), testDuration)
	defer cancel()

	setup := makeVirtualChannelSetup(rng)
	ctest.TestVirtualChannelDispute(ctx, t, setup)
}

func makeVirtualChannelSetup(rng *rand.Rand) ctest.VirtualChannelSetup {
	return ctest.VirtualChannelSetup{
		Clients:           createVirtualChannelClients(rng),
		ChallengeDuration: challengeDuration,
		Asset:             chtest.NewRandomAsset(rng, channel.TestBackendID),
		Balances: ctest.VirtualChannelBalances{
			InitBalsAliceIngrid: []*big.Int{big.NewInt(10), big.NewInt(10)},
			InitBalsBobIngrid:   []*big.Int{big.NewInt(10), big.NewInt(10)},
			InitBalsAliceBob:    []*big.Int{big.NewInt(5), big.NewInt(5)},
			VirtualBalsUpdated:  []*big.Int{big.NewInt(2), big.NewInt(8)},
			FinalBalsAlice:      []*big.Int{big.NewInt(7), big.NewInt(13)},
			FinalBalsBob:        []*big.Int{big.NewInt(13), big.NewInt(7)},
		},
		BalanceDelta:       big.NewInt(0),
		Rng:                rng,
		WaitWatcherTimeout: 100 * time.Millisecond,
	}
}

func createVirtualChannelClients(rng *rand.Rand) [3]ctest.RoleSetup {
	var setupsArray [3]ctest.RoleSetup
	setups := NewSetups(rng, []string{"Alice", "Bob", "Ingrid"}, channel.TestBackendID)
	copy(setupsArray[:], setups)
	return setupsArray
}

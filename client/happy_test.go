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
	"math/rand"
	"testing"

	ctest "perun.network/go-perun/client/test"
)

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

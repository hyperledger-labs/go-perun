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
	"math/rand"
	"testing"

	"perun.network/go-perun/channel"

	chprtest "perun.network/go-perun/channel/persistence/test"
	ctest "perun.network/go-perun/client/test"
)

func TestPersistencePetraRobert(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), twoPartyTestTimeout)
	defer cancel()

	runAliceBobTest(ctx, t, func(rng *rand.Rand) (setups []ctest.RoleSetup, roles [2]ctest.Executer) {
		setups = NewSetupsPersistence(t, rng, []string{"Petra", "Robert"})
		roles = [2]ctest.Executer{
			ctest.NewPetra(t, setups[0]),
			ctest.NewRobert(t, setups[1]),
		}
		return
	})
}

func NewSetupsPersistence(t *testing.T, rng *rand.Rand, names []string) []ctest.RoleSetup {
	t.Helper()
	setups := NewSetups(rng, names, channel.TestBackendID)
	for i := range names {
		setups[i].PR = chprtest.NewPersistRestorer(t)
	}
	return setups
}

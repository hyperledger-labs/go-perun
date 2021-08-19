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
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"perun.network/go-perun/client"
	ctest "perun.network/go-perun/client/test"
	wtest "perun.network/go-perun/wallet/test"
	wiretest "perun.network/go-perun/wire/test"
)

const roleOperationTimeout = 1 * time.Second

func NewSetups(rng *rand.Rand, names []string) []ctest.RoleSetup {
	var (
		bus     = wiretest.NewSerializingLocalBus()
		n       = len(names)
		setup   = make([]ctest.RoleSetup, n)
		backend = ctest.NewMockBackend(rng)
	)

	for i := 0; i < n; i++ {
		setup[i] = ctest.RoleSetup{
			Name:              names[i],
			Identity:          wtest.NewRandomAccount(rng),
			Bus:               bus,
			Funder:            backend,
			Adjudicator:       backend,
			Wallet:            wtest.NewWallet(),
			Timeout:           roleOperationTimeout,
			Backend:           backend,
			ChallengeDuration: 60,
		}
	}

	return setup
}

type Client struct {
	*client.Client
	ctest.RoleSetup
}

func NewClients(rng *rand.Rand, names []string, t *testing.T) []*Client {
	setups := NewSetups(rng, names)
	clients := make([]*Client, len(setups))
	for i, setup := range setups {
		setup.Identity = setup.Wallet.NewRandomAccount(rng)
		cl, err := client.New(setup.Identity.Address(), setup.Bus, setup.Funder, setup.Adjudicator, setup.Wallet)
		assert.NoError(t, err)
		clients[i] = &Client{
			Client:    cl,
			RoleSetup: setup,
		}
	}
	return clients
}

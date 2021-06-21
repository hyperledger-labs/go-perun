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
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/client"
	ctest "perun.network/go-perun/client/test"
	"perun.network/go-perun/log"
	wtest "perun.network/go-perun/wallet/test"
	wiretest "perun.network/go-perun/wire/test"
)

const roleOperationTimeout = 1 * time.Second

func NewSetups(rng *rand.Rand, names []string) []ctest.RoleSetup {
	var (
		bus   = wiretest.NewSerializingLocalBus()
		n     = len(names)
		setup = make([]ctest.RoleSetup, n)
	)

	for i := 0; i < n; i++ {
		acc := wtest.NewRandomAccount(rng)
		backend := &logBackend{
			log: log.WithField("role", names[i]),
			rng: rng,
		}

		setup[i] = ctest.RoleSetup{
			Name:        names[i],
			Identity:    acc,
			Bus:         bus,
			Funder:      backend,
			Adjudicator: backend,
			Wallet:      wtest.NewWallet(),
			Timeout:     roleOperationTimeout,
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

type (
	logBackend struct {
		log         log.Logger
		rng         *rand.Rand
		mu          sync.RWMutex
		latestEvent channel.AdjudicatorEvent
	}
)

func (a *logBackend) Fund(_ context.Context, req channel.FundingReq) error {
	time.Sleep(time.Duration(a.rng.Intn(100)) * time.Millisecond)
	a.log.Infof("Funding: %v", req)
	return nil
}

func (a *logBackend) Register(_ context.Context, req channel.AdjudicatorReq, subChannels []channel.SignedState) error {
	a.log.Infof("Register: %v", req)
	e := channel.NewRegisteredEvent(
		req.Params.ID(),
		&channel.ElapsedTimeout{},
		req.Tx.Version,
	)
	a.setEvent(e)
	return nil
}

func (a *logBackend) Progress(_ context.Context, req channel.ProgressReq) error {
	a.log.Infof("Progress: %v", req)
	a.setEvent(channel.NewProgressedEvent(
		req.Params.ID(),
		&channel.ElapsedTimeout{},
		req.NewState.Clone(),
		req.Idx,
	))
	return nil
}

func (a *logBackend) Withdraw(_ context.Context, req channel.AdjudicatorReq, subStates channel.StateMap) error {
	a.log.Infof("Withdraw: %v, %v", req, subStates)
	return nil
}

func (a *logBackend) Subscribe(_ context.Context, params *channel.Params) (channel.AdjudicatorSubscription, error) {
	a.log.Infof("SubscribeRegistered: %v", params)
	return &simSubscription{a}, nil
}

func (a *logBackend) setEvent(e channel.AdjudicatorEvent) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.latestEvent = e
}

type simSubscription struct {
	a *logBackend
}

func (s *simSubscription) Next() channel.AdjudicatorEvent {
	s.a.mu.RLock()
	defer s.a.mu.RUnlock()
	return s.a.latestEvent
}

func (s *simSubscription) Close() error {
	return nil
}

func (s *simSubscription) Err() error {
	return nil
}

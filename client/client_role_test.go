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
	"time"

	"perun.network/go-perun/channel"
	ctest "perun.network/go-perun/client/test"
	"perun.network/go-perun/log"
	wtest "perun.network/go-perun/wallet/test"
	"perun.network/go-perun/wire"
)

const roleOperationTimeout = 1 * time.Second

func NewSetups(rng *rand.Rand, names []string) []ctest.RoleSetup {
	var (
		bus   = wire.NewLocalBus()
		n     = len(names)
		setup = make([]ctest.RoleSetup, n)
	)

	for i := 0; i < n; i++ {
		acc := wtest.NewRandomAccount(rng)

		// The use of a delayed funder simulates that channel participants may
		// receive their funding confirmation at different times.
		var funder channel.Funder
		if i == 0 {
			funder = &logFunderWithDelay{log.WithField("role", names[i])}
		} else {
			funder = &logFunder{log.WithField("role", names[i])}
		}

		setup[i] = ctest.RoleSetup{
			Name:        names[i],
			Identity:    acc,
			Bus:         bus,
			Funder:      funder,
			Adjudicator: &logAdjudicator{log.WithField("role", names[i]), sync.RWMutex{}, nil},
			Wallet:      wtest.NewWallet(),
			Timeout:     roleOperationTimeout,
		}
	}

	return setup
}

type (
	logFunder struct {
		log log.Logger
	}

	logFunderWithDelay struct {
		log log.Logger
	}

	logAdjudicator struct {
		log         log.Logger
		mu          sync.RWMutex
		latestEvent channel.AdjudicatorEvent
	}
)

func (f *logFunder) Fund(_ context.Context, req channel.FundingReq) error {
	f.log.Infof("Funding: %v", req)
	return nil
}

func (f *logFunderWithDelay) Fund(_ context.Context, req channel.FundingReq) error {
	time.Sleep(100 * time.Millisecond)
	f.log.Infof("Funding: %v", req)
	return nil
}

func (a *logAdjudicator) Register(_ context.Context, req channel.AdjudicatorReq, subChannels []channel.SignedState) error {
	a.log.Infof("Register: %v", req)
	e := channel.NewRegisteredEvent(
		req.Params.ID(),
		&channel.ElapsedTimeout{},
		req.Tx.Version,
	)
	a.setEvent(e)
	return nil
}

func (a *logAdjudicator) Progress(_ context.Context, req channel.ProgressReq) error {
	a.log.Infof("Progress: %v", req)
	a.setEvent(channel.NewProgressedEvent(
		req.Params.ID(),
		&channel.ElapsedTimeout{},
		req.NewState.Clone(),
		req.Idx,
	))
	return nil
}

func (a *logAdjudicator) Withdraw(_ context.Context, req channel.AdjudicatorReq, subStates channel.StateMap) error {
	a.log.Infof("Withdraw: %v, %v", req, subStates)
	return nil
}

func (a *logAdjudicator) Subscribe(_ context.Context, params *channel.Params) (channel.AdjudicatorSubscription, error) {
	a.log.Infof("SubscribeRegistered: %v", params)
	return &simSubscription{a}, nil
}

func (a *logAdjudicator) setEvent(e channel.AdjudicatorEvent) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.latestEvent = e
}

type simSubscription struct {
	a *logAdjudicator
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

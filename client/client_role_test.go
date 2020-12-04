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
	"testing"
	"time"

	"perun.network/go-perun/channel"
	ctest "perun.network/go-perun/client/test"
	"perun.network/go-perun/log"
	wtest "perun.network/go-perun/wallet/test"
	"perun.network/go-perun/wire"
)

const twoPartyTestTimeout = 10 * time.Second
const roleOperationTimeout = 1 * time.Second

func executeTwoPartyTest(t *testing.T, role [2]ctest.Executer, cfg ctest.ExecConfig) {
	log.Info("Starting two-party test")

	// enable stages synchronization
	stages := role[0].EnableStages()
	role[1].SetStages(stages)

	numClients := len(role)
	done := make(chan struct{}, numClients)

	// start clients
	for i := 0; i < numClients; i++ {
		go func(i int) {
			log.Infof("Executing role %d", i)
			role[i].Execute(cfg)
			done <- struct{}{} // signal client done
		}(i)
	}

	// wait for clients to finish or timeout
	timeout := time.After(twoPartyTestTimeout)
	for clientsRunning := numClients; clientsRunning > 0; clientsRunning-- {
		select {
		case <-done: // wait for client done signal
		case <-timeout:
			t.Fatal("twoPartyTest timeout")
		}
	}

	log.Info("Two-party test done")
}

func NewSetups(rng *rand.Rand, names []string) []ctest.RoleSetup {
	var (
		bus   = wire.NewLocalBus()
		n     = len(names)
		setup = make([]ctest.RoleSetup, n)
	)

	for i := 0; i < n; i++ {
		acc := wtest.NewRandomAccount(rng)
		setup[i] = ctest.RoleSetup{
			Name:        names[i],
			Identity:    acc,
			Bus:         bus,
			Funder:      &logFunder{log.WithField("role", names[i])},
			Adjudicator: &logAdjudicator{log.WithField("role", names[i])},
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

	logAdjudicator struct {
		log log.Logger
	}
)

func (f *logFunder) Fund(_ context.Context, req channel.FundingReq) error {
	f.log.Infof("Funding: %v", req)
	return nil
}

func (a *logAdjudicator) Register(_ context.Context, req channel.AdjudicatorReq) (*channel.RegisteredEvent, error) {
	a.log.Infof("Register: %v", req)
	return channel.NewRegisteredEvent(
		req.Params.ID(),
		&channel.ElapsedTimeout{},
		req.Tx.Version,
	), nil
}

func (a *logAdjudicator) Progress(_ context.Context, req channel.ProgressReq) error {
	a.log.Infof("Progress: %v", req)
	return nil
}

func (a *logAdjudicator) Withdraw(_ context.Context, req channel.AdjudicatorReq, subStates channel.StateMap) error {
	a.log.Infof("Withdraw: %v, %v", req, subStates)
	return nil
}

func (a *logAdjudicator) Subscribe(_ context.Context, params *channel.Params) (channel.AdjudicatorSubscription, error) {
	a.log.Infof("SubscribeRegistered: %v", params)
	return nil, nil
}

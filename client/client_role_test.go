// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package client_test

import (
	"context"
	"math/rand"
	"sync"
	"testing"
	"time"

	"perun.network/go-perun/channel"
	ctest "perun.network/go-perun/client/test"
	"perun.network/go-perun/log"
	wtest "perun.network/go-perun/wallet/test"
	"perun.network/go-perun/wire"
)

func executeTwoPartyTest(t *testing.T, role [2]ctest.Executer, cfg ctest.ExecConfig) {
	log.Info("Starting two-party test")

	// enable stages synchronization
	stages := role[0].EnableStages()
	role[1].SetStages(stages)

	var wg sync.WaitGroup
	wg.Add(2)
	for i := 0; i < 2; i++ {
		go func(i int) {
			defer wg.Done()
			log.Infof("Executing role %d", i)
			role[i].Execute(cfg)
		}(i)
	}

	wg.Wait()
	log.Info("Two-party test done")
}

var defaultTimeout = 1 * time.Second

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
			Timeout:     defaultTimeout,
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

func (a *logAdjudicator) Register(ctx context.Context, req channel.AdjudicatorReq) (*channel.RegisteredEvent, error) {
	a.log.Infof("Register: %v", req)
	return &channel.RegisteredEvent{
		ID:      req.Params.ID(),
		Version: req.Tx.Version,
		Timeout: &channel.ElapsedTimeout{},
	}, nil
}

func (a *logAdjudicator) Withdraw(ctx context.Context, req channel.AdjudicatorReq) error {
	a.log.Infof("Withdraw: %v", req)
	return nil
}

func (a *logAdjudicator) SubscribeRegistered(ctx context.Context, params *channel.Params) (channel.RegisteredSubscription, error) {
	a.log.Infof("SubscribeRegistered: %v", params)
	return nil, nil
}

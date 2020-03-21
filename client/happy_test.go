// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package client_test

import (
	"context"
	"math/big"
	"math/rand"
	"sync"
	"testing"
	"time"

	"perun.network/go-perun/channel"
	channeltest "perun.network/go-perun/channel/test"
	clienttest "perun.network/go-perun/client/test"
	"perun.network/go-perun/log"
	"perun.network/go-perun/peer"
	peertest "perun.network/go-perun/peer/test"
	"perun.network/go-perun/wallet"
	wallettest "perun.network/go-perun/wallet/test"
)

var defaultTimeout = 1 * time.Second

func TestHappyAliceBob(t *testing.T) {
	log.Info("Starting happy test")
	rng := rand.New(rand.NewSource(0x1337))

	const A, B = 0, 1 // Indices of Alice and Bob
	var (
		name  = [2]string{"Alice", "Bob"}
		hub   peertest.ConnHub
		acc   [2]wallet.Account
		setup [2]clienttest.RoleSetup
		role  [2]clienttest.Executer
	)

	for i := 0; i < 2; i++ {
		acc[i] = wallettest.NewRandomAccount(rng)
		setup[i] = clienttest.RoleSetup{
			Name:        name[i],
			Identity:    acc[i],
			Dialer:      hub.NewDialer(),
			Listener:    hub.NewListener(acc[i].Address()),
			Funder:      &logFunder{log.WithField("role", name[i])},
			Adjudicator: &logAdjudicator{log.WithField("role", name[i])},
			Timeout:     defaultTimeout,
		}
	}

	role[A] = clienttest.NewAlice(setup[A], t)
	role[B] = clienttest.NewBob(setup[B], t)
	// enable stages synchronization
	stages := role[A].EnableStages()
	role[B].SetStages(stages)

	execConfig := clienttest.ExecConfig{
		PeerAddrs:  [2]peer.Address{acc[A].Address(), acc[B].Address()},
		Asset:      channeltest.NewRandomAsset(rng),
		InitBals:   [2]*big.Int{big.NewInt(100), big.NewInt(100)},
		NumUpdates: [2]int{2, 2},
		TxAmounts:  [2]*big.Int{big.NewInt(5), big.NewInt(3)},
	}

	var wg sync.WaitGroup
	wg.Add(2)
	for i := 0; i < 2; i++ {
		go func(i int) {
			defer wg.Done()
			log.Infof("Starting %s.Execute", name[i])
			role[i].Execute(execConfig)
		}(i)
	}

	wg.Wait()
	log.Info("Happy test done")
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

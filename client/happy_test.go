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

	"github.com/stretchr/testify/assert"

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

	var hub peertest.ConnHub

	aliceAcc := wallettest.NewRandomAccount(rng)
	bobAcc := wallettest.NewRandomAccount(rng)

	setupAlice := clienttest.RoleSetup{
		Name:     "Alice",
		Identity: aliceAcc,
		Dialer:   hub.NewDialer(),
		Listener: hub.NewListener(aliceAcc.Address()),
		Funder:   &logFunder{log.WithField("role", "Alice")},
		Settler:  &logSettler{t, log.WithField("role", "Alice")},
		Timeout:  defaultTimeout,
	}

	setupBob := clienttest.RoleSetup{
		Name:     "Bob",
		Identity: bobAcc,
		Dialer:   hub.NewDialer(),
		Listener: hub.NewListener(bobAcc.Address()),
		Funder:   &logFunder{log.WithField("role", "Bob")},
		Settler:  &logSettler{t, log.WithField("role", "Bob")},
		Timeout:  defaultTimeout,
	}

	execConfig := clienttest.ExecConfig{
		PeerAddrs:       []peer.Address{aliceAcc.Address(), bobAcc.Address()},
		Asset:           channeltest.NewRandomAsset(rng),
		InitBals:        []*big.Int{big.NewInt(100), big.NewInt(100)},
		NumUpdatesBob:   2,
		NumUpdatesAlice: 2,
		TxAmountBob:     big.NewInt(5),
		TxAmountAlice:   big.NewInt(3),
	}

	alice := clienttest.NewAlice(setupAlice, t)
	bob := clienttest.NewBob(setupBob, t)
	// enable stages synchronization
	stages := alice.EnableStages()
	bob.SetStages(stages)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		log.Info("Starting Alice.Execute")
		alice.Execute(execConfig)
	}()

	go func() {
		defer wg.Done()
		log.Info("Starting Bob.Execute")
		bob.Execute(execConfig)
	}()

	wg.Wait()
	log.Info("Happy test done")
}

type (
	logFunder struct {
		log log.Logger
	}

	logSettler struct {
		t   *testing.T
		log log.Logger
	}
)

func (f *logFunder) Fund(_ context.Context, req channel.FundingReq) error {
	f.log.Infof("Funding: %v", req)
	return nil
}

func (s *logSettler) Settle(_ context.Context, req channel.SettleReq, _ wallet.Account) error {
	s.log.Infof("Settling: %v", req)
	for i, sig := range req.Tx.Sigs {
		ok, err := channel.Verify(req.Params.Parts[i], req.Params, req.Tx.State, sig)
		assert.True(s.t, ok)
		assert.NoError(s.t, err)
	}
	return nil
}

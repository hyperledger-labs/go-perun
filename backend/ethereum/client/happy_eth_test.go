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

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/backend/ethereum/channel"
	"perun.network/go-perun/backend/ethereum/channel/test"
	"perun.network/go-perun/backend/ethereum/wallet"
	ethwallettest "perun.network/go-perun/backend/ethereum/wallet/test"
	clienttest "perun.network/go-perun/client/test"
	"perun.network/go-perun/log"
	"perun.network/go-perun/peer"
	peertest "perun.network/go-perun/peer/test"
	wallettest "perun.network/go-perun/wallet/test"
)

var defaultTimeout = 5 * time.Second

func TestHappyAliceBobETH(t *testing.T) {
	log.Info("Starting happy test")
	const A, B = 0, 1 // Indices of Alice and Bob
	var hub peertest.ConnHub
	rng := rand.New(rand.NewSource(0x1337))
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	// Create a new KeyStore
	ks := ethwallettest.GetKeystore()
	// Create alice and bobs account
	aliceAcc := wallettest.NewRandomAccount(rng)
	bobAcc := wallettest.NewRandomAccount(rng)
	aliceAccETH := aliceAcc.(*wallet.Account).Account
	bobAccETH := bobAcc.(*wallet.Account).Account
	// Create SimulatedBackend
	backend := test.NewSimulatedBackend()
	// Fund both accounts
	backend.FundAddress(ctx, aliceAccETH.Address)
	backend.FundAddress(ctx, bobAccETH.Address)
	// Create contract backends
	cbAlice := channel.NewContractBackend(backend, ks, aliceAccETH)
	cbBob := channel.NewContractBackend(backend, ks, bobAccETH)
	// Deploy the contracts
	adjAddr, err := channel.DeployAdjudicator(ctx, cbAlice)
	require.NoError(t, err, "Adjudicator should be deployed successfully")
	assetAddr, err := channel.DeployETHAssetholder(ctx, cbBob, adjAddr)
	require.NoError(t, err, "ETHAssetholder should be deployed successfully")
	// Create the funders
	funderAlice := channel.NewETHFunder(cbAlice, assetAddr)
	funderBob := channel.NewETHFunder(cbBob, assetAddr)
	// Create distinct fund withdrawal receivers
	aliceRecv := wallettest.NewRandomAddress(rng).(*wallet.Address).Address
	bobRecv := wallettest.NewRandomAddress(rng).(*wallet.Address).Address
	// Create the adjudicators
	adjudicatorAlice := channel.NewAdjudicator(cbAlice, adjAddr, aliceRecv)
	adjudicatorBob := channel.NewAdjudicator(cbBob, adjAddr, bobRecv)

	setupAlice := clienttest.RoleSetup{
		Name:        "Alice",
		Identity:    aliceAcc,
		Dialer:      hub.NewDialer(),
		Listener:    hub.NewListener(aliceAcc.Address()),
		Funder:      funderAlice,
		Adjudicator: adjudicatorAlice,
		Timeout:     defaultTimeout,
	}

	setupBob := clienttest.RoleSetup{
		Name:        "Bob",
		Identity:    bobAcc,
		Dialer:      hub.NewDialer(),
		Listener:    hub.NewListener(bobAcc.Address()),
		Funder:      funderBob,
		Adjudicator: adjudicatorBob,
		Timeout:     defaultTimeout,
	}

	execConfig := clienttest.ExecConfig{
		PeerAddrs:  [2]peer.Address{aliceAcc.Address(), bobAcc.Address()},
		InitBals:   [2]*big.Int{big.NewInt(100), big.NewInt(100)},
		Asset:      &wallet.Address{Address: assetAddr},
		NumUpdates: [2]int{2, 2},
		TxAmounts:  [2]*big.Int{big.NewInt(5), big.NewInt(3)},
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

	// Assert correct final balances
	aliceToBob := big.NewInt(int64(execConfig.NumUpdates[A])*execConfig.TxAmounts[A].Int64() -
		int64(execConfig.NumUpdates[B])*execConfig.TxAmounts[B].Int64())
	finalBalAlice := new(big.Int).Sub(execConfig.InitBals[A], aliceToBob)
	finalBalBob := new(big.Int).Add(execConfig.InitBals[B], aliceToBob)
	// reset context timeout
	ctx, cancel = context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	assertBal := func(addr common.Address, bal *big.Int) {
		b, err := backend.BalanceAt(ctx, addr, nil)
		require.NoError(t, err)
		assert.Zero(t, bal.Cmp(b), "ETH balance mismatch")
	}
	assertBal(aliceRecv, finalBalAlice)
	assertBal(bobRecv, finalBalBob)

	log.Info("Happy test done")
}

// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package channel // import "perun.network/go-perun/backend/ethereum/channel"

import (
	"context"
	"math/rand"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"perun.network/go-perun/backend/ethereum/channel/test"
	"perun.network/go-perun/backend/ethereum/wallet"
	ethwallettest "perun.network/go-perun/backend/ethereum/wallet/test"
	"perun.network/go-perun/channel"
	perunwallet "perun.network/go-perun/wallet"
	wallettest "perun.network/go-perun/wallet/test"
)

func newAdjudicator(t *testing.T, rng *rand.Rand, n int) ([]perunwallet.Account, []perunwallet.Address, []*Funder, []*Adjudicator, common.Address) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	simBackend := test.NewSimulatedBackend()
	ks := ethwallettest.GetKeystore()
	deployAccount := wallettest.NewRandomAccount(rng).(*wallet.Account).Account
	simBackend.FundAddress(ctx, deployAccount.Address)
	contractBackend := NewContractBackend(simBackend, ks, deployAccount)
	// Deploy Adjudicator
	adjudicator, err := DeployAdjudicator(context.Background(), contractBackend)
	require.NoError(t, err, "Deploying the adjudicator should not error")
	// Deploy Assetholder
	assetETH, err := DeployETHAssetholder(ctx, contractBackend, adjudicator)
	if err != nil {
		panic(err)
	}
	t.Logf("asset holder address is %v", assetETH)
	t.Logf("adjudicator address is %v", adjudicator)
	accs := make([]perunwallet.Account, n)
	parts := make([]perunwallet.Address, n)
	funders := make([]*Funder, n)
	adjudicators := make([]*Adjudicator, n)
	for i := 0; i < n; i++ {
		acc := wallettest.NewRandomAccount(rng).(*wallet.Account)
		simBackend.FundAddress(ctx, acc.Account.Address)
		accs[i] = acc
		parts[i] = acc.Address()
		cb := NewContractBackend(simBackend, ks, acc.Account)
		funders[i] = NewETHFunder(cb, assetETH)
		adjudicators[i] = NewETHAdjudicator(cb, adjudicator, deployAccount.Address)
	}
	return accs, parts, funders, adjudicators, assetETH
}

func signState(t *testing.T, accounts []perunwallet.Account, params *channel.Params, state *channel.State) channel.Transaction {
	// Sign valid state.
	sigs := make([][]byte, len(accounts))
	for i := 0; i < len(accounts); i++ {
		sig, err := Sign(accounts[i], params, state)
		assert.NoError(t, err, "Sign should not return error")
		sigs[i] = sig
	}
	return channel.Transaction{
		State: state,
		Sigs:  sigs,
	}
}

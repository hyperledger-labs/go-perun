// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package channel // import "perun.network/go-perun/backend/ethereum/channel"

import (
	"context"
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"perun.network/go-perun/backend/ethereum/channel/test"
	"perun.network/go-perun/backend/ethereum/wallet"
	ethwallettest "perun.network/go-perun/backend/ethereum/wallet/test"
	"perun.network/go-perun/channel"
	channeltest "perun.network/go-perun/channel/test"
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
	require.NoError(t, err, "Deploying asset holder failed")
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
		adjudicators[i] = NewAdjudicator(cb, adjudicator, deployAccount.Address)
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

func TestSubscribeRegistered(t *testing.T) {
	seed := time.Now().UnixNano()
	t.Logf("seed is %v", seed)
	rng := rand.New(rand.NewSource(int64(seed)))
	// create new Adjudicator
	accs, parts, funders, adjs, asset := newAdjudicator(t, rng, 1)
	// create valid state and params
	app := channeltest.NewRandomApp(rng)
	params := channel.NewParamsUnsafe(uint64(100*time.Second), parts, app.Def(), big.NewInt(rng.Int63()))
	state := newValidState(rng, params, asset)
	// Set up subscription
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	registered, err := adjs[0].SubscribeRegistered(ctx, params)
	assert.NoError(t, err, "Subscribing to valid params should not error")
	// we need to properly fund the channel
	fundingCtx, funCancel := context.WithTimeout(context.Background(), timeout)
	defer funCancel()
	// fund the contract
	reqFund := channel.FundingReq{
		Params:     params,
		Allocation: &state.Allocation,
		Idx:        channel.Index(0),
	}
	require.NoError(t, funders[0].Fund(fundingCtx, reqFund), "funding should succeed")
	// Now test the register function
	ctxReg, regCancel := context.WithTimeout(context.Background(), timeout)
	defer regCancel()
	tx := signState(t, accs, params, state)
	req := channel.AdjudicatorReq{
		Params: params,
		Acc:    accs[0],
		Idx:    channel.Index(0),
		Tx:     tx,
	}
	event, err := adjs[0].Register(ctxReg, req)
	assert.NoError(t, err, "Registering state should succeed")
	assert.Equal(t, event, registered.Next(), "Events should be equal")
	assert.NoError(t, registered.Close(), "Closing event channel should not error")
	assert.Nil(t, registered.Next(), "Next on closed channel should produce nil")
	assert.NoError(t, registered.Err(), "Closing should produce no error")
	// Setup a new subscription
	registered2, err := adjs[0].SubscribeRegistered(ctx, params)
	assert.NoError(t, err, "registering two subscriptions should not fail")
	assert.Equal(t, event, registered2.Next(), "Events should be equal")
	assert.NoError(t, registered2.Close(), "Closing event channel should not error")
	assert.Nil(t, registered2.Next(), "Next on closed channel should produce nil")
	assert.NoError(t, registered2.Err(), "Closing should produce no error")
}

func newValidState(rng *rand.Rand, params *channel.Params, assetholder common.Address) *channel.State {
	// Create valid state.
	assets := []channel.Asset{
		&Asset{Address: assetholder},
	}
	ofparts := make([][]channel.Bal, len(params.Parts))
	for i := 0; i < len(ofparts); i++ {
		ofparts[i] = make([]channel.Bal, len(assets))
		for k := 0; k < len(assets); k++ {
			ofparts[i][k] = big.NewInt(rng.Int63n(999) + 1)
		}
	}
	allocation := channel.Allocation{
		Assets:  assets,
		OfParts: ofparts,
		Locked:  []channel.SubAlloc{},
	}

	return &channel.State{
		ID:         params.ID(),
		Version:    4,
		App:        params.App,
		Allocation: allocation,
		Data:       channeltest.NewRandomData(rng),
		IsFinal:    false,
	}
}

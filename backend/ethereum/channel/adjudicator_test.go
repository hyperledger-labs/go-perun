// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package channel_test

import (
	"context"
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ethchannel "perun.network/go-perun/backend/ethereum/channel"
	"perun.network/go-perun/backend/ethereum/channel/test"
	ethwallet "perun.network/go-perun/backend/ethereum/wallet"
	"perun.network/go-perun/channel"
	channeltest "perun.network/go-perun/channel/test"
)

const defaultTxTimeout = 2 * time.Second

func signState(t *testing.T, accounts []*ethwallet.Account, params *channel.Params, state *channel.State) channel.Transaction {
	// Sign valid state.
	sigs := make([][]byte, len(accounts))
	for i := 0; i < len(accounts); i++ {
		sig, err := channel.Sign(accounts[i], params, state)
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
	// create test setup
	s := test.NewSetup(t, rng, 1)
	// create valid state and params
	app := channeltest.NewRandomApp(rng)
	params := channel.NewParamsUnsafe(uint64(100*time.Second), s.Parts, app.Def(), big.NewInt(rng.Int63()))
	state := newValidState(rng, params, s.Asset)
	// Set up subscription
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	registered, err := s.Adjs[0].SubscribeRegistered(ctx, params)
	require.NoError(t, err, "Subscribing to valid params should not error")
	// we need to properly fund the channel
	txCtx, txCancel := context.WithTimeout(context.Background(), defaultTxTimeout)
	defer txCancel()
	// fund the contract
	reqFund := channel.FundingReq{
		Params:     params,
		Allocation: &state.Allocation,
		Idx:        channel.Index(0),
	}
	require.NoError(t, s.Funders[0].Fund(txCtx, reqFund), "funding should succeed")
	// Now test the register function
	tx := signState(t, s.Accs, params, state)
	req := channel.AdjudicatorReq{
		Params: params,
		Acc:    s.Accs[0],
		Idx:    channel.Index(0),
		Tx:     tx,
	}
	event, err := s.Adjs[0].Register(txCtx, req)
	assert.NoError(t, err, "Registering state should succeed")
	assert.Equal(t, event, registered.Next(), "Events should be equal")
	assert.NoError(t, registered.Close(), "Closing event channel should not error")
	assert.Nil(t, registered.Next(), "Next on closed channel should produce nil")
	assert.NoError(t, registered.Err(), "Closing should produce no error")
	// Setup a new subscription
	registered2, err := s.Adjs[0].SubscribeRegistered(ctx, params)
	assert.NoError(t, err, "registering two subscriptions should not fail")
	assert.Equal(t, event, registered2.Next(), "Events should be equal")
	assert.NoError(t, registered2.Close(), "Closing event channel should not error")
	assert.Nil(t, registered2.Next(), "Next on closed channel should produce nil")
	assert.NoError(t, registered2.Err(), "Closing should produce no error")
}

func newValidState(rng *rand.Rand, params *channel.Params, assetholder common.Address) *channel.State {
	// Create valid state.
	assets := []channel.Asset{
		&ethchannel.Asset{Address: assetholder},
	}
	ofparts := make([][]channel.Bal, len(params.Parts))
	for i := 0; i < len(ofparts); i++ {
		ofparts[i] = make([]channel.Bal, len(assets))
		for k := 0; k < len(assets); k++ {
			ofparts[i][k] = big.NewInt(rng.Int63n(999) + 1)
		}
	}
	allocation := channel.Allocation{
		Assets:   assets,
		Balances: ofparts,
		Locked:   []channel.SubAlloc{},
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

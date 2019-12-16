// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel

import (
	"context"
	"math/big"
	"math/rand"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"perun.network/go-perun/backend/ethereum/channel/test"
	"perun.network/go-perun/backend/ethereum/wallet"
	"perun.network/go-perun/channel"
	channeltest "perun.network/go-perun/channel/test"
	perunwallet "perun.network/go-perun/wallet"
	wallettest "perun.network/go-perun/wallet/test"
)

func TestSettler_MultipleSettles(t *testing.T) {
	t.Run("Settle 1 party parallel", func(t *testing.T) { settleMultipleConcurrent(t, 1, true) })
	t.Run("Settle 2 party parallel", func(t *testing.T) { settleMultipleConcurrent(t, 2, true) })
	t.Run("Settle 5 party parallel", func(t *testing.T) { settleMultipleConcurrent(t, 5, true) })
	t.Run("Settle 1 party sequential", func(t *testing.T) { settleMultipleConcurrent(t, 1, false) })
	t.Run("Settle 2 party sequential", func(t *testing.T) { settleMultipleConcurrent(t, 2, false) })
	t.Run("Settle 5 party sequential", func(t *testing.T) { settleMultipleConcurrent(t, 5, false) })
}

func settleMultipleConcurrent(t *testing.T, numParts int, parallel bool) {
	seed := 1337 * numParts
	if parallel {
		seed++
	}
	rng := rand.New(rand.NewSource(int64(seed)))
	settler, req, accounts := newSettlerAndRequest(t, rng, numParts, true)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if parallel {
		var wg sync.WaitGroup
		wg.Add(numParts)
		for i := 0; i < numParts; i++ {
			go func(i int) {
				defer wg.Done()
				err := settler.Settle(ctx, req, accounts[i])
				assert.NoError(t, err, "Settling should succeed")
			}(i)
		}
		wg.Wait()
	} else {
		for i := 0; i < numParts; i++ {
			assert.NoError(t, settler.Settle(ctx, req, accounts[i]), "Settling should succeed")
		}
	}
}

func TestSettler_CancelledContext(t *testing.T) {
	rng := rand.New(rand.NewSource(13))
	settler, req, accounts := newSettlerAndRequest(t, rng, 2, true)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	cancel()
	assert.Error(t, settler.Settle(ctx, req, accounts[0]), "Settling on cancelled context should fail")
}

func TestSettler_nonfinalState(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	settler, req, accounts := newSettlerAndRequest(t, rng, 2, false)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	assert.Panics(t, func() { settler.Settle(ctx, req, accounts[0]) }, "Uncooperative settle should panic - not implemented yet")
}

func newSettlerAndRequest(t *testing.T, rng *rand.Rand, numParts int, final bool) (*Settler, channel.SettleReq, []perunwallet.Account) {
	s := newSimulatedSettler()
	f := &Funder{
		ContractBackend: s.ContractBackend,
	}
	adjudicator, err := DeployAdjudicator(context.Background(), s.ContractBackend)
	require.NoError(t, err, "Deploying the adjudicator should not error")
	s.adjAddr = adjudicator
	assetholder, err := DeployETHAssetholder(context.Background(), s.ContractBackend, adjudicator)
	require.NoError(t, err, "Deploying the eth assetholder should not fail")
	f.ethAssetHolder = assetholder
	// Create valid parameters.
	app := channeltest.NewRandomApp(rng)
	accounts := make([]perunwallet.Account, numParts)
	parts := make([]perunwallet.Address, numParts)
	for i := 0; i < numParts; i++ {
		acc := wallettest.NewRandomAccount(rng)
		accounts[i] = acc
		parts[i] = acc.Address()
	}
	params := channel.NewParamsUnsafe(uint64(0), parts, app.Def(), big.NewInt(rng.Int63()))
	state := newValidState(rng, params, assetholder)
	state.IsFinal = final
	// Sign valid state.
	sigs := make([][]byte, numParts)
	for i := 0; i < numParts; i++ {
		sig, err := Sign(accounts[i], params, state)
		assert.NoError(t, err, "Sign should not return error")
		sigs[i] = sig
	}
	tx := channel.Transaction{
		State: state,
		Sigs:  sigs,
	}

	req := channel.SettleReq{
		Params: params,
		Idx:    uint16(0),
		Tx:     tx,
	}
	return s, req, accounts
}

func newSimulatedSettler() *Settler {
	wall := new(wallet.Wallet)
	wall.Connect(keyDir, password)
	acc := wall.Accounts()[0].(*wallet.Account)
	acc.Unlock(password)
	ks := wall.Ks
	simBackend := test.NewSimulatedBackend()
	simBackend.FundAddress(context.Background(), acc.Account.Address)
	return &Settler{
		ContractBackend: ContractBackend{simBackend, ks, acc.Account},
	}
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

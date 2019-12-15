// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel

import (
	"context"
	"math/big"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"perun.network/go-perun/backend/ethereum/channel/test"
	"perun.network/go-perun/backend/ethereum/wallet"
	"perun.network/go-perun/channel"
	channeltest "perun.network/go-perun/channel/test"
	perunwallet "perun.network/go-perun/wallet"
	wallettest "perun.network/go-perun/wallet/test"
)

func TestSettler_Settle(t *testing.T) {
	rng := rand.New(rand.NewSource(1337))
	s := newSimulatedSettler()
	f := &Funder{
		ContractBackend: s.ContractBackend,
	}
	adjudicator, err := DeployAdjudicator(s.ContractBackend)
	assert.NoError(t, err, "Deploying the adjudicator should not error")
	s.adjAddr = adjudicator
	assetholder, err := DeployETHAssetholder(s.ContractBackend, adjudicator)
	assert.NoError(t, err, "Deploying the eth assetholder should not fail")
	f.ethAssetHolder = assetholder
	assert.Panics(t, func() { s.Settle(context.Background(), channel.SettleReq{}, &wallet.Account{}) },
		"Funding with invalid settle request should fail")
	// Create valid parameters.
	app := channeltest.NewRandomApp(rng)
	offChainAcc := wallettest.NewRandomAccount(rng)
	parts := []perunwallet.Address{
		offChainAcc.Address(),
	}
	params := channel.NewParamsUnsafe(uint64(0), parts, app.Def(), big.NewInt(rng.Int63()))
	// Create valid state.
	assets := []channel.Asset{
		&Asset{Address: assetholder},
	}
	ofparts := make([][]channel.Bal, len(parts))
	for i := 0; i < len(ofparts); i++ {
		ofparts[i] = make([]channel.Bal, len(assets))
		for k := 0; k < len(assets); k++ {
			ofparts[i][k] = big.NewInt(0)
		}
	}
	allocation := channel.Allocation{
		Assets:  assets,
		OfParts: ofparts,
		Locked:  []channel.SubAlloc{},
	}

	state := channel.State{
		ID:         params.ID(),
		Version:    4,
		App:        params.App,
		Allocation: allocation,
		Data:       channeltest.NewRandomData(rng),
		IsFinal:    true,
	}
	// Sign valid state.
	sig, err := Sign(offChainAcc, params, &state)
	assert.NoError(t, err, "Sign should not return error")
	tx := channel.Transaction{
		State: &state,
		Sigs:  [][]byte{sig},
	}

	req := channel.SettleReq{
		Params: params,
		Idx:    uint16(0),
		Tx:     tx,
	}

	err = s.Settle(context.Background(), req, offChainAcc)
	assert.NoError(t, err, "Settling with valid state should not produce error")
	err = s.Settle(context.Background(), req, offChainAcc)
	assert.Error(t, err, "Settling twice should fail")
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

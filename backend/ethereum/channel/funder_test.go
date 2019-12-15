// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel

import (
	"context"
	"math/big"
	"math/rand"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"perun.network/go-perun/backend/ethereum/wallet"
	"perun.network/go-perun/channel"

	"perun.network/go-perun/backend/ethereum/channel/test"
	channeltest "perun.network/go-perun/channel/test"
	perunwallet "perun.network/go-perun/wallet"
)

const nodeURL = "ws://localhost:8545"

const (
	keyDir   = "../wallet/testdata"
	password = "secret"

	keystoreAddr = "0x3c5A96FF258B1F4C288068B32474dedC3620233c"
	keyStorePath = "UTC--2019-06-07T12-12-48.775026092Z--3c5a96ff258b1f4c288068b32474dedc3620233c"
)

func TestFunder_Fund(t *testing.T) {
	f := newSimulatedFunder()
	assert.Panics(t, func() { f.Fund(context.Background(), channel.FundingReq{}) }, "Funding with invalid funding req should fail")
	req := channel.FundingReq{
		Params:     &channel.Params{},
		Allocation: &channel.Allocation{},
		Idx:        0,
	}
	assert.NoError(t, f.Fund(context.Background(), req), "Funding with no assets should succeed")
	parts := []perunwallet.Address{
		&wallet.Address{Address: f.account.Address},
	}
	rng := rand.New(rand.NewSource(1337))
	app := channeltest.NewRandomApp(rng)
	params := channel.NewParamsUnsafe(uint64(0), parts, app.Def(), big.NewInt(rng.Int63()))
	allocation := newValidAllocation(f, parts, common.HexToAddress(keystoreAddr))
	req = channel.FundingReq{
		Params:     params,
		Allocation: allocation,
		Idx:        0,
	}
	assert.NoError(t, f.Fund(context.Background(), req), "funding with valid request should succeed")
	assert.NoError(t, f.Fund(context.Background(), req), "multiple funding should succeed")
}

func newSimulatedFunder() *Funder {
	// Set KeyStore
	wall := new(wallet.Wallet)
	wall.Connect(keyDir, password)
	acc := wall.Accounts()[0].(*wallet.Account)
	acc.Unlock(password)
	ks := wall.Ks
	simBackend := test.NewSimulatedBackend()
	simBackend.FundAddress(context.Background(), acc.Account.Address)
	return &Funder{
		ContractBackend: ContractBackend{simBackend, ks, acc.Account}}
}

func newValidAllocation(f *Funder, parts []perunwallet.Address, adjudicatorAddr common.Address) *channel.Allocation {
	// Deploy Assetholder
	assetETH, err := DeployETHAssetholder(f.ContractBackend, adjudicatorAddr)
	if err != nil {
		panic(err)
	}
	f.ethAssetHolder = assetETH
	// Create assets slice
	assets := []channel.Asset{
		&Asset{Address: assetETH},
	}
	ofparts := make([][]channel.Bal, len(parts))
	for i := 0; i < len(ofparts); i++ {
		ofparts[i] = make([]channel.Bal, len(assets))
		for k := 0; k < len(assets); k++ {
			ofparts[i][k] = big.NewInt(1)
		}
	}
	return &channel.Allocation{
		Assets:  assets,
		OfParts: ofparts,
	}
}

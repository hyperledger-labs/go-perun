// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel

import (
	"context"
	"math/big"
	"math/rand"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"perun.network/go-perun/backend/ethereum/bindings/assets"
	"perun.network/go-perun/backend/ethereum/wallet"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/test"
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
	assert.Error(t, f.Fund(context.Background(), channel.FundingReq{}), "Funding with invalid funding req should fail")
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
	app := test.NewRandomApp(rng)
	params := channel.NewParamsUnsafe(uint64(0), parts, app.Def(), big.NewInt(rng.Int63()))
	allocation := newValidAllocation(f, parts, common.HexToAddress(keystoreAddr))
	req = channel.FundingReq{
		Params:     params,
		Allocation: allocation,
		Idx:        0,
	}
	assert.NoError(t, f.Fund(context.Background(), req))
}

type simulatedBackend struct {
	backends.SimulatedBackend
}

func newSimulatedBackend() *simulatedBackend {
	return &simulatedBackend{*backends.NewSimulatedBackend(nil, 8000000)}
}

func (s *simulatedBackend) BlockByNumber(_ context.Context, number *big.Int) (*types.Block, error) {
	if number == nil {
		return s.Blockchain().CurrentBlock(), nil
	}
	return s.Blockchain().GetBlockByNumber(number.Uint64()), nil
}

func (s *simulatedBackend) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	if err := s.SimulatedBackend.SendTransaction(ctx, tx); err != nil {
		return errors.WithStack(err)
	}
	s.Commit()
	return nil
}

func deployETHAssetHolder(f *Funder, adjudicatorAddr common.Address) common.Address {
	auth, err := f.client.newTransactor(context.Background(), f.ks, f.account, big.NewInt(0), 7999999)
	addr, _, _, err := assets.DeployAssetHolderETH(auth, f.client, adjudicatorAddr)
	mixedCase := common.NewMixedcaseAddress(addr)
	ETHAssetHolder = &mixedCase
	if err != nil {
		panic(err)
	}
	return addr
}

func newSimulatedFunder() *Funder {
	f := &Funder{}
	// Set KeyStore
	wall := new(wallet.Wallet)
	wall.Connect(keyDir, password)
	acc := wall.Accounts()[0].(*wallet.Account)
	acc.Unlock(password)
	ks := wall.Ks
	f.ks = ks
	f.account = acc.Account
	f.client = contractBackend{newSimulatedBackend()}
	return f
}

func newValidAllocation(f *Funder, parts []perunwallet.Address, adjudicatorAddr common.Address) *channel.Allocation {
	assetETH := deployETHAssetHolder(f, adjudicatorAddr)
	assets := []channel.Asset{
		&Asset{Address: assetETH},
	}
	ofparts := make([][]channel.Bal, len(parts))
	for i := 0; i < len(ofparts); i++ {
		ofparts[i] = make([]channel.Bal, len(assets))
		for k := 0; k < len(assets); k++ {
			ofparts[i][k] = big.NewInt(0)
		}
	}
	return &channel.Allocation{
		Assets:  assets,
		OfParts: ofparts,
	}
}

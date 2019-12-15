// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel

import (
	"context"
	"math/big"

	"github.com/EthLaika/go-laika/core/types"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"perun.network/go-perun/backend/ethereum/bindings/adjudicator"
	"perun.network/go-perun/backend/ethereum/bindings/assets"
)

// Deployer implements the channel.Deployer interface for Ethereum.
type Deployer struct {
	client  contractBackend
	ks      *keystore.KeyStore
	account *accounts.Account
}

// NewETHDeployer creates a new ethereum deployer.
func NewETHDeployer(client *ethclient.Client, keystore *keystore.KeyStore, account *accounts.Account) Deployer {
	return Deployer{
		client:  contractBackend{client},
		ks:      keystore,
		account: account,
	}
}

// DeployETHAssetholder deploys a new ETHAssetHolder contract
func (d *Deployer) DeployETHAssetholder(adjudicatorAddr common.Address) common.Address {
	auth, err := d.client.newTransactor(context.Background(), d.ks, d.account, big.NewInt(0), 6600000)
	if err != nil {
		panic(err)
	}
	addr, tx, _, err := assets.DeployAssetHolderETH(auth, d.client, adjudicatorAddr)
	if err != nil {
		panic(err)
	}
	receipt, err := bind.WaitMined(context.Background(), d.client, tx)
	if err != nil {
		panic(err)
	}
	if receipt.Status == types.ReceiptStatusFailed {
		panic("could not deploy ethassetholder")
	}
	return addr
}

// DeployAdjudicator deploys a new Adjudicator contract
func (d *Deployer) DeployAdjudicator() common.Address {
	auth, err := d.client.newTransactor(context.Background(), d.ks, d.account, big.NewInt(0), 6600000)
	if err != nil {
		panic(err)
	}
	addr, tx, _, err := adjudicator.DeployAdjudicator(auth, d.client)
	if err != nil {
		panic(err)
	}
	receipt, err := bind.WaitMined(context.Background(), d.client, tx)
	if err != nil {
		panic(err)
	}
	if receipt.Status == types.ReceiptStatusFailed {
		panic("could not deploy adjudicator")
	}
	return addr
}

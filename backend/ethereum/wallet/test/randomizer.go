// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package test // import "perun.network/go-perun/backend/ethereum/wallet/test"

import (
	"io/ioutil"
	"log"
	"math/rand"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"

	ethwallet "perun.network/go-perun/backend/ethereum/wallet"
	"perun.network/go-perun/wallet"
)

// randomizer implements the channel.test.Backend interface.
type randomizer struct {
	wallet *ethwallet.Wallet
}

// NewRandomizer creates a new randomized keystore.
func newRandomizer() *randomizer {
	return &randomizer{wallet: NewTmpWallet()}
}

// NewRandomAddress creates a new random address.
func (r *randomizer) NewRandomAddress(rnd *rand.Rand) wallet.Address {
	addr := NewRandomAddress(rnd)
	return &addr
}

// NewRandomAddress creates a new random account.
func (r *randomizer) NewRandomAccount(rnd *rand.Rand) wallet.Account {
	return r.wallet.NewRandomAccount(rnd)
}

// NewRandomAddress creates a new random ethereum address.
func NewRandomAddress(rnd *rand.Rand) ethwallet.Address {
	var a common.Address
	rnd.Read(a[:])
	return ethwallet.Address(a)
}

// RandomWallet returns the randomizer's wallet that contains all the accounts
// created using NewRandomAccount.
func (r *randomizer) RandomWallet() *ethwallet.Wallet {
	return r.wallet
}

// NewTmpWallet creates a wallet that uses a unique temporary directory to
// store its keys.
func NewTmpWallet() *ethwallet.Wallet {
	const prefix = "go-perun-test-eth-keystore-"
	tmpDir, err := ioutil.TempDir("", prefix)
	if err != nil {
		log.Panicf("Could not create TempDir: %v", err)
	}
	const scryptN = 2
	const scryptP = 1
	w, err := ethwallet.NewWallet(keystore.NewKeyStore(tmpDir, scryptN, scryptP), tmpDir)
	if err != nil {
		log.Panic("Could not create wallet:", err)
	}
	return w
}

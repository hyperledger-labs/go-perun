// Copyright 2019 - See NOTICE file for copyright holders.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

import (
	"io/ioutil"
	"log"
	"math/rand"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"

	ethwallet "perun.network/go-perun/backend/ethereum/wallet"
	"perun.network/go-perun/wallet"
	wtest "perun.network/go-perun/wallet/test"
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
func (r *randomizer) RandomWallet() wtest.Wallet {
	return r.wallet
}

// NewWallet returns a new KeyStore Wallet for testing purposes which is saved
// to a temporary directory.
func (r *randomizer) NewWallet() wtest.Wallet {
	return NewTmpWallet()
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

// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package test // import "perun.network/go-perun/backend/ethereum/wallet/test"

import (
	"crypto/ecdsa"
	"io/ioutil"
	"log"
	"math/rand"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"

	"perun.network/go-perun/backend/ethereum/wallet"
	perunwallet "perun.network/go-perun/wallet"
)

// randomizer implements the channel.test.Backend interface.
type randomizer struct {
	wallet *wallet.Wallet
}

// NewRandomizer creates a new randomized keystore.
func newRandomizer() *randomizer {
	return &randomizer{wallet: NewTmpWallet()}
}

// NewRandomAddress creates a new random address.
func (r *randomizer) NewRandomAddress(rnd *rand.Rand) perunwallet.Address {
	addr := NewRandomAddress(rnd)
	return &addr
}

// NewRandomAddress creates a new random account.
func (r *randomizer) NewRandomAccount(rnd *rand.Rand) perunwallet.Account {
	// Generate a new private key.
	privateKey, err := ecdsa.GenerateKey(secp256k1.S256(), rnd)
	if err != nil {
		log.Panicf("Creation of account failed with error: %v", err)
	}

	keystore := r.wallet.Ks
	// Check if the keystore already holds this key.
	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	if acc, err := keystore.Find(accounts.Account{Address: address}); err == nil {
		return wallet.NewAccountFromEth(r.wallet, &acc)
	}
	// Store the private key in the keystore.
	ethAcc, err := keystore.ImportECDSA(privateKey, "secret")
	if err != nil {
		log.Panicf("Could not store private key in keystore: %v", err)
	}
	acc := wallet.NewAccountFromEth(r.wallet, &ethAcc)
	// Unlock the account before returning it.
	acc.Unlock("secret")
	return acc
}

// NewRandomAddress creates a new random ethereum address.
func NewRandomAddress(rnd *rand.Rand) wallet.Address {
	var a common.Address
	rnd.Read(a[:])
	return wallet.Address{Address: a}
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

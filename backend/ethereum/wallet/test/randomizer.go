// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package test // import "perun.network/go-perun/backend/ethereum/wallet/test"

import (
	"crypto/ecdsa"
	"io/ioutil"
	"log"
	"math/rand"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
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
	const prefix = "go-perun-test-eth-keystore-"
	tmpDir, err := ioutil.TempDir("", prefix)
	if err != nil {
		log.Panicf("Could not create TempDir, error: %v", err)
	}

	const scryptN = 2
	const scryptP = 1
	return &randomizer{
		wallet: wallet.NewWallet(keystore.NewKeyStore(tmpDir, scryptN, scryptP), tmpDir)}
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

	// Store the private key in the keystore.
	keystore := r.wallet.Ks
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

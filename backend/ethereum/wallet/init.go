// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package wallet // import "perun.network/go-perun/backend/ethereum/wallet"

import (
	"log"
	"math/rand"
	"os"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"

	perunwallet "perun.network/go-perun/wallet"
	"perun.network/go-perun/wallet/test"
)

func init() {
	perunwallet.SetBackend(new(Backend))
	test.SetBackend(newRandomizer())
}

// randomizer implements the channel.test.Backend interface.
type randomizer struct {
	wallet Wallet
}

const testDir = "/tmp/tempKeyStoreDir"

// NewRandomizer creates a new randomized keystore.
func newRandomizer() *randomizer {
	// Remove temp keystore if it exists.
	os.RemoveAll(testDir)
	return &randomizer{
		wallet: Wallet{
			ks:        keystore.NewPlaintextKeyStore(testDir),
			directory: testDir,
		},
	}
}

// NewRandomAddress creates a new random address.
func (r *randomizer) NewRandomAddress(rnd *rand.Rand) perunwallet.Address {
	addr := NewRandomAddress(rnd)
	return &addr
}

// NewRandomAddress creates a new random account.
func (r *randomizer) NewRandomAccount(rnd *rand.Rand) perunwallet.Account {
	// Generate a new private key.
	var random [32]byte
	rnd.Read(random[:])
	privateKey, err := crypto.ToECDSA(random[:])
	if err != nil {
		log.Panicf("Creation of account failed with error: %v", err)
	}
	// Store the private key in the keystore.
	keystore := r.wallet.ks
	ethAcc, err := keystore.ImportECDSA(privateKey, "secret")
	if err != nil {
		log.Panicf("Could not store private key in keystore: %v", err)
	}
	acc := newAccountFromEth(&r.wallet, &ethAcc)
	// Unlock the account before returning it.
	acc.Unlock("secret")
	return acc
}

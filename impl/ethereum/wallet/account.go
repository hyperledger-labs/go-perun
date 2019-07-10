// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package wallet

import (
	"sync"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/crypto"
	perun "perun.network/go-perun/wallet"
)

// Account represents an ethereum account
type Account struct {
	address  Address
	account  *accounts.Account
	wallet   *Wallet
	unlocked bool
	mu       sync.RWMutex
}

// Address returns the ethereum address of this account
func (e *Account) Address() perun.Address {
	return &e.address
}

// Unlock unlocks this account
func (e *Account) Unlock(password string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	err := e.wallet.ks.Unlock(*e.account, password)
	if err != nil {
		return err
	}
	e.unlocked = true
	return nil
}

// IsUnlocked checks if this account is unlocked
func (e *Account) IsUnlocked() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.unlocked
}

// Lock locks this account
func (e *Account) Lock() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	err := e.wallet.ks.Lock(e.address.Address)
	if err != nil {
		return err
	}
	e.unlocked = false
	return nil
}

// SignData is used to sign data with this account
func (e *Account) SignData(data []byte) ([]byte, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	hash := crypto.Keccak256(data)
	return e.wallet.ks.SignHash(*e.account, hash)
}

// SignDataWithPW is used to sign a hash with this account and a pw
func (e *Account) SignDataWithPW(password string, data []byte) ([]byte, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	hash := crypto.Keccak256(data)
	return e.wallet.ks.SignHashWithPassphrase(*e.account, password, hash)
}

func fromAccount(wallet *Wallet, account *accounts.Account) perun.Account {
	var acc Account
	acc.address = Address{account.Address}
	acc.account = account
	acc.wallet = wallet
	return &acc
}

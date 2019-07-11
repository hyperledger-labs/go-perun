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
	address Address
	account *accounts.Account
	wallet  *Wallet
	locked  bool
	mu      sync.RWMutex
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
	e.locked = false
	return nil
}

// IsLocked checks if this account is locked
func (e *Account) IsLocked() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.locked
}

// Lock locks this account
func (e *Account) Lock() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	err := e.wallet.ks.Lock(e.address.Address)
	if err != nil {
		return err
	}
	e.locked = true
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

func newAccountFromEth(wallet *Wallet, account *accounts.Account) *Account {
	return &Account{
		address: Address{account.Address},
		account: account,
		wallet:  wallet,
		locked:  true,
	}
}

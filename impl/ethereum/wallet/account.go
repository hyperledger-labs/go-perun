// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package wallet

import (
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/crypto"
	perun "github.com/perun-network/go-perun/wallet"
)

// Account represents an ethereum account
type Account struct {
	address  Address
	account  *accounts.Account
	wallet   *Wallet
	unlocked bool
}

// Address returns the ethereum address of this account
func (e *Account) Address() perun.Address {
	return &e.address
}

// Wallet returns the wallet in which this account is saved
func (e *Account) Wallet() perun.Wallet {
	return e.wallet
}

// Path returns the path to this account
func (e *Account) Path() string {
	return e.account.URL.String()
}

// Unlock unlocks this account
func (e *Account) Unlock(password string) error {
	err := e.wallet.ks.Unlock(*e.account, password)
	if err != nil {
		return err
	}
	e.unlocked = true
	return nil
}

// IsUnlocked checks if this account is unlocked
func (e *Account) IsUnlocked() bool {
	return e.unlocked
}

// Lock locks this account
func (e *Account) Lock() error {
	err := e.wallet.ks.Lock(e.address.Address)
	if err != nil {
		return err
	}
	e.unlocked = false
	return nil
}

// SignData is used to sign data with this account
func (e *Account) SignData(data []byte) ([]byte, error) {
	hash := crypto.Keccak256(data)
	return e.wallet.ks.SignHash(*e.account, hash)
}

// SignDataWithPW is used to sign a hash with this account and a pw
func (e *Account) SignDataWithPW(password string, data []byte) ([]byte, error) {
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

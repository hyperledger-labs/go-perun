// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

// Package wallet defines an etherum wallet.
// It can be used by the framework to interact with a file wallet.
// It uses an ethereum keystore internally which can be found at
// https://github.com/ethereum/go-ethereum/tree/master/accounts/keystore.
package wallet // import "perun.network/go-perun/impl/ethereum/wallet"

import (
	"os"
	"strconv"
	"sync"

	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	perun "perun.network/go-perun/wallet"
)

// Wallet represents an ethereum wallet
// It uses the go-ethereum keystore to store keys.
// Accessing the wallet is threadsafe, however you should not create two wallets from the same key directory
type Wallet struct {
	ks        *keystore.KeyStore
	directory string
	accounts  map[string]*Account
	mu        sync.RWMutex
}

// Path returns the path to this wallet
func (e *Wallet) Path() string {
	return e.directory
}

// refreshAccounts refreshes which accounts are connected to this wallet
func (e *Wallet) refreshAccounts() {
	if e.ks == nil {
		return
	}
	e.mu.Lock()
	defer e.mu.Unlock()

	accounts := e.ks.Accounts()
	for _, tmp := range accounts {
		if _, exists := e.accounts[tmp.Address.String()]; !exists {
			e.accounts[tmp.Address.String()] = newAccountFromEth(e, &tmp)
		}
	}
}

// Connect connects to this wallet
func (e *Wallet) Connect(keyDir, password string) error {
	if _, err := os.Stat(keyDir); os.IsNotExist(err) {
		return errors.New("key directory does not exist")
	}
	e.ks = keystore.NewKeyStore(keyDir, keystore.StandardScryptN, keystore.StandardScryptP)
	e.accounts = make(map[string]*Account)
	e.directory = keyDir

	e.refreshAccounts()

	return nil
}

// Disconnect disconnects from this wallet
func (e *Wallet) Disconnect() error {
	if e.ks == nil {
		return errors.New("keystore not initialized properly")
	}
	e.mu.Lock()
	defer e.mu.Unlock()

	e.ks = nil
	e.accounts = make(map[string]*Account)
	e.directory = ""
	return nil
}

// Status returns the state of this wallet
func (e *Wallet) Status() (string, error) {
	if e.ks == nil {
		return "not initialized", errors.New("keystore not initialized properly")
	}
	return "OK", nil
}

// Accounts returns all accounts held by this wallet
func (e *Wallet) Accounts() []perun.Account {
	e.refreshAccounts()

	e.mu.RLock()
	defer e.mu.RUnlock()

	v := make([]perun.Account, 0, len(e.accounts))
	for _, value := range e.accounts {
		v = append(v, value)
	}
	return v
}

// Contains checks whether this wallet holds this account
func (e *Wallet) Contains(a perun.Account) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if a == nil {
		return false
	}

	// check cache first
	if _, exists := e.accounts[a.Address().String()]; exists {
		return true
	}

	// if not found, query the keystore
	if acc, ok := a.(*Account); ok {
		found := e.ks.HasAddress(acc.address.Address)
		// add to the cache
		if found {
			e.mu.Lock()
			e.accounts[a.Address().String()] = acc
			e.mu.Unlock()
		}
		return found
	}
	panic("account is not an ethereum account")
}

// Lock locks this wallet and all keys
func (e *Wallet) Lock() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.ks == nil {
		return errors.New("keystore not initialized properly")
	}

	for _, acc := range e.accounts {
		if err := acc.Lock(); err != nil {
			return errors.Wrap(err, "lock all accounts failed")
		}
	}
	return nil
}

// Helper implements the utility interface defined in the wallet package.
type Helper struct{}

// NewAddressFromString creates a new address from a string
func (w *Helper) NewAddressFromString(s string) (perun.Address, error) {
	addr, err := common.NewMixedcaseAddressFromString(s)
	if err != nil {
		zeroAddr := common.BytesToAddress(make([]byte, 20, 20))
		return &Address{zeroAddr}, err
	}
	return &Address{addr.Address()}, nil
}

// NewAddressFromBytes creates a new address from a byte array
func (w *Helper) NewAddressFromBytes(data []byte) (perun.Address, error) {
	if len(data) != 20 {
		errString := "could not create address from bytes of length: " + strconv.Itoa(len(data))
		return &Address{ZeroAddr}, errors.New(errString)
	}
	return &Address{common.BytesToAddress(data)}, nil
}

// VerifySignature verifies if a signature was made by this account
func (w *Helper) VerifySignature(msg, sig []byte, a perun.Address) (bool, error) {
	hash := crypto.Keccak256(msg)
	pk, err := crypto.SigToPub(hash, sig)
	if err != nil {
		return false, err
	}
	addr := crypto.PubkeyToAddress(*pk)
	return a.Equals(&Address{addr}), nil
}

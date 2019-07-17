// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

// Package wallet defines an etherum wallet.
// It can be used by the framework to interact with a file wallet.
// It uses an ethereum keystore internally which can be found at
// https://github.com/ethereum/go-ethereum/tree/master/accounts/keystore.
package wallet // import "perun.network/go-perun/backend/ethereum/wallet"

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

// Wallet represents an ethereum wallet.
// It uses the go-ethereum keystore to store keys.
// Accessing the wallet is threadsafe, however you should not create two wallets from the same key directory.
type Wallet struct {
	ks        *keystore.KeyStore
	directory string
	accounts  map[string]*Account
	mu        sync.RWMutex
}

// Path returns the path to this wallet.
func (w *Wallet) Path() string {
	return w.directory
}

// refreshAccounts refreshes which accounts are connected to this wallet.
func (w *Wallet) refreshAccounts() {
	if w.ks == nil {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()

	accounts := w.ks.Accounts()
	for _, tmp := range accounts {
		if _, exists := w.accounts[tmp.Address.String()]; !exists {
			w.accounts[tmp.Address.String()] = newAccountFromEth(w, &tmp)
		}
	}
}

// Connect connects to this wallet.
func (w *Wallet) Connect(keyDir, password string) error {
	if _, err := os.Stat(keyDir); os.IsNotExist(err) {
		return errors.New("key directory does not exist")
	}
	w.ks = keystore.NewKeyStore(keyDir, keystore.StandardScryptN, keystore.StandardScryptP)
	w.accounts = make(map[string]*Account)
	w.directory = keyDir

	w.refreshAccounts()

	return nil
}

// Disconnect disconnects from this wallet.
func (w *Wallet) Disconnect() error {
	if w.ks == nil {
		return errors.New("keystore not initialized properly")
	}
	w.mu.Lock()
	defer w.mu.Unlock()

	w.ks = nil
	w.accounts = make(map[string]*Account)
	w.directory = ""
	return nil
}

// Status returns the state of this wallet.
func (w *Wallet) Status() (string, error) {
	if w.ks == nil {
		return "not initialized", errors.New("keystore not initialized properly")
	}
	return "OK", nil
}

// Accounts returns all accounts held by this wallet.
func (w *Wallet) Accounts() []perun.Account {
	w.refreshAccounts()

	w.mu.RLock()
	defer w.mu.RUnlock()

	v := make([]perun.Account, 0, len(w.accounts))
	for _, value := range w.accounts {
		v = append(v, value)
	}
	return v
}

// Contains checks whether this wallet holds this account.
func (w *Wallet) Contains(a perun.Account) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if a == nil {
		return false
	}

	// check cache first
	if _, exists := w.accounts[a.Address().String()]; exists {
		return true
	}

	// if not found, query the keystore
	if acc, ok := a.(*Account); ok {
		found := w.ks.HasAddress(acc.address.Address)
		// add to the cache
		if found {
			w.mu.Lock()
			w.accounts[a.Address().String()] = acc
			w.mu.Unlock()
		}
		return found
	}
	panic("account is not an ethereum account")
}

// Lock locks this wallet and all keys.
func (w *Wallet) Lock() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.ks == nil {
		return errors.New("keystore not initialized properly")
	}

	for _, acc := range w.accounts {
		if err := acc.Lock(); err != nil {
			return errors.Wrap(err, "lock all accounts failed")
		}
	}
	return nil
}

// Backend implements the utility interface defined in the wallet package.
type Backend struct{}

// NewAddressFromString creates a new address from a string.
func (h *Backend) NewAddressFromString(s string) (perun.Address, error) {
	addr, err := common.NewMixedcaseAddressFromString(s)
	if err != nil {
		zeroAddr := common.BytesToAddress(make([]byte, 20, 20))
		return &Address{zeroAddr}, err
	}
	return &Address{addr.Address()}, nil
}

// NewAddressFromBytes creates a new address from a byte array.
func (h *Backend) NewAddressFromBytes(data []byte) (perun.Address, error) {
	if len(data) != 20 {
		errString := "could not create address from bytes of length: " + strconv.Itoa(len(data))
		return &Address{ZeroAddr}, errors.New(errString)
	}
	return &Address{common.BytesToAddress(data)}, nil
}

// VerifySignature verifies if a signature was made by this account.
func (h *Backend) VerifySignature(msg, sig []byte, a perun.Address) (bool, error) {
	hash := crypto.Keccak256(msg)
	pk, err := crypto.SigToPub(hash, sig)
	if err != nil {
		return false, err
	}
	addr := crypto.PubkeyToAddress(*pk)
	return a.Equals(&Address{addr}), nil
}

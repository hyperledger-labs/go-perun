// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package wallet

import (
	"errors"
	"os"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	perun "github.com/perun-network/go-perun/wallet"
)

// Wallet represents an ethereum wallet
// It uses the go-ethereum keystore to store keys.
// Accessing the wallet is threadsafe, however you should not create two wallets from the same key directory
type Wallet struct {
	ks        *keystore.KeyStore
	directory string
	accounts  map[string]perun.Account
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
			e.accounts[tmp.Address.String()] = fromAccount(e, &tmp)
		}
	}
}

// Connect connects to this wallet
func (e *Wallet) Connect(keyDir, password string) error {
	if _, err := os.Stat(keyDir); os.IsNotExist(err) {
		return errors.New("Keyfile does not exist")
	}
	e.ks = keystore.NewKeyStore(keyDir, keystore.StandardScryptN, keystore.StandardScryptP)
	e.accounts = make(map[string]perun.Account)
	e.directory = keyDir

	e.refreshAccounts()

	return nil
}

// Disconnect disconnects from this wallet
func (e *Wallet) Disconnect() error {
	if e.ks == nil {
		return errors.New("Keystore not initialized properly")
	}
	e.mu.Lock()
	defer e.mu.Unlock()

	e.ks = nil
	e.accounts = make(map[string]perun.Account)
	e.directory = ""
	return nil
}

// Status returns the state of this wallet
func (e *Wallet) Status() (string, error) {
	if e.ks == nil {
		return "ERROR", errors.New("Keystore not initialized properly")
	}
	return "OK", nil
}

// Accounts returns all accounts held by this wallet
func (e *Wallet) Accounts() []perun.Account {
	e.refreshAccounts()

	e.mu.Lock()
	defer e.mu.Unlock()

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
	acc, ok := a.(*Account)
	if ok {
		return e.ks.HasAddress(acc.address.Address)
	}
	return false
}

// Lock locks this wallet and all keys
func (e *Wallet) Lock() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.ks == nil {
		return errors.New("Keystore not initialized properly")
	}

	for _, acc := range e.accounts {
		err := acc.Lock()
		if err != nil {
			return err
		}
	}
	return nil
}

// Helper is a helper struct
type Helper struct{}

// NewAddressFromString creates a new address from a string
func (w *Helper) NewAddressFromString(s string) (perun.Address, error) {
	tmp, err := common.NewMixedcaseAddressFromString(s)
	if err != nil {
		zeroAddr := common.BytesToAddress(make([]byte, 20, 20))
		return &Address{zeroAddr}, err
	}
	return &Address{tmp.Address()}, nil
}

// NewAddressFromBytes creates a new address from a byte array
func (w *Helper) NewAddressFromBytes(data []byte) (perun.Address, error) {
	if len(data) != 20 {
		zeroAddr := common.BytesToAddress(make([]byte, 20, 20))
		errString := "Could not create Address from bytes: " + string(len(data))
		return &Address{zeroAddr}, errors.New(errString)
	}
	return &Address{common.BytesToAddress(data)}, nil
}

// VerifySignature verifies if a signature was made by this account
func (w *Helper) VerifySignature(msg, sign []byte, a perun.Address) (bool, error) {
	hash := crypto.Keccak256(msg)
	pk, err := crypto.SigToPub(hash, sign)
	if err != nil {
		return false, err
	}
	addr := crypto.PubkeyToAddress(*pk)
	return a.Equals(&Address{addr}), nil
}

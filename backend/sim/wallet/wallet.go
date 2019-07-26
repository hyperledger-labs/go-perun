// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

// Package wallet provides a simulated wallet.
// The simulated wallet can be used for internal testing.
// DO NOT use this simulated wallet in production.
package wallet // import "perun.network/go-perun/backend/sim/wallet"

import (
	"sync"

	"github.com/pkg/errors"

	perun "perun.network/go-perun/wallet"
)

// Wallet represents a simulated wallet.
type Wallet struct {
	directory string
	account   Account
	mu        sync.RWMutex
	connected bool
}

// Path returns the path to this wallet.
func (w *Wallet) Path() string {
	return ""
}

// Connect connects to this wallet.
func (w *Wallet) Connect(keyDir, password string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.account = NewAccount()
	w.connected = true
	return nil
}

// Disconnect disconnects from this wallet.
func (w *Wallet) Disconnect() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.connected {
		return errors.New("Double disconnect")
	}
	w.connected = false
	return nil
}

// Status returns the state of this wallet.
func (w *Wallet) Status() (string, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if !w.connected {
		return "", errors.New("Not connected")
	}
	return "OK", nil
}

// Accounts returns all accounts held by this wallet.
func (w *Wallet) Accounts() []perun.Account {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if !w.connected {
		return []perun.Account{}
	}
	return []perun.Account{&w.account}
}

// Contains checks whether this wallet holds this account.
func (w *Wallet) Contains(a perun.Account) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()

	acc, ok := a.(*Account)
	if !ok {
		return false
	}
	return w.account.Address().Equals(acc.Address())
}

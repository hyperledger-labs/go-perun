// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package wallet // import "perun.network/go-perun/backend/sim/wallet"

import (
	"sync"

	"github.com/pkg/errors"

	"perun.network/go-perun/log"
	"perun.network/go-perun/wallet"
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
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.directory
}

// Connect connects to this wallet.
func (w *Wallet) Connect(keyDir, password string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

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
func (w *Wallet) Accounts() []wallet.Account {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if !w.connected {
		return []wallet.Account{}
	}

	return []wallet.Account{&w.account}
}

// Contains checks whether this wallet holds this account.
func (w *Wallet) Contains(a wallet.Account) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if !w.connected || a == nil {
		return false
	}

	acc, ok := a.(*Account)
	if !ok {
		log.Panic("Wrong account type passed to wallet.Contains")
	}

	return w.account.Address().Equals(acc.Address())
}

// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package wallet

import (
	"math/rand"
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"

	"perun.network/go-perun/wallet"
)

var _ wallet.Wallet = (*Wallet)(nil)

// NewWallet creates a new empty wallet.
func NewWallet() *Wallet {
	return &Wallet{accs: make(map[[64]byte]*Account)}
}

// NewRestoredWallet creates a wallet with a list of preexisting accounts which
// are initially locked. This simulates a wallet that has just been restored
// from persistent storage, and Unlock() has to be called to make accounts
// usable.
func NewRestoredWallet(accounts ...*Account) *Wallet {
	w := NewWallet()
	for _, acc := range accounts {
		acc.locked.Set()
		w.AddAccount(acc)
	}

	return w
}

// Wallet is a collection of accounts. Query accounts using Unlock, track their
// usage using IncrementUsage and DecrementUsage, and lock them using LockAll.
// Create new accounts using NewRandomAccount, and add existing accounts using
// AddAccount. Check whether the wallet owns a particular account via
// HasAccount.
type Wallet struct {
	accMutex sync.RWMutex
	accs     map[[64]byte]*Account
}

// Unlock retrieves the account belonging to the supplied address, and unlocks
// it. If the address does not have a corresponding account in the wallet,
// returns an error.
func (w *Wallet) Unlock(a wallet.Address) (wallet.Account, error) {
	w.accMutex.RLock()
	defer w.accMutex.RUnlock()

	acc, ok := w.accs[a.(*Address).ByteArray()]
	if !ok {
		return nil, errors.New("unlock unknown address")
	}

	acc.locked.Unset()
	return acc, nil
}

// LockAll locks all of a wallet's accounts.
func (w *Wallet) LockAll() {
	w.accMutex.RLock()
	defer w.accMutex.RUnlock()

	for _, acc := range w.accs {
		acc.locked.Set()
	}
}

// IncrementUsage increases an account's usage count, which is used for
// resource management. Panics if the wallet does not have an account that
// corresponds to the supplied address.
func (w *Wallet) IncrementUsage(a wallet.Address) {
	w.accMutex.RLock()
	defer w.accMutex.RUnlock()

	acc, ok := w.accs[a.(*Address).ByteArray()]
	if !ok {
		panic("invalid address")
	}

	atomic.AddInt32(&acc.references, 1)
}

// DecrementUsage decreases an account's usage count, and if it reaches 0,
// locks and deletes the account from the wallet. Panics if the call is not
// matched to another preceding IncrementUsage call or if the supplied address
// does not correspond to any of the wallet's accounts.
func (w *Wallet) DecrementUsage(a wallet.Address) {
	bytes := a.(*Address).ByteArray()

	w.accMutex.Lock()
	defer w.accMutex.Unlock()

	acc, ok := w.accs[bytes]
	if !ok {
		panic("invalid address")
	}

	newCount := atomic.AddInt32(&acc.references, -1)
	if newCount < 0 {
		panic("unmatched DecrementUsage call")
	}

	if newCount == 0 {
		acc.locked.Set()
		delete(w.accs, bytes)
	}
}

// UsageCount retrieves an account's usage count (controlled via IncrementUsage
// and DecrementUsage). Panics if the supplied address does not correspond to
// any of the wallet's accounts.
func (w *Wallet) UsageCount(a wallet.Address) int {
	bytes := a.(*Address).ByteArray()

	w.accMutex.RLock()
	defer w.accMutex.RUnlock()

	acc, ok := w.accs[bytes]
	if !ok {
		panic("invalid address")
	}

	return int(atomic.LoadInt32(&acc.references))
}

// NewRandomAccount creates and a new random account from the provided
// randomness stream. The account is automatically added to the wallet. Returns
// the generated account. The returned account is already unlocked.
func (w *Wallet) NewRandomAccount(rng *rand.Rand) wallet.Account {
	acc := NewRandomAccount(rng)
	w.AddAccount(acc)
	return acc
}

// AddAccount registers an externally generated account to the wallet. If the
// account was already registered beforehand, an error is returned. Does not
// lock or unlock the account.
func (w *Wallet) AddAccount(acc *Account) error {
	bytes := acc.Address().(*Address).ByteArray()

	w.accMutex.Lock()
	defer w.accMutex.Unlock()

	if _, ok := w.accs[bytes]; ok {
		return errors.New("duplicate insertion")
	}
	w.accs[bytes] = acc

	return nil
}

// HasAccount checks whether a Wallet has an account. This is only useful for
// easier testing.
func (w *Wallet) HasAccount(acc *Account) bool {
	w.accMutex.RLock()
	defer w.accMutex.RUnlock()

	_, ok := w.accs[acc.Address().(*Address).ByteArray()]
	return ok
}

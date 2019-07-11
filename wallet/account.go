// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package wallet

// Account represents a single account.
type Account interface {
	// Address used by this account.
	Address() Address

	// Unlocks this account with the given passphrase.
	// Returns an error if unlocking failed.
	// Unlocking an unlocked account should not fail.
	Unlock(password string) error

	// Returns a bool indicating whether this account is currently locked.
	IsLocked() bool

	// Locks this account.
	// Locking an already locked account should not fail.
	Lock() error

	// SignData requests a signature from this account.
	// It returns the signature or an error.
	SignData(data []byte) ([]byte, error)

	// SignDataWithPW requests a signature from this account.
	// It returns the signature or an error.
	// If the account is locked, it will unlock the account, sign the data and lock the account again.
	SignDataWithPW(password string, data []byte) ([]byte, error)
}

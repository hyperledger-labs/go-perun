// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

// Package wallet defines an abstraction to wallet providers.
// It provides an interface to connect other packages to a wallet provider.
// Wallet providers can be hardware, software remote or local wallets.
package wallet // import "perun.network/go-perun/wallet"

// Wallet represents a single or multiple accounts on a hardware or software wallet.
type Wallet interface {
	// Path returns an identifier under which this wallet is located.
	// Should return an empty string if the wallet was not properly initialized.
	Path() string

	// Connect establishes a connection to a wallet.
	// It should not decrypt the keys.
	// Returns an error if a connection cannot be established.
	Connect(path, password string) error

	// Disconnect closes a connection to a wallet and locks all accounts.
	// It returns an error if no connection is currently established to the wallet.
	Disconnect() error

	// Status returns the current status of the wallet.
	// Returns an error if the wallet is in a non-usable state (e.g. if no connection is established).
	Status() (string, error)

	// Accounts returns all accounts associated with this wallet.
	// Should return an empty byteslice if no accounts are found.
	Accounts() []Account

	// Contains checks whether this wallet contains this account.
	Contains(a Account) bool
}

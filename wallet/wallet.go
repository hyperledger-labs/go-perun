// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package wallet

// Wallet is a collection of Accounts, i.e., secret keys. The interface defines
// a method Unlock, which the framework calls to get an Account for an Address.
// The other methods may, but don't need to, be implemented to gain more
// efficient resource handling by the Wallet implementation.
type Wallet interface {
	// Unlock requests an unlocked Account for an Address from the Wallet.
	// The returned account must be able to sign messages at least until
	// LockAll has been called, or a matching count of IncrementUsage and
	// DecrementUsage calls on the account's address has been made. Unlock may
	// be called multiple times for the same Address by the Perun SDK.
	Unlock(Address) (Account, error)

	// LockAll is called by the framework when a Client shuts down. This should
	// release all temporary resources held by the wallet, and accesses to
	// accounts after this call are no longer expected to succeed by the Perun
	// SDK. Implementing this function with any behavior is not essential.
	LockAll()

	// IncrementUsage is called whenever a new channel is created or restored.
	// The address passed to the function belongs to the Account the Client is
	// using to participate in the channel. Implementing this function with any
	// behavior is not essential.
	IncrementUsage(Address)

	// DecrementUsage is called whenever a channel is settled. The address
	// passed to the function belongs to the Account the Client is using to
	// participate in the channel. It is guaranteed by the Perun SDK that when
	// an account had the same number of DecrementUsage calls as prior
	// IncrementUsage calls made to it, it can be safely deleted permanently by
	// the wallet implementation. In that event, the affected account does not
	// have to be able to sign messages anymore. Implementing this function with
	// any behavior is not essential.
	DecrementUsage(Address)
}

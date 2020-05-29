// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package wallet

import (
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/pkg/errors"
	"perun.network/go-perun/wallet"
)

// Account represents an ethereum account.
type Account struct {
	Account accounts.Account
	wallet  *Wallet
}

// Address returns the ethereum address of this account.
func (a *Account) Address() wallet.Address {
	return (*Address)(&a.Account.Address)
}

// SignData is used to sign data with this account.
func (a *Account) SignData(data []byte) ([]byte, error) {
	hash := prefixedHash(data)
	sig, err := a.wallet.Ks.SignHash(a.Account, hash)
	if err != nil {
		return nil, errors.Wrap(err, "SignHash")
	}
	sig[64] += 27
	return sig, nil
}

// NewAccountFromEth creates a new perun account from a given ethereum account.
func NewAccountFromEth(wallet *Wallet, account *accounts.Account) *Account {
	return &Account{
		Account: *account,
		wallet:  wallet,
	}
}

// Copyright 2019 - See NOTICE file for copyright holders.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

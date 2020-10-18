// Copyright 2020 - See NOTICE file for copyright holders.
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

package hd

import (
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"

	ethwallet "perun.network/go-perun/backend/ethereum/wallet"
	"perun.network/go-perun/wallet"
)

// Account represents an account held in the HD wallet.
type Account struct {
	wallet  accounts.Wallet
	Account accounts.Account
}

// Address returns the address of this account.
func (a *Account) Address() wallet.Address {
	return ethwallet.AsWalletAddr(a.Account.Address)
}

// SignData is used to sign data with this account.
func (a *Account) SignData(data []byte) ([]byte, error) {
	hash := crypto.Keccak256(data)
	sig, err := a.wallet.SignText(a.Account, hash)
	if err != nil {
		return nil, errors.Wrap(err, "SignText")
	}
	sig[64] += 27
	return sig, nil
}

// NewAccountFromEth creates a new perun account from a given ethereum account.
func NewAccountFromEth(wallet accounts.Wallet, account accounts.Account) *Account {
	return &Account{
		wallet:  wallet,
		Account: account,
	}
}

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

package keystore

import (
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
)

// Transactor can be used to make TransactOpts for accounts stored in a keystore.
type Transactor struct {
	Ks     *keystore.KeyStore
	Signer types.Signer
}

// NewTransactor returns a TransactOpts for the given account. It errors if the account is
// not contained in the keystore used for initializing transactOpts backend.
func (t *Transactor) NewTransactor(account accounts.Account) (*bind.TransactOpts, error) {
	if !t.Ks.HasAddress(account.Address) {
		return nil, errors.New("the wallet does not contain the keys for the given account")
	}
	return &bind.TransactOpts{
		From: account.Address,
		Signer: func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
			keystore, signer := t.Ks, t.Signer
			if address != account.Address {
				return nil, bind.ErrNotAuthorized
			}
			signature, err := keystore.SignHash(account, signer.Hash(tx).Bytes())
			if err != nil {
				return nil, err
			}
			return tx.WithSignature(signer, signature)
		},
	}, nil
}

// NewTransactor returns a backend that can make TransactOpts for accounts contained in the given keystore.
func NewTransactor(w Wallet, s types.Signer) *Transactor {
	return &Transactor{Ks: w.Ks, Signer: s}
}

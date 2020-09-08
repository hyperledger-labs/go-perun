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
	"github.com/pkg/errors"
)

// Transactor can be used to make TransactOpts for accounts stored in a keystore.
type Transactor struct {
	ks *keystore.KeyStore
}

// NewTransactor returns a TransactOpts for the given account. It errors if the account is
// not contained in the keystore used for initializing transactOpts backend.
func (t *Transactor) NewTransactor(account accounts.Account) (*bind.TransactOpts, error) {
	tr, err := bind.NewKeyStoreTransactor(t.ks, account)
	return tr, errors.WithStack(err)
}

// NewTransactor returns a backend that can make TransactOpts for accounts contained in the given keystore.
func NewTransactor(w Wallet) *Transactor {
	return &Transactor{ks: w.Ks}
}

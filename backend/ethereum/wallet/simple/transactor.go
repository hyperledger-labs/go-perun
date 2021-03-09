// Copyright 2021 - See NOTICE file for copyright holders.
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

package simple

import (
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"

	ethwallet "perun.network/go-perun/backend/ethereum/wallet"
)

// Transactor can be used to make TransactOpts for accounts stored in a wallet.
type Transactor struct {
	*Wallet
	types.Signer
}

// NewTransactor returns a TransactOpts for the given account. It errors if the
// account is not contained in the wallet of the transactor factory.
func (t *Transactor) NewTransactor(account accounts.Account) (*bind.TransactOpts, error) {
	walletAcc, err := t.Wallet.Unlock(ethwallet.AsWalletAddr(account.Address))
	if err != nil {
		return nil, err
	}
	acc := walletAcc.(*Account)

	return &bind.TransactOpts{
		From: account.Address,
		Signer: func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
			if address != account.Address {
				return nil, errors.New("not authorized to sign this account")
			}

			signature, err := acc.SignHash(t.Signer.Hash(tx).Bytes())
			if err != nil {
				return nil, err
			}
			return tx.WithSignature(t.Signer, signature)
		},
	}, nil
}

// NewTransactor returns a Transactor that can make TransactOpts for
// accounts contained in the given simple wallet.
func NewTransactor(w *Wallet, signer types.Signer) *Transactor {
	return &Transactor{Wallet: w, Signer: signer}
}

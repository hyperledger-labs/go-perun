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
	"math/big"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
)

// Transactor can be used to make TransactOpts for accounts stored in a HD wallet.
type Transactor struct {
	wallet accounts.Wallet
}

// NewTransactor returns a TransactOpts for the given account. It errors if the account is
// not contained in the wallet used for initializing transactor backend.
func (t *Transactor) NewTransactor(account accounts.Account) (*bind.TransactOpts, error) {
	if !t.wallet.Contains(account) {
		return nil, errors.New("account not found in wallet")
	}
	return &bind.TransactOpts{
		From: account.Address,
		Signer: func(signer types.Signer, address common.Address, tx *types.Transaction) (*types.Transaction, error) {
			if address != account.Address {
				return nil, errors.New("not authorized to sign this account")
			}
			// Last parameter (chainID) is only relevant when making EIP155 compliant signatures.
			// Since we use only non EIP155 signatures, set this to zero value.
			// For more details, see here (https://github.com/ethereum/EIPs/blob/master/EIPS/eip-155.md).
			return t.wallet.SignTx(account, tx, big.NewInt(0))
		},
	}, nil
}

// NewTransactor returns a backend that can make TransactOpts for accounts
// contained in the given ethereum wallet.
func NewTransactor(w accounts.Wallet) *Transactor {
	return &Transactor{wallet: w}
}

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
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
)

// Transactor can be used to make TransactOpts for accounts stored in a HD wallet.
type Transactor struct {
	Wallet accounts.Wallet
	Signer types.Signer
}

type hashSigner interface {
	SignHash(account accounts.Account, hash []byte) ([]byte, error)
}

// NewTransactor returns a TransactOpts for the given account. It errors if the account is
// not contained in the wallet used for initializing transactor backend.
func (t *Transactor) NewTransactor(account accounts.Account) (*bind.TransactOpts, error) {
	if !t.Wallet.Contains(account) {
		return nil, errors.New("account not found in wallet")
	}
	return &bind.TransactOpts{
		From: account.Address,
		Signer: func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
			if address != account.Address {
				return nil, errors.New("not authorized to sign this account")
			}

			hs, ok := t.Wallet.(hashSigner)
			if !ok {
				// signer.Hash returns the hash of the tx according to the chain
				// configuration but the accounts.Wallet interface only contains methods
				// SignData and SignText, which also do hashing. So there's no way to
				// properly sign a transaction with a Wallet that doesn't have the
				// SignHash method, if the tx needs to be signed according to EIP155
				// rules. So in this case, use the Wallet's SignTx method.
				return t.Wallet.SignTx(account, tx, tx.ChainId())
			}

			signature, err := hs.SignHash(account, t.Signer.Hash(tx).Bytes())
			if err != nil {
				return nil, err
			}
			return tx.WithSignature(t.Signer, signature)
		},
	}, nil
}

// NewTransactor returns a backend that can make TransactOpts for accounts
// contained in the given ethereum wallet.
func NewTransactor(w accounts.Wallet, signer types.Signer) *Transactor {
	return &Transactor{Wallet: w, Signer: signer}
}

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
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"

	ethwallet "perun.network/go-perun/backend/ethereum/wallet"
	"perun.network/go-perun/wallet"
)

// Account represents an account held in the simple wallet.
type Account struct {
	accounts.Account
	key *ecdsa.PrivateKey
}

// Address returns the Ethereum address of this account.
func (a *Account) Address() wallet.Address {
	return ethwallet.AsWalletAddr(a.Account.Address)
}

// SignData is used to sign data with this account.
func (a *Account) SignData(data []byte) ([]byte, error) {
	hash := ethwallet.PrefixedHash(data)
	sig, err := a.SignHash(hash)
	if err != nil {
		return nil, errors.Wrap(err, "SignHash")
	}
	sig[64] += 27
	return sig, nil
}

// SignHash is used to sign an already prefixed hash with this account.
func (a *Account) SignHash(hash []byte) ([]byte, error) {
	return crypto.Sign(hash, a.key)
}

// createAccount creates an account using the given private key.
func createAccount(k *ecdsa.PrivateKey) *Account {
	return &Account{
		Account: accounts.Account{Address: crypto.PubkeyToAddress(k.PublicKey)},
		key:     k,
	}
}

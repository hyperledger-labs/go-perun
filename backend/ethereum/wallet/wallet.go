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
	"crypto/ecdsa"
	"math/rand"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/pkg/errors"

	"perun.network/go-perun/log"
	"perun.network/go-perun/wallet"
)

var _ wallet.Wallet = (*Wallet)(nil)

// Wallet represents an ethereum wallet.
// It uses the go-ethereum keystore to store keys.
// Accessing the wallet is threadsafe, however you should not create two wallets from the same key directory.
type Wallet struct {
	Ks *keystore.KeyStore
	pw string
}

// NewWallet creates a new Wallet from a keystore and password.
func NewWallet(ks *keystore.KeyStore, pw string) (*Wallet, error) {
	if accs := ks.Accounts(); len(accs) != 0 {
		// Check that the accounts in the wallet can be unlocked with the
		// password (assuming that all accounts use the same password).
		if err := ks.Update(accs[0], pw, pw); err != nil {
			return nil, errors.Wrap(err, "invalid password")
		}
	}

	// Check that the password

	return &Wallet{
		Ks: ks,
		pw: pw,
	}, nil
}

// Contains checks whether this wallet holds this account.
func (w *Wallet) Contains(a wallet.Address) bool {
	if a == nil {
		return false
	}

	return w.Ks.HasAddress(AsEthAddr(a))
}

// NewAccount creates a new random account which is already unlocked.
func (w *Wallet) NewAccount() *Account {
	acc, err := w.Ks.NewAccount(w.pw)
	if err != nil || w.Ks.Unlock(acc, w.pw) != nil {
		panic("failed to create random account")
	}
	log.Debugf("Created new account %v", acc.Address)
	return NewAccountFromEth(w, &acc)
}

// NewRandomAccount creates a new pseudorandom account using the provided
// randomness. The returned account is already unlocked.
func (w *Wallet) NewRandomAccount(rnd *rand.Rand) wallet.Account {
	privateKey, err := ecdsa.GenerateKey(secp256k1.S256(), rnd)
	if err != nil {
		log.Panicf("Creating account: %v", err)
	}
	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	if acc, err := w.Ks.Find(accounts.Account{Address: address}); err == nil {
		w.Unlock((*Address)(&address))
		return NewAccountFromEth(w, &acc)
	}

	ethAcc, err := w.Ks.ImportECDSA(privateKey, w.pw)
	if err != nil {
		log.Panicf("Storing private key: %v", err)
	}
	w.Unlock((*Address)(&address))
	log.Debugf("Created new random account %v", ethAcc.Address)

	return NewAccountFromEth(w, &ethAcc)
}

// Unlock retrieves the account with the given address and unlocks it. If there
// is no matching account or unlocking fails, returns an error.
func (w *Wallet) Unlock(addr wallet.Address) (wallet.Account, error) {
	log.Debugf("Unlocking account %v", addr)
	// Hack: create ethereum account from ethereum address.
	acc := accounts.Account{Address: common.Address(*addr.(*Address))}

	if err := w.Ks.Unlock(acc, w.pw); err != nil {
		return nil, errors.Wrapf(err, "unlocking %v", addr)
	}
	return &Account{
		Account: acc,
		wallet:  w,
	}, nil
}

// LockAll locks all the wallet's keys and releases all its resources. It is no
// longer usable after this call.
func (w *Wallet) LockAll() {
	log.Debug("Locking wallet")
	if w.Ks == nil {
		return
	}

	for _, acc := range w.Ks.Accounts() {
		if err := w.Ks.Lock(acc.Address); err != nil {
			log.WithError(err).Errorf("failed to lock account %v", acc.Address)
		}
	}

	w.Ks = nil
}

// IncrementUsage currently does nothing. In the future, it will track the usage of keys.
func (w *Wallet) IncrementUsage(a wallet.Address) {
	log.Trace("IncrementUsage ", a)
}

// DecrementUsage currently does nothing. In the future, it will track the usage of keys and release unused keys.
func (w *Wallet) DecrementUsage(a wallet.Address) {
	log.Trace("DecrementUsage ", a)
}

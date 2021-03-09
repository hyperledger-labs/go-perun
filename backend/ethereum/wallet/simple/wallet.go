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
	"math/rand"
	"sync"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/pkg/errors"

	ethwallet "perun.network/go-perun/backend/ethereum/wallet"
	"perun.network/go-perun/log"
	"perun.network/go-perun/wallet"
)

var _ wallet.Wallet = (*Wallet)(nil)

// Wallet is a simple wallet.Wallet implementation holding a map of all included accounts.
type Wallet struct {
	accounts map[common.Address]*Account
	mutex    sync.RWMutex
}

// NewWallet creates a new Wallet with accounts corresponding to the privateKeys.
func NewWallet(privateKeys ...*ecdsa.PrivateKey) *Wallet {
	accs := make(map[common.Address]*Account)
	for _, key := range privateKeys {
		addr := crypto.PubkeyToAddress(key.PublicKey)
		accs[addr] = createAccount(key)
	}
	return &Wallet{accounts: accs}
}

// Contains checks whether this wallet contains the account corresponding to the given address.
func (w *Wallet) Contains(addr common.Address) bool {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.contains(addr)
}

// contains checks whether this wallet contains the account corresponding to the
// given address. Assumes that the mutex is locked.
func (w *Wallet) contains(addr common.Address) bool {
	_, ok := w.accounts[addr]
	return ok
}

// NewRandomAccount creates a new pseudorandom account using the provided
// randomness. The returned account is already unlocked.
func (w *Wallet) NewRandomAccount(prng *rand.Rand) wallet.Account {
	privateKey, err := ecdsa.GenerateKey(secp256k1.S256(), prng)
	if err != nil {
		log.Panicf("Creating account: %v", err)
	}

	w.mutex.Lock()
	defer w.mutex.Unlock()

	addr := crypto.PubkeyToAddress(privateKey.PublicKey)
	if w.contains(addr) {
		log.Panicf("Randomly generated account already exists: %v", addr)
	}

	acc := Account{
		Account: accounts.Account{Address: addr},
		key:     privateKey,
	}
	w.accounts[addr] = &acc
	return w.accounts[addr]
}

// Unlock returns the account corresponding to the given address if the wallet
// contains this account.
func (w *Wallet) Unlock(address wallet.Address) (wallet.Account, error) {
	if _, ok := address.(*ethwallet.Address); !ok {
		return nil, errors.New("address must be ethwallet.Address")
	}

	w.mutex.RLock()
	defer w.mutex.RUnlock()
	if acc, ok := w.accounts[ethwallet.AsEthAddr(address)]; ok {
		return acc, nil
	}
	return nil, errors.New("account not found in wallet")
}

// LockAll is called by the framework when a Client shuts down.
func (w *Wallet) LockAll() {}

// IncrementUsage is called whenever a new channel is created or restored.
func (w *Wallet) IncrementUsage(address wallet.Address) {}

// DecrementUsage is called whenever a channel is settled.
func (w *Wallet) DecrementUsage(address wallet.Address) {}

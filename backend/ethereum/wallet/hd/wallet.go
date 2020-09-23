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
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"

	ethwallet "perun.network/go-perun/backend/ethereum/wallet"
	"perun.network/go-perun/wallet"
)

var _ wallet.Wallet = (*Wallet)(nil)

// DefaultRootDerivationPath represent the default root derivation path for ethereum accounts as per BIP32.
var DefaultRootDerivationPath = accounts.DefaultRootDerivationPath

// Wallet is a wallet.Wallet implementation for using HD wallets. It supports any
// implementation of the HD wallet interface (accounts.Wallet) defined in go-ethereum project.
type Wallet struct {
	numDerivedAccs     uint
	rootDerivationPath accounts.DerivationPath
	wallet             accounts.Wallet

	// mutex ensures numDerivedAccs can be safely modified in concurrent calls for adding/removing accounts.
	mutex sync.Mutex
}

// NewWallet returns a new perun wallet that uses the given HD wallet.
//
// Use the DefaultRootDerivationPath for accessing ethereum on-chain accounts.
// numUsedAccs should be the number of used accounts in the wallet.
//
// All of these accounts will be retreived from the wallet, unlocked and made accesible for making signatures.
func NewWallet(hdwallet accounts.Wallet, derivationPath string, numUsedAccs uint) (*Wallet, error) {
	if hdwallet == nil {
		return nil, errors.New("wallet must not be nil")
	}
	path, err := accounts.ParseDerivationPath(derivationPath)
	if err != nil {
		return nil, errors.Wrap(err, "parsing derivation path")
	}
	wallet := &Wallet{
		numDerivedAccs:     0,
		rootDerivationPath: path,
		wallet:             hdwallet,
	}
	return wallet, wallet.deriveAccounts(numUsedAccs)
}

func (w *Wallet) deriveAccounts(numUsedAccs uint) (err error) {
	for i := uint(0); i < numUsedAccs; i++ {
		if _, err = w.newAccount(); err != nil {
			break
		}
	}
	return errors.WithMessage(err, "deriving keys")
}

// Contains checks whether this wallet contains the account corresponding to the given address.
func (w *Wallet) Contains(addr common.Address) bool {
	return w.wallet.Contains(accounts.Account{Address: addr})
}

// NewAccount creates a new account which is unlocked and ready to use.
// It will be derived at the index numDerivedAccs (as the index starts from 0) and
// numDerivedAccs will be incremented.
func (w *Wallet) NewAccount() (*Account, error) {
	return w.newAccount()
}

func (w *Wallet) newAccount() (*Account, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	newAccountPath := fmt.Sprintf("%v%d", w.rootDerivationPath, w.numDerivedAccs)
	path, err := accounts.ParseDerivationPath(newAccountPath)
	if err != nil {
		return nil, errors.Wrap(err, "deriving path for new account")
	}
	acc, err := w.wallet.Derive(path, true)
	if err != nil {
		return nil, errors.Wrap(err, "deriving a new account")
	}
	w.numDerivedAccs++
	return &Account{
		wallet:  w.wallet,
		account: acc,
	}, nil
}

// Unlock checks if the wallet contains the account corresponding to the given address.
// There is no concept of unlocking in software only hd wallet.
func (w *Wallet) Unlock(addr wallet.Address) (wallet.Account, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	acc := accounts.Account{Address: ethwallet.AsEthAddr(addr)}
	if !w.wallet.Contains(acc) {
		return nil, errors.New("account not found in wallet")
	}
	return &Account{
		wallet:  w.wallet,
		account: acc,
	}, nil
}

// LockAll implements wallet.LockAll. It is noop.
func (w *Wallet) LockAll() {}

// IncrementUsage implements wallet.Wallet. It is a noop.
func (w *Wallet) IncrementUsage(a wallet.Address) {}

// DecrementUsage implements wallet.Wallet. It is a noop.
func (w *Wallet) DecrementUsage(a wallet.Address) {}

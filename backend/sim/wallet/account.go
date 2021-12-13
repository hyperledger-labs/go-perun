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
	"crypto/rand"
	"io"

	"github.com/pkg/errors"

	"perun.network/go-perun/log"
	"perun.network/go-perun/wallet"
	"polycry.pt/poly-go/sync/atomic"
)

// Account represents a mocked account.
type Account struct {
	privKey *ecdsa.PrivateKey

	locked     atomic.Bool
	references int32
}

const (
	// how many points a sig consists of.
	pointsPerSig = 2
	// how many bits are in a byte.
	bitsPerByte = 8
)

// NewRandomAccount generates a new account, reading randomness form the given
// rng. It is not saved to any wallet.
func NewRandomAccount(rng io.Reader) *Account {
	privateKey, err := ecdsa.GenerateKey(curve, rng)
	if err != nil {
		log.Panicf("Creation of account failed with error", err)
	}

	return &Account{
		privKey: privateKey,
	}
}

// Address returns the address of this account.
func (a *Account) Address() wallet.Address {
	return wallet.Address((*Address)(&a.privKey.PublicKey))
}

// SignData is used to sign data with this account. If the account is locked,
// returns an error instead of a signature.
func (a *Account) SignData(data []byte) (wallet.Sig, error) {
	if a.locked.IsSet() {
		return nil, errors.New("account locked")
	}

	// escda.Sign needs a digest as input
	// ref https://golang.org/pkg/crypto/ecdsa/#Sign
	r, s, err := ecdsa.Sign(rand.Reader, a.privKey, digest(data))
	return &Sig{r: r, s: s}, errors.Wrap(err, "account could not sign data")
}

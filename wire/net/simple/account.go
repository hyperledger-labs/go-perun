// Copyright 2022 - See NOTICE file for copyright holders.
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
	"crypto"
	crypto_rand "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"math/rand"

	"github.com/pkg/errors"
	"perun.network/go-perun/wire"
)

// Account is a wire account.
type Account struct {
	addr       wire.Address
	privateKey *rsa.PrivateKey
}

// Address returns the account's address.
func (acc *Account) Address() map[int]wire.Address {
	return map[int]wire.Address{0: acc.addr}
}

// Sign signs the given message with the account's private key.
func (acc *Account) Sign(msg []byte) ([]byte, error) {
	if acc.privateKey == nil {
		return nil, errors.New("private key is nil")
	}
	hashed := sha256.Sum256(msg)
	signature, err := rsa.SignPKCS1v15(crypto_rand.Reader, acc.privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return nil, err
	}
	return signature, nil
}

// NewRandomAccount generates a new random account.
func NewRandomAccount(rng *rand.Rand) *Account {
	keySize := 2048
	privateKey, err := rsa.GenerateKey(rng, keySize)
	if err != nil {
		panic(err)
	}

	address := NewRandomAddress(rng)
	address.PublicKey = &privateKey.PublicKey

	return &Account{
		addr:       address,
		privateKey: privateKey,
	}
}

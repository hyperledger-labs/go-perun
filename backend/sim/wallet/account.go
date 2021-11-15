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
	"math/big"

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
func (a *Account) SignData(data []byte) ([]byte, error) {
	if a.locked.IsSet() {
		return nil, errors.New("account locked")
	}

	// escda.Sign needs a digest as input
	// ref https://golang.org/pkg/crypto/ecdsa/#Sign
	r, s, err := ecdsa.Sign(rand.Reader, a.privKey, digest(data))

	if err != nil {
		return nil, errors.Wrap(err, "account could not sign data")
	}

	return serializeSignature(r, s)
}

// serializeSignature serializes a signature into a []byte or returns an error.
// The length of the []byte is dictated by the curves parameters and padded with 0 bytes if necessary.
func serializeSignature(r, s *big.Int) ([]byte, error) {
	pointSize := curve.Params().BitSize / bitsPerByte
	rBytes := append(make([]byte, pointSize-len(r.Bytes())), r.Bytes()...)
	sBytes := append(make([]byte, pointSize-len(s.Bytes())), s.Bytes()...)

	return append(rBytes, sBytes...), nil
}

// deserializeSignature deserializes a signature from a byteslice and returns `r` and `s`
// or an error.
func deserializeSignature(b []byte) (*big.Int, *big.Int, error) {
	pointSize := curve.Params().BitSize / bitsPerByte
	sigSize := pointsPerSig * pointSize
	if len(b) != sigSize {
		return nil, nil, errors.Errorf("expected %d bytes for a signature but got: %d", sigSize, len(b))
	}

	var r, s big.Int
	rBytes := b[0:pointSize]
	sBytes := b[pointSize:sigSize]
	r.SetBytes(rBytes)
	s.SetBytes(sBytes)

	return &r, &s, nil
}

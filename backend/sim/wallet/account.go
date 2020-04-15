// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package wallet // import "perun.network/go-perun/backend/sim/wallet"

import (
	"crypto/ecdsa"
	"crypto/rand"
	"io"
	"math/big"

	"github.com/pkg/errors"

	"perun.network/go-perun/log"
	"perun.network/go-perun/wallet"
)

// Account represents a mocked account.
type Account struct {
	privKey *ecdsa.PrivateKey
}

// NewRandomAccount creates a new account using the randomness
// provided by rng
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

// SignData is used to sign data with this account.
func (a *Account) SignData(data []byte) ([]byte, error) {
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
	pointSize := curve.Params().BitSize / 8
	rBytes := append(make([]byte, pointSize-len(r.Bytes())), r.Bytes()...)
	sBytes := append(make([]byte, pointSize-len(s.Bytes())), s.Bytes()...)

	return append(rBytes, sBytes...), nil
}

// deserializeSignature deserializes a signature from a byteslice and returns `r` and `s`
// or an error.
func deserializeSignature(b []byte) (*big.Int, *big.Int, error) {
	pointSize := curve.Params().BitSize / 8
	if len(b) != pointSize*2 {
		return nil, nil, errors.Errorf("expected %d bytes for a signature but got: %d", pointSize*2, len(b))
	}

	var r, s big.Int
	rBytes := b[0:pointSize]
	sBytes := b[pointSize : pointSize*2]
	r.SetBytes(rBytes)
	s.SetBytes(sBytes)

	return &r, &s, nil
}

// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package wallet

import (
	"crypto/ecdsa"
	"crypto/rand"
	"math/big"

	"github.com/pkg/errors"

	"perun.network/go-perun/log"
	perun "perun.network/go-perun/wallet"
)

// Account represents a mocked account.
type Account struct {
	address Address
	privKey *ecdsa.PrivateKey
}

// NewAccount creates a new account using the randomness
// provided by crypto/rand.
func NewAccount() Account {
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panicf("Creation of account failed with error", err)
	}

	return Account{
		address: pubKeyToAddress(&privateKey.PublicKey),
		privKey: privateKey,
	}
}

// Address returns the address of this account.
func (a Account) Address() perun.Address {
	return perun.Address(&a.address)
}

// SignData is used to sign data with this account.
func (a Account) SignData(data []byte) ([]byte, error) {
	r, s, err := ecdsa.Sign(rand.Reader, a.privKey, hash(data))

	if err != nil {
		return nil, errors.Wrap(err, "account could not sign data")
	}

	return serializeSignature(r, s), err
}

// serializeSignature serializes a signature given as two points on
// a curve into a byte slice.
// The serialized format is:
// len(r) : len(s) : r.Bytes() : s.Bytes()
// This function only works for curves where no point is > 2^256 - 1
func serializeSignature(r, s *big.Int) []byte {
	bytesR := r.Bytes()
	bytesS := s.Bytes()

	lens := []byte{byte(len(bytesR)), byte(len(bytesS))}
	lens = append(lens, bytesR...)
	return append(lens, bytesS...)
}

func deserializeSignature(b []byte) (r, s *big.Int, err error) {
	lenR := int(b[0])
	lenS := int(b[1])

	if lenR+lenS+2 != len(b) {
		return nil, nil, errors.New("Could not deserialize signature")
	}

	r = new(big.Int).SetBytes(b[2 : 2+lenR])
	s = new(big.Int).SetBytes(b[2+lenR:])
	return
}

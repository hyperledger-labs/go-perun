// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package wallet // import "perun.network/go-perun/backend/sim/wallet"

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/asn1"
	"io"
	"math/big"

	"github.com/pkg/errors"

	"perun.network/go-perun/log"
	"perun.network/go-perun/wallet"
)

// Account represents a mocked account.
type Account struct {
	address Address
	privKey *ecdsa.PrivateKey
}

// NewRandomAccount creates a new account using the randomness
// provided by rng
func NewRandomAccount(rng io.Reader) Account {
	privateKey, err := ecdsa.GenerateKey(curve, rng)

	if err != nil {
		log.Panicf("Creation of account failed with error", err)
	}

	return Account{
		address: Address(privateKey.PublicKey),
		privKey: privateKey,
	}
}

// Address returns the address of this account.
func (a Account) Address() wallet.Address {
	return wallet.Address(&a.address)
}

// SignData is used to sign data with this account.
func (a Account) SignData(data []byte) ([]byte, error) {
	// escda.Sign needs a digest as input
	// ref https://golang.org/pkg/crypto/ecdsa/#Sign
	r, s, err := ecdsa.Sign(rand.Reader, a.privKey, digest(data))

	if err != nil {
		return nil, errors.Wrap(err, "account could not sign data")
	}

	return serializeSignature(r, s)
}

type ecdsaSignature struct {
	R, S *big.Int
}

// serializeSignature serializes a r and s in the manner of golang
// by creating a ecdsaSignature and marshalling it with asn1.
// ref https://en.wikipedia.org/wiki/Abstract_Syntax_Notation_One
// ref https://golang.org/pkg/encoding/asn1/
func serializeSignature(r, s *big.Int) ([]byte, error) {
	data, err := asn1.Marshal(ecdsaSignature{r, s})

	if err != nil {
		return nil, errors.Wrap(err, "asn1.Marshall error")
	}

	return data, nil
}

// deserializeSignature deserializes a r and s in the manner of golang
// by creating a ecdsaSignature and unmarshalling it with asn1.
// ref https://en.wikipedia.org/wiki/Abstract_Syntax_Notation_One
// ref https://golang.org/pkg/encoding/asn1/
func deserializeSignature(b []byte) (r, s *big.Int, err error) {
	var sig ecdsaSignature
	rest, err := asn1.Unmarshal(b, &sig)

	if err != nil {
		return nil, nil, errors.Wrap(err, "asn1.Unmarshall error")
	}
	if len(rest) != 0 {
		return nil, nil, errors.New("asn1.Unmarshall error")
	}

	return sig.R, sig.S, nil
}

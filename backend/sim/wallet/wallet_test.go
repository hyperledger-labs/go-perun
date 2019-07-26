// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package wallet

import (
	"math/big"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	perun "perun.network/go-perun/wallet"
	"perun.network/go-perun/wallet/test"
)

// TestSignatureSerialize tests serializeSignature and deserializeSignature since
// a signature is only a []byte, we cant use io.serializable here
func TestSignatureSerialize(t *testing.T) {
	a := assert.New(t)
	// Constant seed for determinism
	rand.Seed(1337)

	for i := 0; i < 100; i++ {
		rBytes := make([]byte, AddressLength)
		sBytes := make([]byte, AddressLength)

		// These always return nil error
		rand.Read(rBytes)
		rand.Read(sBytes)

		r := new(big.Int).SetBytes(rBytes)
		s := new(big.Int).SetBytes(sBytes)

		sig := serializeSignature(r, s)
		R, S, err := deserializeSignature(sig)

		a.Nil(err, "Deserialization should not fail")
		a.Equal(r, R, "Serialized and deserialized r values should be equal")
		a.Equal(s, S, "Serialized and deserialized s values should be equal")
	}
}

func TestGenericTests(t *testing.T) {
	t.Run("Generic Signature Test", func(t *testing.T) {
		t.Parallel()
		test.GenericWalletTest(t, newSetup())
	})
	t.Run("Generic Signature Test", func(t *testing.T) {
		t.Parallel()
		test.GenericSignatureTest(t, newSetup())
	})
	t.Run("Generic Signature Test", func(t *testing.T) {
		t.Parallel()
		test.GenericAddressTest(t, newSetup())
	})
}

func newSetup() *test.Setup {
	account := NewAccount()
	initWallet := func(w perun.Wallet) error { return w.Connect("", "") }
	unlockedAccount := func() (perun.Account, error) { return &account, nil }

	secondAccount := NewAccount()

	return &test.Setup{
		Wallet:          new(Wallet),
		Backend:         new(Backend),
		UnlockedAccount: unlockedAccount,
		InitWallet:      initWallet,
		AddrString:      secondAccount.Address().String(),
	}
}

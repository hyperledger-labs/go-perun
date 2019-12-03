// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package test // import "perun.network/go-perun/wallet/test"

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/pkg/io/test"
	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wire"
)

// InitWallet initializes a wallet.
type InitWallet func(wallet.Wallet) error

// UnlockedAccount provides an unlocked account.
type UnlockedAccount func() (wallet.Account, error)

// Setup provides all objects needed for the generic tests
type Setup struct {
	//Wallet tests
	Wallet          wallet.Wallet   // wallet implementation, should be uninitialized
	UnlockedAccount UnlockedAccount // provides an account that is ready to sign
	InitWallet      InitWallet      // function that initializes a wallet.
	//Address tests
	AddressBytes []byte         // valid address, should not be in wallet
	Backend      wallet.Backend // backend implementation
	// Signature tests
	DataToSign []byte
}

// GenericWalletTest runs a test suite designed to test the general functionality of an implementation of wallet.
// This function should be called by every implementation of the wallet interface.
func GenericWalletTest(t *testing.T, s *Setup) {
	testUninitializedWallet(t, s)
	testInitializedWallet(t, s)
	testUninitializedWallet(t, s)
}

func testUninitializedWallet(t *testing.T, s *Setup) {
	assert.NotNil(t, s.Wallet, "Wallet should not be nil")
	assert.Equal(t, "", s.Wallet.Path(), "Expected path not to be initialized")

	_, err := s.Wallet.Status()
	assert.NotNil(t, err, "Expected error on not connected wallet")
	assert.NotNil(t, s.Wallet.Disconnect(), "Disconnect of not connected wallet should return an error")
	assert.NotNil(t, s.Wallet.Accounts(), "Expected empty byteslice")
	assert.Equal(t, 0, len(s.Wallet.Accounts()), "Expected empty byteslice")
	assert.False(t, s.Wallet.Contains(*new(wallet.Account)), "Uninitalized wallet should not contain account")
}

func testInitializedWallet(t *testing.T, s *Setup) {
	assert.Nil(t, s.InitWallet(s.Wallet), "Expected connect to succeed")

	_, err := s.Wallet.Status()
	assert.Nil(t, err, "Unlocked wallet should not produce errors")
	assert.NotNil(t, s.Wallet.Accounts(), "Expected accounts")
	assert.False(t, s.Wallet.Contains(*new(wallet.Account)), "Expected wallet not to contain an empty account")
	assert.Equal(t, 1, len(s.Wallet.Accounts()), "Expected one account")

	acc := s.Wallet.Accounts()[0]
	assert.True(t, s.Wallet.Contains(acc), "Expected wallet to contain account")

	assert.Nil(t, s.Wallet.Disconnect(), "Expected disconnect to succeed")
}

// GenericSignatureTest runs a test suite designed to test the general functionality of an account.
// This function should be called by every implementation of the wallet interface.
func GenericSignatureTest(t *testing.T, s *Setup) {
	acc, err := s.UnlockedAccount()
	assert.Nil(t, err)
	// Check unlocked account
	sign, err := acc.SignData(s.DataToSign)
	assert.Nil(t, err, "Sign with unlocked account should succeed")
	valid, err := s.Backend.VerifySignature(s.DataToSign, sign, acc.Address())
	assert.True(t, valid, "Verification should succeed")
	assert.Nil(t, err, "Verification should not produce error")

	addr, err := s.Backend.NewAddressFromBytes(s.AddressBytes)
	assert.Nil(t, err, "Byte deserialization of address should work")
	valid, err = s.Backend.VerifySignature(s.DataToSign, sign, addr)
	assert.False(t, valid, "Verification with wrong address should fail")
	assert.Nil(t, err, "Verification of valid signature should not produce error")

	tampered := make([]byte, len(sign))
	copy(tampered, sign)
	// Invalidate the signature and check for error
	tampered[0] = ^sign[0]
	valid, err = s.Backend.VerifySignature(s.DataToSign, tampered, acc.Address())
	if valid && err == nil {
		t.Error("Verification of invalid signature should produce error or return false")
	}
	// Truncate the signature and check for error
	tampered = sign[:len(sign)-1]
	valid, err = s.Backend.VerifySignature(s.DataToSign, tampered, acc.Address())
	if valid && err != nil {
		t.Error("Verification of invalid signature should produce error or return false")
	}
	// Expand the signature and check for error
	tampered = append(sign, 0)
	valid, err = s.Backend.VerifySignature(s.DataToSign, tampered, acc.Address())
	if valid && err != nil {
		t.Error("Verification of invalid signature should produce error or return false")
	}

	// Test DecodeSig
	sign, err = acc.SignData(s.DataToSign)
	require.NoError(t, err, "Sign with unlocked account should succeed")

	buff := new(bytes.Buffer)
	wire.Encode(buff, sign)
	sign2, err := s.Backend.DecodeSig(buff)
	assert.NoError(t, err, "Decoded signature should work")
	assert.Equal(t, sign, sign2, "Decoded signature should be equal to the original")

	// Test DecodeSig on short stream
	wire.Encode(buff, sign)
	shortBuff := bytes.NewBuffer(buff.Bytes()[:len(buff.Bytes())-1]) // remove one byte
	_, err = s.Backend.DecodeSig(shortBuff)
	assert.Error(t, err, "DecodeSig on short stream should error")
}

// GenericSignatureSizeTest tests that the size of the signatures produced by
// Account.Sign(…) does not vary between executions (tested with 2048 samples).
func GenericSignatureSizeTest(t *testing.T, s *Setup) {
	acc, err := s.UnlockedAccount()
	require.NoError(t, err)
	// get a signature
	sign, err := acc.SignData(s.DataToSign)
	require.NoError(t, err, "Sign with unlocked account should succeed")

	// Test that signatures have constant length
	l := len(sign)
	for i := 0; i < 8; i++ {
		t.Run("parallel signing", func(t *testing.T) {
			t.Parallel()
			for i := 0; i < 256; i++ {
				sign, err := acc.SignData(s.DataToSign)
				require.NoError(t, err, "Sign with unlocked account should succeed")
				require.Equal(t, l, len(sign), "Signatures should have constant length: %d vs %d", l, len(sign))
			}
		})
	}
}

// GenericAddressTest runs a test suite designed to test the general functionality of addresses.
// This function should be called by every implementation of the wallet interface.
func GenericAddressTest(t *testing.T, s *Setup) {
	addrLen := len(s.AddressBytes)
	null, err := s.Backend.NewAddressFromBytes(make([]byte, addrLen, addrLen))
	assert.Nil(t, err, "Byte deserialization of zero address should work")
	addr, err := s.Backend.NewAddressFromBytes(s.AddressBytes)
	assert.Nil(t, err, "Byte deserialization of address should work")

	addrString := addr.String()
	if addrString[:2] == "0x" {
		addrString = addrString[2:]
	}
	fullAddrString := hex.EncodeToString(s.AddressBytes)
	assert.Greater(t, len(addrString), 0)
	assert.Contains(t, fullAddrString, addrString)

	assert.Equal(t, s.AddressBytes, addr.Bytes(), "Expected equality of address bytes")
	assert.False(t, addr.Equals(null), "Expected inequality of zero, nonzero address")
	assert.True(t, null.Equals(null), "Expected equality of zero address to itself")

	t.Run("Generic Serializable Test", func(t *testing.T) {
		test.GenericSerializableTest(t, addr)
	})
}

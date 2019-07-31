// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package test // import "perun.network/go-perun/wallet/test"

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/wallet"
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
	AddrString string         // valid address, should not be in wallet
	Backend    wallet.Backend // backend implementation
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

	addr, err := s.Backend.NewAddressFromString(s.AddrString)
	assert.Nil(t, err, "Byte deserialization of Address should work")
	valid, err = s.Backend.VerifySignature(s.DataToSign, sign, addr)
	assert.False(t, valid, "Verification with wrong address should fail")
	assert.Nil(t, err, "Verification of valid signature should not produce error")

	sign[0] = ^sign[0] // invalidate signature
	valid, err = s.Backend.VerifySignature(s.DataToSign, sign, acc.Address())
	assert.False(t, valid, "Verification should fail")
	assert.NotNil(t, err, "Verification of invalid signature should produce error")
}

// GenericAddressTest runs a test suite designed to test the general functionality of addresses.
// This function should be called by every implementation of the wallet interface.
func GenericAddressTest(t *testing.T, s *Setup) {
	init, err := s.Backend.NewAddressFromString(s.AddrString)
	assert.Nil(t, err, "String parsing of Address should work")
	unInit, err := s.Backend.NewAddressFromBytes(make([]byte, len(init.Bytes()), len(init.Bytes())))
	assert.Nil(t, err, "Byte deserialization of Address should work")
	addr, err := s.Backend.NewAddressFromBytes(init.Bytes())
	assert.Nil(t, err, "Byte deserialization of Address should work")
	assert.Equal(t, init, addr, "Expected equality to serialized byte array")
	addr, err = s.Backend.NewAddressFromString(init.String())
	assert.Nil(t, err, "String parsing of Address should work")

	assert.Equal(t, init, addr, "Expected equality to serialized string array")
	assert.True(t, init.Equals(init), "Expected equality of non-zero to itself")
	assert.False(t, init.Equals(unInit), "Expected non-equality to other")
	assert.True(t, unInit.Equals(unInit), "Expected equality of zero to itself")

	var buf bytes.Buffer
	addr.Encode(&buf)
	assert.Equal(t, buf.Bytes(), addr.Bytes(), "Encoding address into buffer")
	addrDec, err := s.Backend.DecodeAddress(&buf)
	assert.Nil(t, err, "Decoding address from buffer")
	assert.Equal(t, addr.Bytes(), addrDec.Bytes(), "Decoding address from buffer")

	_, err = s.Backend.DecodeAddress(&buf)
	assert.NotNil(t, err, "Decoding address from empty stream should fail")
}

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

package test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/pkg/io"
	"perun.network/go-perun/wallet"
)

// InitWallet initializes a wallet.

// UnlockedAccount provides an unlocked account.
type UnlockedAccount func() (wallet.Account, error)

// Setup provides all objects needed for the generic tests.
type Setup struct {
	Backend wallet.Backend // backend implementation
	Wallet  wallet.Wallet  // the wallet instance used for testing
	Address wallet.Address // an address of an account in the test wallet
	// Address tests
	AddressEncoded []byte // a valid nonzero address not in the wallet
	// Signature tests
	DataToSign []byte
}

// GenericSignatureTest runs a test suite designed to test the general functionality of an account.
// This function should be called by every implementation of the wallet interface.
func GenericSignatureTest(t *testing.T, s *Setup) {
	acc, err := s.Wallet.Unlock(s.Address)
	assert.NoError(t, err)
	// Check unlocked account
	sig, err := acc.SignData(s.DataToSign)
	assert.NoError(t, err, "Sign with unlocked account should succeed")
	valid, err := s.Backend.VerifySignature(s.DataToSign, sig, acc.Address())
	assert.True(t, valid, "Verification should succeed")
	assert.NoError(t, err, "Verification should not produce error")

	addr, err := s.Backend.DecodeAddress(bytes.NewBuffer(s.AddressEncoded))
	assert.NoError(t, err, "Byte deserialization of address should work")
	valid, err = s.Backend.VerifySignature(s.DataToSign, sig, addr)
	assert.False(t, valid, "Verification with wrong address should fail")
	assert.NoError(t, err, "Verification of valid signature should not produce error")

	tampered := make([]byte, len(sig))
	copy(tampered, sig)
	// Invalidate the signature and check for error
	tampered[0] = ^sig[0]
	valid, err = s.Backend.VerifySignature(s.DataToSign, tampered, acc.Address())
	if valid && err == nil {
		t.Error("Verification of invalid signature should produce error or return false")
	}
	// Truncate the signature and check for error
	tampered = sig[:len(sig)-1]
	valid, err = s.Backend.VerifySignature(s.DataToSign, tampered, acc.Address())
	if valid && err != nil {
		t.Error("Verification of invalid signature should produce error or return false")
	}
	// Expand the signature and check for error
	// nolint:gocritic
	tampered = append(sig, 0)
	valid, err = s.Backend.VerifySignature(s.DataToSign, tampered, acc.Address())
	if valid && err != nil {
		t.Error("Verification of invalid signature should produce error or return false")
	}

	// Test DecodeSig
	sig, err = acc.SignData(s.DataToSign)
	require.NoError(t, err, "Sign with unlocked account should succeed")

	buff := new(bytes.Buffer)
	err = io.Encode(buff, sig)
	require.NoError(t, err, "encode sig")
	sign2, err := s.Backend.DecodeSig(buff)
	assert.NoError(t, err, "Decoded signature should work")
	assert.Equal(t, sig, sign2, "Decoded signature should be equal to the original")

	// Test DecodeSig on short stream
	err = io.Encode(buff, sig)
	require.NoError(t, err, "encode sig")
	shortBuff := bytes.NewBuffer(buff.Bytes()[:len(buff.Bytes())-1]) // remove one byte
	_, err = s.Backend.DecodeSig(shortBuff)
	assert.Error(t, err, "DecodeSig on short stream should error")
}

// GenericSignatureSizeTest tests that the size of the signatures produced by
// Account.Sign(…) does not vary between executions (tested with 2048 samples).
func GenericSignatureSizeTest(t *testing.T, s *Setup) {
	acc, err := s.Wallet.Unlock(s.Address)
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

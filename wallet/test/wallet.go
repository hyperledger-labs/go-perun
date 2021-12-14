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

	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wire/perunio"
)

// InitWallet initializes a wallet.

// UnlockedAccount provides an unlocked account.
type UnlockedAccount func() (wallet.Account, error)

// Setup provides all objects needed for the generic tests.
type Setup struct {
	Backend         wallet.Backend // backend implementation
	Wallet          wallet.Wallet  // the wallet instance used for testing
	AddressInWallet wallet.Address // an address of an account in the test wallet
	ZeroAddress     wallet.Address // an address that is less or equal to any other address
	DataToSign      []byte         // some data to sign
	AddressEncoded  []byte         // a valid nonzero address not in the wallet
}

// TestAccountWithWalletAndBackend tests an account implementation together with
// a corresponding wallet and backend implementation.
// This function should be called by every implementation of the wallet interface.
func TestAccountWithWalletAndBackend(t *testing.T, s *Setup) { //nolint:revive // `test.Test...` stutters, but we accept that here.
	acc, err := s.Wallet.Unlock(s.AddressInWallet)
	assert.NoError(t, err)
	// Check unlocked account
	sig, err := acc.SignData(s.DataToSign)
	assert.NoError(t, err, "Sign with unlocked account should succeed")
	valid, err := sig.Verify(s.DataToSign, acc.Address())
	assert.True(t, valid, "Verification should succeed")
	assert.NoError(t, err, "Verification should not produce error")

	addr := s.Backend.NewAddress()
	err = perunio.Decode(bytes.NewReader(s.AddressEncoded), addr)
	assert.NoError(t, err, "Byte deserialization of address should work")
	valid, err = sig.Verify(s.DataToSign, addr)
	assert.False(t, valid, "Verification with wrong address should fail")
	assert.NoError(t, err, "Verification of valid signature should not produce error")

	// Invalidate the signature and check for error
	tamperedBytes := marshalAndTamper(t, sig, func(sigBytes *[]byte) { (*sigBytes)[0] = ^(*sigBytes)[0] })
	tampered := s.Backend.NewSig()
	require.NoError(t, tampered.UnmarshalBinary(tamperedBytes))
	valid, err = tampered.Verify(s.DataToSign, acc.Address())
	if valid && err == nil {
		t.Error("Verification of invalid signature should produce error or return false")
	}
	// Truncate the signature and check for error
	tamperedBytes = marshalAndTamper(t, sig, func(sigBytes *[]byte) { *sigBytes = (*sigBytes)[:len(*sigBytes)-1] })
	tampered = s.Backend.NewSig()
	require.Error(t, tampered.UnmarshalBinary(tamperedBytes), "unmarshalling should fail when length is incorrect")

	// Expand the signature and check for error
	tamperedBytes = marshalAndTamper(t, sig, func(sigBytes *[]byte) { *sigBytes = append(*sigBytes, 0) })
	tampered = s.Backend.NewSig()
	require.Error(t, tampered.UnmarshalBinary(tamperedBytes), "unmarshalling should fail when length is incorrect")

	// Test DecodeSig
	sig, err = acc.SignData(s.DataToSign)
	require.NoError(t, err, "Sign with unlocked account should succeed")

	buff := new(bytes.Buffer)
	err = perunio.Encode(buff, sig)
	require.NoError(t, err, "encode sig")
	sign2 := s.Backend.NewSig()
	err = perunio.Decode(buff, sign2)
	assert.NoError(t, err, "Decoded signature should work")
	assert.Equal(t, sig, sign2, "Decoded signature should be equal to the original")

	// Test DecodeSig on short stream
	err = perunio.Encode(buff, sig)
	require.NoError(t, err, "encode sig")
	shortBuff := bytes.NewBuffer(buff.Bytes()[:len(buff.Bytes())-1]) // remove one byte
	sign3 := s.Backend.NewSig()
	err = perunio.Decode(shortBuff, sign3)
	assert.Error(t, err, "DecodeSig on short stream should error")
}

func marshalAndTamper(t *testing.T, sig wallet.Sig, tamperer func(sigBytes *[]byte)) []byte {
	t.Helper()

	sigBytes, err := sig.MarshalBinary()
	require.NoError(t, err)
	tamperer(&sigBytes)
	return sigBytes
}

// GenericSignatureSizeTest tests that the size of the signatures produced by
// Account.Sign(…) does not vary between executions (tested with 2048 samples).
func GenericSignatureSizeTest(t *testing.T, s *Setup) {
	t.Helper()
	acc, err := s.Wallet.Unlock(s.AddressInWallet)
	require.NoError(t, err)
	// get a signature
	sign, err := acc.SignData(s.DataToSign)
	require.NoError(t, err, "Sign with unlocked account should succeed")

	signRaw, err := sign.MarshalBinary()
	require.NoError(t, err)

	// Test that signatures have constant length
	l := len(signRaw)
	for i := 0; i < 8; i++ {
		t.Run("parallel signing", func(t *testing.T) {
			t.Parallel()
			for i := 0; i < 256; i++ {
				sign, err := acc.SignData(s.DataToSign)
				require.NoError(t, err, "Sign with unlocked account should succeed")
				signRaw, err := sign.MarshalBinary()
				require.NoError(t, err)
				require.Equal(t, l, len(signRaw), "Signatures should have constant length: %d vs %d", l, len(signRaw))
			}
		})
	}
}

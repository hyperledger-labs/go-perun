// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package wallet

import (
	"encoding/hex"
	"testing"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	perun "perun.network/go-perun/wallet"
	"perun.network/go-perun/wallet/test"
)

const (
	keyDir      = "testdata"
	password    = "secret"
	sampleAddr  = "1234560000000000000000000000000000000000"
	invalidAddr = "0x12345600000000000000000000000000000000001"
	dataToSign  = "SomeLongDataThatShouldBeSignedPlease"

	keystoreAddr = "0x647ec26ae49b14060660504f4DA1c2059E1C5Ab6"
	keyStorePath = "UTC--2019-12-11T17-00-07.156850888Z--647ec26ae49b14060660504f4da1c2059e1c5ab6"
)

func TestGenericWalletTests(t *testing.T) {
	setup := newSetup()
	test.GenericWalletTest(t, setup)
}

func TestGenericSignatureTests(t *testing.T) {
	setup := newSetup()
	test.GenericSignatureTest(t, setup)
	test.GenericSignatureSizeTest(t, setup)
}

func TestGenericAddressTests(t *testing.T) {
	setup := newSetup()
	test.GenericAddressTest(t, setup)
}

func TestAddress(t *testing.T) {
	w := connectTmpKeystore(t)
	perunAcc := w.Accounts()[0]
	ethAcc := new(accounts.Account)

	unsetAccount := new(Account)
	assert.Equal(t, make([]byte, common.AddressLength), unsetAccount.Address().Bytes(),
		"Unset address should be zero array")

	ethAcc.Address.SetBytes(perunAcc.Address().Bytes())
	assert.Equal(t, ethAcc.Address.Bytes(), perunAcc.Address().Bytes(),
		"Ethereum account set to perun address bytes should be the same")
	assert.NotEqual(t, [common.AddressLength]byte{}, [common.AddressLength]byte(ethAcc.Address),
		"Ethereum account should not be zero array")
	assert.NotEqual(t, nil, ethAcc.Address.Bytes(), "Set address should not be nil")
}

func TestKeyStore(t *testing.T) {
	w := new(Wallet)
	assert.NotNil(t, w.Connect("", ""), "Expected connect to fail")
	assert.NotNil(t, w.Connect("Invalid_Directory", ""), "Expected connect to fail")
	assert.NotNil(t, w.Disconnect(), "Expected disconnect on uninitialized wallet to fail")
	assert.NotNil(t, w.Lock(), "Expected lock on uninitialized wallet to fail")
	w = connectTmpKeystore(t)

	unsetAccount := new(Account)
	assert.False(t, w.Contains(unsetAccount), "Keystore should not contain empty account")
}

func TestLocking(t *testing.T) {
	t.Run("WrongUnlock", func(t *testing.T) {
		t.Parallel()
		w := connectTmpKeystore(t)
		acc := w.Accounts()[0].(*Account)
		assert.True(t, w.Contains(acc), "Expected wallet to contain account")
		assert.True(t, acc.IsLocked(), "Account should be locked")
		assert.NotNil(t, acc.Unlock(""), "Unlock with wrong pw should fail")
	})
	t.Run("Unlock&Lock", func(t *testing.T) {
		t.Parallel()
		w := connectTmpKeystore(t)
		acc := w.Accounts()[0].(*Account)
		assert.True(t, w.Contains(acc), "Expected wallet to contain account")
		assert.Nil(t, acc.Unlock(password), "Expected unlock to work")
		assert.False(t, acc.IsLocked(), "Account should be unlocked")
		assert.Nil(t, acc.Lock(), "Expected lock to work")
		assert.True(t, acc.IsLocked(), "Account should be locked")
	})
	t.Run("Unlock&LockWallet", func(t *testing.T) {
		t.Parallel()
		w := connectTmpKeystore(t)
		acc := w.Accounts()[0].(*Account)
		assert.True(t, w.Contains(acc), "Expected wallet to contain account")
		assert.Nil(t, acc.Unlock(password), "Expected unlock to work")
		assert.False(t, acc.IsLocked(), "Account should be unlocked")
		assert.Nil(t, w.Lock(), "Expected lock to succeed")
		assert.True(t, acc.IsLocked(), "Account should be locked")
	})
}

func TestSignatures(t *testing.T) {
	w := connectTmpKeystore(t)
	acc := w.Accounts()[0].(*Account)
	_, err := acc.SignData([]byte(dataToSign))
	assert.NotNil(t, err, "Sign with locked account should fail")
	sign, err := acc.SignDataWithPW(password, []byte(dataToSign))
	assert.Nil(t, err, "SignPW with locked account should succeed")
	assert.Equal(t, len(sign), SignatureLength, "Ethereum signature has wrong length")
	valid, err := new(Backend).VerifySignature([]byte(dataToSign), sign, acc.Address())
	assert.True(t, valid, "Verification should succeed")
	assert.Nil(t, err, "Verification should succeed")
	assert.True(t, acc.IsLocked(), "Account should not be unlocked")
	_, err = acc.SignDataWithPW("", []byte(dataToSign))
	assert.NotNil(t, err, "SignPW with wrong pw should fail")
}

func TestBackend(t *testing.T) {
	backend := new(Backend)

	s := newSetup()

	addr, err := backend.NewAddressFromBytes(s.AddressBytes)
	assert.Nil(t, err, "NewAddress from Bytes should work")
	assert.Equal(t, s.AddressBytes, addr.Bytes())

	_, err = backend.NewAddressFromBytes([]byte(invalidAddr))
	assert.NotNil(t, err, "Conversion from wrong address should fail")
}

func newSetup() *test.Setup {
	wallet := new(Wallet)
	wallet.Connect(keyDir, password)
	acc := wallet.Accounts()[0].(*Account)
	acc.Unlock(password)

	sampleBytes, err := hex.DecodeString(sampleAddr)
	if err != nil {
		panic("invalid sample address")
	}

	initWallet := func(w perun.Wallet) error { return w.Connect("./"+keyDir, password) }
	unlockedAccount := func() (perun.Account, error) { return acc, nil }

	return &test.Setup{
		Wallet:          new(Wallet),
		InitWallet:      initWallet,
		UnlockedAccount: unlockedAccount,
		Backend:         new(Backend),
		AddressBytes:    sampleBytes,
		DataToSign:      []byte(dataToSign),
	}
}

func connectTmpKeystore(t *testing.T) *Wallet {
	w := new(Wallet)
	assert.Nil(t, w.Connect(keyDir, password), "Unable to open keystore")
	assert.NotEqual(t, len(w.Accounts()), 0, "Wallet contains no accounts")
	return w
}

// Benchmarking part starts here

func BenchmarkGenericAccount(b *testing.B) {
	setup := newSetup()
	test.GenericAccountBenchmark(b, setup)
}

func BenchmarkGenericWallet(b *testing.B) {
	setup := newSetup()
	test.GenericWalletBenchmark(b, setup)
}

func BenchmarkGenericBackend(b *testing.B) {
	setup := newSetup()
	test.GenericBackendBenchmark(b, setup)
}

func BenchmarkEthereumAccounts(b *testing.B) {
	s := newSetup()
	b.Run("Lock", func(b *testing.B) { benchAccountLock(b, s) })
	b.Run("Unlock", func(b *testing.B) { benchAccountUnlock(b, s) })
}

func benchAccountLock(b *testing.B, s *test.Setup) {
	perunAcc, err := s.UnlockedAccount()
	require.Nil(b, err)
	acc := perunAcc.(*Account)

	for n := 0; n < b.N; n++ {
		err := acc.Lock()

		if err != nil {
			b.Fatal(err)
		}
	}
}

func benchAccountUnlock(b *testing.B, s *test.Setup) {
	perunAcc, err := s.UnlockedAccount()
	require.Nil(b, err)
	acc := perunAcc.(*Account)

	for n := 0; n < b.N; n++ {
		err := acc.Unlock(password)

		if err != nil {
			b.Fatal(err)
		}
	}
}

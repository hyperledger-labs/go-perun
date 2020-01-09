// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package wallet

import (
	"bytes"
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
	setup := newSetup(t)
	test.GenericWalletTest(t, setup)
}

func TestGenericSignatureTests(t *testing.T) {
	setup := newSetup(t)
	test.GenericSignatureTest(t, setup)
	test.GenericSignatureSizeTest(t, setup)
}

func TestGenericAddressTests(t *testing.T) {
	setup := newSetup(t)
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
	assert.Equal(t, len(sign), SigLen, "Ethereum signature has wrong length")
	valid, err := new(Backend).VerifySignature([]byte(dataToSign), sign, acc.Address())
	assert.True(t, valid, "Verification should succeed")
	assert.Nil(t, err, "Verification should succeed")
	assert.True(t, acc.IsLocked(), "Account should not be unlocked")
	_, err = acc.SignDataWithPW("", []byte(dataToSign))
	assert.NotNil(t, err, "SignPW with wrong pw should fail")
}

func TestBackend(t *testing.T) {
	backend := new(Backend)

	s := newSetup(t)

	buff := bytes.NewReader(s.AddressBytes)
	addr, err := backend.DecodeAddress(buff)

	assert.Nil(t, err, "NewAddress from Bytes should work")
	assert.Equal(t, s.AddressBytes, addr.Bytes())

	buff = bytes.NewReader([]byte(invalidAddr))
	_, err = backend.DecodeAddress(buff)
	assert.NotNil(t, err, "Conversion from wrong address should fail")
}

func newSetup(t require.TestingT) *test.Setup {
	wallet := new(Wallet)
	require.NoError(t, wallet.Connect(keyDir, password))
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
	setup := newSetup(b)
	test.GenericAccountBenchmark(b, setup)
}

func BenchmarkGenericWallet(b *testing.B) {
	setup := newSetup(b)
	test.GenericWalletBenchmark(b, setup)
}

func BenchmarkGenericBackend(b *testing.B) {
	setup := newSetup(b)
	test.GenericBackendBenchmark(b, setup)
}

func BenchmarkEthereumAccounts(b *testing.B) {
	s := newSetup(b)
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

func Test_EthSignature(t *testing.T) {
	msg, err := hex.DecodeString("f27b90711d11d10a155fc8ba0eed1ffbf449cf3730d88c0cb77b98f61750ab34000000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000a0000000000000000000000000000000000000000000000000000000000000022000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000160000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000000a0000000000000000000000000000000000000000000000000000000000000014000000000000000000000000000000000000000000000000000000000000000010000000000000000000000002c2b9c9a4a25e24b174f26114e8926a9f2128fe40000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000000000000000000000000000000de0b6b3a7640000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000")
	require.NoError(t, err, "decode msg should not error")
	sig, err := hex.DecodeString("538da6430f7915832de165f89c69239020461b80861559a00d4f5a2a7705765219eb3969eb7095f8addb6bf9c9f96f6adf44cfd4a8136516f88b337a428bf1bb1b")
	require.NoError(t, err, "decode sig should not error")
	addr := &Address{
		Address: common.HexToAddress("f17f52151EbEF6C7334FAD080c5704D77216b732"),
	}
	b, err := VerifySignature(msg, sig, addr)
	assert.NoError(t, err, "VerifySignature should not error")
	assert.True(t, b, "VerifySignature")
}

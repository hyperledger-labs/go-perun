// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package wallet

import (
	"testing"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"perun.network/go-perun/wallet/test"
	perun "perun.network/go-perun/wallet"
)

const (
	keyDir      = "testdata"
	password    = "secret"
	sampleAddr  = "0x1234560000000000000000000000000000000000"
	invalidAddr = "0x12345600000000000000000000000000000000001"
	dataToSign  = "SomeLongDataThatShouldBeSignedPlease"

	keystoreAddr = "0xf4c288068b32474dedc3620233c"
	keyStorePath = "UTC--2019-06-07T12-12-48.775026092Z--3c5a96ff258b1f4c288068b32474dedc3620233c"
)

func TestGenericWalletTests(t *testing.T) {
	t.Parallel()
	setup := newSetup()
	test.GenericWalletTest(t, setup)
}

func TestGenericSignatureTests(t *testing.T) {
	t.Parallel()
	setup := newSetup()
	test.GenericSignatureTest(t, setup)
}

func TestGenericAddressTests(t *testing.T) {
	t.Parallel()
	setup := newSetup()
	test.GenericAddressTest(t, setup)
}

func TestAddress(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
	w := connectTmpKeystore(t)
	acc := w.Accounts()[0].(*Account)
	assert.True(t, w.Contains(acc), "Expected wallet to contain account")
	// Check unlock account
	assert.True(t, acc.IsLocked(), "Account should be locked")
	assert.NotNil(t, acc.Unlock(""), "Unlock with wrong pw should fail")
	assert.Nil(t, acc.Unlock(password), "Expected unlock to work")
	assert.False(t, acc.IsLocked(), "Account should be unlocked")
	assert.Nil(t, acc.Lock(), "Expected lock to work")
	assert.True(t, acc.IsLocked(), "Account should be locked")
	// Check lock all accounts
	assert.Nil(t, acc.Unlock(password), "Expected unlock to work")
	assert.False(t, acc.IsLocked(), "Account should be unlocked")
	assert.Nil(t, w.Lock(), "Expected lock to succeed")
	assert.True(t, acc.IsLocked(), "Account should be locked")
}

func TestSignatures(t *testing.T) {
	t.Parallel()
	w := connectTmpKeystore(t)
	acc := w.Accounts()[0].(*Account)
	_, err := acc.SignData([]byte(dataToSign))
	assert.NotNil(t, err, "Sign with locked account should fail")
	sign, err := acc.SignDataWithPW(password,  []byte(dataToSign))
	assert.Nil(t, err, "SignPW with locked account should succeed")
	valid, err := new(Backend).VerifySignature( []byte(dataToSign), sign, acc.Address())
	assert.True(t, valid, "Verification should succeed")
	assert.Nil(t, err, "Verification should succeed")
	assert.True(t, acc.IsLocked(), "Account should not be unlocked")
}

func TestBackend(t *testing.T) {
	t.Parallel()
	backend := new(Backend)

	addrStr, err := backend.NewAddressFromString(sampleAddr)
	assert.Nil(t, err, "NewAddress from valid address string should work")
	assert.Equal(t, addrStr.String(), sampleAddr, "NewAddress from valid address string should be the same")

	addrBytes, err := backend.NewAddressFromBytes(addrStr.Bytes())
	assert.Nil(t, err, "NewAddress from Bytes should work")
	assert.True(t, addrStr.Equals(addrBytes), "Address from bytes or string should be the same")

	_, err = backend.NewAddressFromBytes([]byte(invalidAddr))
	assert.NotNil(t, err, "Conversion from wrong address should fail")

	_, err = backend.NewAddressFromString(invalidAddr)
	assert.NotNil(t, err, "Conversion from wrong address should fail")
}

func newSetup() *test.Setup {
	wallet := new(Wallet)
	wallet.Connect(keyDir, password)
	acc := wallet.Accounts()[0].(*Account)
	acc.Unlock(password)

	initWallet := func (w perun.Wallet) (error) { return w.Connect("./" + keyDir, password)}
	unlockedAccount := func () (perun.Account, error) {return acc, nil}

	return &test.Setup{
		Wallet:     new(Wallet),
		InitWallet: initWallet,
		UnlockedAccount: unlockedAccount,
		Backend:    new(Backend),
		AddrString: sampleAddr,
		DataToSign: []byte(dataToSign),
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
	b.Run("Lock", func(t *testing.B) { benchAccountLock(t, s) })
	b.Run("Unlock", func(t *testing.B) { benchAccountUnlock(t, s) })
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


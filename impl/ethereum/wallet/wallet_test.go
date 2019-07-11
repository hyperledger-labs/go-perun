// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package wallet

import (
	"testing"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"perun.network/go-perun/wallet/test"
)

const (
	keyDir      = "testdata"
	password    = "secret"
	sampleAddr  = "0x1234560000000000000000000000000000000000"
	invalidAddr = "0x12345600000000000000000000000000000000001"
	dataToSign  = "SomeLongDataThatShouldBeSignedPlease"
	signedData  = ""

	keystoreAddr = "0xf4c288068b32474dedc3620233c"
	keyStorePath = "UTC--2019-06-07T12-12-48.775026092Z--3c5a96ff258b1f4c288068b32474dedc3620233c"
)

func TestGenericWalletTests(t *testing.T) {
	t.Parallel()
	setup := newTestSetup(t)
	test.GenericWalletTest(setup)
}

func TestGenericSignatureTests(t *testing.T) {
	t.Parallel()
	setup := newTestSetup(t)
	test.GenericSignatureTest(setup)
}

func TestGenericAddressTests(t *testing.T) {
	t.Parallel()
	setup := newTestSetup(t)
	test.GenericAddressTest(setup)
}

func TestAddress(t *testing.T) {
	t.Parallel()
	w := connectTmpKeystore(t)
	perunAcc := w.Accounts()[0]
	ethAcc := new(accounts.Account)

	unsetAccount := new(Account)
	nilAddr := common.BytesToAddress(make([]byte, 40, 40))

	assert.Equal(t, nilAddr.Bytes(), unsetAccount.Address().Bytes(), "Unset address should be nil")

	ethAcc.Address.SetBytes(perunAcc.Address().Bytes())

	assert.Equal(t, ethAcc.Address.Bytes(), perunAcc.Address().Bytes(), "Bytes should return same value as internal structure")

	assert.NotEqual(t, nil, ethAcc.Address.Bytes(), "Set address should not be nil")
}

func TestKeyStore(t *testing.T) {
	t.Parallel()
	w := new(Wallet)
	assert.NotNil(t, w.Connect("", ""), "Expected connect to fail")

	w = connectTmpKeystore(t)

	unsetAccount := new(Account)
	assert.False(t, w.Contains(unsetAccount), "Keystore should not contain empty account")
}

func TestHelper(t *testing.T) {
	t.Parallel()
	helper := new(Helper)
	addr, err := helper.NewAddressFromString(sampleAddr)

	assert.Nil(t, err, "Conversion of valid address should work")

	_, err = helper.NewAddressFromBytes(addr.Bytes())

	assert.Nil(t, err, "Conversion of valid address should work")

	_, err = helper.NewAddressFromBytes([]byte(invalidAddr))

	assert.NotNil(t, err, "Conversion from wrong address should fail")

	_, err = helper.NewAddressFromString(invalidAddr)

	assert.NotNil(t, err, "Conversion from wrong address should fail")
}

func newTestSetup(t *testing.T) *test.Setup {
	setup := new(test.Setup)
	setup.T = t
	setup.Wallet = new(Wallet)
	setup.Path = "./" + keyDir
	setup.WalletPW = password
	setup.AccountPW = password
	setup.Helper = new(Helper)
	setup.AddrString = sampleAddr
	setup.DataToSign = []byte(dataToSign)
	setup.SignedData = []byte(signedData)
	return setup
}

func connectTmpKeystore(t *testing.T) *Wallet {
	w := new(Wallet)
	assert.Nil(t, w.Connect(keyDir, password), "Unable to open keystore")
	assert.NotEqual(t, len(w.Accounts()), 0, "Wallet contains no accounts")
	return w
}

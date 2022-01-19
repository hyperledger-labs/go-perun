// Copyright 2020 - See NOTICE file for copyright holders.
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

package hd_test

import (
	"math/rand"
	"testing"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	hdwalletimpl "github.com/miguelmota/go-ethereum-hdwallet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ethwallet "perun.network/go-perun/backend/ethereum/wallet"
	"perun.network/go-perun/backend/ethereum/wallet/hd"
	ethwallettest "perun.network/go-perun/backend/ethereum/wallet/test"
	"perun.network/go-perun/wallet/test"
	pkgtest "polycry.pt/poly-go/test"
)

var dataToSign = []byte("SomeLongDataThatShouldBeSignedPlease")

func TestGenericSignatureTests(t *testing.T) {
	s, _, _ := newSetup(t, pkgtest.Prng(t))
	test.TestAccountWithWalletAndBackend(t, s)
	test.GenericSignatureSizeTest(t, s)
	test.TestAddress(t, s)
}

func TestNewWallet(t *testing.T) {
	prng := pkgtest.Prng(t)

	walletSeed := make([]byte, 20)
	prng.Read(walletSeed)
	mnemonic, err := hdwalletimpl.NewMnemonicFromEntropy(walletSeed)
	require.NoError(t, err)

	rawHDWallet, err := hdwalletimpl.NewFromMnemonic(mnemonic)
	require.NoError(t, err)

	t.Run("happy", func(t *testing.T) {
		hdWallet, err := hd.NewWallet(rawHDWallet, hd.DefaultRootDerivationPath.String(), 0)
		require.NoError(t, err)
		require.NotNil(t, hdWallet)
	})

	t.Run("err_DerivationPath", func(t *testing.T) {
		_, err := hd.NewWallet(rawHDWallet, "invalid-derivation-path", 0)
		require.Error(t, err)
	})

	t.Run("err_nilWallet", func(t *testing.T) {
		_, err := hd.NewWallet(nil, hd.DefaultRootDerivationPath.String(), 0)
		require.Error(t, err)
	})
}

func TestSignWithMissingKey(t *testing.T) {
	setup, accsWallet, _ := newSetup(t, pkgtest.Prng(t))

	missingAddr := common.BytesToAddress(setup.AddressMarshalled)
	acc := hd.NewAccountFromEth(accsWallet, accounts.Account{Address: missingAddr})
	require.NotNil(t, acc)
	_, err := acc.SignData(setup.DataToSign)
	assert.Error(t, err, "Sign with missing account should fail")
}

func TestUnlock(t *testing.T) {
	setup, _, hdWallet := newSetup(t, pkgtest.Prng(t))

	missingAddr := common.BytesToAddress(setup.AddressMarshalled)
	_, err := hdWallet.Unlock(ethwallet.AsWalletAddr(missingAddr))
	assert.Error(t, err, "should error on unlocking missing address")

	acc, err := hdWallet.Unlock(setup.AddressInWallet)
	assert.NoError(t, err, "should not error on unlocking valid address")
	assert.NotNil(t, acc, "account should be non nil when error is nil")
}

func TestContains(t *testing.T) {
	setup, _, hdWallet := newSetup(t, pkgtest.Prng(t))

	assert.False(t, hdWallet.Contains(common.Address{}), "should not contain nil account")

	missingAddr := common.BytesToAddress(setup.AddressMarshalled)
	assert.False(t, hdWallet.Contains(missingAddr), "should not contain address of the missing account")

	assert.True(t, hdWallet.Contains(ethwallet.AsEthAddr(setup.AddressInWallet)), "should contain valid account")
}

func newSetup(t require.TestingT, prng *rand.Rand) (*test.Setup, accounts.Wallet, *hd.Wallet) {
	walletSeed := make([]byte, 20)
	prng.Read(walletSeed)
	mnemonic, err := hdwalletimpl.NewMnemonicFromEntropy(walletSeed)
	require.NoError(t, err)

	rawHDWallet, err := hdwalletimpl.NewFromMnemonic(mnemonic)
	require.NoError(t, err)
	require.NotNil(t, rawHDWallet, "hdwallet must not be nil")

	hdWallet, err := hd.NewWallet(rawHDWallet, hd.DefaultRootDerivationPath.String(), 0)
	require.NoError(t, err)
	require.NotNil(t, hdWallet)

	acc, err := hdWallet.NewAccount()
	require.NoError(t, err)
	require.NotNil(t, acc)

	addressNotInWallet := ethwallettest.NewRandomAddress(prng)
	addrMarshalled, err := addressNotInWallet.MarshalBinary()
	require.NoError(t, err)

	return &test.Setup{
		Wallet:            hdWallet,
		AddressInWallet:   acc.Address(),
		Backend:           new(ethwallet.Backend),
		AddressMarshalled: addrMarshalled,
		ZeroAddress:       ethwallet.AsWalletAddr(common.Address{}),
		DataToSign:        dataToSign,
	}, rawHDWallet, hdWallet
}

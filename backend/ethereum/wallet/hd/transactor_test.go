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
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/backend/ethereum/channel/test"
	"perun.network/go-perun/backend/ethereum/wallet"
	"perun.network/go-perun/backend/ethereum/wallet/hd"
	pkgtest "perun.network/go-perun/pkg/test"
)

// missingAddr is the address for which key will not be contained in the wallet.
const missingAddr = "0x1"

func TestTransactor(t *testing.T) {
	prng := pkgtest.Prng(t)
	test.GenericTransactorTest(t, newTransactorSetup(t, prng))
}

// nolint:interfacer // rand.Rand is preferred over io.Reader here.
func newTransactorSetup(t require.TestingT, prng *rand.Rand) test.TransactorSetup {
	walletSeed := make([]byte, 20)
	prng.Read(walletSeed)
	mnemonic, err := hdwalletimpl.NewMnemonicFromEntropy(walletSeed)
	require.NoError(t, err)

	rawHDWallet, err := hdwalletimpl.NewFromMnemonic(mnemonic)
	require.NoError(t, err)
	require.NotNil(t, rawHDWallet)

	hdWallet, err := hd.NewWallet(rawHDWallet, hd.DefaultRootDerivationPath.String(), 0)
	require.NoError(t, err)
	require.NotNil(t, hdWallet)

	validAcc, err := hdWallet.NewAccount()
	require.NoError(t, err)
	require.NotNil(t, validAcc)

	return test.TransactorSetup{
		Tr:         hd.NewTransactor(hdWallet.Wallet()),
		ValidAcc:   accounts.Account{Address: wallet.AsEthAddr(validAcc.Address())},
		MissingAcc: accounts.Account{Address: common.HexToAddress(missingAddr)},
	}
}

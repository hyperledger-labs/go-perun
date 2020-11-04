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
	"math/big"
	"math/rand"
	"testing"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/backend/ethereum/channel/test"
	"perun.network/go-perun/backend/ethereum/wallet"
	"perun.network/go-perun/backend/ethereum/wallet/hd"
	pkgtest "perun.network/go-perun/pkg/test"
)

// missingAddr is the address for which key will not be contained in the wallet.
const missingAddr = "0x1"

// noSignHash is used to hide the SignHash function of the embedded wallet.
type noSignHash struct {
	*hdwallet.Wallet
	SignTxCalled bool
}

func TestTransactor(t *testing.T) {
	rng := pkgtest.Prng(t)
	// The normal transactor should be able to deal with all signers.
	s := newTransactorSetup(t, rng, false)
	test.GenericLegacyTransactorTest(t, rng, s)
	test.GenericEIP155TransactorTest(t, rng, s)

	// The noSignHasher should only work with Frontier and Homestead.
	s = newTransactorSetup(t, rng, true)
	test.GenericLegacyTransactorTest(t, rng, s)
	assert.True(t, s.Tr.(*hd.Transactor).Wallet.(*noSignHash).SignTxCalled)
}

// nolint:interfacer // rand.Rand is preferred over io.Reader here.
func newTransactorSetup(t *testing.T, prng *rand.Rand, hideSignHash bool) test.TransactorSetup {
	walletSeed := make([]byte, 20)
	prng.Read(walletSeed)
	mnemonic, err := hdwallet.NewMnemonicFromEntropy(walletSeed)
	require.NoError(t, err)

	rawHDWallet, err := hdwallet.NewFromMnemonic(mnemonic)
	require.NoError(t, err)
	require.NotNil(t, rawHDWallet)

	var wrappedWallet accounts.Wallet = rawHDWallet
	if hideSignHash {
		wrappedWallet = &noSignHash{rawHDWallet, false}
	}
	hdWallet, err := hd.NewWallet(wrappedWallet, hd.DefaultRootDerivationPath.String(), 0)
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

// SignHash hides the public SignHash.
func (*noSignHash) SignHash() {
}

// SignTx calls SignTx of the embedded wallet and sets SignTxCalled to true.
func (n *noSignHash) SignTx(account accounts.Account, tx *types.Transaction, chainID *big.Int) (*types.Transaction, error) {
	n.SignTxCalled = true
	return n.Wallet.SignTx(account, tx, chainID)
}

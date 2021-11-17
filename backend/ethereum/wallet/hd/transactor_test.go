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
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/backend/ethereum/channel/test"
	"perun.network/go-perun/backend/ethereum/wallet"
	"perun.network/go-perun/backend/ethereum/wallet/hd"
	pkgtest "polycry.pt/poly-go/test"
)

// missingAddr is the address for which key will not be contained in the wallet.
const missingAddr = "0x1"

// noSignHash is used to hide the SignHash function of the embedded wallet.
type noSignHash struct {
	*hdwallet.Wallet
}

func TestTransactor(t *testing.T) {
	rng := pkgtest.Prng(t)
	chainID := rng.Int63()

	tests := []struct {
		title        string
		signer       types.Signer
		chainID      int64
		hideSignHash bool
	}{
		{
			title:   "FrontierSigner",
			signer:  &types.FrontierSigner{},
			chainID: 0,
		},
		{
			title:   "HomesteadSigner",
			signer:  &types.HomesteadSigner{},
			chainID: 0,
		},
		{
			title:        "FrontierSigner (hideSignHash)",
			signer:       &types.FrontierSigner{},
			chainID:      0,
			hideSignHash: true,
		},
		{
			title:        "HomesteadSigner (hideSignHash)",
			signer:       &types.HomesteadSigner{},
			chainID:      0,
			hideSignHash: true,
		},
		{
			title:   "EIP155Signer",
			signer:  types.NewEIP155Signer(big.NewInt(chainID)),
			chainID: chainID,
		},
	}

	for _, _t := range tests {
		_t := _t
		t.Run(_t.title, func(t *testing.T) {
			s := newTransactorSetup(t, rng, _t.hideSignHash, _t.signer, _t.chainID)
			test.GenericSignerTest(t, rng, s)
		})
	}
}

// rand.Rand is preferred over io.Reader here.
func newTransactorSetup(t *testing.T, prng *rand.Rand, hideSignHash bool, signer types.Signer, chainID int64) test.TransactorSetup {
	t.Helper()
	walletSeed := make([]byte, 20)
	prng.Read(walletSeed)
	mnemonic, err := hdwallet.NewMnemonicFromEntropy(walletSeed)
	require.NoError(t, err)

	rawHDWallet, err := hdwallet.NewFromMnemonic(mnemonic)
	require.NoError(t, err)
	require.NotNil(t, rawHDWallet)

	var wrappedWallet accounts.Wallet = rawHDWallet
	if hideSignHash {
		wrappedWallet = &noSignHash{rawHDWallet}
	}
	hdWallet, err := hd.NewWallet(wrappedWallet, hd.DefaultRootDerivationPath.String(), 0)
	require.NoError(t, err)
	require.NotNil(t, hdWallet)

	validAcc, err := hdWallet.NewAccount()
	require.NoError(t, err)
	require.NotNil(t, validAcc)

	return test.TransactorSetup{
		Signer:     signer,
		ChainID:    chainID,
		Tr:         hd.NewTransactor(hdWallet.Wallet(), signer),
		ValidAcc:   accounts.Account{Address: wallet.AsEthAddr(validAcc.Address())},
		MissingAcc: accounts.Account{Address: common.HexToAddress(missingAddr)},
	}
}

// SignHash hides the public SignHash.
func (*noSignHash) SignHash() {
}

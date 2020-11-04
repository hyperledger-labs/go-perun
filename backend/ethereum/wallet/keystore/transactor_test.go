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

package keystore_test

import (
	"math/rand"
	"testing"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	ethchanneltest "perun.network/go-perun/backend/ethereum/channel/test"
	"perun.network/go-perun/backend/ethereum/wallet"
	"perun.network/go-perun/backend/ethereum/wallet/keystore"
	_ "perun.network/go-perun/backend/ethereum/wallet/test"
	pkgtest "perun.network/go-perun/pkg/test"
	wallettest "perun.network/go-perun/wallet/test"
)

// Random address for which key will not be contained in the wallet.
const randomAddr = "0x1"

func TestTxOptsBackend(t *testing.T) {
	prng := pkgtest.Prng(t)
	s := newTransactorSetup(t, prng)
	ethchanneltest.GenericEIP155TransactorTest(t, prng, s)
	ethchanneltest.GenericLegacyTransactorTest(t, prng, s)
}

func newTransactorSetup(t require.TestingT, prng *rand.Rand) ethchanneltest.TransactorSetup {
	ksWallet, ok := wallettest.RandomWallet().(*keystore.Wallet)
	require.Truef(t, ok, "random wallet in wallettest should be a keystore wallet")
	acc := wallettest.NewRandomAccount(prng)
	return ethchanneltest.TransactorSetup{
		Tr:         keystore.NewTransactor(*ksWallet),
		ValidAcc:   accounts.Account{Address: wallet.AsEthAddr(acc.Address())},
		MissingAcc: accounts.Account{Address: common.HexToAddress(randomAddr)},
	}
}

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

package test

import (
	"context"
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"

	ethchannel "perun.network/go-perun/backend/ethereum/channel"
	ethwallet "perun.network/go-perun/backend/ethereum/wallet"
	"perun.network/go-perun/backend/ethereum/wallet/keystore"
	"perun.network/go-perun/wallet"
	wallettest "perun.network/go-perun/wallet/test"
)

const defaultTxTimeout = 5 * time.Second

type (
	// SimSetup holds the test setup for a simulated backend.
	SimSetup struct {
		SimBackend *SimulatedBackend           // A simulated blockchain backend
		TxSender   *keystore.Account           // funded account for sending transactions
		CB         *ethchannel.ContractBackend // contract backend bound to the TxSender
	}

	// Setup holds a complete test setup for channel backend testing.
	Setup struct {
		SimSetup
		Accs    []*keystore.Account  // on-chain funders and channel participant accounts
		Parts   []wallet.Address     // channel participants
		Recvs   []*ethwallet.Address // on-chain receivers of withdrawn funds
		Funders []*ethchannel.Funder // funders, bound to respective account
		Adjs    []*SimAdjudicator    // adjudicator, withdrawal bound to respecive receivers
		Asset   common.Address       // the asset
	}
)

// NewSimSetup return a simulated backend test setup. The rng is used to
// generate the random account for sending of transaction.
func NewSimSetup(rng *rand.Rand, txFinalityDepth uint64) *SimSetup {
	simBackend := NewSimulatedBackend()
	ksWallet := wallettest.RandomWallet().(*keystore.Wallet)
	txAccount := ksWallet.NewRandomAccount(rng).(*keystore.Account)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	simBackend.FundAddress(ctx, txAccount.Account.Address)

	contractBackend := ethchannel.NewContractBackend(
		simBackend,
		keystore.NewTransactor(*ksWallet, types.NewEIP155Signer(big.NewInt(1337))),
		txFinalityDepth,
	)

	return &SimSetup{
		SimBackend: simBackend,
		TxSender:   txAccount,
		CB:         &contractBackend,
	}
}

// NewSetup returns a channel backend testing setup. When the adjudicator and
// asset holder contract are deployed and an error occurs, Fatal is called on
// the passed *testing.T. Parameter n determines how many accounts, receivers
// adjudicators and funders are created. The Parts are the Addresses of the
// Accs.
// `blockInterval` enables the auto-mining feature if set to a value != 0.
func NewSetup(t *testing.T, rng *rand.Rand, n int, blockInterval time.Duration, txFinalityDepth uint64) *Setup {
	s := &Setup{
		SimSetup: *NewSimSetup(rng, txFinalityDepth),
		Accs:     make([]*keystore.Account, n),
		Parts:    make([]wallet.Address, n),
		Recvs:    make([]*ethwallet.Address, n),
		Funders:  make([]*ethchannel.Funder, n),
		Adjs:     make([]*SimAdjudicator, n),
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTxTimeout)
	defer cancel()
	adjudicator, err := ethchannel.DeployAdjudicator(ctx, *s.CB, s.TxSender.Account)
	require.NoError(t, err)
	s.Asset, err = ethchannel.DeployETHAssetholder(ctx, *s.CB, adjudicator, s.TxSender.Account)
	require.NoError(t, err)
	asset := ethchannel.Asset(s.Asset)

	ksWallet := wallettest.RandomWallet().(*keystore.Wallet)
	require.NoErrorf(t, err, "initializing wallet from test keystore")
	for i := 0; i < n; i++ {
		s.Accs[i] = ksWallet.NewRandomAccount(rng).(*keystore.Account)
		s.Parts[i] = s.Accs[i].Address()
		s.SimBackend.FundAddress(ctx, s.Accs[i].Account.Address)
		s.Recvs[i] = ksWallet.NewRandomAccount(rng).Address().(*ethwallet.Address)
		cb := ethchannel.NewContractBackend(
			s.SimBackend,
			keystore.NewTransactor(*ksWallet, types.NewEIP155Signer(big.NewInt(1337))),
			txFinalityDepth,
		)
		s.Funders[i] = ethchannel.NewFunder(cb)
		require.True(t, s.Funders[i].RegisterAsset(asset, ethchannel.NewETHDepositor(), s.Accs[i].Account))
		s.Adjs[i] = NewSimAdjudicator(cb, adjudicator, common.Address(*s.Recvs[i]), s.Accs[i].Account)
	}

	if blockInterval != 0 {
		s.SimBackend.StartMining(blockInterval)
		t.Cleanup(s.SimBackend.StopMining)
	}

	return s
}

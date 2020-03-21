// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package test

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	ethchannel "perun.network/go-perun/backend/ethereum/channel"
	ethwallet "perun.network/go-perun/backend/ethereum/wallet"
	ethwallettest "perun.network/go-perun/backend/ethereum/wallet/test"
	"perun.network/go-perun/wallet"
	wallettest "perun.network/go-perun/wallet/test"
)

const defaultTxTimeout = 5 * time.Second

type (
	// SimSetup holds the test setup for a simulated backend.
	SimSetup struct {
		SimBackend *SimulatedBackend           // A simulated blockchain backend
		TxSender   *ethwallet.Account          // funded account for sending transactions
		CB         *ethchannel.ContractBackend // contract backend bound to the TxSender
	}

	// Setup holds a complete test setup for channel backend testing.
	Setup struct {
		SimSetup
		Accs    []*ethwallet.Account // on-chain funders and channel participant accounts
		Parts   []wallet.Address     // channel participants
		Recvs   []*ethwallet.Address // on-chain receivers of withdrawn funds
		Funders []*ethchannel.Funder // funders, bound to respective account
		Adjs    []*SimAdjudicator    // adjudicator, withdrawal bound to respecive receivers
		Asset   common.Address       // the asset
	}
)

// NewSimSetup return a simulated backend test setup. The rng is used to
// generate the random account for sending of transaction.
func NewSimSetup(rng *rand.Rand) *SimSetup {
	simBackend := NewSimulatedBackend()
	ks := ethwallettest.GetKeystore()
	txAccount := wallettest.NewRandomAccount(rng).(*ethwallet.Account)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	simBackend.FundAddress(ctx, txAccount.Account.Address)
	contractBackend := ethchannel.NewContractBackend(simBackend, ks, txAccount.Account)

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
func NewSetup(t *testing.T, rng *rand.Rand, n int) *Setup {
	s := &Setup{
		SimSetup: *NewSimSetup(rng),
		Accs:     make([]*ethwallet.Account, n),
		Parts:    make([]wallet.Address, n),
		Recvs:    make([]*ethwallet.Address, n),
		Funders:  make([]*ethchannel.Funder, n),
		Adjs:     make([]*SimAdjudicator, n),
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTxTimeout)
	defer cancel()
	adjudicator, err := ethchannel.DeployAdjudicator(ctx, *s.CB)
	require.NoError(t, err)
	s.Asset, err = ethchannel.DeployETHAssetholder(ctx, *s.CB, adjudicator)
	require.NoError(t, err)
	t.Logf("asset holder address is %v", s.Asset)
	t.Logf("adjudicator address is %v", adjudicator)

	for i := 0; i < n; i++ {
		s.Accs[i] = wallettest.NewRandomAccount(rng).(*ethwallet.Account)
		s.Parts[i] = s.Accs[i].Address()
		s.SimBackend.FundAddress(ctx, s.Accs[i].Account.Address)
		s.Recvs[i] = wallettest.NewRandomAddress(rng).(*ethwallet.Address)
		cb := ethchannel.NewContractBackend(s.SimBackend, ethwallettest.GetKeystore(), s.Accs[i].Account)
		s.Funders[i] = ethchannel.NewETHFunder(cb, s.Asset)
		s.Adjs[i] = NewSimAdjudicator(cb, adjudicator, s.Recvs[i].Address)
	}

	return s
}

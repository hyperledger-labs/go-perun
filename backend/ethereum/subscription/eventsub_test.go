// Copyright 2021 - See NOTICE file for copyright holders.
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

package subscription_test

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/backend/ethereum/bindings"
	"perun.network/go-perun/backend/ethereum/bindings/assetholder"
	"perun.network/go-perun/backend/ethereum/bindings/assetholdereth"
	"perun.network/go-perun/backend/ethereum/bindings/peruntoken"
	ethchannel "perun.network/go-perun/backend/ethereum/channel"
	"perun.network/go-perun/backend/ethereum/channel/test"
	"perun.network/go-perun/backend/ethereum/subscription"
	"perun.network/go-perun/backend/ethereum/wallet/keystore"
	channeltest "perun.network/go-perun/channel/test"
	"perun.network/go-perun/log"
	wallettest "perun.network/go-perun/wallet/test"
	pkgtest "polycry.pt/poly-go/test"
)

const (
	txGasLimit      = 100000
	txFinalityDepth = 1
)

// TestEventSub tests the `EventSub` by:
// 1. Emit `1/4 n` events
// 2. Starting up the `EventSub`
// 3. Emit `3/4 n` events
// 4. Checking that `EventSub` contains `n` distinct events.
func TestEventSub(t *testing.T) {
	n := 1000
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	rng := pkgtest.Prng(t)

	// Simulated chain setup.
	sb := test.NewSimulatedBackend()
	ksWallet := wallettest.RandomWallet().(*keystore.Wallet)
	account := &ksWallet.NewRandomAccount(rng).(*keystore.Account).Account
	sb.FundAddress(ctx, account.Address)
	cb := ethchannel.NewContractBackend(
		sb,
		keystore.NewTransactor(*ksWallet, types.NewEIP155Signer(big.NewInt(1337))),
		txFinalityDepth,
	)

	// Setup Perun Token.
	tokenAddr, err := ethchannel.DeployPerunToken(ctx, cb, *account, []common.Address{account.Address}, channeltest.MaxBalance)
	require.NoError(t, err)
	token, err := peruntoken.NewERC20(tokenAddr, cb)
	require.NoError(t, err)
	ct := pkgtest.NewConcurrent(t)

	// Sync channel to ensure that at least n/4 events were sent.
	waitSent := make(chan interface{})
	go ct.Stage("emitter", func(t pkgtest.ConcT) {
		for i := 0; i < n; i++ {
			if i == n/4 {
				close(waitSent)
			}
			log.Debug("Sending ", i)
			// Send the transaction.
			opts, err := cb.NewTransactor(ctx, txGasLimit, *account)
			require.NoError(t, err)
			tx, err := token.IncreaseAllowance(opts, account.Address, big.NewInt(1))
			require.NoError(t, err)
			// Wait for the TX to be mined.
			_, err = cb.ConfirmTransaction(ctx, tx, *account)
			require.NoError(t, err)
		}
	})
	sink := make(chan *subscription.Event, 10)
	eFact := func() *subscription.Event {
		return &subscription.Event{
			Name: bindings.Events.ERC20Approval,
			Data: new(peruntoken.ERC20Approval),
		}
	}
	// Setup the event sub after some events have been sent.
	<-waitSent
	contract := bind.NewBoundContract(tokenAddr, bindings.ABI.ERC20Token, cb, cb, cb)
	sub, err := subscription.NewEventSub(ctx, cb, contract, eFact, 10000)
	require.NoError(t, err)
	go ct.Stage("sub", func(t pkgtest.ConcT) {
		defer close(sink)
		require.NoError(t, sub.Read(context.Background(), sink))
	})

	go ct.Stage("receiver", func(t pkgtest.ConcT) {
		var lastTx common.Hash
		// Receive `n` unique events.
		for i := 0; i < n; i++ {
			e := <-sink
			log.Debug("Read ", i)
			require.NotNil(t, e)
			// It is possible to receive the same event twice.
			if e.Log.TxHash == lastTx {
				i--
			}
			lastTx = e.Log.TxHash

			want := &peruntoken.ERC20Approval{
				Owner:   account.Address,
				Spender: account.Address,
				Value:   big.NewInt(int64(i + 1)),
			}
			require.Equal(t, want, e.Data)
			require.False(t, e.Log.Removed)
		}
		sub.Close()
	})

	ct.Wait("emitter", "sub", "receiver")
	// Check that read terminated.
	require.Nil(t, <-sink)
}

// TestEventSub_Filter checks that the EventSub filters transactions.
func TestEventSub_Filter(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	rng := pkgtest.Prng(t)

	// Simulated chain setup.
	sb := test.NewSimulatedBackend()
	ksWallet := wallettest.RandomWallet().(*keystore.Wallet)
	account := &ksWallet.NewRandomAccount(rng).(*keystore.Account).Account
	sb.FundAddress(ctx, account.Address)
	cb := ethchannel.NewContractBackend(
		sb,
		keystore.NewTransactor(*ksWallet, types.NewEIP155Signer(big.NewInt(1337))),
		txFinalityDepth,
	)

	// Setup ETH AssetHolder.
	adjAddr, err := ethchannel.DeployAdjudicator(ctx, cb, *account)
	require.NoError(t, err)
	ahAddr, err := ethchannel.DeployETHAssetholder(ctx, cb, adjAddr, *account)
	require.NoError(t, err)
	ah, err := assetholdereth.NewAssetHolder(ahAddr, cb)
	require.NoError(t, err)
	ct := pkgtest.NewConcurrent(t)

	// Send the transaction.
	fundingID := channeltest.NewRandomChannelID(rng)
	opts, err := cb.NewTransactor(ctx, txGasLimit, *account)
	require.NoError(t, err)
	opts.Value = big.NewInt(1)
	tx, err := ah.Deposit(opts, fundingID, big.NewInt(1))
	require.NoError(t, err)
	// Wait for the TX to be mined.
	_, err = cb.ConfirmTransaction(ctx, tx, *account)
	require.NoError(t, err)

	// Create the filter.
	Filter := []interface{}{fundingID}
	// Setup the event sub.
	sink := make(chan *subscription.Event, 1)
	eFact := func() *subscription.Event {
		return &subscription.Event{
			Name:   bindings.Events.AhDeposited,
			Data:   new(assetholder.AssetHolderDeposited),
			Filter: [][]interface{}{Filter},
		}
	}
	contract := bind.NewBoundContract(ahAddr, bindings.ABI.AssetHolder, cb, cb, cb)
	sub, err := subscription.NewEventSub(ctx, cb, contract, eFact, 100)
	require.NoError(t, err)
	go ct.Stage("sub", func(t pkgtest.ConcT) {
		defer close(sink)
		require.NoError(t, sub.Read(context.Background(), sink))
	})

	// Receive 1 event.
	e := <-sink
	require.NotNil(t, e)
	want := &assetholder.AssetHolderDeposited{
		FundingID: fundingID,
		Amount:    big.NewInt(int64(1)),
	}
	log.Debug("TX0 Hash: ", e.Log.TxHash)
	require.Equal(t, want, e.Data)
	require.False(t, e.Log.Removed)
	sub.Close()
	// We do not check here that <-sink returns nil, since the EventSub
	// can receive events more than once.
}

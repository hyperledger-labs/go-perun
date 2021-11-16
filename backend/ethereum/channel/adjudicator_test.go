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

package channel_test

import (
	"context"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ethchannel "perun.network/go-perun/backend/ethereum/channel"
	"perun.network/go-perun/backend/ethereum/channel/test"
	"perun.network/go-perun/backend/ethereum/wallet/keystore"
	ethwallettest "perun.network/go-perun/backend/ethereum/wallet/test"
	"perun.network/go-perun/channel"
	channeltest "perun.network/go-perun/channel/test"
	pkgtest "polycry.pt/poly-go/test"
)

const defaultTxTimeout = 2 * time.Second

func testSignState(t *testing.T, accounts []*keystore.Account, state *channel.State) channel.Transaction {
	tx, err := signState(accounts, state)
	assert.NoError(t, err, "Sign should not return error")
	return tx
}

func signState(accounts []*keystore.Account, state *channel.State) (channel.Transaction, error) {
	sigs := make([][]byte, len(accounts))
	for i := range accounts {
		sig, err := channel.Sign(accounts[i], state)
		if err != nil {
			return channel.Transaction{}, errors.WithMessagef(err, "signing with account %d", i)
		}
		sigs[i] = sig
	}
	return channel.Transaction{
		State: state.Clone(),
		Sigs:  sigs,
	}, nil
}

func TestSubscribeRegistered(t *testing.T) {
	rng := pkgtest.Prng(t)
	// create test setup
	s := test.NewSetup(t, rng, 1, blockInterval, TxFinalityDepth)
	// create valid state and params
	params, state := channeltest.NewRandomParamsAndState(
		rng,
		channeltest.WithChallengeDuration(uint64(100*time.Second)),
		channeltest.WithParts(s.Parts...),
		channeltest.WithAssets((*ethchannel.Asset)(&s.Asset)),
		channeltest.WithIsFinal(false),
		channeltest.WithLedgerChannel(true),
		channeltest.WithVirtualChannel(false),
	)
	// Set up subscription
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	registered, err := s.Adjs[0].Subscribe(ctx, params.ID())
	require.NoError(t, err, "Subscribing to valid params should not error")
	// we need to properly fund the channel
	txCtx, txCancel := context.WithTimeout(context.Background(), defaultTxTimeout)
	defer txCancel()
	// fund the contract
	reqFund := channel.NewFundingReq(params, state, channel.Index(0), state.Balances)
	require.NoError(t, s.Funders[0].Fund(txCtx, *reqFund), "funding should succeed")
	// create subscription
	adj := s.Adjs[0]
	sub, err := adj.Subscribe(ctx, params.ID())
	require.NoError(t, err)
	defer sub.Close()
	// Now test the register function
	tx := testSignState(t, s.Accs, state)
	req := channel.AdjudicatorReq{
		Params: params,
		Acc:    s.Accs[0],
		Idx:    channel.Index(0),
		Tx:     tx,
	}
	assert.NoError(t, adj.Register(txCtx, req, nil), "Registering state should succeed")
	event := sub.Next()
	assert.Equal(t, event, registered.Next(), "Events should be equal")
	assert.NoError(t, registered.Close(), "Closing event channel should not error")
	assert.Nil(t, registered.Next(), "Next on closed channel should produce nil")
	assert.NoError(t, registered.Err(), "Closing should produce no error")
	// Setup a new subscription
	registered2, err := adj.Subscribe(ctx, params.ID())
	assert.NoError(t, err, "registering two subscriptions should not fail")
	assert.Equal(t, event, registered2.Next(), "Events should be equal")
	assert.NoError(t, registered2.Close(), "Closing event channel should not error")
	assert.Nil(t, registered2.Next(), "Next on closed channel should produce nil")
	assert.NoError(t, registered2.Err(), "Closing should produce no error")
}

func TestValidateAdjudicator(t *testing.T) {
	// Test setup
	rng := pkgtest.Prng(t)
	s := test.NewSimSetup(t, rng, TxFinalityDepth, blockInterval)

	t.Run("no_adj_code", func(t *testing.T) {
		randomAddr := (common.Address)(ethwallettest.NewRandomAddress(rng))
		ctx, cancel := context.WithTimeout(context.Background(), defaultTxTimeout)
		defer cancel()
		require.True(t, ethchannel.IsErrInvalidContractCode(ethchannel.ValidateAdjudicator(ctx, *s.CB, randomAddr)))
	})
	t.Run("incorrect_adj_code", func(t *testing.T) {
		randomAddr := (common.Address)(ethwallettest.NewRandomAddress(rng))
		ctx, cancel := context.WithTimeout(context.Background(), defaultTxTimeout)
		defer cancel()
		incorrectCodeAddr, err := ethchannel.DeployETHAssetholder(ctx, *s.CB, randomAddr, s.TxSender.Account)
		require.NoError(t, err)
		require.True(t, ethchannel.IsErrInvalidContractCode(ethchannel.ValidateAdjudicator(ctx, *s.CB, incorrectCodeAddr)))
	})
	t.Run("correct_adj_code", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), defaultTxTimeout)
		defer cancel()
		adjudicatorAddr, err := ethchannel.DeployAdjudicator(ctx, *s.CB, s.TxSender.Account)
		require.NoError(t, err)
		t.Logf("adjudicator address is %v", adjudicatorAddr)
		require.NoError(t, ethchannel.ValidateAdjudicator(ctx, *s.CB, adjudicatorAddr))
	})
}

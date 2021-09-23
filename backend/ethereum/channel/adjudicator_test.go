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
	"math/rand"
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
	pkgtest "perun.network/go-perun/pkg/test"
)

// TestAdjudicator tests the adjudicator.
func TestAdjudicator(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTestTimeout)
	defer cancel()

	rng := pkgtest.Prng(t)
	numParts := 2 + rng.Intn(maxNumParts-2)
	s := test.NewSetup(t, rng, numParts, blockInterval)
	a := newAdjudicatorSetup(s)
	channeltest.TestAdjudicator(ctx, t, a)
}

type adjudicatorSetup struct {
	setup *test.Setup
}

func newAdjudicatorSetup(setup *test.Setup) *adjudicatorSetup {
	return &adjudicatorSetup{
		setup: setup,
	}
}

type testAdjudicator struct {
	*test.SimAdjudicator
}

// Progress is a no-op because app channels are not supported yet.
func (a *testAdjudicator) Progress(ctx context.Context, req channel.ProgressReq) error {
	return nil
}

func (a *adjudicatorSetup) Adjudicator() channel.Adjudicator {
	return &testAdjudicator{a.setup.Adjs[0]}
}

func (a *adjudicatorSetup) NewRegisterReq(ctx context.Context, rng *rand.Rand) (*channel.AdjudicatorReq, []channel.SignedState) {
	req := a.newAdjudicatorReq(ctx, rng, channeltest.WithIsFinal(false))
	return req, nil
}

func (a *adjudicatorSetup) newAdjudicatorReq(ctx context.Context, rng *rand.Rand, opts ...channeltest.RandomOpt) *channel.AdjudicatorReq {
	params, state := a.newRandomParamsAndState(rng, opts...)

	// Fund the channel.
	numParts := len(params.Parts)
	errs := make(chan error, numParts)
	for i, funder := range a.setup.Funders {
		req := channel.NewFundingReq(params, state, channel.Index(i), state.Balances)
		go func(funder channel.Funder, req channel.FundingReq) {
			errs <- funder.Fund(ctx, req)
		}(funder, *req)
	}
	for range a.setup.Funders {
		select {
		case err := <-errs:
			if err != nil {
				panic(err)
			}
		case <-ctx.Done():
			panic(ctx.Err())
		}
	}

	tx, err := signState(a.setup.Accs, state)
	if err != nil {
		panic(err)
	}
	return &channel.AdjudicatorReq{
		Acc:    a.setup.Accs[0],
		Idx:    0,
		Params: params,
		Tx:     tx,
	}
}

func (a *adjudicatorSetup) newRandomParamsAndState(rng *rand.Rand, opts ...channeltest.RandomOpt) (*channel.Params, *channel.State) {
	_opts := a.defaultOpts()
	_opts = append(_opts, opts...)
	return channeltest.NewRandomParamsAndState(rng, _opts...)
}

func (a *adjudicatorSetup) defaultOpts() []channeltest.RandomOpt {
	return []channeltest.RandomOpt{
		channeltest.WithChallengeDuration(100),
		channeltest.WithParts(a.setup.Parts...),
		channeltest.WithAssets((*ethchannel.Asset)(&a.setup.Asset)),
		channeltest.WithLedgerChannel(true),
		channeltest.WithVirtualChannel(false),
		channeltest.WithNumLocked(0),
	}
}

func (a *adjudicatorSetup) NewProgressReq(context.Context, *rand.Rand) *channel.ProgressReq {
	return &channel.ProgressReq{} // Progression test is not implemented.
}

func (a *adjudicatorSetup) NewWithdrawReq(ctx context.Context, rng *rand.Rand) (*channel.AdjudicatorReq, channel.StateMap) {
	adj := a.Adjudicator()
	req := a.newAdjudicatorReq(ctx, rng, channeltest.WithoutApp())
	subChannels := []channel.SignedState{}

	if !req.Tx.IsFinal {
		// Register.
		err := adj.Register(ctx, *req, subChannels)
		if err != nil {
			panic(err)
		}

		// Wait until concludable.
		sub, err := adj.Subscribe(ctx, req.Params.ID())
		if err != nil {
			panic(err)
		}
		err = sub.Next().Timeout().Wait(ctx)
		if err != nil {
			panic(err)
		}
	}

	return req, channeltest.MakeStateMapFromSignedStates(subChannels...)
}

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
	s := test.NewSetup(t, rng, 1, blockInterval)
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
	s := test.NewSimSetup(rng)

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

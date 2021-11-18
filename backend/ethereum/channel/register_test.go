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
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ethchannel "perun.network/go-perun/backend/ethereum/channel"
	"perun.network/go-perun/backend/ethereum/channel/test"
	"perun.network/go-perun/channel"
	channeltest "perun.network/go-perun/channel/test"
	pkgtest "polycry.pt/poly-go/test"
)

func TestAdjudicator_MultipleRegisters(t *testing.T) {
	testParallel := func(n int) {
		t.Run(fmt.Sprintf("Register %d party parallel", n), func(t *testing.T) { registerMultiple(t, n, true) })
	}
	testSequential := func(n int) {
		t.Run(fmt.Sprintf("Register %d party sequential", n), func(t *testing.T) { registerMultiple(t, n, false) })
	}

	for _, n := range []int{1, 2, 5} {
		testParallel(n)
		testSequential(n)
	}
}

func registerMultiple(t *testing.T, numParts int, parallel bool) {
	t.Helper()
	rng := pkgtest.Prng(t)
	// create test setup
	s := test.NewSetup(t, rng, numParts, blockInterval, TxFinalityDepth)
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

	// we need to properly fund the channel
	fundingCtx, funCancel := context.WithTimeout(context.Background(), defaultTxTimeout*time.Duration(numParts))
	defer funCancel()
	// fund the contract
	ct := pkgtest.NewConcurrent(t)
	for i, funder := range s.Funders {
		sleepTime := time.Millisecond * time.Duration(rng.Int63n(10)+1)
		i, funder := i, funder
		go ct.StageN("funding loop", numParts, func(rt pkgtest.ConcT) {
			time.Sleep(sleepTime)
			req := channel.NewFundingReq(params, state, channel.Index(i), state.Balances)
			require.NoError(rt, funder.Fund(fundingCtx, *req), "funding should succeed")
		})
	}
	ct.Wait("funding loop")

	// Now test the register function
	ctx, cancel := context.WithTimeout(context.Background(), defaultTxTimeout)
	defer cancel()
	var wg sync.WaitGroup
	startBarrier := make(chan struct{})
	if parallel {
		wg.Add(numParts)
	}

	txs := make([]*channel.Transaction, numParts)
	subs := make([]channel.AdjudicatorSubscription, numParts)
	for i := 0; i < numParts; i++ {
		state.Version = uint64(int(state.Version) + i) // manipulate the state
		tx := testSignState(t, s.Accs, state)
		txs[i] = &tx

		sleepDuration := time.Duration(rng.Int63n(10)+1) * time.Millisecond
		reg := func(i int, tx channel.Transaction) {
			if parallel {
				defer wg.Done()
				<-startBarrier
				time.Sleep(sleepDuration)
			}

			// create subscription
			adj := s.Adjs[i]
			sub, err := adj.Subscribe(ctx, params.ID())
			require.NoError(t, err)
			subs[i] = sub

			// register
			req := channel.AdjudicatorReq{
				Params: params,
				Tx:     tx,
			}
			err = adj.Register(ctx, req, nil)
			require.True(t, err == nil || ethchannel.IsErrTxFailed(err), "Registering peer %d", i)
		}
		if parallel {
			go reg(i, tx)
		} else {
			reg(i, tx)
		}
	}

	if parallel {
		close(startBarrier)
		wg.Wait()
	}

	time.Sleep(100 * time.Millisecond) // Give the subscriptions time to receive the latest event.

	// Check result.
	for i := 0; i < numParts; i++ {
		tx, sub := txs[i], subs[i]
		event := sub.Next()
		sub.Close()
		require.NotEqual(t, event, &channel.RegisteredEvent{}, "registering should return valid event")
		require.GreaterOrEqualf(t, event.Version(), tx.State.Version, "peer %d: expected version >= %d, got %d", i, event.Version(), tx.State.Version)
		require.False(t, event.Timeout().IsElapsed(ctx),
			"registering non-final state should return unelapsed timeout")
		t.Logf("Peer[%d] registered successfully", i)
	}
}

func TestRegister_FinalState(t *testing.T) {
	rng := pkgtest.Prng(t)
	// create new Adjudicator
	s := test.NewSetup(t, rng, 1, blockInterval, TxFinalityDepth)
	// create valid state and params
	params, state := channeltest.NewRandomParamsAndState(rng, channeltest.WithChallengeDuration(uint64(100*time.Second)), channeltest.WithParts(s.Parts...), channeltest.WithAssets((*ethchannel.Asset)(&s.Asset)), channeltest.WithIsFinal(true), channeltest.WithLedgerChannel(true))
	// we need to properly fund the channel
	fundingCtx, funCancel := context.WithTimeout(context.Background(), defaultTxTimeout)
	defer funCancel()
	// fund the contract
	reqFund := channel.NewFundingReq(params, state, channel.Index(0), state.Balances)
	require.NoError(t, s.Funders[0].Fund(fundingCtx, *reqFund), "funding should succeed")
	// Now test the register function
	ctx, cancel := context.WithTimeout(context.Background(), defaultTxTimeout)
	defer cancel()
	// create subscription
	adj := s.Adjs[0]
	sub, err := adj.Subscribe(ctx, params.ID())
	require.NoError(t, err)
	defer sub.Close()
	// register
	tx := testSignState(t, s.Accs, state)
	req := channel.AdjudicatorReq{
		Params: params,
		Acc:    s.Accs[0],
		Idx:    channel.Index(0),
		Tx:     tx,
	}
	assert.NoError(t, adj.Register(ctx, req, nil), "Registering final state should succeed")
	event := sub.Next()
	assert.NotEqual(t, event, &channel.RegisteredEvent{}, "registering should return valid event")
	assert.True(t, event.Timeout().IsElapsed(ctx), "registering final state should return elapsed timeout")
	t.Logf("Peer[%d] registered successful", 0)
}

func TestRegister_CancelledContext(t *testing.T) {
	rng := pkgtest.Prng(t)
	// create test setup
	s := test.NewSetup(t, rng, 1, blockInterval, TxFinalityDepth)
	// create valid state and params
	params, state := channeltest.NewRandomParamsAndState(rng, channeltest.WithChallengeDuration(uint64(100*time.Second)), channeltest.WithParts(s.Parts...), channeltest.WithAssets((*ethchannel.Asset)(&s.Asset)), channeltest.WithIsFinal(false), channeltest.WithLedgerChannel(true))
	// we need to properly fund the channel
	fundingCtx, funCancel := context.WithTimeout(context.Background(), defaultTxTimeout)
	defer funCancel()
	// fund the contract
	reqFund := channel.NewFundingReq(params, state, channel.Index(0), state.Balances)
	require.NoError(t, s.Funders[0].Fund(fundingCtx, *reqFund), "funding should succeed")
	// Now test the register function
	ctx, cancel := context.WithTimeout(context.Background(), defaultTxTimeout)
	// directly cancel timeout
	cancel()
	// create subscription
	adj := s.Adjs[0]
	sub, err := adj.Subscribe(ctx, params.ID())
	require.NoError(t, err)
	defer sub.Close()
	// register
	tx := testSignState(t, s.Accs, state)
	req := channel.AdjudicatorReq{
		Params: params,
		Acc:    s.Accs[0],
		Idx:    channel.Index(0),
		Tx:     tx,
	}
	assert.Error(t, adj.Register(ctx, req, nil), "Registering with canceled context should error")
}

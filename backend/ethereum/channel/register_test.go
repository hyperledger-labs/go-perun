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
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ethchannel "perun.network/go-perun/backend/ethereum/channel"
	"perun.network/go-perun/backend/ethereum/channel/test"
	"perun.network/go-perun/channel"
	channeltest "perun.network/go-perun/channel/test"
	pkgtest "perun.network/go-perun/pkg/test"
)

func TestAdjudicator_MultipleRegisters(t *testing.T) {
	t.Run("Register 1 party parallel", func(t *testing.T) { registerMultipleConcurrent(t, 1, true) })
	t.Run("Register 2 party parallel", func(t *testing.T) { registerMultipleConcurrent(t, 2, true) })
	t.Run("Register 5 party parallel", func(t *testing.T) { registerMultipleConcurrent(t, 5, true) })
	t.Run("Register 10 party parallel", func(t *testing.T) { registerMultipleConcurrent(t, 10, true) })
	t.Run("Register 1 party sequential", func(t *testing.T) { registerMultipleConcurrent(t, 1, false) })
	t.Run("Register 2 party sequential", func(t *testing.T) { registerMultipleConcurrent(t, 2, false) })
	t.Run("Register 5 party sequential", func(t *testing.T) { registerMultipleConcurrent(t, 5, false) })
}

func registerMultipleConcurrent(t *testing.T, numParts int, parallel bool) {
	rng := pkgtest.Prng(t)
	// create test setup
	s := test.NewSetup(t, rng, numParts)
	// create valid state and params
	params, state := channeltest.NewRandomParamsAndState(rng, channeltest.WithChallengeDuration(uint64(100*time.Second)), channeltest.WithParts(s.Parts...), channeltest.WithAssets((*ethchannel.Asset)(&s.Asset)), channeltest.WithIsFinal(false))

	// we need to properly fund the channel
	fundingCtx, funCancel := context.WithTimeout(context.Background(), defaultTxTimeout)
	defer funCancel()
	// fund the contract
	ct := pkgtest.NewConcurrent(t)
	for i, funder := range s.Funders {
		sleepTime := time.Duration(rng.Int63n(10) + 1)
		i, funder := i, funder
		go ct.StageN("funding loop", numParts, func(rt pkgtest.ConcT) {
			time.Sleep(sleepTime * time.Millisecond)
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
	for i := 0; i < numParts; i++ {
		sleepDuration := time.Duration(rng.Int63n(10)+1) * time.Millisecond
		// manipulate the state
		state.Version = uint64(int(state.Version) + i)
		tx := testSignState(t, s.Accs, params, state)
		reg := func(i int, tx channel.Transaction) {
			if parallel {
				defer wg.Done()
				<-startBarrier
				time.Sleep(sleepDuration)
			}
			// create subscription
			adj := s.Adjs[i]
			sub, err := adj.Subscribe(ctx, params)
			require.NoError(t, err)
			defer sub.Close()
			// register
			req := channel.AdjudicatorReq{
				Params: params,
				Acc:    s.Accs[i],
				Idx:    channel.Index(i),
				Tx:     tx,
			}
			assert.NoError(t, adj.Register(ctx, req), "Registering should succeed")
			event := sub.Next()
			assert.NotEqual(t, event, &channel.RegisteredEvent{}, "registering should return valid event")
			assert.False(t, event.Timeout().IsElapsed(ctx),
				"registering non-final state should return unelapsed timeout")
			t.Logf("Peer[%d] registered successful", i)
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
}

func TestRegister_FinalState(t *testing.T) {
	rng := pkgtest.Prng(t)
	// create new Adjudicator
	s := test.NewSetup(t, rng, 1)
	// create valid state and params
	params, state := channeltest.NewRandomParamsAndState(rng, channeltest.WithChallengeDuration(uint64(100*time.Second)), channeltest.WithParts(s.Parts...), channeltest.WithAssets((*ethchannel.Asset)(&s.Asset)), channeltest.WithIsFinal(true))
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
	sub, err := adj.Subscribe(ctx, params)
	require.NoError(t, err)
	defer sub.Close()
	// register
	tx := testSignState(t, s.Accs, params, state)
	req := channel.AdjudicatorReq{
		Params: params,
		Acc:    s.Accs[0],
		Idx:    channel.Index(0),
		Tx:     tx,
	}
	assert.NoError(t, adj.Register(ctx, req), "Registering final state should succeed")
	event := sub.Next()
	assert.NotEqual(t, event, &channel.RegisteredEvent{}, "registering should return valid event")
	assert.True(t, event.Timeout().IsElapsed(ctx), "registering final state should return elapsed timeout")
	t.Logf("Peer[%d] registered successful", 0)
}

func TestRegister_CancelledContext(t *testing.T) {
	rng := pkgtest.Prng(t)
	// create test setup
	s := test.NewSetup(t, rng, 1)
	// create valid state and params
	params, state := channeltest.NewRandomParamsAndState(rng, channeltest.WithChallengeDuration(uint64(100*time.Second)), channeltest.WithParts(s.Parts...), channeltest.WithAssets((*ethchannel.Asset)(&s.Asset)), channeltest.WithIsFinal(false))
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
	sub, err := adj.Subscribe(ctx, params)
	require.NoError(t, err)
	defer sub.Close()
	// register
	tx := testSignState(t, s.Accs, params, state)
	req := channel.AdjudicatorReq{
		Params: params,
		Acc:    s.Accs[0],
		Idx:    channel.Index(0),
		Tx:     tx,
	}
	assert.Error(t, adj.Register(ctx, req), "Registering with canceled context should error")
}

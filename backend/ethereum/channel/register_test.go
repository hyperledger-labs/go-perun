// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package channel_test

import (
	"context"
	"math/rand"
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
	seed := time.Now().UnixNano()
	t.Logf("seed is %v", seed)
	if parallel {
		seed++
	}
	rng := rand.New(rand.NewSource(int64(seed)))
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
		go ct.StageN("funding loop", numParts, func(rt require.TestingT) {
			time.Sleep(sleepTime * time.Millisecond)
			req := channel.FundingReq{
				Params: params,
				State:  state,
				Idx:    channel.Index(i),
			}
			require.NoError(rt, funder.Fund(fundingCtx, req), "funding should succeed")
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
		tx := signState(t, s.Accs, params, state)
		reg := func(i int, tx channel.Transaction) {
			if parallel {
				defer wg.Done()
				<-startBarrier
				time.Sleep(sleepDuration)
			}
			req := channel.AdjudicatorReq{
				Params: params,
				Acc:    s.Accs[i],
				Idx:    channel.Index(i),
				Tx:     tx,
			}
			event, err := s.Adjs[i].Register(ctx, req)
			assert.NoError(t, err, "Registering should succeed")
			assert.NotEqual(t, event, &channel.RegisteredEvent{}, "registering should return valid event")
			assert.False(t, event.Timeout.IsElapsed(ctx),
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
	seed := time.Now().UnixNano()
	t.Logf("seed is %v", seed)
	rng := rand.New(rand.NewSource(int64(seed)))
	// create new Adjudicator
	s := test.NewSetup(t, rng, 1)
	// create valid state and params
	params, state := channeltest.NewRandomParamsAndState(rng, channeltest.WithChallengeDuration(uint64(100*time.Second)), channeltest.WithParts(s.Parts...), channeltest.WithAssets((*ethchannel.Asset)(&s.Asset)), channeltest.WithIsFinal(true))
	// we need to properly fund the channel
	fundingCtx, funCancel := context.WithTimeout(context.Background(), defaultTxTimeout)
	defer funCancel()
	// fund the contract
	reqFund := channel.FundingReq{
		Params: params,
		State:  state,
		Idx:    channel.Index(0),
	}
	require.NoError(t, s.Funders[0].Fund(fundingCtx, reqFund), "funding should succeed")
	// Now test the register function
	ctx, cancel := context.WithTimeout(context.Background(), defaultTxTimeout)
	defer cancel()
	tx := signState(t, s.Accs, params, state)
	req := channel.AdjudicatorReq{
		Params: params,
		Acc:    s.Accs[0],
		Idx:    channel.Index(0),
		Tx:     tx,
	}
	event, err := s.Adjs[0].Register(ctx, req)
	assert.NoError(t, err, "Registering final state should succeed")
	assert.NotEqual(t, event, &channel.RegisteredEvent{}, "registering should return valid event")
	assert.True(t, event.Timeout.IsElapsed(ctx), "registering final state should return elapsed timeout")
	t.Logf("Peer[%d] registered successful", 0)
}

func TestRegister_CancelledContext(t *testing.T) {
	seed := time.Now().UnixNano()
	t.Logf("seed is %v", seed)
	rng := rand.New(rand.NewSource(int64(seed)))
	// create test setup
	s := test.NewSetup(t, rng, 1)
	// create valid state and params
	params, state := channeltest.NewRandomParamsAndState(rng, channeltest.WithChallengeDuration(uint64(100*time.Second)), channeltest.WithParts(s.Parts...), channeltest.WithAssets((*ethchannel.Asset)(&s.Asset)), channeltest.WithIsFinal(false))
	// we need to properly fund the channel
	fundingCtx, funCancel := context.WithTimeout(context.Background(), defaultTxTimeout)
	defer funCancel()
	// fund the contract
	reqFund := channel.FundingReq{
		Params: params,
		State:  state,
		Idx:    channel.Index(0),
	}
	require.NoError(t, s.Funders[0].Fund(fundingCtx, reqFund), "funding should succeed")
	// Now test the register function
	ctx, cancel := context.WithTimeout(context.Background(), defaultTxTimeout)
	// directly cancel timeout
	cancel()
	tx := signState(t, s.Accs, params, state)
	req := channel.AdjudicatorReq{
		Params: params,
		Acc:    s.Accs[0],
		Idx:    channel.Index(0),
		Tx:     tx,
	}
	event, err := s.Adjs[0].Register(ctx, req)
	assert.Error(t, err, "Registering with canceled context should error")
	assert.Nil(t, event, "should not produce valid event")
}

// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package channel

import (
	"context"
	"math/big"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"perun.network/go-perun/channel"
	channeltest "perun.network/go-perun/channel/test"
	pkgtest "perun.network/go-perun/pkg/test"
)

func TestAdjudicator_MultipleWithdraws_FinalState(t *testing.T) {
	t.Run("Withdraw 1 party parallel", func(t *testing.T) { withdrawMultipleConcurrentFinal(t, 1, true) })
	t.Run("Withdraw 2 party parallel", func(t *testing.T) { withdrawMultipleConcurrentFinal(t, 2, true) })
	t.Run("Withdraw 5 party parallel", func(t *testing.T) { withdrawMultipleConcurrentFinal(t, 5, true) })
	t.Run("Withdraw 10 party parallel", func(t *testing.T) { withdrawMultipleConcurrentFinal(t, 10, true) })
	t.Run("Withdraw 1 party sequential", func(t *testing.T) { withdrawMultipleConcurrentFinal(t, 1, false) })
	t.Run("Withdraw 2 party sequential", func(t *testing.T) { withdrawMultipleConcurrentFinal(t, 2, false) })
	t.Run("Withdraw 5 party sequential", func(t *testing.T) { withdrawMultipleConcurrentFinal(t, 5, false) })
}

func withdrawMultipleConcurrentFinal(t *testing.T, numParts int, parallel bool) {
	seed := time.Now().UnixNano()
	t.Logf("seed is %v", seed)
	if parallel {
		seed++
	}
	rng := rand.New(rand.NewSource(int64(seed)))
	// create new Adjudicators
	accs, parts, funders, adjs, asset := newAdjudicator(t, rng, numParts)
	// create valid state and params
	app := channeltest.NewRandomApp(rng)
	params := channel.NewParamsUnsafe(uint64(0), parts, app.Def(), big.NewInt(rng.Int63()))
	state := newValidState(rng, params, asset)
	// we need to properly fund the channel
	fundingCtx, funCancel := context.WithTimeout(context.Background(), timeout)
	defer funCancel()
	// fund the contract
	ct := pkgtest.NewConcurrent(t)
	for i, funder := range funders {
		sleepTime := time.Duration(rng.Int63n(10) + 1)
		i, funder := i, funder
		go ct.StageN("funding loop", numParts, func(rt require.TestingT) {
			time.Sleep(sleepTime * time.Millisecond)
			req := channel.FundingReq{
				Params:     params,
				Allocation: &state.Allocation,
				Idx:        channel.Index(i),
			}
			require.NoError(rt, funder.Fund(fundingCtx, req), "funding should succeed")
		})
	}
	ct.Wait("funding loop")
	// manipulate the state
	state.IsFinal = true
	tx := signState(t, accs, params, state)

	// Now test the withdraw function
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if parallel {
		startBarrier := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(numParts)
		for i := 0; i < numParts; i++ {
			sleepDuration := time.Duration(rng.Int63n(10)+1) * time.Millisecond
			go func(i int) {
				defer wg.Done()
				<-startBarrier
				time.Sleep(sleepDuration)
				req := channel.AdjudicatorReq{
					Params: params,
					Acc:    accs[i],
					Idx:    channel.Index(i),
					Tx:     tx,
				}
				err := adjs[i].Withdraw(ctx, req)
				assert.NoError(t, err, "Withdrawing should succeed")
			}(i)
		}
		close(startBarrier)
		wg.Wait()
	} else {
		for i := 0; i < numParts; i++ {
			req := channel.AdjudicatorReq{
				Params: params,
				Acc:    accs[i],
				Idx:    channel.Index(i),
				Tx:     tx,
			}
			err := adjs[i].Withdraw(ctx, req)
			assert.NoError(t, err, "Withdrawing should succeed")
		}
	}
	assertHoldingsZero(ctx, t, funders, params, state)
}

func TestWithdrawNonFinalState(t *testing.T) {
	seed := time.Now().UnixNano()
	t.Logf("seed is %v", seed)
	rng := rand.New(rand.NewSource(int64(seed)))
	// create new Adjudicators
	accs, parts, funders, adjs, asset := newAdjudicator(t, rng, 1)
	// create valid state and params
	app := channeltest.NewRandomApp(rng)
	params := channel.NewParamsUnsafe(uint64(0), parts, app.Def(), big.NewInt(rng.Int63()))
	state := newValidState(rng, params, asset)
	// we need to properly fund the channel
	fundingCtx, funCancel := context.WithTimeout(context.Background(), timeout)
	defer funCancel()
	// fund the contract
	fundingReq := channel.FundingReq{
		Params:     params,
		Allocation: &state.Allocation,
		Idx:        channel.Index(0),
	}
	require.NoError(t, funders[0].Fund(fundingCtx, fundingReq), "funding should succeed")
	// try to withdraw non-final state
	tx := signState(t, accs, params, state)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	req := channel.AdjudicatorReq{
		Params: params,
		Acc:    accs[0],
		Idx:    channel.Index(0),
		Tx:     tx,
	}
	err := adjs[0].Withdraw(ctx, req)
	assert.Error(t, err, "Withdrawing non-final state should fail")
}

func assertHoldingsZero(ctx context.Context, t *testing.T, funders []*Funder, params *channel.Params, state *channel.State) {
	for i, funder := range funders {
		req := channel.FundingReq{
			Params:     params,
			Allocation: &state.Allocation,
			Idx:        channel.Index(i),
		}
		alloc, err := getOnChainAllocation(ctx, funder, req)
		assert.NoError(t, err, "Getting on-chain allocs should succeed")
		sum := big.NewInt(0)
		for x := range alloc {
			for y := range alloc[x] {
				sum = sum.Add(sum, alloc[x][y])
			}
		}
		assert.Equal(t, big.NewInt(0), sum, "Holdings should be set to zero after withdraw")
	}
}

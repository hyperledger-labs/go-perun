// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package channel // import "perun.network/go-perun/backend/ethereum/channel"

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
	// create new Adjudicators
	accs, parts, funders, adjs, asset := newAdjudicator(t, rng, numParts)
	// create valid state and params
	app := channeltest.NewRandomApp(rng)
	params := channel.NewParamsUnsafe(uint64(100*time.Second), parts, app.Def(), big.NewInt(rng.Int63()))
	state := newValidState(rng, params, asset)
	// we need to properly fund the channel
	fundingCtx, funCancel := context.WithTimeout(context.Background(), timeout)
	defer funCancel()
	// fund the contract
	var wgFund sync.WaitGroup
	wgFund.Add(numParts)
	for i, funder := range funders {
		sleepTime := time.Duration(rng.Int63n(10) + 1)
		go func(i int, funder *Funder) {
			defer wgFund.Done()
			time.Sleep(sleepTime * time.Millisecond)
			req := channel.FundingReq{
				Params:     params,
				Allocation: &state.Allocation,
				Idx:        channel.Index(i),
			}
			require.NoError(t, funders[i].Fund(fundingCtx, req), "funding should succeed")
		}(i, funder)
	}
	wgFund.Wait()

	// Now test the register function
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
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
		tx := signState(t, accs, params, state)
		reg := func(i int, tx channel.Transaction) {
			if parallel {
				defer wg.Done()
				<-startBarrier
				time.Sleep(sleepDuration)
			}
			req := channel.AdjudicatorReq{
				Params: params,
				Acc:    accs[i],
				Idx:    channel.Index(i),
				Tx:     tx,
			}
			event, err := adjs[i].Register(ctx, req)
			assert.NoError(t, err, "Registering should succeed")
			assert.NotEqual(t, event, &channel.Registered{}, "registering should return valid event")
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
	accs, parts, funders, adjs, asset := newAdjudicator(t, rng, 1)
	// create valid state and params
	app := channeltest.NewRandomApp(rng)
	params := channel.NewParamsUnsafe(uint64(100*time.Second), parts, app.Def(), big.NewInt(rng.Int63()))
	state := newValidState(rng, params, asset)
	state.IsFinal = true
	// we need to properly fund the channel
	fundingCtx, funCancel := context.WithTimeout(context.Background(), timeout)
	defer funCancel()
	// fund the contract
	reqFund := channel.FundingReq{
		Params:     params,
		Allocation: &state.Allocation,
		Idx:        channel.Index(0),
	}
	require.NoError(t, funders[0].Fund(fundingCtx, reqFund), "funding should succeed")
	// Now test the register function
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	tx := signState(t, accs, params, state)
	req := channel.AdjudicatorReq{
		Params: params,
		Acc:    accs[0],
		Idx:    channel.Index(0),
		Tx:     tx,
	}
	event, err := adjs[0].Register(ctx, req)
	assert.NoError(t, err, "Registering final state should succeed")
	assert.NotEqual(t, event, &channel.Registered{}, "registering should return valid event")
	assert.Equal(t, time.Time{}, event.Timeout, "Registering final state should produce zero timeout")
	t.Logf("Peer[%d] registered successful", 0)
}

func TestRegister_CancelledContext(t *testing.T) {
	seed := time.Now().UnixNano()
	t.Logf("seed is %v", seed)
	rng := rand.New(rand.NewSource(int64(seed)))
	// create new Adjudicator
	accs, parts, funders, adjs, asset := newAdjudicator(t, rng, 1)
	// create valid state and params
	app := channeltest.NewRandomApp(rng)
	params := channel.NewParamsUnsafe(uint64(100*time.Second), parts, app.Def(), big.NewInt(rng.Int63()))
	state := newValidState(rng, params, asset)
	// we need to properly fund the channel
	fundingCtx, funCancel := context.WithTimeout(context.Background(), timeout)
	defer funCancel()
	// fund the contract
	reqFund := channel.FundingReq{
		Params:     params,
		Allocation: &state.Allocation,
		Idx:        channel.Index(0),
	}
	require.NoError(t, funders[0].Fund(fundingCtx, reqFund), "funding should succeed")
	// Now test the register function
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	// directly cancel timeout
	cancel()
	tx := signState(t, accs, params, state)
	req := channel.AdjudicatorReq{
		Params: params,
		Acc:    accs[0],
		Idx:    channel.Index(0),
		Tx:     tx,
	}
	event, err := adjs[0].Register(ctx, req)
	assert.Error(t, err, "Registering with canceled context should error")
	assert.Nil(t, event, "should not produce valid event")
}

// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package channel_test

import (
	"context"
	"math/big"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ethchannel "perun.network/go-perun/backend/ethereum/channel"
	"perun.network/go-perun/backend/ethereum/channel/test"
	ethwallet "perun.network/go-perun/backend/ethereum/wallet"
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
	// create test setup
	s := test.NewSetup(t, rng, numParts)
	// create valid state and params
	app := channeltest.NewRandomApp(rng)
	params := channel.NewParamsUnsafe(uint64(0), s.Parts, app.Def(), big.NewInt(rng.Int63()))
	state := newValidState(rng, params, s.Asset)
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
	tx := signState(t, s.Accs, params, state)

	// Now test the withdraw function
	ctx, cancel := context.WithTimeout(context.Background(), defaultTxTimeout)
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
					Acc:    s.Accs[i],
					Idx:    channel.Index(i),
					Tx:     tx,
				}
				err := s.Adjs[i].Withdraw(ctx, req)
				assert.NoError(t, err, "Withdrawing should succeed")
			}(i)
		}
		close(startBarrier)
		wg.Wait()
	} else {
		for i := 0; i < numParts; i++ {
			req := channel.AdjudicatorReq{
				Params: params,
				Acc:    s.Accs[i],
				Idx:    channel.Index(i),
				Tx:     tx,
			}
			err := s.Adjs[i].Withdraw(ctx, req)
			assert.NoError(t, err, "Withdrawing should succeed")
		}
	}
	assertHoldingsZero(ctx, t, s.CB, params, state.Assets)
}

func TestWithdrawZeroBalance(t *testing.T) {
	t.Run("1 Participant", func(t *testing.T) {
		testWithdrawZeroBalance(t, 1)
	})
	t.Run("2 Participant", func(t *testing.T) {
		testWithdrawZeroBalance(t, 2)
	})
}

// shouldFunders decides who should fund. 1 indicates funding, 0 indicates skipping.
func testWithdrawZeroBalance(t *testing.T, n int) {
	rng := rand.New(rand.NewSource(int64(0xDDD)))
	s := test.NewSetup(t, rng, n)
	// create valid state and params
	app := channeltest.NewRandomApp(rng)
	params := channel.NewParamsUnsafe(0, s.Parts, app.Def(), big.NewInt(rng.Int63()))
	state := newValidState(rng, params, s.Asset)
	state.IsFinal = true

	for i := range params.Parts {
		if i%2 == 0 {
			state.Allocation.Balances[0][i].Set(big.NewInt(0))
		} // is != 0 otherwise
		t.Logf("Part: %d ShouldFund: %t Bal: %v", i, i%2 == 1, state.Allocation.Balances[0][i])
	}

	// fund
	ct := pkgtest.NewConcurrent(t)
	for i, funder := range s.Funders {
		i, funder := i, funder
		go ct.StageN("funding loop", n, func(rt require.TestingT) {
			req := channel.FundingReq{
				Params:     params,
				Allocation: &state.Allocation,
				Idx:        channel.Index(i),
			}
			require.NoError(rt, funder.Fund(context.Background(), req), "funding should succeed")
		})
	}
	ct.Wait("funding loop")

	// register
	req := channel.AdjudicatorReq{
		Params: params,
		Acc:    s.Accs[0],
		Tx:     signState(t, s.Accs, params, state),
		Idx:    0,
	}
	_, err := s.Adjs[0].Register(context.Background(), req)
	require.NoError(t, err)
	// we don't need to wait for a timeout since we registered a final state

	// withdraw
	for i, adj := range s.Adjs {
		req.Acc = s.Accs[i]
		req.Idx = channel.Index(i)
		// check that the nonce stays the same for zero balance withdrawals
		diff, err := test.NonceDiff(s.Accs[i].Address(), adj, func() error {
			return adj.Withdraw(context.Background(), req)
		})
		require.NoError(t, err)
		if i%2 == 0 {
			assert.Zero(t, diff, "Nonce should stay the same")
		} else {
			assert.Equal(t, int(1), diff, "Nonce should increase by 1")
		}
	}
	assertHoldingsZero(context.Background(), t, s.CB, params, state.Assets)
}

func TestWithdraw(t *testing.T) {
	rng := rand.New(rand.NewSource(0xc007))
	// create test setup
	s := test.NewSetup(t, rng, 1)
	// create valid state and params
	app := channeltest.NewRandomApp(rng)
	params := channel.NewParamsUnsafe(uint64(0), s.Parts, app.Def(), big.NewInt(rng.Int63()))
	state := newValidState(rng, params, s.Asset)
	// we need to properly fund the channel
	fundingCtx, funCancel := context.WithTimeout(context.Background(), defaultTxTimeout)
	defer funCancel()
	// fund the contract
	fundingReq := channel.FundingReq{
		Params:     params,
		Allocation: &state.Allocation,
		Idx:        channel.Index(0),
	}
	require.NoError(t, s.Funders[0].Fund(fundingCtx, fundingReq), "funding should succeed")
	req := channel.AdjudicatorReq{
		Params: params,
		Acc:    s.Accs[0],
		Idx:    channel.Index(0),
	}

	testWithdraw := func(t *testing.T, shouldWork bool) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		req.Tx = signState(t, s.Accs, params, state)
		err := s.Adjs[0].Withdraw(ctx, req)

		if shouldWork {
			assert.NoError(t, err, "Withdrawing should work")
		} else {
			assert.Error(t, err, "Withdrawing should fail")
		}
	}

	t.Run("Withdraw non-final state", func(t *testing.T) {
		testWithdraw(t, false)
	})

	t.Run("Withdraw final state", func(t *testing.T) {
		state.IsFinal = true
		testWithdraw(t, true)
	})

	t.Run("Withdrawal idempotence", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			// get nonce
			oldNonce, err := s.Adjs[0].PendingNonceAt(context.Background(), s.Accs[0].Address().(*ethwallet.Address).Address)
			require.NoError(t, err)
			// withdraw
			testWithdraw(t, true)
			// get nonce
			nonce, err := s.Adjs[0].PendingNonceAt(context.Background(), s.Accs[0].Address().(*ethwallet.Address).Address)
			require.NoError(t, err)
			assert.Equal(t, oldNonce, nonce, "Nonce must not change in subsequent withdrawals")
		}
	})
}

func TestWithdrawNonFinal(t *testing.T) {
	assert := assert.New(t)
	rng := rand.New(rand.NewSource(0x7070707))
	// create test setup
	s := test.NewSetup(t, rng, 1)
	// create valid state and params
	app := channeltest.NewRandomApp(rng)
	params := channel.NewParamsUnsafe(60, s.Parts, app.Def(), big.NewInt(rng.Int63()))
	state := newValidState(rng, params, s.Asset)
	state.IsFinal = false // make sure the state is non-final

	ctx, cancel := context.WithTimeout(context.Background(), defaultTxTimeout)
	defer cancel()
	fundingReq := channel.FundingReq{
		Params:     params,
		Allocation: &state.Allocation,
		Idx:        channel.Index(0),
	}
	require.NoError(t, s.Funders[0].Fund(ctx, fundingReq), "funding should succeed")

	req := channel.AdjudicatorReq{
		Params: params,
		Acc:    s.Accs[0],
		Idx:    0,
		Tx:     signState(t, s.Accs, params, state),
	}
	reg, err := s.Adjs[0].Register(ctx, req)
	require.NoError(t, err)
	t.Log("Registered ", reg)
	assert.False(reg.Timeout.IsElapsed(),
		"registering non-final state should have non-elapsed timeout")
	assert.NoError(reg.Timeout.Wait(ctx))
	assert.True(reg.Timeout.IsElapsed(), "timeout should have elapsed after Wait()")
	assert.NoError(s.Adjs[0].Withdraw(ctx, req),
		"withdrawing should succeed after waiting for timeout")
}

func assertHoldingsZero(ctx context.Context, t *testing.T, cb *ethchannel.ContractBackend, params *channel.Params, _assets []channel.Asset) {
	alloc, err := getOnChainAllocation(ctx, cb, params, _assets)
	require.NoError(t, err, "Getting on-chain allocs should succeed")
	for i, assetalloc := range alloc {
		for j, a := range assetalloc {
			assert.Zerof(t, a.Sign(), "Allocation of asset[%d] and part[%d] non-zero.", j, i)
		}
	}
}
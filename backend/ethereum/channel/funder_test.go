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

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/backend/ethereum/bindings/assets"
	ethchannel "perun.network/go-perun/backend/ethereum/channel"
	"perun.network/go-perun/backend/ethereum/channel/test"
	ethwallet "perun.network/go-perun/backend/ethereum/wallet"
	ethwallettest "perun.network/go-perun/backend/ethereum/wallet/test"
	"perun.network/go-perun/channel"
	channeltest "perun.network/go-perun/channel/test"
	pkgtest "perun.network/go-perun/pkg/test"
	"perun.network/go-perun/wallet"
	wallettest "perun.network/go-perun/wallet/test"
)

const (
	keyDir   = "../wallet/testdata"
	password = "secret"
)

func TestFunderZeroBalance(t *testing.T) {
	t.Run("1 Participant", func(t *testing.T) {
		testFunderZeroBalance(t, 1)
	})
	t.Run("2 Participant", func(t *testing.T) {
		testFunderZeroBalance(t, 2)
	})
}

func testFunderZeroBalance(t *testing.T, n int) {
	rng := rand.New(rand.NewSource(0xDDD))
	parts, funders, _, params, allocation := newNFunders(context.Background(), t, rng, n)

	for i := range parts {
		if i%2 == 0 {
			allocation.Balances[0][i].Set(big.NewInt(0))
		} // is != 0 otherwise
		t.Logf("Part: %d ShouldFund: %t Bal: %v", i, i%2 == 1, allocation.Balances[0][i])
	}
	// fund
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		req := channel.FundingReq{
			Params:     params,
			Allocation: allocation,
			Idx:        channel.Index(i),
		}

		// Check that the funding only changes the nonce when the balance is not zero
		go func(i int) {
			defer wg.Done()
			diff, err := test.NonceDiff(parts[i], funders[i], func() error {
				return funders[i].Fund(context.Background(), req)
			})
			require.NoError(t, err)
			if i%2 == 0 {
				assert.Zero(t, diff, "Nonce should stay the same")
			} else {
				assert.Equal(t, int(1), diff, "Nonce should increase by 1")
			}
		}(i)
	}
	wg.Wait()
	// Check result balances
	assert.NoError(t, compareOnChainAlloc(params, *allocation, &funders[0].ContractBackend))
}

func TestFunder_Fund(t *testing.T) {
	// Need unique seed per run.
	seed := time.Now().UnixNano()
	t.Logf("seed is %d", seed)
	rng := rand.New(rand.NewSource(seed))
	ctx, cancel := context.WithTimeout(context.Background(), defaultTxTimeout)
	defer cancel()
	parts, funders, _, params, allocation := newNFunders(ctx, t, rng, 1)
	// Test invalid funding request
	assert.Panics(t, func() { funders[0].Fund(ctx, channel.FundingReq{}) }, "Funding with invalid funding req should fail")
	// Test funding without assets
	req := channel.FundingReq{
		Params:     &channel.Params{},
		Allocation: &channel.Allocation{},
		Idx:        0,
	}
	assert.NoError(t, funders[0].Fund(ctx, req), "Funding with no assets should succeed")
	// Test with valid request
	req = channel.FundingReq{
		Params:     params,
		Allocation: allocation,
		Idx:        0,
	}

	t.Run("Funding idempotence", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			nonce := 0
			if i == 0 {
				nonce = 1
			}
			diff, err := test.NonceDiff(parts[0], funders[0], func() error {
				return funders[0].Fund(context.Background(), req)
			})
			require.NoError(t, err, "Subsequent Fund calls should not modify the nonce")
			assert.Equal(t, nonce, diff)
		}
	})
	// Test already closed context
	cancel()
	assert.Error(t, funders[0].Fund(ctx, req), "funding with already cancelled context should fail")
	// Check result balances
	assert.NoError(t, compareOnChainAlloc(params, *allocation, &funders[0].ContractBackend))
}

func TestPeerTimedOutFundingError(t *testing.T) {
	t.Run("peer 0 faulty out of 2", func(t *testing.T) { testFundingTimeout(t, 0, 2) })
	t.Run("peer 1 faulty out of 2", func(t *testing.T) { testFundingTimeout(t, 1, 2) })
	t.Run("peer 0 faulty out of 3", func(t *testing.T) { testFundingTimeout(t, 0, 3) })
	t.Run("peer 1 faulty out of 3", func(t *testing.T) { testFundingTimeout(t, 1, 3) })
	t.Run("peer 2 faulty out of 3", func(t *testing.T) { testFundingTimeout(t, 2, 3) })
}

func testFundingTimeout(t *testing.T, faultyPeer, peers int) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), defaultTxTimeout)
	defer cancel()
	// Need unique seed per run.
	seed := time.Now().UnixNano()
	t.Logf("seed is %d", seed)
	rng := rand.New(rand.NewSource(seed))

	ct := pkgtest.NewConcurrent(t)

	_, funders, _, params, allocation := newNFunders(ctx, t, rng, peers)

	for i, funder := range funders {
		sleepTime := time.Duration(rng.Int63n(10) + 1)
		i, funder := i, funder
		go ct.StageN("funding loop", peers, func(rt require.TestingT) {
			// Faulty peer does not fund the channel.
			if i == faultyPeer {
				return
			}
			time.Sleep(sleepTime * time.Millisecond)
			req := channel.FundingReq{
				Params:     params,
				Allocation: allocation,
				Idx:        uint16(i),
			}
			ctx2, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			defer cancel()
			err := funder.Fund(ctx2, req)
			require.True(rt, channel.IsFundingTimeoutError(err), "funder should return FundingTimeoutError")
			pErr := errors.Cause(err).(*channel.FundingTimeoutError) // unwrap error
			assert.Equal(t, pErr.Errors[0].Asset, 0, "Wrong asset set")
			assert.Equal(t, uint16(faultyPeer), pErr.Errors[0].TimedOutPeers[0], "Peer should be detected as erroneous")
		})
	}
	ct.Wait("funding loop")
}

func TestFunder_Fund_multi(t *testing.T) {
	t.Run("1-party funding", func(t *testing.T) { testFunderFunding(t, 1) })
	t.Run("2-party funding", func(t *testing.T) { testFunderFunding(t, 2) })
	t.Run("3-party funding", func(t *testing.T) { testFunderFunding(t, 3) })
	t.Run("10-party funding", func(t *testing.T) { testFunderFunding(t, 10) })
}

func testFunderFunding(t *testing.T, n int) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTxTimeout)
	defer cancel()
	// Need unique seed per run.
	seed := time.Now().UnixNano()
	t.Logf("seed is %d", seed)
	rng := rand.New(rand.NewSource(seed))

	ct := pkgtest.NewConcurrent(t)

	_, funders, _, params, allocation := newNFunders(ctx, t, rng, n)

	for i, funder := range funders {
		sleepTime := time.Duration(rng.Int63n(10) + 1)
		i, funder := i, funder
		go ct.StageN("funding", n, func(rt require.TestingT) {
			time.Sleep(sleepTime * time.Millisecond)
			req := channel.FundingReq{
				Params:     params,
				Allocation: allocation,
				Idx:        channel.Index(i),
			}
			err := funder.Fund(ctx, req)
			require.NoError(rt, err, "funding should succeed")
		})
	}

	ct.Wait("funding")
	// Check result balances
	assert.NoError(t, compareOnChainAlloc(params, *allocation, &funders[0].ContractBackend))
}

func newNFunders(
	ctx context.Context,
	t *testing.T,
	rng *rand.Rand,
	n int,
) (
	parts []wallet.Address,
	funders []*ethchannel.Funder,
	app channel.App,
	params *channel.Params,
	allocation *channel.Allocation,
) {
	simBackend := test.NewSimulatedBackend()
	ks := ethwallettest.GetKeystore()
	deployAccount := wallettest.NewRandomAccount(rng).(*ethwallet.Account).Account
	simBackend.FundAddress(ctx, deployAccount.Address)
	contractBackend := ethchannel.NewContractBackend(simBackend, ks, deployAccount)
	// Deploy Assetholder
	assetETH, err := ethchannel.DeployETHAssetholder(ctx, contractBackend, deployAccount.Address)
	require.NoError(t, err, "Deployment should succeed")
	t.Logf("asset holder address is %v", assetETH)
	parts = make([]wallet.Address, n)
	funders = make([]*ethchannel.Funder, n)
	for i := 0; i < n; i++ {
		acc := wallettest.NewRandomAccount(rng).(*ethwallet.Account)
		simBackend.FundAddress(ctx, acc.Account.Address)
		parts[i] = acc.Address()
		cb := ethchannel.NewContractBackend(simBackend, ks, acc.Account)
		funders[i] = ethchannel.NewETHFunder(cb, assetETH)
	}
	app = channeltest.NewRandomApp(rng)
	params = channel.NewParamsUnsafe(rng.Uint64(), parts, app.Def(), big.NewInt(rng.Int63()))
	allocation = newValidAllocation(parts, assetETH)
	return
}

// newSimulatedFunder creates a new funder
func newSimulatedFunder(t *testing.T) *ethchannel.Funder {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTxTimeout)
	defer cancel()
	// Need unique seed per run.
	seed := time.Now().UnixNano()
	t.Logf("seed is %d", seed)
	rng := rand.New(rand.NewSource(seed))
	_, funder, _, _, _ := newNFunders(ctx, t, rng, 1)
	return funder[0]
}

func newValidAllocation(parts []wallet.Address, assetETH common.Address) *channel.Allocation {
	// Create assets slice
	assets := []channel.Asset{
		&ethchannel.Asset{Address: assetETH},
	}
	rng := rand.New(rand.NewSource(1337))
	balances := make([][]channel.Bal, len(assets))
	for a := range assets {
		balances[a] = make([]channel.Bal, len(parts))
		for p := range parts {
			// create new random balance in range [1,999]
			balances[a][p] = big.NewInt(rng.Int63n(999) + 1)
		}
	}
	return &channel.Allocation{
		Assets:   assets,
		Balances: balances,
	}
}

// compareOnChainAlloc returns error if `alloc` differs from the on-chain allocation.
func compareOnChainAlloc(params *channel.Params, alloc channel.Allocation, cb *ethchannel.ContractBackend) error {
	onChain, err := getOnChainAllocation(context.Background(), cb, params, alloc.Assets)
	if err != nil {
		return errors.WithMessage(err, "getting on-chain allocation")
	}
	for a := range onChain {
		for p := range onChain[a] {
			if alloc.Balances[a][p].Cmp(onChain[a][p]) != 0 {
				return errors.Errorf("Balances[%d][%d] differ. Expected: %v, on-chain: %v", a, p, alloc.Balances[a][p], onChain[a][p])
			}
		}
	}
	return nil
}

func getOnChainAllocation(ctx context.Context, cb *ethchannel.ContractBackend, params *channel.Params, _assets []channel.Asset) ([][]channel.Bal, error) {
	partIDs := ethchannel.FundingIDs(params.ID(), params.Parts...)
	alloc := make([][]channel.Bal, len(_assets))

	for k, asset := range _assets {
		alloc[k] = make([]channel.Bal, len(params.Parts))
		contract, err := assets.NewAssetHolder(asset.(*ethchannel.Asset).Address, cb)
		if err != nil {
			return nil, err
		}

		for i, id := range partIDs {
			opts := bind.CallOpts{
				Pending: false,
				Context: ctx,
			}
			val, err := contract.Holdings(&opts, id)
			if err != nil {
				return nil, err
			}
			alloc[k][i] = val
		}
	}
	return alloc, nil
}

// Copyright 2019 - See NOTICE file for copyright holders.
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
	"math/big"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/backend/ethereum/bindings/assets"
	ethchannel "perun.network/go-perun/backend/ethereum/channel"
	"perun.network/go-perun/backend/ethereum/channel/test"
	"perun.network/go-perun/backend/ethereum/wallet/keystore"
	"perun.network/go-perun/channel"
	channeltest "perun.network/go-perun/channel/test"
	pkgtest "perun.network/go-perun/pkg/test"
	"perun.network/go-perun/wallet"
	wallettest "perun.network/go-perun/wallet/test"
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
	rng := pkgtest.Prng(t)
	parts, funders, params, allocation := newNFunders(context.Background(), t, rng, n)

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
			Params: params,
			State:  &channel.State{Allocation: *allocation},
			Idx:    channel.Index(i),
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
	rng := pkgtest.Prng(t)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTxTimeout)
	defer cancel()
	parts, funders, params, allocation := newNFunders(ctx, t, rng, 1)
	// Test invalid funding request
	assert.Panics(t, func() { funders[0].Fund(ctx, channel.FundingReq{}) }, "Funding with invalid funding req should fail")
	// Test funding without assets
	req := channel.FundingReq{
		Params: &channel.Params{},
		State:  &channel.State{},
		Idx:    0,
	}
	assert.NoError(t, funders[0].Fund(ctx, req), "Funding with no assets should succeed")
	// Test with valid request
	req = channel.FundingReq{
		Params: params,
		State:  &channel.State{Allocation: *allocation},
		Idx:    0,
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
	rng := pkgtest.Prng(t)
	ct := pkgtest.NewConcurrent(t)

	_, funders, params, allocation := newNFunders(ctx, t, rng, peers)

	for i, funder := range funders {
		sleepTime := time.Duration(rng.Int63n(10) + 1)
		i, funder := i, funder
		go ct.StageN("funding loop", peers, func(rt pkgtest.ConcT) {
			// Faulty peer does not fund the channel.
			if i == faultyPeer {
				return
			}
			time.Sleep(sleepTime * time.Millisecond)
			req := channel.FundingReq{
				Params: params,
				State:  &channel.State{Allocation: *allocation},
				Idx:    uint16(i),
			}
			defer cancel()
			err := funder.Fund(ctx, req)
			require.True(rt, channel.IsFundingTimeoutError(err), "funder should return FundingTimeoutError")
			pErr := errors.Cause(err).(*channel.FundingTimeoutError) // unwrap error
			assert.Equal(t, pErr.Errors[0].Asset, channel.Index(0), "Wrong asset set")
			assert.Equal(t, uint16(faultyPeer), pErr.Errors[0].TimedOutPeers[0], "Peer should be detected as erroneous")
		})
	}

	time.Sleep(1 * time.Second) // give all funders enough time to fund
	// Hackily extract SimulatedBackend from funder
	sb, ok := funders[0].ContractInterface.(*test.SimulatedBackend)
	require.True(t, ok)
	// advance block time so that funding fails for non-funders
	require.NoError(t, sb.AdjustTime(time.Duration(params.ChallengeDuration)*time.Second))
	sb.Commit()

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
	rng := pkgtest.Prng(t)
	ct := pkgtest.NewConcurrent(t)

	_, funders, params, allocation := newNFunders(ctx, t, rng, n)

	for i, funder := range funders {
		sleepTime := time.Duration(rng.Int63n(10) + 1)
		i, funder := i, funder
		go ct.StageN("funding", n, func(rt pkgtest.ConcT) {
			time.Sleep(sleepTime * time.Millisecond)
			req := channel.FundingReq{
				Params: params,
				State:  &channel.State{Allocation: *allocation},
				Idx:    channel.Index(i),
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
	params *channel.Params,
	allocation *channel.Allocation,
) {
	simBackend := test.NewSimulatedBackend()
	ksWallet := wallettest.RandomWallet().(*keystore.Wallet)

	deployAccount := &ksWallet.NewRandomAccount(rng).(*keystore.Account).Account
	simBackend.FundAddress(ctx, deployAccount.Address)
	contractBackend := ethchannel.NewContractBackend(simBackend, keystore.NewTransactor(*ksWallet))

	// Deploy Assetholder
	assetAddr, err := ethchannel.DeployETHAssetholder(ctx, contractBackend, deployAccount.Address, *deployAccount)
	require.NoError(t, err, "Deployment should succeed")
	t.Logf("asset holder address is %v", assetAddr)
	asset := ethchannel.Asset(assetAddr)

	parts = make([]wallet.Address, n)
	funders = make([]*ethchannel.Funder, n)
	for i := 0; i < n; i++ {
		acc := ksWallet.NewRandomAccount(rng).(*keystore.Account)
		simBackend.FundAddress(ctx, acc.Account.Address)
		parts[i] = acc.Address()
		cb := ethchannel.NewContractBackend(simBackend, keystore.NewTransactor(*ksWallet))
		accounts := map[ethchannel.Asset]accounts.Account{asset: acc.Account}
		depositors := map[ethchannel.Asset]ethchannel.Depositor{asset: new(ethchannel.ETHDepositor)}
		funders[i] = ethchannel.NewFunder(cb, accounts, depositors)
	}

	// The SimBackend advances 10 sec per transaction/block, so generously add 20
	// sec funding duration per participant
	params = channeltest.NewRandomParams(rng, channeltest.WithParts(parts...), channeltest.WithChallengeDuration(uint64(n)*20))
	allocation = channeltest.NewRandomAllocation(rng, channeltest.WithNumParts(n), channeltest.WithAssets((*ethchannel.Asset)(&assetAddr)))
	return
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
				return errors.Errorf("balances[%d][%d] differ. Expected: %v, on-chain: %v", a, p, alloc.Balances[a][p], onChain[a][p])
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
		contract, err := assets.NewAssetHolder(common.Address(*asset.(*ethchannel.Asset)), cb)
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

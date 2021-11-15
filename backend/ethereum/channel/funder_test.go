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
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/backend/ethereum/bindings/assetholder"
	"perun.network/go-perun/backend/ethereum/bindings/peruntoken"
	ethchannel "perun.network/go-perun/backend/ethereum/channel"
	"perun.network/go-perun/backend/ethereum/channel/test"
	ethwallet "perun.network/go-perun/backend/ethereum/wallet"
	"perun.network/go-perun/backend/ethereum/wallet/keystore"
	"perun.network/go-perun/channel"
	channeltest "perun.network/go-perun/channel/test"
	"perun.network/go-perun/wallet"
	wallettest "perun.network/go-perun/wallet/test"
	pkgtest "polycry.pt/poly-go/test"
)

const (
	txGasLimit = 100000
)

func TestFunder_RegisterAsset_IsAssetRegistered(t *testing.T) {
	rng := pkgtest.Prng(t)

	funder, assets, depositors, accs := newFunderSetup(rng)
	n := len(assets)

	for i := 0; i < n; i++ {
		_, _, ok := funder.IsAssetRegistered(assets[i])
		require.False(t, ok, "on a newly initialzed funder, no assets are registered")
	}

	for i := 0; i < n; i++ {
		require.True(t, funder.RegisterAsset(assets[i], depositors[i], accs[i]), "should not error on registering a new asset")
	}

	for i := 0; i < n; i++ {
		depositor, acc, ok := funder.IsAssetRegistered(assets[i])
		require.True(t, ok, "registered asset should be returned")
		assert.Equal(t, depositors[i], depositor)
		assert.Equal(t, accs[i], acc)
	}
}

func newFunderSetup(rng *rand.Rand) (
	*ethchannel.Funder, []ethchannel.Asset, []ethchannel.Depositor, []accounts.Account) {
	n := 2
	simBackend := test.NewSimulatedBackend()
	ksWallet := wallettest.RandomWallet().(*keystore.Wallet)
	cb := ethchannel.NewContractBackend(
		simBackend,
		keystore.NewTransactor(*ksWallet, types.NewEIP155Signer(big.NewInt(1337))),
		TxFinalityDepth,
	)
	funder := ethchannel.NewFunder(cb)
	assets := make([]ethchannel.Asset, n)
	depositors := make([]ethchannel.Depositor, n)
	accs := make([]accounts.Account, n)

	for i := 0; i < n; i++ {
		assets[i] = *(wallettest.NewRandomAddress(rng).(*ethwallet.Address))
		accs[i] = accounts.Account{Address: ethwallet.AsEthAddr(wallettest.NewRandomAddress(rng))}
	}
	// Use an ETH depositor with random addresses at index 0.
	depositors[0] = ethchannel.NewETHDepositor()
	// Use an ERC20 depositor with random addresses at index 1.
	token := wallettest.NewRandomAddress(rng)
	depositors[1] = ethchannel.NewERC20Depositor(ethwallet.AsEthAddr(token))
	return funder, assets, depositors, accs
}

func TestFunder_OneForAllFunding(t *testing.T) {
	// One party will fund the complete FundingAgreement and the other parties
	// do nothing.
	t.Run("One for all 1", func(t *testing.T) { testFunderOneForAllFunding(t, 1) })
	t.Run("One for all 2", func(t *testing.T) { testFunderOneForAllFunding(t, 2) })
	t.Run("One for all 5", func(t *testing.T) { testFunderOneForAllFunding(t, 5) })
}

func testFunderOneForAllFunding(t *testing.T, n int) {
	t.Helper()
	t.Parallel()
	rng := pkgtest.Prng(t, n)
	ct := pkgtest.NewConcurrent(t)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTxTimeout*time.Duration(n))
	defer cancel()
	parts, funders, params, alloc := newNFunders(ctx, t, rng, n)
	agreement := alloc.Balances.Clone()
	richy := rng.Intn(n) // `richy` will fund for all.

	for j, a := range agreement {
		for i := 0; i < n; i++ {
			if i == richy {
				a[i] = alloc.Balances.Sum()[j]
			} else {
				a[i].SetInt64(0)
			}
		}
	}
	require.Equal(t, agreement.Sum(), alloc.Balances.Sum())

	for i := 0; i < n; i++ {
		i := i
		go ct.StageN("funding", n, func(rt pkgtest.ConcT) {
			req := channel.NewFundingReq(params, &channel.State{Allocation: *alloc}, channel.Index(i), agreement)
			diff, err := test.NonceDiff(parts[i], funders[i], func() error {
				return funders[i].Fund(ctx, *req)
			})
			require.NoError(rt, err)
			if i == richy {
				numTx, err := funders[i].NumTX(*req)
				require.NoError(t, err)
				assert.Equal(rt, int(numTx), diff, "%d transactions should have been sent", numTx)
			} else {
				assert.Zero(rt, diff, "Nonce should stay the same")
			}
		})
	}
	ct.Wait("funding")
	// Check on-chain balances.
	assert.NoError(t, compareOnChainAlloc(ctx, params, agreement, alloc.Assets, &funders[0].ContractBackend))
}

func TestFunder_CrossOverFunding(t *testing.T) {
	// Peers will randomly fund for each other.
	t.Run("Cross over 1", func(t *testing.T) { testFunderCrossOverFunding(t, 1) })
	t.Run("Cross over 2", func(t *testing.T) { testFunderCrossOverFunding(t, 2) })
	t.Run("Cross over 5", func(t *testing.T) { testFunderCrossOverFunding(t, 5) })
	t.Run("Cross over 10", func(t *testing.T) { testFunderCrossOverFunding(t, 10) })
}

func testFunderCrossOverFunding(t *testing.T, n int) {
	t.Helper()
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), defaultTxTimeout*time.Duration(n))
	defer cancel()
	rng := pkgtest.Prng(t, n)
	ct := pkgtest.NewConcurrent(t)
	parts, funders, params, alloc := newNFunders(ctx, t, rng, n)

	// Shuffle the balances.
	agreement := channeltest.ShuffleBalances(rng, alloc.Balances)
	require.Equal(t, agreement.Sum(), alloc.Balances.Sum())

	for i, funder := range funders {
		i, funder := i, funder
		go ct.StageN("funding", n, func(rt pkgtest.ConcT) {
			req := channel.NewFundingReq(params, &channel.State{Allocation: *alloc}, channel.Index(i), agreement)
			numTx, err := funders[i].NumTX(*req)
			require.NoError(t, err)
			diff, err := test.NonceDiff(parts[i], funder, func() error {
				return funder.Fund(ctx, *req)
			})
			require.NoError(rt, err, "funding should succeed")
			assert.Equal(rt, int(numTx), diff, "%d transactions should have been sent", numTx)
		})
	}

	ct.Wait("funding")
	// Check result balances
	assert.NoError(t, compareOnChainAlloc(ctx, params, agreement, alloc.Assets, &funders[0].ContractBackend))
}

func TestFunder_ZeroBalance(t *testing.T) {
	t.Run("1 Participant", func(t *testing.T) { testFunderZeroBalance(t, 1) })
	t.Run("2 Participant", func(t *testing.T) { testFunderZeroBalance(t, 2) })
	t.Run("5 Participant", func(t *testing.T) { testFunderZeroBalance(t, 5) })
	t.Run("10 Participant", func(t *testing.T) { testFunderZeroBalance(t, 10) })
}

func testFunderZeroBalance(t *testing.T, n int) {
	t.Helper()
	t.Parallel()
	rng := pkgtest.Prng(t, n)
	ct := pkgtest.NewConcurrent(t)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTxTimeout*time.Duration(n))
	defer cancel()
	parts, funders, params, alloc := newNFunders(ctx, t, rng, n)
	agreement := alloc.Balances.Clone()

	for i := range agreement {
		for j := range parts {
			if j%2 == 0 {
				alloc.Balances[i][j].SetInt64(0)
				agreement[i][j].SetInt64(0)
			}
			t.Logf("Part: %d ShouldFund: %t Bal: %v", j, j%2 == 1, alloc.Balances[0][j])
		}
	}
	// fund
	for i := 0; i < n; i++ {
		i := i
		// Check that the funding only changes the nonce when the balance is not zero
		go ct.StageN("funding", n, func(rt pkgtest.ConcT) {
			req := channel.NewFundingReq(params, &channel.State{Allocation: *alloc}, channel.Index(i), agreement)

			diff, err := test.NonceDiff(parts[i], funders[i], func() error {
				return funders[i].Fund(ctx, *req)
			})
			require.NoError(rt, err)
			if i%2 == 0 {
				assert.Zero(rt, diff, "Nonce should stay the same")
			} else {
				numTx, err := funders[i].NumTX(*req)
				require.NoError(t, err)
				assert.Equal(rt, int(numTx), diff, "%d transactions should have been sent", numTx)
			}
		})
	}
	ct.Wait("funding")
	// Check result balances
	assert.NoError(t, compareOnChainAlloc(ctx, params, agreement, alloc.Assets, &funders[0].ContractBackend))
}

func TestFunder_Multiple(t *testing.T) {
	rng := pkgtest.Prng(t)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTxTimeout)
	defer cancel()
	parts, funders, params, alloc := newNFunders(ctx, t, rng, 1)
	// Test invalid funding request
	assert.Panics(t, func() { funders[0].Fund(ctx, channel.FundingReq{}) }, "Funding with invalid funding req should fail") //nolint:errcheck
	// Test funding without assets
	req := channel.NewFundingReq(&channel.Params{}, &channel.State{}, 0, make(channel.Balances, 0))
	require.NoError(t, funders[0].Fund(ctx, *req), "Funding with no assets should succeed")
	// Test with valid request
	req = channel.NewFundingReq(params, &channel.State{Allocation: *alloc}, 0,
		alloc.Balances)

	t.Run("Funding idempotence", func(t *testing.T) {
		for i := 0; i < 1; i++ {
			var err error
			numTx := uint32(0)
			if i == 0 {
				numTx, err = funders[0].NumTX(*req)
				require.NoError(t, err)
			}
			diff, err := test.NonceDiff(parts[0], funders[0], func() error {
				return funders[0].Fund(ctx, *req)
			})
			require.NoError(t, err)
			assert.Equal(t, int(numTx), diff, "Nonce should increase")
		}
	})
	// Test already closed context
	cancel()
	assert.Error(t, funders[0].Fund(ctx, *req), "funding with already cancelled context should fail")
	// Check result balances
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	assert.NoError(t, compareOnChainAlloc(ctx, params, alloc.Balances, alloc.Assets, &funders[0].ContractBackend))
}

func TestFunder_PeerTimeout(t *testing.T) {
	t.Run("peer 0 faulty out of 2", func(t *testing.T) { testFundingTimeout(t, 0, 2) })
	t.Run("peer 1 faulty out of 2", func(t *testing.T) { testFundingTimeout(t, 1, 2) })
	t.Run("peer 0 faulty out of 3", func(t *testing.T) { testFundingTimeout(t, 0, 3) })
	t.Run("peer 1 faulty out of 3", func(t *testing.T) { testFundingTimeout(t, 1, 3) })
	t.Run("peer 2 faulty out of 3", func(t *testing.T) { testFundingTimeout(t, 2, 3) })
}

func testFundingTimeout(t *testing.T, faultyPeer, n int) {
	t.Helper()
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), defaultTxTimeout*time.Duration(n))
	defer cancel()
	rng := pkgtest.Prng(t, faultyPeer, n)
	ct := pkgtest.NewConcurrent(t)

	_, funders, params, alloc := newNFunders(ctx, t, rng, n)

	for i, funder := range funders {
		sleepTime := time.Duration(rng.Int63n(10) + 1)
		i, funder := i, funder
		go ct.StageN("funding loop", n, func(rt pkgtest.ConcT) {
			// Faulty peer does not fund the channel.
			if i == faultyPeer {
				return
			}
			time.Sleep(sleepTime * time.Millisecond)
			req := channel.NewFundingReq(params, &channel.State{Allocation: *alloc}, channel.Index(i), alloc.Balances)
			err := funder.Fund(ctx, *req)
			require.Error(t, err)
			require.True(rt, channel.IsFundingTimeoutError(err), "funder should return FundingTimeoutError")
			pErr := errors.Cause(err).(channel.FundingTimeoutError) // unwrap error
			// Check that `faultyPeer` is reported as faulty.
			require.Len(t, pErr.Errors, len(alloc.Assets))
			for _, e := range pErr.Errors {
				require.Len(t, e.TimedOutPeers, 1)
				assert.Equal(t, channel.Index(faultyPeer), e.TimedOutPeers[0], "Peer should be detected as erroneous")
			}
		outer:
			for a := 0; a < len(alloc.Assets); a++ {
				for _, e := range pErr.Errors {
					if e.Asset == channel.Index(a) {
						continue outer
					}
				}
				require.Fail(t, "asset should be reported as underfunded")
			}
		})
	}

	// Give each funder `numAssets * numPeers * 200` ms time to fund.
	time.Sleep(time.Duration(n*len(alloc.Balances)) * 200 * time.Millisecond)
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
	t.Helper()
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), defaultTxTimeout*time.Duration(n))
	defer cancel()
	rng := pkgtest.Prng(t, n)
	ct := pkgtest.NewConcurrent(t)

	_, funders, params, alloc := newNFunders(ctx, t, rng, n)

	for i, funder := range funders {
		sleepTime := time.Duration(rng.Int63n(10) + 1)
		i, funder := i, funder
		go ct.StageN("funding", n, func(rt pkgtest.ConcT) {
			time.Sleep(sleepTime * time.Millisecond)
			req := channel.NewFundingReq(params, &channel.State{Allocation: *alloc}, channel.Index(i), alloc.Balances)
			err := funder.Fund(ctx, *req)
			require.NoError(rt, err, "funding should succeed")
		})
	}

	ct.Wait("funding")
	// Check result balances
	assert.NoError(t, compareOnChainAlloc(ctx, params, alloc.Balances, alloc.Assets, &funders[0].ContractBackend))
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
	t.Helper()
	simBackend := test.NewSimulatedBackend()
	// Start the auto-mining of blocks.
	simBackend.StartMining(blockInterval)
	t.Cleanup(simBackend.StopMining)
	ksWallet := wallettest.RandomWallet().(*keystore.Wallet)

	deployAccount := &ksWallet.NewRandomAccount(rng).(*keystore.Account).Account
	simBackend.FundAddress(ctx, deployAccount.Address)
	tokenAcc := &ksWallet.NewRandomAccount(rng).(*keystore.Account).Account
	simBackend.FundAddress(ctx, tokenAcc.Address)
	cb := ethchannel.NewContractBackend(
		simBackend,
		keystore.NewTransactor(*ksWallet, types.NewEIP155Signer(big.NewInt(1337))),
		TxFinalityDepth,
	)

	// Deploy ETHAssetholder
	assetAddr1, err := ethchannel.DeployETHAssetholder(ctx, cb, deployAccount.Address, *deployAccount)
	require.NoError(t, err, "Deployment should succeed")
	t.Logf("asset holder #1 address is %s", assetAddr1.Hex())
	asset1 := ethchannel.Asset(assetAddr1)
	// Deploy PerunToken + ETHAssetholder.

	token, err := ethchannel.DeployPerunToken(ctx, cb, *deployAccount, []common.Address{tokenAcc.Address}, channeltest.MaxBalance)
	require.NoError(t, err, "Deployment should succeed")
	assetAddr2, err := ethchannel.DeployERC20Assetholder(ctx, cb, common.Address{}, token, *deployAccount)
	require.NoError(t, err, "Deployment should succeed")
	t.Logf("asset holder #2 address is %s", assetAddr2.Hex())
	asset2 := ethchannel.Asset(assetAddr2)

	parts = make([]wallet.Address, n)
	funders = make([]*ethchannel.Funder, n)
	for i := range parts {
		acc := ksWallet.NewRandomAccount(rng).(*keystore.Account).Account
		parts[i] = ethwallet.AsWalletAddr(acc.Address)

		simBackend.FundAddress(ctx, ethwallet.AsEthAddr(parts[i]))
		err = fundERC20(ctx, cb, *tokenAcc, ethwallet.AsEthAddr(parts[i]), token, asset2)
		require.NoError(t, err)

		funders[i] = ethchannel.NewFunder(cb)
		require.True(t, funders[i].RegisterAsset(asset1, ethchannel.NewETHDepositor(), acc))
		require.True(t, funders[i].RegisterAsset(asset2, ethchannel.NewERC20Depositor(token), acc))
	}

	// The challenge duration needs to be really large, since the auto-mining of
	// SimBackend advances the block time with 100 seconds/second.
	// By using a large value, we make sure that longer running tests work.
	params = channeltest.NewRandomParams(rng, channeltest.WithParts(parts...), channeltest.WithChallengeDuration(uint64(n)*40000))
	allocation = channeltest.NewRandomAllocation(rng, channeltest.WithNumParts(n), channeltest.WithAssets((*ethchannel.Asset)(&assetAddr1), (*ethchannel.Asset)(&assetAddr2)))
	return
}

// fundERC20 funds `to` with ERC20 tokens from account `from`.
func fundERC20(ctx context.Context, cb ethchannel.ContractBackend, from accounts.Account, to common.Address, token common.Address, asset ethchannel.Asset) error {
	contract, err := peruntoken.NewERC20(token, cb)
	if err != nil {
		return errors.WithMessagef(err, "binding AssetHolderERC20 contract at: %v", asset)
	}
	// Transfer.
	opts, err := cb.NewTransactor(ctx, txGasLimit, from)
	if err != nil {
		return errors.WithMessagef(err, "creating transactor for asset: %v", asset)
	}
	amount := new(big.Int).Rsh(channeltest.MaxBalance, 10)
	tx, err := contract.Transfer(opts, to, amount)
	if err != nil {
		return errors.WithMessage(err, "transferring tokens")
	}
	_, err = cb.ConfirmTransaction(ctx, tx, from)
	return err
}

// compareOnChainAlloc returns error if `alloc` differs from the on-chain allocation.
func compareOnChainAlloc(ctx context.Context, params *channel.Params, balances channel.Balances, assets []channel.Asset, cb *ethchannel.ContractBackend) error {
	onChain, err := getOnChainAllocation(ctx, cb, params, assets)
	if err != nil {
		return errors.WithMessage(err, "getting on-chain allocation")
	}
	for a := range onChain {
		for p := range onChain[a] {
			if balances[a][p].Cmp(onChain[a][p]) != 0 {
				return errors.Errorf("balances[%d][%d] differ. Expected: %v, on-chain: %v", a, p, balances[a][p], onChain[a][p])
			}
		}
	}
	return nil
}

func getOnChainAllocation(ctx context.Context, cb *ethchannel.ContractBackend, params *channel.Params, _assets []channel.Asset) (channel.Balances, error) {
	partIDs := ethchannel.FundingIDs(params.ID(), params.Parts...)
	alloc := make(channel.Balances, len(_assets))

	for k, asset := range _assets {
		alloc[k] = make([]channel.Bal, len(params.Parts))
		contract, err := assetholder.NewAssetHolder(common.Address(*asset.(*ethchannel.Asset)), cb)
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

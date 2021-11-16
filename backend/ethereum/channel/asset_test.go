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
	"testing"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	ethchannel "perun.network/go-perun/backend/ethereum/channel"
	"perun.network/go-perun/backend/ethereum/channel/test"
	ethwallettest "perun.network/go-perun/backend/ethereum/wallet/test"
	pkgtest "polycry.pt/poly-go/test"
)

func TestValidateAssetHolderETH(t *testing.T) {
	testValidateAssetHolder(t, ethchannel.DeployETHAssetholder, ethchannel.ValidateAssetHolderETH)
}

func TestValidateAssetHolderERC20(t *testing.T) {
	var (
		rng      = pkgtest.Prng(t)
		token    = common.Address(ethwallettest.NewRandomAddress(rng))
		deployer = func(ctx context.Context,
			backend ethchannel.ContractBackend,
			adjudicatorAddr common.Address,
			deployer accounts.Account) (common.Address, error) {
			return ethchannel.DeployERC20Assetholder(ctx, backend, adjudicatorAddr, token, deployer)
		}
		verifier = func(ctx context.Context,
			backend bind.ContractBackend, assetHolderETH, adjudicator common.Address) error {
			return ethchannel.ValidateAssetHolderERC20(ctx, backend, assetHolderETH, adjudicator, token)
		}
	)

	testValidateAssetHolder(t, deployer, verifier)
}

func testValidateAssetHolder(t *testing.T,
	deployer func(ctx context.Context,
		backend ethchannel.ContractBackend,
		adjudicatorAddr common.Address,
		deployer accounts.Account) (common.Address, error),
	validator func(ctx context.Context,
		backend bind.ContractBackend, assetHolderETH, adjudicator common.Address) error) {
	// Test setup
	rng := pkgtest.Prng(t, "validate")
	s := test.NewSimSetup(t, rng, TxFinalityDepth, blockInterval)
	ctx, cancel := context.WithTimeout(context.Background(), 4*defaultTxTimeout)
	defer cancel()

	t.Run("no_asset_code", func(t *testing.T) {
		randomAddr1 := (common.Address)(ethwallettest.NewRandomAddress(rng))
		randomAddr2 := (common.Address)(ethwallettest.NewRandomAddress(rng))
		require.True(t, ethchannel.IsErrInvalidContractCode(validator(ctx, s.CB, randomAddr1, randomAddr2)))
	})

	t.Run("incorrect_asset_code", func(t *testing.T) {
		randomAddr1 := (common.Address)(ethwallettest.NewRandomAddress(rng))
		incorrectCodeAddr, err := ethchannel.DeployAdjudicator(ctx, *s.CB, s.TxSender.Account)
		require.NoError(t, err)
		require.True(t, ethchannel.IsErrInvalidContractCode(validator(ctx, s.CB, incorrectCodeAddr, randomAddr1)))
	})

	t.Run("incorrect_adj_addr", func(t *testing.T) {
		adjAddrToSet := (common.Address)(ethwallettest.NewRandomAddress(rng))
		adjAddrToExpect := (common.Address)(ethwallettest.NewRandomAddress(rng))
		assetHolderAddr, err := deployer(ctx, *s.CB, adjAddrToSet, s.TxSender.Account)
		require.NoError(t, err)
		require.True(t, ethchannel.IsErrInvalidContractCode(validator(ctx, s.CB, assetHolderAddr, adjAddrToExpect)))
	})

	t.Run("correct_adj_addr_with_invalid_contract", func(t *testing.T) {
		adjudicatorAddr := (common.Address)(ethwallettest.NewRandomAddress(rng))
		assetHolderAddr, err := deployer(ctx, *s.CB, adjudicatorAddr, s.TxSender.Account)
		require.NoError(t, err)
		require.NoError(t, validator(ctx, s.CB, assetHolderAddr, adjudicatorAddr))
	})

	t.Run("all_correct", func(t *testing.T) {
		adjudicatorAddr, err := ethchannel.DeployAdjudicator(ctx, *s.CB, s.TxSender.Account)
		require.NoError(t, err)
		assetHolderAddr, err := deployer(ctx, *s.CB, adjudicatorAddr, s.TxSender.Account)
		require.NoError(t, err)
		require.NoError(t, validator(ctx, s.CB, assetHolderAddr, adjudicatorAddr))
	})
}

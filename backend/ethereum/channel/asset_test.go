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

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	ethchannel "perun.network/go-perun/backend/ethereum/channel"
	"perun.network/go-perun/backend/ethereum/channel/test"
	ethwallettest "perun.network/go-perun/backend/ethereum/wallet/test"
	pkgtest "perun.network/go-perun/pkg/test"
)

func TestValidateAssetHolderETH(t *testing.T) {
	// Test setup
	rng := pkgtest.Prng(t)
	s := test.NewSimSetup(rng)

	t.Run("no_asset_code", func(t *testing.T) {
		randomAddr1 := (common.Address)(ethwallettest.NewRandomAddress(rng))
		randomAddr2 := (common.Address)(ethwallettest.NewRandomAddress(rng))
		ctx, cancel := context.WithTimeout(context.Background(), defaultTxTimeout)
		defer cancel()
		require.True(t, ethchannel.IsContractBytecodeError(ethchannel.ValidateAssetHolderETH(ctx, *s.CB, randomAddr1, randomAddr2)))
	})
	t.Run("incorrect_asset_code", func(t *testing.T) {
		randomAddr1 := (common.Address)(ethwallettest.NewRandomAddress(rng))
		ctx, cancel := context.WithTimeout(context.Background(), defaultTxTimeout)
		defer cancel()
		incorrectCodeAddr, err := ethchannel.DeployAdjudicator(ctx, *s.CB)
		require.NoError(t, err)
		require.True(t, ethchannel.IsContractBytecodeError(ethchannel.ValidateAssetHolderETH(ctx, *s.CB, incorrectCodeAddr, randomAddr1)))
	})
	t.Run("incorrect_adj_addr", func(t *testing.T) {
		adjAddrToSet := (common.Address)(ethwallettest.NewRandomAddress(rng))
		adjAddrToExpect := (common.Address)(ethwallettest.NewRandomAddress(rng))
		ctx, cancel := context.WithTimeout(context.Background(), defaultTxTimeout)
		defer cancel()
		assetHolderAddr, err := ethchannel.DeployETHAssetholder(ctx, *s.CB, adjAddrToSet)
		require.NoError(t, err)
		t.Logf("assetholder address is %v", assetHolderAddr)
		require.True(t, ethchannel.IsContractBytecodeError(ethchannel.ValidateAssetHolderETH(ctx, *s.CB, assetHolderAddr, adjAddrToExpect)))
	})
	t.Run("all_correct", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), defaultTxTimeout)
		defer cancel()
		adjudicatorAddr, err := ethchannel.DeployAdjudicator(ctx, *s.CB)
		require.NoError(t, err)
		assetHolderAddr, err := ethchannel.DeployETHAssetholder(ctx, *s.CB, adjudicatorAddr)
		require.NoError(t, err)
		t.Logf("adjudicator address is %v", adjudicatorAddr)
		t.Logf("assetholder address is %v", assetHolderAddr)
		require.NoError(t, ethchannel.ValidateAssetHolderETH(ctx, *s.CB, assetHolderAddr, adjudicatorAddr))
	})
}

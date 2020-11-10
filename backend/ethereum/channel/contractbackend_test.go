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
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/backend/ethereum/bindings/adjudicator"
	ethchannel "perun.network/go-perun/backend/ethereum/channel"
	"perun.network/go-perun/backend/ethereum/channel/test"
	ethwallet "perun.network/go-perun/backend/ethereum/wallet"
	ethwallettest "perun.network/go-perun/backend/ethereum/wallet/test"
	pkgtest "perun.network/go-perun/pkg/test"
	"perun.network/go-perun/wallet"
)

func fromEthAddr(a common.Address) wallet.Address {
	return (*ethwallet.Address)(&a)
}

func Test_calcFundingIDs(t *testing.T) {
	tests := []struct {
		name         string
		participants []wallet.Address
		channelID    [32]byte
		want         [][32]byte
	}{
		{"Test nil array, empty channelID", nil, [32]byte{}, make([][32]byte, 0)},
		{"Test nil array, non-empty channelID", nil, [32]byte{1}, make([][32]byte, 0)},
		{"Test empty array, non-empty channelID", []wallet.Address{}, [32]byte{1}, make([][32]byte, 0)},
		// Tests based on actual data from contracts.
		{"Test non-empty array, empty channelID", []wallet.Address{&ethwallet.Address{}},
			[32]byte{}, [][32]byte{{173, 50, 40, 182, 118, 247, 211, 205, 66, 132, 165, 68, 63, 23, 241, 150, 43, 54, 228, 145, 179, 10, 64, 178, 64, 88, 73, 229, 151, 186, 95, 181}}},
		{"Test non-empty array, non-empty channelID", []wallet.Address{&ethwallet.Address{}},
			[32]byte{1}, [][32]byte{{130, 172, 39, 157, 178, 106, 32, 109, 155, 165, 169, 76, 7, 255, 148, 10, 234, 75, 59, 253, 232, 130, 14, 201, 95, 78, 250, 10, 207, 208, 213, 188}}},
		{"Test non-empty array, non-empty channelID", []wallet.Address{fromEthAddr(common.BytesToAddress([]byte{}))},
			[32]byte{1}, [][32]byte{{130, 172, 39, 157, 178, 106, 32, 109, 155, 165, 169, 76, 7, 255, 148, 10, 234, 75, 59, 253, 232, 130, 14, 201, 95, 78, 250, 10, 207, 208, 213, 188}}},
	}
	for _, _tt := range tests {
		tt := _tt
		t.Run(tt.name, func(t *testing.T) {
			got := ethchannel.FundingIDs(tt.channelID, tt.participants...)
			assert.Equal(t, got, tt.want, "FundingIDs not as expected")
		})
	}
}

func Test_NewTransactor(t *testing.T) {
	rng := pkgtest.Prng(t)
	s := test.NewSimSetup(rng)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tests := []struct {
		name     string
		ctx      context.Context
		gasLimit uint64
	}{
		{"Test without context", nil, uint64(0)},
		{"Test valid transactor", ctx, uint64(0)},
		{"Test valid transactor", ctx, uint64(12345)},
	}
	for _, _tt := range tests {
		tt := _tt
		t.Run(tt.name, func(t *testing.T) {
			transactor, err := s.CB.NewTransactor(tt.ctx, tt.gasLimit, s.TxSender.Account)
			assert.NoError(t, err, "Creating Transactor should succeed")
			assert.Equal(t, s.TxSender.Account.Address, transactor.From, "Transactor address not properly set")
			assert.Equal(t, tt.ctx, transactor.Context, "Context not set properly")
			assert.Equal(t, tt.gasLimit, transactor.GasLimit, "Gas limit not set properly")
		})
	}
}

func Test_NewWatchOpts(t *testing.T) {
	rng := pkgtest.Prng(t)
	s := test.NewSimSetup(rng)
	watchOpts, err := s.CB.NewWatchOpts(context.Background())
	require.NoError(t, err, "Creating watchopts on valid ContractBackend should succeed")
	assert.Equal(t, context.Background(), watchOpts.Context, "context should be set")
	assert.Equal(t, uint64(1), *watchOpts.Start, "startblock should be 1")
	key := "foo"
	ctx := context.WithValue(context.Background(), &key, "bar")
	watchOpts, err = s.CB.NewWatchOpts(ctx)
	require.NoError(t, err, "Creating watchopts on valid ContractBackend should succeed")
	assert.Equal(t, context.WithValue(context.Background(), &key, "bar"), watchOpts.Context, "context should be set")
	assert.Equal(t, uint64(1), *watchOpts.Start, "startblock should be 1")
}

func TestFetchCodeAtAddr(t *testing.T) {
	// Test setup
	rng := pkgtest.Prng(t)
	s := test.NewSimSetup(rng)

	t.Run("no_code", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), defaultTxTimeout)
		defer cancel()
		randomAddr := (common.Address)(ethwallettest.NewRandomAddress(rng))
		t.Logf("address with no contract code - %v", randomAddr)
		code, err := ethchannel.FetchCodeAtAddr(ctx, *s.CB, randomAddr)
		require.True(t, ethchannel.IsContractBytecodeError(err))
		require.Nil(t, code)
	})
	t.Run("valid_bytecode", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), defaultTxTimeout)
		defer cancel()
		adjudicatorAddr, err := ethchannel.DeployAdjudicator(ctx, *s.CB, s.TxSender.Account)
		require.NoError(t, err)
		t.Logf("contract deployed at address - %v", adjudicatorAddr)
		code, err := ethchannel.FetchCodeAtAddr(ctx, *s.CB, adjudicatorAddr)
		require.NoError(t, err)
		require.Equal(t, fmt.Sprintf("%x", code), adjudicator.AdjudicatorBinRuntime)
	})
}

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
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ethchannel "perun.network/go-perun/backend/ethereum/channel"
	"perun.network/go-perun/backend/ethereum/channel/test"
	ethwallet "perun.network/go-perun/backend/ethereum/wallet"
	"perun.network/go-perun/wallet"
	pkgtest "polycry.pt/poly-go/test"
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
		{
			"Test non-empty array, empty channelID",
			[]wallet.Address{&ethwallet.Address{}},
			[32]byte{},
			[][32]byte{{173, 50, 40, 182, 118, 247, 211, 205, 66, 132, 165, 68, 63, 23, 241, 150, 43, 54, 228, 145, 179, 10, 64, 178, 64, 88, 73, 229, 151, 186, 95, 181}},
		},
		{
			"Test non-empty array, non-empty channelID",
			[]wallet.Address{&ethwallet.Address{}},
			[32]byte{1},
			[][32]byte{{130, 172, 39, 157, 178, 106, 32, 109, 155, 165, 169, 76, 7, 255, 148, 10, 234, 75, 59, 253, 232, 130, 14, 201, 95, 78, 250, 10, 207, 208, 213, 188}},
		},
		{
			"Test non-empty array, non-empty channelID",
			[]wallet.Address{fromEthAddr(common.BytesToAddress([]byte{}))},
			[32]byte{1},
			[][32]byte{{130, 172, 39, 157, 178, 106, 32, 109, 155, 165, 169, 76, 7, 255, 148, 10, 234, 75, 59, 253, 232, 130, 14, 201, 95, 78, 250, 10, 207, 208, 213, 188}},
		},
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
	s := test.NewSimSetup(t, rng, TxFinalityDepth, blockInterval)
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
	s := test.NewSimSetup(t, rng, TxFinalityDepth, blockInterval)
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

// Test_ConfirmTransaction tests that a transaction is confirmed after exactly
// `TxBlockFinality` blocks when using `ConfirmTransaction`.
// Does not test reorgs.
func Test_ConfirmTransaction(t *testing.T) {
	rng := pkgtest.Prng(t)
	s := test.NewSimSetup(t, rng, TxFinalityDepth, 0)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTestTimeout)
	defer cancel()

	// Create the Transaction.
	rawTx := types.NewTx(&types.DynamicFeeTx{
		Nonce:     0,
		GasFeeCap: big.NewInt(test.InitialGasBaseFee),
		Gas:       params.TxGas,
		To:        &common.Address{},
		Value:     big.NewInt(1),
	})
	opts, err := s.CB.NewTransactor(ctx, params.TxGas, s.TxSender.Account)
	require.NoError(t, err)
	signed, err := opts.Signer(s.TxSender.Account.Address, rawTx)
	require.NoError(t, err)

	// Send the TX.
	require.NoError(t, s.SimBackend.SimulatedBackend.SendTransaction(ctx, signed))

	// Write receipt into `confirmed` when the TX is confirmed.
	confirmed := make(chan *types.Receipt)
	go func() {
		// Confirm.
		r, err := s.CB.ConfirmTransaction(ctx, signed, s.TxSender.Account)
		require.NoError(t, err)
		confirmed <- r
	}()

	// Create new blocks.
	for i := 0; i < int(TxFinalityDepth); i++ {
		// Check that it is not yet confirmed.
		select {
		case <-time.After(100 * time.Millisecond):
		case <-confirmed:
			t.Error("TX should not be confirmed yet.")
			t.FailNow()
		}
		// Mine new block.
		s.SimBackend.Commit()
	}
	// Wait for confirm.
	r := <-confirmed
	// Get current block height.
	h, err := s.CB.BlockByNumber(ctx, nil)
	require.NoError(t, err)
	// Assert that it got included in `TxBlockFinality` many blocks.
	assert.Equal(t, uint64(TxFinalityDepth), (h.NumberU64()-r.BlockNumber.Uint64())+1)
}

// Test_ReorgConfirmTransaction tests that a TX is confirmed correctly after a
// reorg.
func Test_ReorgConfirmTransaction(t *testing.T) {
	// Test does not make sense for Finality < 2.
	require.Greater(t, uint64(TxFinalityDepth), uint64(1))
	rng := pkgtest.Prng(t)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTestTimeout)
	defer cancel()
	s := test.NewTokenSetup(ctx, t, rng, TxFinalityDepth)

	// Send the TX and mine one block.
	tx := s.IncAllowance(ctx)

	// Wait `TxFinalityDepth - 2` many blocks.
	for i := uint64(0); i < TxFinalityDepth-2; i++ {
		s.SB.Commit()
	}

	// Check that the TX is not confirmed yet.
	s.ConfirmTx(tx, false)

	// Do a reorg and add two more blocks. Move the TX one block forward.
	// The TX should now be included in `TxFinalityDepth` many blocks.
	err := s.SB.Reorg(ctx, TxFinalityDepth-1, func(txs []types.Transactions) []types.Transactions {
		ret := make([]types.Transactions, TxFinalityDepth+1)
		ret[1] = txs[0]
		return ret
	})
	require.NoError(t, err)

	// Confirm
	s.ConfirmTx(tx, true)
}

// Test_ReorgRemoveTransaction tests that a TX is not confirmed if a reorg
// removes it from the canonical chain before `TxFinalityDepth` is reached.
func Test_ReorgRemoveTransaction(t *testing.T) {
	// Test does not make sense for Finality < 2.
	require.Greater(t, uint64(TxFinalityDepth), uint64(1))
	rng := pkgtest.Prng(t)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTestTimeout)
	defer cancel()
	s := test.NewTokenSetup(ctx, t, rng, TxFinalityDepth)

	// Send the TX and mine one block.
	tx := s.IncAllowance(ctx)

	// Wait `TxFinalityDepth - 2` many blocks.
	for i := uint64(0); i < TxFinalityDepth-2; i++ {
		s.SB.Commit()
	}

	// Check that the TX is not confirmed yet.
	s.ConfirmTx(tx, false)

	// Do a reorg by adding two more blocks and removing the TX.
	// The `TxFinalityDepth` would now be reached.
	err := s.SB.Reorg(ctx, TxFinalityDepth-1, func(txs []types.Transactions) []types.Transactions {
		return make([]types.Transactions, TxFinalityDepth+1)
	})
	require.NoError(t, err)

	// Still not confirmed.
	s.ConfirmTx(tx, false)
}

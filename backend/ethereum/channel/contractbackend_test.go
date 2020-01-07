// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package channel

import (
	"context"
	"io"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"perun.network/go-perun/backend/ethereum/wallet"
	perunwallet "perun.network/go-perun/wallet"
)

type testInvalidAsset [33]byte

func (t *testInvalidAsset) Encode(w io.Writer) error {
	return errors.New("Unimplemented")
}

func (t *testInvalidAsset) Decode(r io.Reader) error {
	return errors.New("Unimplemented")
}

func Test_calcFundingIDs(t *testing.T) {
	tests := []struct {
		name         string
		participants []perunwallet.Address
		channelID    [32]byte
		want         [][32]byte
	}{
		{"Test nil array, empty channelID", nil, [32]byte{}, make([][32]byte, 0)},
		{"Test nil array, non-empty channelID", nil, [32]byte{1}, make([][32]byte, 0)},
		{"Test empty array, non-empty channelID", []perunwallet.Address{}, [32]byte{1}, make([][32]byte, 0)},
		// Tests based on actual data from contracts.
		{"Test non-empty array, empty channelID", []perunwallet.Address{&wallet.Address{}},
			[32]byte{}, [][32]byte{{173, 50, 40, 182, 118, 247, 211, 205, 66, 132, 165, 68, 63, 23, 241, 150, 43, 54, 228, 145, 179, 10, 64, 178, 64, 88, 73, 229, 151, 186, 95, 181}}},
		{"Test non-empty array, non-empty channelID", []perunwallet.Address{&wallet.Address{}},
			[32]byte{1}, [][32]byte{{130, 172, 39, 157, 178, 106, 32, 109, 155, 165, 169, 76, 7, 255, 148, 10, 234, 75, 59, 253, 232, 130, 14, 201, 95, 78, 250, 10, 207, 208, 213, 188}}},
		{"Test non-empty array, non-empty channelID", []perunwallet.Address{&wallet.Address{Address: common.BytesToAddress([]byte{})}},
			[32]byte{1}, [][32]byte{{130, 172, 39, 157, 178, 106, 32, 109, 155, 165, 169, 76, 7, 255, 148, 10, 234, 75, 59, 253, 232, 130, 14, 201, 95, 78, 250, 10, 207, 208, 213, 188}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calcFundingIDs(tt.participants, tt.channelID)
			assert.Equal(t, got, tt.want, "FundingIDs not as expected")
		})
	}
}

func Test_NewTransactor(t *testing.T) {
	f := &ContractBackend{}
	assert.Panics(t,
		func() { f.newTransactor(context.Background(), big.NewInt(0), uint64(0)) },
		"Creating transactor on invalid backend should fail")
	// Test on valid contract backend
	sf := newSimulatedFunder(t)
	f = &sf.ContractBackend
	tests := []struct {
		name     string
		ctx      context.Context
		value    *big.Int
		gasLimit uint64
	}{
		{"Test without context", nil, big.NewInt(0), uint64(0)},
		{"Test valid transactor", context.Background(), big.NewInt(0), uint64(0)},
		{"Test valid transactor", context.Background(), big.NewInt(1220), uint64(12345)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transactor, err := f.newTransactor(tt.ctx, tt.value, tt.gasLimit)
			assert.NoError(t, err, "Creating Transactor should succeed")
			assert.Equal(t, sf.account.Address, transactor.From, "Transactor address not properly set")
			assert.Equal(t, uint64(tt.gasLimit), transactor.GasLimit, "Gas limit not set properly")
			assert.Equal(t, tt.value, transactor.Value, "Transaction value not set properly")
			assert.Equal(t, big.NewInt(1), transactor.GasPrice, "Invalid gas price")
		})
	}
}

func Test_NewWatchOpts(t *testing.T) {
	f := &ContractBackend{}
	assert.Panics(t, func() { f.newWatchOpts(context.Background()) }, "Creating watchopts on invalid backend should panic")
	sf := newSimulatedFunder(t)
	f = &ContractBackend{sf.ContractBackend, sf.ks, sf.account}
	watchOpts, err := f.newWatchOpts(context.Background())
	assert.NoError(t, err, "Creating watchopts on valid ContractBackend should succeed")
	assert.Equal(t, context.Background(), watchOpts.Context, "context should be set")
	assert.Equal(t, uint64(1), *watchOpts.Start, "startblock should be 1")
	key := "foo"
	ctx := context.WithValue(context.Background(), &key, "bar")
	watchOpts, err = f.newWatchOpts(ctx)
	assert.NoError(t, err, "Creating watchopts on valid ContractBackend should succeed")
	assert.Equal(t, context.WithValue(context.Background(), &key, "bar"), watchOpts.Context, "context should be set")
	assert.Equal(t, uint64(1), *watchOpts.Start, "startblock should be 1")
}

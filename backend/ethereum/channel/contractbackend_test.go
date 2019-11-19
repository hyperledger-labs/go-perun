// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel

import (
	"context"
	"io"
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"perun.network/go-perun/backend/ethereum/wallet"
	"perun.network/go-perun/channel/test"
	perunio "perun.network/go-perun/pkg/io"
	perunwallet "perun.network/go-perun/wallet"
)

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
		{"Test non-empty array, empty channelID", []perunwallet.Address{&wallet.Address{}},
			[32]byte{}, [][32]byte{[32]byte{173, 50, 40, 182, 118, 247, 211, 205, 66, 132, 165, 68, 63, 23, 241, 150, 43, 54, 228, 145, 179, 10, 64, 178, 64, 88, 73, 229, 151, 186, 95, 181}}},
		{"Test non-empty array, non-empty channelID", []perunwallet.Address{&wallet.Address{}},
			[32]byte{1}, [][32]byte{[32]byte{130, 172, 39, 157, 178, 106, 32, 109, 155, 165, 169, 76, 7, 255, 148, 10, 234, 75, 59, 253, 232, 130, 14, 201, 95, 78, 250, 10, 207, 208, 213, 188}}},
		{"Test non-empty array, non-empty channelID", []perunwallet.Address{&wallet.Address{Address: common.BytesToAddress([]byte{})}},
			[32]byte{1}, [][32]byte{[32]byte{130, 172, 39, 157, 178, 106, 32, 109, 155, 165, 169, 76, 7, 255, 148, 10, 234, 75, 59, 253, 232, 130, 14, 201, 95, 78, 250, 10, 207, 208, 213, 188}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calcFundingIDs(tt.participants, tt.channelID)
			if err != nil {
				t.Errorf("calculating PartIDs should not produce errors.")
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("calcFundingIDs() = %v, want %v", got, tt.want)
			}
		})
	}
}

type testInvalidAsset [33]byte

func (t *testInvalidAsset) Encode(w io.Writer) error {
	return errors.New("Unimplemented")
}

func (t *testInvalidAsset) Decode(r io.Reader) error {
	return errors.New("Unimplemented")
}

func Test_assetToAddress(t *testing.T) {

	var invAsset testInvalidAsset
	tests := []struct {
		name    string
		asset   perunio.Serializable
		want    common.Address
		wantErr bool
	}{
		{"Test invalid address", &invAsset, common.Address{}, true},
		{"Test valid address", &test.Asset{Address: &wallet.Address{}}, common.Address{}, false},
		{"Test valid address",
			&test.Asset{Address: &wallet.Address{Address: common.Address{1, 2, 3, 4, 5}}},
			common.Address{1, 2, 3, 4, 5}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				assert.Panics(t, func() { assetToAddress(tt.asset) })
				return
			}
			got := assetToAddress(tt.asset)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("assetToAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_NewTransactor(t *testing.T) {
	f := &contractBackend{}
	_, err := f.newTransactor(nil, nil, nil, big.NewInt(0), 1000)
	assert.Error(t, err, "Funder has to have a context set")
	_, err = f.newTransactor(nil, nil, nil, big.NewInt(0), 1000)
	assert.Error(t, err, "Creating a transactor without a connection should fail")
	sf := newSimulatedFunder()
	f = &contractBackend{sf.client}
	transactor, err := f.newTransactor(context.Background(), sf.ks, sf.account, big.NewInt(0), 1000)
	assert.NoError(t, err, "Creating Transactor should succeed")
	assert.Equal(t, sf.account.Address, transactor.From, "Transactor address not properly set")
	assert.Equal(t, uint64(1000), transactor.GasLimit, "Gas limit not set properly")
	assert.Equal(t, big.NewInt(0), transactor.Value, "Transaction value not set properly")
	transactor, err = f.newTransactor(context.Background(), sf.ks, sf.account, big.NewInt(12345), 12345)
	assert.NoError(t, err, "Creating Transactor should succeed")
	assert.Equal(t, sf.account.Address, transactor.From, "Transactor address not properly set")
	assert.Equal(t, uint64(12345), transactor.GasLimit, "Gas limit not set properly")
	assert.Equal(t, big.NewInt(12345), transactor.Value, "Transaction value not set properly")
}

func Test_NewWatchOpts(t *testing.T) {
	f := &contractBackend{}
	watchOpts := f.newWatchOpts(context.Background(), 0)
	assert.Equal(t, uint64(0), *watchOpts.Start, "Creating watchopts with no context should succeed")
	watchOpts = f.newWatchOpts(context.Background(), 123)
	assert.Equal(t, context.Background(), watchOpts.Context, "Creating watchopts with context should succeed")
	assert.Equal(t, uint64(123), *watchOpts.Start, "Creating watchopts with no context should succeed")
}

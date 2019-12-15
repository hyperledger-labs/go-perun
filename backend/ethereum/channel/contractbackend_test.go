// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel

import (
	"context"
	"crypto/ecdsa"
	"io"
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"perun.network/go-perun/backend/ethereum/wallet"
	perunwallet "perun.network/go-perun/wallet"
)

type simulatedBackend struct {
	backends.SimulatedBackend
	faucetSK   *ecdsa.PrivateKey
	faucetAddr common.Address
}

func newSimulatedBackend() *simulatedBackend {
	sk, err := crypto.GenerateKey()
	if err != nil {
		panic(err)
	}
	faucetAddr := crypto.PubkeyToAddress(sk.PublicKey)
	addr := map[common.Address]core.GenesisAccount{
		common.BytesToAddress([]byte{1}): {Balance: big.NewInt(1)}, // ECRecover
		common.BytesToAddress([]byte{2}): {Balance: big.NewInt(1)}, // SHA256
		common.BytesToAddress([]byte{3}): {Balance: big.NewInt(1)}, // RIPEMD
		common.BytesToAddress([]byte{4}): {Balance: big.NewInt(1)}, // Identity
		common.BytesToAddress([]byte{5}): {Balance: big.NewInt(1)}, // ModExp
		common.BytesToAddress([]byte{6}): {Balance: big.NewInt(1)}, // ECAdd
		common.BytesToAddress([]byte{7}): {Balance: big.NewInt(1)}, // ECScalarMul
		common.BytesToAddress([]byte{8}): {Balance: big.NewInt(1)}, // ECPairing
		faucetAddr:                       {Balance: new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(9))},
	}
	alloc := core.GenesisAlloc(addr)
	return &simulatedBackend{*backends.NewSimulatedBackend(alloc, 8000000), sk, faucetAddr}
}

func (s *simulatedBackend) BlockByNumber(_ context.Context, number *big.Int) (*types.Block, error) {
	if number == nil {
		return s.Blockchain().CurrentBlock(), nil
	}
	if block := s.Blockchain().GetBlockByNumber(number.Uint64()); block != nil {
		return block, nil
	}
	return nil, errors.New("got nil block from blockchain")
}

func (s *simulatedBackend) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	if err := s.SimulatedBackend.SendTransaction(ctx, tx); err != nil {
		return errors.WithStack(err)
	}
	s.Commit()
	return nil
}

func (s *simulatedBackend) fundAddress(ctx context.Context, addr common.Address) {
	nonce, err := s.PendingNonceAt(context.Background(), s.faucetAddr)
	if err != nil {
		panic(err)
	}
	value := new(big.Int).Lsh(big.NewInt(1), 64) // 10 eth in wei
	tx := types.NewTransaction(nonce, addr, value, gasLimit, big.NewInt(1), nil)
	signer := types.NewEIP155Signer(big.NewInt(1337))
	signedTX, err := types.SignTx(tx, signer, s.faucetSK)
	if err != nil {
		panic(err)
	}
	if err := s.SendTransaction(ctx, signedTX); err != nil {
		panic(err)
	}
	bind.WaitMined(context.Background(), s, signedTX)
}

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

func Test_NewTransactor(t *testing.T) {
	f := &ContractBackend{}
	_, err := f.newTransactor(nil, nil, nil, big.NewInt(0), 1000)
	assert.Error(t, err, "Funder has to have a context set")
	sf := newSimulatedFunder()
	f = &ContractBackend{sf.ContractBackend, sf.ks, sf.account}
	transactor, err := f.newTransactor(context.Background(), sf.ks, sf.account, big.NewInt(0), 1000)
	assert.NoError(t, err, "Creating Transactor should succeed")
	assert.Equal(t, sf.account.Address, transactor.From, "Transactor address not properly set")
	assert.Equal(t, uint64(1000), transactor.GasLimit, "Gas limit not set properly")
	assert.Equal(t, big.NewInt(0), transactor.Value, "Transaction value not set properly")
	assert.Equal(t, big.NewInt(1), transactor.GasPrice, "Invalid gas price")
	transactor, err = f.newTransactor(context.Background(), sf.ks, sf.account, big.NewInt(12345), 12345)
	assert.NoError(t, err, "Creating Transactor should succeed")
	assert.Equal(t, sf.account.Address, transactor.From, "Transactor address not properly set")
	assert.Equal(t, uint64(12345), transactor.GasLimit, "Gas limit not set properly")
	assert.Equal(t, big.NewInt(12345), transactor.Value, "Transaction value not set properly")
	assert.Equal(t, big.NewInt(1), transactor.GasPrice, "Invalid gas price")
}

func Test_NewWatchOpts(t *testing.T) {
	f := &ContractBackend{}
	assert.Panics(t, func() { f.newWatchOpts(context.Background()) }, "Creating watchopts on invalid backend should panic")
	sf := newSimulatedFunder()
	f = &ContractBackend{sf.ContractBackend, sf.ks, sf.account}
	watchOpts, err := f.newWatchOpts(context.Background())
	assert.NoError(t, err, "Creating watchopts on valid ContractBackend should succeed")
	assert.Equal(t, context.Background(), watchOpts.Context, "Creating watchopts with context should succeed")
	assert.Equal(t, uint64(1), *watchOpts.Start, "Creating watchopts with no context should succeed")
}

func BenchmarkFunder(b *testing.B) {
	var t testing.T
	for i := 0; i < b.N; i++ {
		TestFunder_Fund(&t)
	}
}

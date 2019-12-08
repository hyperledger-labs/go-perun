// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"perun.network/go-perun/backend/ethereum/wallet"
	"perun.network/go-perun/channel"
	perunwallet "perun.network/go-perun/wallet"
)

type contractInterface interface {
	bind.ContractBackend
	BlockByNumber(context.Context, *big.Int) (*types.Block, error)
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
}

type contractBackend struct {
	contractInterface
}

func (c *contractBackend) newWatchOpts(ctx context.Context, startBlock uint64) *bind.WatchOpts {
	return &bind.WatchOpts{
		Start:   &startBlock,
		Context: ctx,
	}
}

func (c *contractBackend) newTransactor(ctx context.Context, ks *keystore.KeyStore, acc *accounts.Account, value *big.Int, gasLimit uint64) (*bind.TransactOpts, error) {
	if ks == nil {
		return nil, errors.New("contract backend is not configured properly")
	}
	nonce, err := c.PendingNonceAt(ctx, acc.Address)
	if err != nil {
		return nil, err
	}

	_, err = c.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	auth, err := bind.NewKeyStoreTransactor(ks, *acc)
	if err != nil {
		return nil, err
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = value       // in wei
	auth.GasLimit = gasLimit // in units
	auth.GasPrice = big.NewInt(0)

	return auth, nil
}

func calcFundingIDs(participants []perunwallet.Address, channelID channel.ID) ([][32]byte, error) {
	partIDs := make([][32]byte, len(participants))
	args := abi.Arguments{{Type: abibytes32}, {Type: abiaddress}}
	for idx, pID := range participants {
		address := pID.(*wallet.Address)
		bytes, err := args.Pack(channelID, address.Address)
		if err != nil {
			return nil, errors.Wrap(err, "Could not pack values")
		}
		partIDs[idx] = crypto.Keccak256Hash(bytes)
	}
	return partIDs, nil
}

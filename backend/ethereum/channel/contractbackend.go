// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package channel

import (
	"context"
	"log"
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

// How many blocks we query into the past for events.
const startBlockOffset = 100

// GasLimit is the max amount of gas we want to send per transaction.
const GasLimit = 200000

type ContractInterface interface {
	bind.ContractBackend
	BlockByNumber(context.Context, *big.Int) (*types.Block, error)
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
}

type ContractBackend struct {
	ContractInterface
	ks      *keystore.KeyStore
	account *accounts.Account
}

// NewContractBackend creates a new ContractBackend with the given parameters.
func NewContractBackend(cf ContractInterface, ks *keystore.KeyStore, acc *accounts.Account) ContractBackend {
	return ContractBackend{
		ContractInterface: cf,
		ks:                ks,
		account:           acc,
	}
}

func (c *ContractBackend) newWatchOpts(ctx context.Context) (*bind.WatchOpts, error) {
	latestBlock, err := c.BlockByNumber(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Could not retrieve latest block")
	}
	// max(1, latestBlock - offset)
	var blockNum uint64
	if latestBlock.NumberU64() > startBlockOffset {
		blockNum = latestBlock.NumberU64() - startBlockOffset
	} else {
		blockNum = 1
	}

	return &bind.WatchOpts{
		Start:   &blockNum,
		Context: ctx,
	}, nil
}

func (c *ContractBackend) newTransactor(ctx context.Context, value *big.Int, gasLimit uint64) (*bind.TransactOpts, error) {
	nonce, err := c.PendingNonceAt(ctx, c.account.Address)
	if err != nil {
		return nil, err
	}

	gasPrice, err := c.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	auth, err := bind.NewKeyStoreTransactor(c.ks, *c.account)
	if err != nil {
		return nil, err
	}

	auth.Nonce = new(big.Int).SetUint64(nonce)
	auth.Value = value       // in wei
	auth.GasLimit = gasLimit // in units
	auth.GasPrice = gasPrice

	return auth, nil
}

func calcFundingIDs(participants []perunwallet.Address, channelID channel.ID) [][32]byte {
	partIDs := make([][32]byte, len(participants))
	args := abi.Arguments{{Type: abibytes32}, {Type: abiaddress}}
	for idx, pID := range participants {
		address := pID.(*wallet.Address)
		bytes, err := args.Pack(channelID, address.Address)
		if err != nil {
			log.Panicf("error packing values: %v", err)
		}
		partIDs[idx] = crypto.Keccak256Hash(bytes)
	}
	return partIDs
}

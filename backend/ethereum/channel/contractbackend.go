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
const GasLimit = 500000

// ContractInterface provides all functions needed by an ethereum backend.
// Both test.SimulatedBackend and ethclient.Client implement this interface.
type ContractInterface interface {
	bind.ContractBackend
	BlockByNumber(context.Context, *big.Int) (*types.Block, error)
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
}

// ContractBackend adds a keystore and an on-chain account to the ContractInterface.
// This is needed to send on-chain transaction to interact with the smart contracts.
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
	blockNum, err := c.getBlockNum(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "new watch opts")
	}

	return &bind.WatchOpts{
		Start:   &blockNum,
		Context: ctx,
	}, nil
}

func (c *ContractBackend) newFilterOpts(ctx context.Context) (*bind.FilterOpts, error) {
	blockNum, err := c.getBlockNum(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "new filter opts")
	}
	return &bind.FilterOpts{
		Start:   blockNum,
		End:     nil,
		Context: ctx,
	}, nil
}

func (c *ContractBackend) getBlockNum(ctx context.Context) (uint64, error) {
	latestBlock, err := c.BlockByNumber(ctx, nil)
	if err != nil {
		return uint64(0), errors.Wrap(err, "Could not retrieve latest block")
	}
	// max(1, latestBlock - offset)
	var blockNum uint64
	if latestBlock.NumberU64() > startBlockOffset {
		blockNum = latestBlock.NumberU64() - startBlockOffset
	} else {
		blockNum = 1
	}
	return blockNum, nil
}

func (c *ContractBackend) newTransactor(ctx context.Context, valueWei *big.Int, gasLimit uint64) (*bind.TransactOpts, error) {
	nonce, err := c.PendingNonceAt(ctx, c.account.Address)
	if err != nil {
		return nil, errors.Wrap(err, "querying pending nonce")
	}

	gasPrice, err := c.SuggestGasPrice(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "querying suggested gas price")
	}

	auth, err := bind.NewKeyStoreTransactor(c.ks, *c.account)
	if err != nil {
		return nil, errors.Wrap(err, "creating transactor")
	}

	auth.Nonce = new(big.Int).SetUint64(nonce)
	auth.Value = valueWei    // in wei
	auth.GasLimit = gasLimit // in units
	auth.GasPrice = gasPrice

	return auth, nil
}

func calcFundingIDs(channelID channel.ID, participants ...perunwallet.Address) [][32]byte {
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

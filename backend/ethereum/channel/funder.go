// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel // import "perun.network/go-perun/backend/ethereum/channel"

import (
	"bytes"
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/event"
	"github.com/pkg/errors"

	"perun.network/go-perun/backend/ethereum/bindings/assets"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/log"
)

const gasLimit = 200000

var (
	// Declaration for abi-encoding.
	abibytes32, _ = abi.NewType("bytes32", nil)
	abiaddress, _ = abi.NewType("address", nil)
)

type contract struct {
	*assets.AssetHolder
	*common.Address
}

// Funder implements the channel.Funder interface for Ethereum.
type Funder struct {
	client  contractBackend
	ks      *keystore.KeyStore
	account *accounts.Account
	// ETHAssetHolder is the on-chain address of the ETH asset holder.
	// This is needed to distinguish between ETH and ERC-20 transactions.
	ethAssetHolder common.Address
}

// compile time check that we implement the perun funder interface
var _ channel.Funder = (*Funder)(nil)

// NewETHFunder creates a new ethereum funder.
func NewETHFunder(client *ethclient.Client, keystore *keystore.KeyStore, account *accounts.Account, ethAssetHolder common.Address) Funder {
	return Funder{
		client:         contractBackend{client},
		ks:             keystore,
		account:        account,
		ethAssetHolder: ethAssetHolder,
	}
}

// Fund implements the funder interface.
// It can be used to fund state channels on the ethereum blockchain.
func (f *Funder) Fund(ctx context.Context, request channel.FundingReq) error {
	if request.Params == nil || request.Allocation == nil {
		panic("invalid funding request")
	}
	var channelID = request.Params.ID()
	log.Debugf("Funding Channel with ChannelID %d", channelID)

	partIDs, err := calcFundingIDs(request.Params.Parts, channelID)
	if err != nil {
		return err
	}

	contracts, err := f.connectToContracts(request)
	if err != nil {
		return errors.Wrap(err, "Connecting to contracts failed")
	}

	confirmation := make(chan error)
	go func() {
		confirmation <- f.waitForFundingConfirmations(ctx, request, contracts, partIDs)
	}()

	if err := f.fundAssets(ctx, request, contracts, partIDs); err != nil {
		return errors.Wrap(err, "Funding assets failed")
	}

	return <-confirmation
}

func (f *Funder) connectToContracts(request channel.FundingReq) ([]contract, error) {
	contracts := make([]contract, len(request.Allocation.Assets))
	// Connect to all AssetHolder contracts.
	for assetIndex, asset := range request.Allocation.Assets {
		// Decode and set the asset address.
		assetAddr := asset.(*Asset).Address
		ctr, err := assets.NewAssetHolder(assetAddr, f.client)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Could not connect to asset holder %d", assetIndex))
		}
		contracts[assetIndex] = contract{ctr, &assetAddr}
	}
	return contracts, nil
}

func (f *Funder) fundAssets(ctx context.Context, request channel.FundingReq, contracts []contract, partIDs [][32]byte) (err error) {
	// Connect to all AssetHolder contracts.
	for assetIndex, asset := range contracts {
		// Create a new transaction (needs to be cloned because of go-ethereum bug).
		balance := new(big.Int).Set(request.Allocation.OfParts[request.Idx][assetIndex])
		var auth *bind.TransactOpts
		// If we want to fund the channel with ether, send eth in transaction.
		if bytes.Equal(asset.Bytes(), f.ethAssetHolder.Bytes()) {
			auth, err = f.client.newTransactor(ctx, f.ks, f.account, balance, gasLimit)
		} else {
			auth, err = f.client.newTransactor(ctx, f.ks, f.account, big.NewInt(0), gasLimit)
		}
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Could not create Transactor for asset %d", assetIndex))
		}
		// Call the asset holder contract.
		tx, err := contracts[assetIndex].Deposit(auth, partIDs[request.Idx], balance)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf(("Deposit failed for asset %d"), assetIndex))
		}
		log.Debugf("Sending transaction to the blockchain with txHash: ", tx.Hash().Hex())
	}
	return nil
}

// waitForFundingConfirmations waits for the confirmations events on the blockchain that
// both we and all peers sucessfully funded the channel.
func (f *Funder) waitForFundingConfirmations(ctx context.Context, request channel.FundingReq, contracts []contract, partIDs [][32]byte) error {
	deposited := make(chan *assets.AssetHolderDeposited)
	subs := make([]event.Subscription, len(contracts))
	// Wait for confirmation on each asset.
	for assetIndex := range contracts {
		watchOpts, err := f.client.newWatchOpts(ctx)
		if err != nil {
			return errors.Wrap(err, "error creating watchopts")
		}
		subs[assetIndex], err = contracts[assetIndex].WatchDeposited(watchOpts, deposited, partIDs)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("WatchDeposit on asset %d failed", assetIndex))
		}
	}

	allocation := request.Allocation.Clone()
	for i := 0; i < len(request.Params.Parts)*len(contracts); i++ {
		select {
		case event := <-deposited:
			// Calculate the position in the participant array.
			idx := -1
			for k, ele := range partIDs {
				if ele == event.FundingID {
					idx = k
				}
			}
			// Retrieve the position in the asset array.
			assetIdx := -1
			for k, ele := range contracts {
				if *ele.Address == event.Raw.Address {
					assetIdx = k
				}
			}
			// Check if the participant send the correct amounts of funds.
			if allocation.OfParts[idx][assetIdx].Cmp(event.Amount) != 0 {
				return errors.New("deposit in asset %d from pariticipant %d does not match agreed upon asset")
			}
			allocation.OfParts[idx][assetIdx] = big.NewInt(0)
		case <-ctx.Done():
			return errors.Wrap(ctx.Err(), "Waiting for events cancelled by context")
		case err := <-subs[i].Err():
			return errors.Wrap(err, "Error while waiting for events")
		}
	}
	// Check if everyone funded correctly.
	for i := 0; i < len(allocation.OfParts); i++ {
		for k := 0; k < len(allocation.OfParts[i]); k++ {
			if allocation.OfParts[i][k].Cmp(big.NewInt(0)) != 0 {
				var err channel.PeerTimedOutFundingError
				err.TimedOutPeerIdx = uint16(i)
				return err
			}
		}
	}
	return nil
}

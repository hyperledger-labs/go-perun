// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package channel // import "perun.network/go-perun/backend/ethereum/channel"

import (
	"bytes"
	"context"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/pkg/errors"

	"perun.network/go-perun/backend/ethereum/bindings/assets"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/log"
)

var (
	// Declaration for abi-encoding.
	abibytes32, _ = abi.NewType("bytes32", nil)
	abiaddress, _ = abi.NewType("address", nil)
	// Error that is returned if an event was not found in the past.
	errEventNotFound = errors.New("Event not found")
)

type assetHolder struct {
	*assets.AssetHolder
	*common.Address
}

// Funder implements the channel.Funder interface for Ethereum.
type Funder struct {
	ContractBackend
	mu sync.Mutex
	// ETHAssetHolder is the on-chain address of the ETH asset holder.
	// This is needed to distinguish between ETH and ERC-20 transactions.
	ethAssetHolder common.Address
}

// compile time check that we implement the perun funder interface
var _ channel.Funder = (*Funder)(nil)

// NewETHFunder creates a new ethereum funder.
func NewETHFunder(backend ContractBackend, ethAssetHolder common.Address) *Funder {
	return &Funder{
		ContractBackend: backend,
		ethAssetHolder:  ethAssetHolder,
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

	partIDs := calcFundingIDs(request.Params.Parts, channelID)

	contracts, err := f.connectToContracts(request.Allocation.Assets)
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

func (f *Funder) connectToContracts(assetHolders []channel.Asset) ([]assetHolder, error) {
	contracts := make([]assetHolder, len(assetHolders))
	// Connect to all AssetHolder contracts.
	for assetIndex, asset := range assetHolders {
		// Decode and set the asset address.
		assetAddr := asset.(*Asset).Address
		ctr, err := assets.NewAssetHolder(assetAddr, f)
		if err != nil {
			return nil, errors.Wrapf(err, "connecting to assetholder %d", assetIndex)
		}
		contracts[assetIndex] = assetHolder{ctr, &assetAddr}
	}
	return contracts, nil
}

func (f *Funder) fundAssets(ctx context.Context, request channel.FundingReq, contracts []assetHolder, partIDs [][32]byte) (err error) {
	// Connect to all AssetHolder contracts.
	for assetIndex, asset := range contracts {
		// Create a new transaction (needs to be cloned because of go-ethereum bug).
		// See https://github.com/ethereum/go-ethereum/pull/20412
		balance := new(big.Int).Set(request.Allocation.OfParts[request.Idx][assetIndex])
		var auth *bind.TransactOpts
		// If we want to fund the channel with ether, send eth in transaction.
		tx, err := func() (*types.Transaction, error) {
			f.mu.Lock()
			defer f.mu.Unlock()
			if bytes.Equal(asset.Bytes(), f.ethAssetHolder.Bytes()) {
				auth, err = f.newTransactor(ctx, balance, GasLimit)
			} else {
				auth, err = f.newTransactor(ctx, big.NewInt(0), GasLimit)
			}
			if err != nil {
				return nil, errors.Wrapf(err, "creating transactor for asset %d", assetIndex)
			}
			// Call the asset holder contract.
			tx, err := contracts[assetIndex].Deposit(auth, partIDs[request.Idx], balance)
			return tx, errors.WithStack(err)
		}()

		if err != nil {
			return errors.WithMessagef(err, "depositing asset %d", assetIndex)
		}
		if err := execSuccessful(ctx, f.ContractBackend, tx); err != nil {
			return errors.WithMessage(err, "mining transaction")
		}
		log.Debugf("Sending transaction to the blockchain with txHash: %v", tx.Hash().Hex())
	}
	return nil
}

func filterOldEvents(ctx context.Context, asset assetHolder, deposited chan *assets.AssetHolderDeposited, partIDs [][32]byte) error {
	// Filter
	filterOpts := bind.FilterOpts{
		Start:   uint64(1),
		End:     nil,
		Context: ctx}
	iter, err := asset.FilterDeposited(&filterOpts, partIDs)
	if err != nil {
		return errors.Wrap(err, "filtering deposited events")
	}
	if !iter.Next() {
		// Event not found.
		return errEventNotFound
	}
	deposited <- iter.Event
	return nil
}

// waitForFundingConfirmations waits for the confirmations events on the blockchain that
// both we and all peers sucessfully funded the channel.
func (f *Funder) waitForFundingConfirmations(ctx context.Context, request channel.FundingReq, contracts []assetHolder, partIDs [][32]byte) error {
	deposited := make(chan *assets.AssetHolderDeposited)
	subs := make([]event.Subscription, len(contracts))
	defer func() {
		for _, sub := range subs {
			sub.Unsubscribe()
		}
	}()
	// Wait for confirmation on each asset.
	for assetIndex := range contracts {
		// Watch new events
		watchOpts, err := f.newWatchOpts(ctx)
		if err != nil {
			return errors.WithMessage(err, "error creating watchopts")
		}
		sub, err := contracts[assetIndex].WatchDeposited(watchOpts, deposited, partIDs)
		if err != nil {
			return errors.Wrapf(err, "WatchDeposit on asset %d failed", assetIndex)
		}
		subs[assetIndex] = sub
	}

	// Query old events
	go func() {
		for assetIndex := range contracts {
			if err := filterOldEvents(ctx, contracts[assetIndex], deposited, partIDs); err != nil && err != errEventNotFound {
				log.Warnf("Error filtering old Deposited events for asset[%d]: %v", assetIndex, err)
			}
		}
	}()

	allocation := request.Allocation.Clone()
	for i := 0; i < len(contracts); i++ {
		for k := 0; k < len(request.Params.Parts); k++ {
			select {
			case event := <-deposited:
				// Calculate the position in the participant array.
				idx := -1
				for h, id := range partIDs {
					if id == event.FundingID {
						idx = h
						break
					}
				}
				// Retrieve the position in the asset array.
				assetIdx := -1
				for h, ctr := range contracts {
					if *ctr.Address == event.Raw.Address {
						assetIdx = h
						break
					}
				}
				log.Debugf(
					"Deposited event received for asset %d and participant %d, id: %v",
					assetIdx, idx, event.FundingID)
				// Check if the participant sent the correct amounts of funds.
				if allocation.OfParts[idx][assetIdx].Cmp(event.Amount) != 0 {
					return errors.Errorf("deposit in asset %d from pariticipant %d does not match agreed upon asset", assetIdx, idx)
				}
				allocation.OfParts[idx][assetIdx] = big.NewInt(0)
			case <-ctx.Done():
				return errors.Wrap(ctx.Err(), "Waiting for events cancelled by context")
			case err := <-subs[i].Err():
				return errors.Wrap(err, "Error while waiting for events")
			}
		}
	}

	// Check if everyone funded correctly.
	for i := 0; i < len(allocation.OfParts); i++ {
		for k := 0; k < len(allocation.OfParts[i]); k++ {
			if allocation.OfParts[i][k].Cmp(big.NewInt(0)) != 0 {
				return channel.NewPeerTimedOutFundingError(uint16(1))
			}
		}
	}
	return nil
}

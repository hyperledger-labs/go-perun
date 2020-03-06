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
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"

	"perun.network/go-perun/backend/ethereum/bindings/assets"
	"perun.network/go-perun/backend/ethereum/wallet"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/log"
	perunwallet "perun.network/go-perun/wallet"
)

type assetHolder struct {
	*assets.AssetHolder
	*common.Address
	assetIndex int
}

// Funder implements the channel.Funder interface for Ethereum.
type Funder struct {
	ContractBackend
	mu  sync.Mutex
	log log.Logger // structured logger
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
		log:             log.WithField("account", backend.account.Address),
	}
}

// Fund implements the funder interface.
// It can be used to fund state channels on the ethereum blockchain.
func (f *Funder) Fund(ctx context.Context, request channel.FundingReq) error {
	var channelID = request.Params.ID()
	f.log.WithField("channel", channelID).Debug("Funding Channel.")

	partIDs := FundingIDs(channelID, request.Params.Parts...)

	errChan := make(chan error, len(request.Allocation.Assets))
	errs := make([]*channel.AssetFundingError, len(request.Allocation.Assets))
	var wg sync.WaitGroup
	wg.Add(len(request.Allocation.Assets))
	for index, asset := range request.Allocation.Assets {
		go func(index int, asset channel.Asset) {
			defer wg.Done()
			if err := f.fundAsset(ctx, request, index, asset, partIDs, errs); err != nil {
				errChan <- errors.WithMessage(err, "fund asset")
			}
		}(index, asset)
	}
	wg.Wait()
	close(errChan)
	if err := <-errChan; err != nil {
		return err
	}

	prunedErrs := errs[:0]
	for _, e := range errs {
		if e != nil {
			prunedErrs = append(prunedErrs, e)
		}
	}
	return channel.NewFundingTimeoutError(prunedErrs)
}

func (f *Funder) fundAsset(ctx context.Context, request channel.FundingReq, index int, asset channel.Asset, partIDs [][32]byte, errs []*channel.AssetFundingError) error {
	contract, err := f.connectToContract(asset, index)
	if err != nil {
		return errors.Wrap(err, "connecting to contracts")
	}

	confirmation := make(chan error)
	go func() {
		confirmation <- f.waitForFundingConfirmation(ctx, request, contract, partIDs)
	}()

	if err := f.sendFundingTransaction(ctx, request, contract, partIDs); err != nil {
		return errors.Wrap(err, "sending funding tx")
	}
	err = <-confirmation
	if channel.IsAssetFundingError(err) {
		errs[index] = err.(*channel.AssetFundingError)
	} else if err != nil {
		return err
	}
	return nil
}

func (f *Funder) connectToContract(asset channel.Asset, assetIndex int) (assetHolder, error) {
	// Decode and set the asset address.
	assetAddr := asset.(*Asset).Address
	ctr, err := assets.NewAssetHolder(assetAddr, f)
	if err != nil {
		return assetHolder{}, errors.Wrapf(err, "connecting to assetholder")
	}
	return assetHolder{ctr, &assetAddr, assetIndex}, nil
}

func (f *Funder) sendFundingTransaction(ctx context.Context, request channel.FundingReq, asset assetHolder, partIDs [][32]byte) error {
	tx, err := f.createFundingTx(ctx, request, asset, partIDs)
	if err != nil {
		return errors.WithMessagef(err, "depositing asset %d", asset.assetIndex)
	}
	if err := confirmTransaction(ctx, f.ContractBackend, tx); err != nil {
		return errors.WithMessage(err, "mining transaction")
	}
	f.log.Debugf("peer[%d] Transaction with txHash: [%v] executed successful", request.Idx, tx.Hash().Hex())
	return nil
}

func (f *Funder) createFundingTx(ctx context.Context, request channel.FundingReq, asset assetHolder, partIDs [][32]byte) (*types.Transaction, error) {
	// Create a new transaction (needs to be cloned because of go-ethereum bug).
	// See https://github.com/ethereum/go-ethereum/pull/20412
	balance := new(big.Int).Set(request.Allocation.Balances[request.Idx][asset.assetIndex])
	// Lock the funder for correct nonce usage.
	f.mu.Lock()
	defer f.mu.Unlock()
	var auth *bind.TransactOpts
	var errI error
	if bytes.Equal(asset.Bytes(), f.ethAssetHolder.Bytes()) {
		// If we want to fund the channel with ether, send eth in transaction.
		auth, errI = f.NewTransactor(ctx, balance, GasLimit)
	} else {
		auth, errI = f.NewTransactor(ctx, big.NewInt(0), GasLimit)
	}
	if errI != nil {
		return nil, errors.Wrapf(errI, "creating transactor for asset %d", asset.assetIndex)
	}
	// Call the asset holder contract.
	tx, err := asset.Deposit(auth, partIDs[request.Idx], balance)
	f.log.Debugf("peer[%d] Created funding transaction with txHash: %v, amount %d", request.Idx, tx.Hash().Hex(), balance)
	return tx, errors.WithStack(err)
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
	for iter.Next() {
		deposited <- iter.Event
	}
	return nil
}

// waitForFundingConfirmation waits for the confirmation events on the blockchain that
// both we and all peers successfully funded the channel.
func (f *Funder) waitForFundingConfirmation(ctx context.Context, request channel.FundingReq, asset assetHolder, partIDs [][32]byte) error {
	deposited := make(chan *assets.AssetHolderDeposited)
	// Watch new events
	watchOpts, err := f.NewWatchOpts(ctx)
	if err != nil {
		return errors.WithMessage(err, "error creating watchopts")
	}
	sub, err := asset.WatchDeposited(watchOpts, deposited, partIDs)
	if err != nil {
		return errors.Wrapf(err, "WatchDeposit on asset %d failed", asset.assetIndex)
	}
	defer sub.Unsubscribe()

	// we let the filter queries and all subscription errors write into this error
	// channel.
	errChan := make(chan error, 1)
	go func() {
		errChan <- errors.Wrapf(<-sub.Err(), "subscription for asset %d", asset.assetIndex)
	}()

	// Query old events
	go func() {
		if err := filterOldEvents(ctx, asset, deposited, partIDs); err != nil {
			errChan <- errors.WithMessagef(err, "filtering old Deposited events for asset %d", asset.assetIndex)
		}
	}()

	allocation := request.Allocation.Clone()
	N := len(request.Params.Parts)
	for N > 0 {
		select {
		case event := <-deposited:
			log := f.log.WithField("fundingID", event.FundingID)
			log.Debugf("peer[%d] Received event with amount %v", request.Idx, event.Amount)

			// Calculate the position in the participant array.
			idx := -1
			for h, id := range partIDs {
				if id == event.FundingID {
					idx = h
					break
				}
			}

			amount := allocation.Balances[idx][asset.assetIndex]
			if amount.Sign() == 0 {
				continue // ignore double events
			}

			log.Debugf("Deposited event received for asset %d and participant %d", asset.assetIndex, idx)

			amount.Sub(amount, event.Amount)
			if amount.Sign() != 1 {
				// participant funded successfully
				N--
				allocation.Balances[idx][asset.assetIndex] = big.NewInt(0)
			}

		case <-ctx.Done():
			var indices []channel.Index
			for k, bals := range allocation.Balances {
				if bals[asset.assetIndex].Sign() == 1 {
					indices = append(indices, channel.Index(k))
				}
			}
			if len(indices) != 0 {
				return &channel.AssetFundingError{Asset: asset.assetIndex, TimedOutPeers: indices}
			}
			return nil
		case err := <-errChan:
			return err
		}
	}
	return nil
}

// FundingIDs returns a slice the same size as the number of passed participants
// where each entry contains the hash Keccak256(channel id || participant address).
func FundingIDs(channelID channel.ID, participants ...perunwallet.Address) [][32]byte {
	partIDs := make([][32]byte, len(participants))
	args := abi.Arguments{{Type: abiBytes32}, {Type: abiAddress}}
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

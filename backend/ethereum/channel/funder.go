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

package channel

import (
	"context"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"

	"perun.network/go-perun/backend/ethereum/bindings/assets"
	"perun.network/go-perun/backend/ethereum/wallet"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/log"
	pcontext "perun.network/go-perun/pkg/context"
	perunwallet "perun.network/go-perun/wallet"
)

type assetHolder struct {
	*assets.AssetHolder
	*common.Address
	assetIndex channel.Index
}

// Funder implements the channel.Funder interface for Ethereum.
type Funder struct {
	ContractBackend
	// accounts associates an Account to every AssetIndex.
	accounts map[Asset]accounts.Account
	// depositors associates a Depositor to every AssetIndex.
	depositors map[Asset]Depositor
	log        log.Logger // structured logger
}

// compile time check that we implement the perun funder interface.
var _ channel.Funder = (*Funder)(nil)

// NewFunder creates a new ethereum funder.
func NewFunder(backend ContractBackend, accounts map[Asset]accounts.Account, depositors map[Asset]Depositor) *Funder {
	return &Funder{
		ContractBackend: backend,
		accounts:        accounts,
		depositors:      depositors,
		log:             log.Get(),
	}
}

// Fund implements the channel.Funder interface. It funds all assets in
// parallel. If not all participants successfully fund within a timeframe of
// ChallengeDuration seconds, Fund returns a FundingTimeoutError.
//
// If funding on a real blockchain, make sure that the passed context doesn't
// cancel before the funding period of length ChallengeDuration elapses, or
// funding will be canceled prematurely.
func (f *Funder) Fund(ctx context.Context, request channel.FundingReq) error {
	var channelID = request.Params.ID()
	f.log.WithField("channel", channelID).Debug("Funding Channel.")

	// We wait for the funding timeout in a go routine and cancel the funding
	// context if the timeout elapses.
	timeout, err := NewBlockTimeoutDuration(ctx, f.ContractInterface, request.Params.ChallengeDuration)
	if err != nil {
		return errors.WithMessage(err, "creating block timeout")
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel() // in case we return before block timeout
	go func() {
		if err := timeout.Wait(ctx); err != nil && !pcontext.IsContextError(err) {
			f.log.Warn("Fund: BlockTimeout.Wait runtime error: ", err)
		}
		cancel() // cancel funding context on funding timeout
	}()

	fundingIDs := FundingIDs(channelID, request.Params.Parts...)

	errChan := make(chan error, len(request.State.Assets))
	errs := make([]*channel.AssetFundingError, len(request.State.Assets))
	var wg sync.WaitGroup
	wg.Add(len(request.State.Assets))
	for index, perunAsset := range request.State.Assets {
		go func(index channel.Index, perunAsset channel.Asset) {
			defer wg.Done()
			asset := *perunAsset.(*Asset)
			if err := f.fundAsset(ctx, request, index, asset, fundingIDs, errs); err != nil {
				f.log.WithField("asset", asset).WithError(err).Error("Could not fund asset")
				errChan <- errors.WithMessage(err, "fund asset")
			}
		}(channel.Index(index), perunAsset)
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

func (f *Funder) fundAsset(ctx context.Context, request channel.FundingReq, assetIndex channel.Index, asset Asset, fundingIDs [][32]byte, errs []*channel.AssetFundingError) error {
	contract, err := f.bindContract(asset, assetIndex)
	if err != nil {
		return errors.Wrap(err, "connecting to contracts")
	}

	// start watching for deposit events, also including past ones
	confirmation := make(chan error)
	go func() {
		confirmation <- f.waitForFundingConfirmation(ctx, request, contract, fundingIDs)
	}()

	bal := request.State.Balances[assetIndex][request.Idx]
	if bal.Sign() <= 0 {
		f.log.WithFields(log.Fields{"channel": request.Params.ID(), "idx": request.Idx}).Debug("Skipped zero funding.")
	} else if alreadyFunded, err := checkFunded(ctx, request, contract, fundingIDs[request.Idx]); err != nil {
		return errors.WithMessage(err, "checking funded")
	} else if alreadyFunded {
		f.log.WithFields(log.Fields{"channel": request.Params.ID(), "idx": request.Idx}).Debug("Skipped second funding.")
	} else if err := f.deposit(ctx, bal, asset, fundingIDs[request.Idx]); err != nil {
		return errors.WithMessage(err, "depositing funds")
	}

	err = <-confirmation
	if channel.IsAssetFundingError(err) {
		errs[assetIndex] = err.(*channel.AssetFundingError)
	} else if err != nil {
		return err
	}
	return nil
}

// deposit deposits funds for one funding-ID by calling the associated Depositor.
// Returns an error if no matching Depositor or Account could be found.
func (f *Funder) deposit(ctx context.Context, bal *big.Int, asset Asset, fundingID [32]byte) error {
	depositor, ok := f.depositors[asset]
	if !ok {
		return errors.Errorf("could not find Depositor for asset #%d", asset)
	}
	acc, ok := f.accounts[asset]
	if !ok {
		return errors.Errorf("could not find account for asset #%d", asset)
	}

	dReq := NewDepositReq(bal, f.ContractBackend, asset, acc, fundingID)
	txs, err := depositor.Deposit(ctx, *dReq)
	if err != nil {
		return errors.WithMessage(err, "Depositor.deposit")
	}

	for i, tx := range txs {
		if err := f.confirmTransaction(ctx, tx, acc); err != nil {
			return errors.WithMessagef(err, "sending %dth Depositor-tx", i)
		}
		f.log.Debugf("Mined TX: %v", tx.Hash().Hex())
	}
	return nil
}

// checkFunded returns whether the funding for `request` was already complete.
func checkFunded(ctx context.Context, request channel.FundingReq, asset assetHolder, partID [32]byte) (bool, error) {
	iter, err := filterFunds(ctx, asset, partID)
	if err != nil {
		return false, errors.WithMessagef(err, "filtering old Funding events for asset %d", asset.assetIndex)
	}
	// nolint:errcheck
	defer iter.Close()

	amount := new(big.Int).Set(request.State.Balances[asset.assetIndex][request.Idx])
	for iter.Next() {
		amount.Sub(amount, iter.Event.Amount)
	}
	return amount.Sign() != 1, iter.Error()
}

func (f *Funder) bindContract(asset Asset, assetIndex channel.Index) (assetHolder, error) {
	// Decode and set the asset address.
	assetAddr := common.Address(asset)
	ctr, err := assets.NewAssetHolder(assetAddr, f)
	if err != nil {
		return assetHolder{}, errors.Wrapf(err, "connecting to assetholder")
	}
	return assetHolder{ctr, &assetAddr, assetIndex}, nil
}

func filterFunds(ctx context.Context, asset assetHolder, fundingIDs ...[32]byte) (*assets.AssetHolderDepositedIterator, error) {
	// Filter
	filterOpts := bind.FilterOpts{
		Start:   uint64(1),
		End:     nil,
		Context: ctx}
	iter, err := asset.FilterDeposited(&filterOpts, fundingIDs)
	if err != nil {
		return nil, errors.Wrap(err, "filtering deposited events")
	}

	return iter, nil
}

// waitForFundingConfirmation waits for the confirmation events on the blockchain that
// both we and all peers successfully funded the channel.
// nolint: funlen
func (f *Funder) waitForFundingConfirmation(ctx context.Context, request channel.FundingReq, asset assetHolder, fundingIDs [][32]byte) error {
	deposited := make(chan *assets.AssetHolderDeposited)
	// Watch new events
	watchOpts, err := f.NewWatchOpts(ctx)
	if err != nil {
		return errors.WithMessage(err, "error creating watchopts")
	}
	sub, err := asset.WatchDeposited(watchOpts, deposited, fundingIDs)
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

	// Query all old funding events
	go func() {
		iter, err := filterFunds(ctx, asset, fundingIDs...)
		if err != nil {
			errChan <- errors.WithMessagef(err, "filtering old Deposited events for asset %d", asset.assetIndex)
			return
		}
		defer iter.Close() // nolint: errcheck
		for iter.Next() {
			deposited <- iter.Event
		}
	}()

	allocation := request.State.Allocation.Clone()
	// Count how many zero balance funding requests are there
	N := len(request.Params.Parts) - countZeroBalances(&allocation, asset.assetIndex)

	// Wait for all non-zero funding requests
	for N > 0 {
		select {
		case event := <-deposited:
			log := f.log.WithField("fundingID", event.FundingID)
			log.Debugf("peer[%d] Received event with amount %v", request.Idx, event.Amount)

			// Calculate the position in the participant array.
			idx := getPartIdx(event.FundingID, fundingIDs)

			amount := allocation.Balances[asset.assetIndex][idx]
			if amount.Sign() == 0 {
				continue // ignore double events
			}

			log.Debugf("Deposited event received for asset %d and participant %d", asset.assetIndex, idx)

			amount.Sub(amount, event.Amount)
			if amount.Sign() != 1 {
				// participant funded successfully
				N--
				allocation.Balances[asset.assetIndex][idx].SetUint64(0)
			}

		case <-ctx.Done():
			var indices []channel.Index
			for k, bals := range allocation.Balances[asset.assetIndex] {
				if bals.Sign() == 1 {
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

func getPartIdx(partID [32]byte, fundingIDs [][32]byte) int {
	for i, id := range fundingIDs {
		if id == partID {
			return i
		}
	}
	return -1
}

func countZeroBalances(alloc *channel.Allocation, assetIndex channel.Index) (n int) {
	for _, part := range alloc.Balances[assetIndex] {
		if part.Sign() == 0 {
			n++
		}
	}
	return
}

// FundingIDs returns a slice the same size as the number of passed participants
// where each entry contains the hash Keccak256(channel id || participant address).
func FundingIDs(channelID channel.ID, participants ...perunwallet.Address) [][32]byte {
	ids := make([][32]byte, len(participants))
	args := abi.Arguments{{Type: abiBytes32}, {Type: abiAddress}}
	for idx, pID := range participants {
		address := pID.(*wallet.Address)
		bytes, err := args.Pack(channelID, common.Address(*address))
		if err != nil {
			log.Panicf("error packing values: %v", err)
		}
		ids[idx] = crypto.Keccak256Hash(bytes)
	}
	return ids
}

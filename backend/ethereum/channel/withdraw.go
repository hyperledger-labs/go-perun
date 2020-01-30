// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package channel // import "perun.network/go-perun/backend/ethereum/channel"

import (
	"context"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"perun.network/go-perun/backend/ethereum/bindings/assets"
	"perun.network/go-perun/backend/ethereum/wallet"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/log"
)

func (a *Adjudicator) withdraw(ctx context.Context, request channel.AdjudicatorReq) error {
	assets := request.Tx.Allocation.Assets

	errChan := make(chan error, len(assets))
	var wg sync.WaitGroup
	wg.Add(len(assets))
	for index, asset := range assets {
		go func(index int, asset channel.Asset) {
			defer wg.Done()
			contract, err := connectToAssetHolder(a.ContractBackend, asset, index)
			if err != nil {
				errChan <- errors.Wrap(err, "Connecting to contracts failed")
				return
			}

			confirmation := make(chan error)
			go func() {
				confirmation <- a.waitForWithdrawnEvent(ctx, request, contract)
			}()

			if err := a.withdrawAsset(ctx, request, contract); err != nil {
				errChan <- errors.Wrap(err, "Funding assets failed")
				return
			}
			errChan <- <-confirmation
		}(index, asset)
	}
	wg.Wait()
	close(errChan)
	return <-errChan
}

func connectToAssetHolder(backend ContractBackend, asset channel.Asset, assetIndex int) (assetHolder, error) {
	// Decode and set the asset address.
	assetAddr := asset.(*Asset).Address
	ctr, err := assets.NewAssetHolder(assetAddr, backend)
	if err != nil {
		return assetHolder{}, errors.Wrapf(err, "connecting to assetholder")
	}
	return assetHolder{ctr, &assetAddr, assetIndex}, nil
}

func (a *Adjudicator) waitForWithdrawnEvent(ctx context.Context, request channel.AdjudicatorReq, asset assetHolder) error {
	withdrawn := make(chan *assets.AssetHolderWithdrawn)
	participant := request.Params.Parts[request.Idx].(*wallet.Address).Address
	// Watch new events
	watchOpts, err := a.newWatchOpts(ctx)
	if err != nil {
		return errors.WithMessage(err, "error creating watchopts")
	}
	sub, err := asset.WatchWithdrawn(watchOpts, withdrawn, []common.Address{participant})
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
		// Setup filter
		filterOpts, err := a.newFilterOpts(ctx)
		if err != nil {
			errChan <- err
		}
		iter, err := asset.FilterWithdrawn(filterOpts, []common.Address{participant})
		if err != nil {
			errChan <- errors.WithStack(err)
		}
		if iter.Next() {
			withdrawn <- iter.Event
		}
		// No event found, just return normally
	}()

	select {
	case event := <-withdrawn:
		log.Debugf("peer[%d] has successfully withdrawn %v", request.Idx, event.Amount)
		return nil
	case <-ctx.Done():
		errors.New("Timed out while withdrawing")
	case err := <-errChan:
		return err
	}
	return nil
}

func (a *Adjudicator) withdrawAsset(ctx context.Context, request channel.AdjudicatorReq, asset assetHolder) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	trans, err := a.newTransactor(ctx, big.NewInt(0), GasLimit)
	if err != nil {
		return errors.Wrapf(err, "creating transactor for asset %d", asset.assetIndex)
	}
	// Create a new Withdrawal authorization.
	auth, sig, err := a.newWithdrawalAuth(request, asset)
	if err != nil {
		return errors.Wrap(err, "creating withdrawal auth")
	}
	// Call the asset holder contract.
	ethTx, err := asset.Withdraw(trans, auth, sig)
	if err != nil {
		return errors.Wrapf(err, "depositing asset %d", asset.assetIndex)
	}
	return errors.WithMessage(execSuccessful(ctx, a.ContractBackend, ethTx), "mining transaction")
}

func (a *Adjudicator) newWithdrawalAuth(request channel.AdjudicatorReq, asset assetHolder) (assets.AssetHolderWithdrawalAuth, []byte, error) {
	auth := assets.AssetHolderWithdrawalAuth{
		ChannelID:   request.Params.ID(),
		Participant: request.Acc.Address().(*wallet.Address).Address,
		Receiver:    a.OnChainAddress,
		Amount:      request.Tx.Allocation.OfParts[request.Idx][asset.assetIndex],
	}
	enc, err := encodeAssetHolderWithdrawalAuth(auth)
	if err != nil {
		return assets.AssetHolderWithdrawalAuth{}, nil, errors.WithMessage(err, "encoding withdrawal auth")
	}

	sig, err := request.Acc.SignData(enc)
	return auth, sig, errors.WithMessage(err, "sign data")
}

func encodeAssetHolderWithdrawalAuth(auth assets.AssetHolderWithdrawalAuth) ([]byte, error) {
	// encodeAssetHolderWithdrawalAuth encodes the AssetHolderWithdrawalAuth as with abi.encode() in the smart contracts.
	args := abi.Arguments{
		{Type: abiBytes32},
		{Type: abiAddress},
		{Type: abiAddress},
		{Type: abiUint256},
	}
	enc, err := args.Pack(
		auth.ChannelID,
		auth.Participant,
		auth.Receiver,
		auth.Amount,
	)
	return enc, errors.WithStack(err)
}

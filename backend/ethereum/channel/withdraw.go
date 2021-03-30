// Copyright 2020 - See NOTICE file for copyright holders.
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

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"perun.network/go-perun/backend/ethereum/bindings/assetholder"
	"perun.network/go-perun/backend/ethereum/wallet"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/client"
	"perun.network/go-perun/log"
)

// Withdraw ensures that a channel has been concluded and the final outcome
// withdrawn from the asset holders.
func (a *Adjudicator) Withdraw(ctx context.Context, req channel.AdjudicatorReq, subStates channel.StateMap) error {
	if err := a.ensureConcluded(ctx, req, subStates); err != nil {
		return errors.WithMessage(err, "ensure Concluded")
	}

	return errors.WithMessage(a.ensureWithdrawn(ctx, req), "ensure Withdrawn")
}

func (a *Adjudicator) ensureWithdrawn(ctx context.Context, req channel.AdjudicatorReq) error {
	g, ctx := errgroup.WithContext(ctx)

	// nolint:scopelint
	for index, asset := range req.Tx.Allocation.Assets {
		// Skip zero balance withdrawals
		if req.Tx.Allocation.Balances[index][req.Idx].Sign() == 0 {
			a.log.WithFields(log.Fields{"channel": req.Params.ID, "idx": req.Idx}).Debug("Skipped zero withdrawing.")
			continue
		}
		g.Go(func() error {
			// Create subscription
			watchOpts, err := a.NewWatchOpts(ctx)
			if err != nil {
				return errors.WithMessage(err, "creating watchOpts")
			}
			fundingID := FundingIDs(req.Params.ID(), req.Params.Parts[req.Idx])[0]
			withdrawn := make(chan *assetholder.AssetHolderWithdrawn)
			contract, err := bindAssetHolder(a.ContractBackend, asset, channel.Index(index))
			if err != nil {
				return errors.WithMessage(err, "connecting asset holder")
			}
			sub, err := contract.WatchWithdrawn(watchOpts, withdrawn, [][32]byte{fundingID})
			if err != nil {
				err = checkIsChainNotReachableError(err)
				return errors.WithMessage(err, "WatchWithdrawn failed")
			}
			defer sub.Unsubscribe()

			// Filter past events
			if found, err := a.filterWithdrawn(ctx, fundingID, contract); err != nil {
				return errors.WithMessage(err, "filtering old Withdrawn events")
			} else if found {
				return nil
			}

			// No withdrawn event found in the past, send transaction.
			if err := a.callAssetWithdraw(ctx, req, contract); err != nil {
				return errors.WithMessage(err, "withdrawing assets failed")
			}

			// Wait for event
			select {
			case <-withdrawn:
				return nil
			case <-ctx.Done():
				return errors.Wrap(ctx.Err(), "context cancelled")
			case err = <-sub.Err():
				err = checkIsChainNotReachableError(err)
				return errors.Wrap(err, "subscription error")
			}
		})
	}
	return g.Wait()
}

func bindAssetHolder(backend ContractBackend, asset channel.Asset, assetIndex channel.Index) (assetHolder, error) {
	// Decode and set the asset address.
	assetAddr := common.Address(*asset.(*Asset))
	ctr, err := assetholder.NewAssetHolder(assetAddr, backend)
	if err != nil {
		return assetHolder{}, errors.Wrap(err, "connecting to assetholder")
	}
	return assetHolder{ctr, &assetAddr, assetIndex}, nil
}

// filterWithdrawn returns whether there has been a Withdrawn event in the past.
func (a *Adjudicator) filterWithdrawn(ctx context.Context, fundingID [32]byte, asset assetHolder) (bool, error) {
	filterOpts, err := a.NewFilterOpts(ctx)
	if err != nil {
		return false, err
	}
	iter, err := asset.FilterWithdrawn(filterOpts, [][32]byte{fundingID})
	if err != nil {
		err = checkIsChainNotReachableError(err)
		return false, errors.WithMessage(err, "creating iterator")
	}
	// nolint:errcheck
	defer iter.Close()

	if !iter.Next() {
		err = checkIsChainNotReachableError(iter.Error())
		return false, errors.WithMessage(err, "iterating")
	}
	// Event found
	return true, nil
}

func (a *Adjudicator) callAssetWithdraw(ctx context.Context, request channel.AdjudicatorReq, asset assetHolder) error {
	auth, sig, err := a.newWithdrawalAuth(request, asset)
	if err != nil {
		return errors.WithMessage(err, "creating withdrawal auth")
	}
	tx, err := func() (*types.Transaction, error) {
		if !a.mu.TryLockCtx(ctx) {
			return nil, errors.Wrap(ctx.Err(), "context canceled while acquiring tx lock")
		}
		defer a.mu.Unlock()
		trans, err := a.NewTransactor(ctx, GasLimit, a.txSender)
		if err != nil {
			return nil, errors.WithMessagef(err, "creating transactor for asset %d", asset.assetIndex)
		}

		tx, err := asset.Withdraw(trans, auth, sig)
		if err != nil {
			err = checkIsChainNotReachableError(err)
			return nil, errors.WithMessagef(err, "withdrawing asset %d", asset.assetIndex)
		}
		log.Debugf("Sent transaction %v", tx.Hash().Hex())
		return tx, nil
	}()
	if err != nil {
		return err
	}
	_, err = a.ConfirmTransaction(ctx, tx, a.txSender)
	if err != nil && errors.Is(err, errTxTimedOut) {
		err = client.NewTxTimedoutError(Withdraw.String(), tx.Hash().Hex(), err.Error())
	}
	return errors.WithMessage(err, "mining transaction")
}

func (a *Adjudicator) newWithdrawalAuth(request channel.AdjudicatorReq, asset assetHolder) (assetholder.AssetHolderWithdrawalAuth, []byte, error) {
	auth := assetholder.AssetHolderWithdrawalAuth{
		ChannelID:   request.Params.ID(),
		Participant: wallet.AsEthAddr(request.Acc.Address()),
		Receiver:    a.Receiver,
		Amount:      request.Tx.Allocation.Balances[asset.assetIndex][request.Idx],
	}
	enc, err := encodeAssetHolderWithdrawalAuth(auth)
	if err != nil {
		return assetholder.AssetHolderWithdrawalAuth{}, nil, errors.WithMessage(err, "encoding withdrawal auth")
	}

	sig, err := request.Acc.SignData(enc)
	return auth, sig, errors.WithMessage(err, "sign data")
}

func encodeAssetHolderWithdrawalAuth(auth assetholder.AssetHolderWithdrawalAuth) ([]byte, error) {
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

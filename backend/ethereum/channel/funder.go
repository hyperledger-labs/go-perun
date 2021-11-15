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
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"

	"perun.network/go-perun/backend/ethereum/bindings"
	"perun.network/go-perun/backend/ethereum/bindings/assetholder"
	"perun.network/go-perun/backend/ethereum/subscription"
	"perun.network/go-perun/backend/ethereum/wallet"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/client"
	"perun.network/go-perun/log"
	perunwallet "perun.network/go-perun/wallet"
	pcontext "polycry.pt/poly-go/context"
	perror "polycry.pt/poly-go/errors"
)

type assetHolder struct {
	*assetholder.AssetHolder
	*common.Address
	contract   *bind.BoundContract
	assetIndex channel.Index
}

// Funder implements the channel.Funder interface for Ethereum.
//
// In addition to the `Fund` method required by the `Funder` interface, it also
// provides additional functions for convenience.
//
// All the exported methods are thread-safe and can be invoked concurrently.
type Funder struct {
	mtx sync.RWMutex

	ContractBackend
	// accounts associates an Account to every AssetIndex.
	accounts map[Asset]accounts.Account
	// depositors associates a Depositor to every AssetIndex.
	depositors map[Asset]Depositor
	log        log.Logger // structured logger
}

const funderEventBufSize = 10

// compile time check that we implement the perun funder interface.
var _ channel.Funder = (*Funder)(nil)

// NewFunder creates a new ethereum funder.
func NewFunder(backend ContractBackend) *Funder {
	return &Funder{
		ContractBackend: backend,
		accounts:        make(map[wallet.Address]accounts.Account),
		depositors:      make(map[wallet.Address]Depositor),
		log:             log.Get(),
	}
}

// RegisterAsset registers the depositor and account for the specified asset in
// the funder.
//
// Deposits for this asset will be sent using the depositors from the
// specified account when funding. Hence, it is the responsibility of the
// caller to ensure, the account has sufficient balance in the asset.
//
// It returns true if the asset was successfully registered, false if it was already
// present.
func (f *Funder) RegisterAsset(asset Asset, d Depositor, acc accounts.Account) bool {
	f.mtx.Lock()
	defer f.mtx.Unlock()

	// Both the maps (f.accounts & f.assets) are always modified togethe such
	// that they will have the same set of keys. Hence, it is okay to check one
	// of the two.
	if _, ok := f.accounts[asset]; ok {
		return false
	}
	f.accounts[asset] = acc
	f.depositors[asset] = d
	return true
}

// IsAssetRegistered returns if the specified asset is registered in the funder or not.
// If is registered, then the corresponding depositor and account will also be
// returned.
func (f *Funder) IsAssetRegistered(asset Asset) (Depositor, accounts.Account, bool) {
	f.mtx.RLock()
	defer f.mtx.RUnlock()

	// Both the maps (f.accounts & f.assets) are always modified togethe such
	// that they will have the same set of keys. Hence, it is okay to check one
	// of the two.
	if acc, ok := f.accounts[asset]; ok {
		return f.depositors[asset], acc, true
	}
	return nil, accounts.Account{}, false
}

// Fund implements the channel.Funder interface. It funds all assets in
// parallel. If not all participants successfully fund within a timeframe of
// ChallengeDuration seconds, Fund returns a FundingTimeoutError.
//
// If funding on a real blockchain, make sure that the passed context doesn't
// cancel before the funding period of length ChallengeDuration elapses, or
// funding will be canceled prematurely.
func (f *Funder) Fund(ctx context.Context, request channel.FundingReq) error {
	f.mtx.RLock()
	defer f.mtx.RUnlock()

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

	// Fund each asset, saving the TX in `txs` and the errors in `errg`.
	txs, errg := f.fundAssets(ctx, channelID, request)

	// Wait for the TXs to be mined.
	for a, asset := range request.State.Assets {
		for i, tx := range txs[a] {
			acc := f.accounts[*asset.(*Asset)]
			if _, err := f.ConfirmTransaction(ctx, tx, acc); err != nil {
				if errors.Is(err, errTxTimedOut) {
					err = client.NewTxTimedoutError(Fund.String(), tx.Hash().Hex(), err.Error())
				}
				return errors.WithMessagef(err, "sending %dth funding TX for asset %d", i, a)
			}
			f.log.Debugf("Mined TX: %v", tx.Hash().Hex())
		}
	}

	// Wait for the funding events or timeout.
	var fundingErrs []*channel.AssetFundingError
	nonFundingErrg := perror.NewGatherer()
	for _, err := range perror.Causes(errg.Wait()) {
		if channel.IsAssetFundingError(err) && err != nil {
			fundingErrs = append(fundingErrs, err.(*channel.AssetFundingError))
		} else if err != nil {
			nonFundingErrg.Add(err)
		}
	}
	// Prioritize funding errors over other errors.
	if len(fundingErrs) != 0 {
		return channel.NewFundingTimeoutError(fundingErrs)
	}
	return nonFundingErrg.Err()
}

// fundAssets funds each asset of the funding agreement in the `req`.
// Sends the transactions and returns them. Wait on the returned gatherer
// to ensure that all `funding` events were received.
func (f *Funder) fundAssets(ctx context.Context, channelID channel.ID, req channel.FundingReq) ([]types.Transactions, *perror.Gatherer) {
	txs := make([]types.Transactions, len(req.State.Assets))
	errg := perror.NewGatherer()
	fundingIDs := FundingIDs(channelID, req.Params.Parts...)

	for index, asset := range req.State.Assets {
		// Bind contract.
		contract := bindAssetHolder(f.ContractBackend, asset, channel.Index(index))
		// Wait for the funding event.
		errg.Go(func() error {
			return f.waitForFundingConfirmation(ctx, req, contract, fundingIDs)
		})

		// Send the funding TX.
		tx, err := f.sendFundingTx(ctx, req, contract, fundingIDs[req.Idx])
		if err != nil {
			f.log.WithField("asset", asset).WithError(err).Error("Could not fund asset")
			errg.Add(errors.WithMessage(err, "funding asset"))
			continue
		}
		txs[index] = tx
	}
	return txs, errg
}

// sendFundingTx sends and returns the TXs that are needed to fulfill the
// funding request. It is idempotent.
func (f *Funder) sendFundingTx(ctx context.Context, request channel.FundingReq, contract assetHolder, fundingID [32]byte) (txs []*types.Transaction, fatal error) {
	bal := request.Agreement[contract.assetIndex][request.Idx]
	// nolint: gocritic
	if bal == nil || bal.Sign() <= 0 {
		f.log.WithFields(log.Fields{"channel": request.Params.ID(), "idx": request.Idx}).Debug("Skipped zero funding.")
	} else if alreadyFunded, err := f.checkFunded(ctx, bal, contract, fundingID); err != nil {
		return nil, errors.WithMessage(err, "checking funded")
	} else if alreadyFunded {
		f.log.WithFields(log.Fields{"channel": request.Params.ID(), "idx": request.Idx}).Debug("Skipped second funding.")
	} else {
		return f.deposit(ctx, bal, wallet.Address(*contract.Address), fundingID)
	}
	return nil, nil
}

// deposit deposits funds for one funding-ID by calling the associated Depositor.
// Returns an error if no matching Depositor or Account could be found.
func (f *Funder) deposit(ctx context.Context, bal *big.Int, asset Asset, fundingID [32]byte) (types.Transactions, error) {
	depositor, ok := f.depositors[asset]
	if !ok {
		return nil, errors.Errorf("could not find Depositor for asset #%d", asset)
	}
	acc, ok := f.accounts[asset]
	if !ok {
		return nil, errors.Errorf("could not find account for asset #%d", asset)
	}

	return depositor.Deposit(ctx, *NewDepositReq(bal, f.ContractBackend, asset, acc, fundingID))
}

// checkFunded returns whether `fundingID` holds at least `amount` funds.
func (f *Funder) checkFunded(ctx context.Context, amount *big.Int, asset assetHolder, fundingID [32]byte) (bool, error) {
	deposited := make(chan *subscription.Event, funderEventBufSize)
	subErr := make(chan error, 1)
	// Subscribe to events.
	sub, err := f.depositedSub(ctx, asset.contract, fundingID)
	if err != nil {
		return false, errors.WithMessage(err, "subscribing to deposited event")
	}
	defer sub.Close()
	// Read from the sub.
	go func() {
		defer close(deposited)
		subErr <- sub.ReadPast(ctx, deposited)
	}()

	left := new(big.Int).Set(amount)
	for _event := range deposited {
		event := _event.Data.(*assetholder.AssetHolderDeposited)
		left.Sub(left, event.Amount)
	}
	return left.Sign() != 1, errors.WithMessagef(<-subErr, "filtering old Funding events for asset %d", asset.assetIndex)
}

func (f *Funder) depositedSub(ctx context.Context, contract *bind.BoundContract, fundingIDs ...[32]byte) (*subscription.ResistantEventSub, error) {
	filter := make([]interface{}, len(fundingIDs))
	for i, fundingID := range fundingIDs {
		filter[i] = fundingID
	}
	event := func() *subscription.Event {
		return &subscription.Event{
			Name:   bindings.Events.AhDeposited,
			Data:   new(assetholder.AssetHolderDeposited),
			Filter: [][]interface{}{filter},
		}
	}
	sub, err := subscription.Subscribe(ctx, f, contract, event, startBlockOffset, f.txFinalityDepth)
	return sub, errors.WithMessage(err, "subscribing to deposited event")
}

// waitForFundingConfirmation waits for the confirmation events on the blockchain that
// both we and all peers successfully funded the channel for the specified asset
// according to the funding agreement.
func (f *Funder) waitForFundingConfirmation(ctx context.Context, request channel.FundingReq, asset assetHolder, fundingIDs [][32]byte) error {
	deposited := make(chan *subscription.Event)
	subErr := make(chan error, 1)
	// Subscribe to events.
	sub, err := f.depositedSub(ctx, asset.contract, fundingIDs...)
	if err != nil {
		return errors.WithMessage(err, "subscribing to deposited event")
	}
	defer sub.Close()
	// Read from the sub.
	go func() {
		subErr <- sub.Read(ctx, deposited)
	}()

	// The allocation that all participants agreed on.
	agreement := request.Agreement.Clone()[asset.assetIndex]
	// Count how many zero balance funding requests are there
	N := len(request.Params.Parts) - countZeroBalances(agreement)

	// Wait for all non-zero funding requests
	for N > 0 {
		select {
		case rawEvent := <-deposited:
			event := rawEvent.Data.(*assetholder.AssetHolderDeposited)
			log := f.log.WithField("fundingID", event.FundingID)

			// Calculate the position in the participant array.
			idx := getPartIdx(event.FundingID, fundingIDs)

			amount := agreement[idx]
			if amount.Sign() == 0 {
				continue // ignore double events
			}

			amount.Sub(amount, event.Amount)
			if amount.Sign() != 1 {
				// participant funded successfully
				N--
				agreement[idx].SetUint64(0)
			}
			log.Debugf("peer[%d]: got: %v, remaining for [%d,%d] = %v. N: %d", request.Idx, event.Amount, asset.assetIndex, idx, amount, N)

		case <-ctx.Done():
			var indices []channel.Index
			for k, bals := range agreement {
				if bals.Sign() == 1 {
					indices = append(indices, channel.Index(k))
				}
			}
			if len(indices) != 0 {
				return &channel.AssetFundingError{Asset: asset.assetIndex, TimedOutPeers: indices}
			}
			return nil
		case err := <-subErr:
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

func countZeroBalances(bals []channel.Bal) (n int) {
	for _, part := range bals {
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

// NumTX returns how many Transactions are needed for the funding request.
func (f *Funder) NumTX(req channel.FundingReq) (sum uint32, err error) {
	f.mtx.RLock()
	defer f.mtx.RUnlock()

	for _, a := range req.State.Assets {
		depositor, ok := f.depositors[*a.(*Asset)]
		if !ok {
			return 0, errors.Errorf("could not find Depositor for asset #%d", a)
		}
		sum += depositor.NumTX()
	}
	return
}

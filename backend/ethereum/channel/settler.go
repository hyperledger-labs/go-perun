// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package channel

import (
	"context"
	stderrors "errors"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"

	"perun.network/go-perun/backend/ethereum/bindings/adjudicator"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/log"
	perunwallet "perun.network/go-perun/wallet"
)

// Settler implements the channel.Settler interface for Ethereum.
type Settler struct {
	ContractBackend
	// Address of the adjudicator contract.
	adjAddr     common.Address
	adjInstance *adjudicator.Adjudicator
	// Mutex prevents race on transaction nonces.
	mu sync.Mutex
}

// compile time check that we implement the perun settler interface
var _ channel.Settler = (*Settler)(nil)

// Error that is returned if an event was not found in the past.
var errConcludedNotFound = stderrors.New("concluded event not found")

// NewETHSettler creates a new ethereum funder.
func NewETHSettler(backend ContractBackend, adjAddr common.Address) *Settler {
	return &Settler{
		ContractBackend: backend,
		adjAddr:         adjAddr,
	}
}

// Settle calls pushOutcome on the Adjudicator to set the channel outcomes in the asset holders.
// Withdrawal from the asset holders is not implemented yet.
// The parameter acc is currently ignored, as it is only used to sign withdrawal authorizations.
func (s *Settler) Settle(ctx context.Context, req channel.SettleReq, acc perunwallet.Account) error {
	if req.Params == nil || req.Tx.State == nil {
		panic("invalid settlement request")
	}
	if err := s.checkAdjInstance(); err != nil {
		return errors.WithMessage(err, "connecting to adjudicator")
	}
	if req.Tx.State.IsFinal {
		return s.cooperativeSettle(ctx, req)
	}
	return s.uncooperativeSettle(ctx, req)
}

func (s *Settler) cooperativeSettle(ctx context.Context, req channel.SettleReq) error {
	// Listen for blockchain events.
	watchOpts, err := s.newWatchOpts(ctx)
	if err != nil {
		return errors.WithMessage(err, "creating watchOpts")
	}
	concluded := make(chan *adjudicator.AdjudicatorFinalConcluded)
	sub, err := s.adjInstance.WatchFinalConcluded(watchOpts, concluded, [][32]byte{req.Params.ID()})
	if err != nil {
		return errors.Wrap(err, "WatchFinalConcluded failed")
	}
	defer sub.Unsubscribe()

	if err := s.filterOldConfirmations(ctx, req.Params.ID()); err != errConcludedNotFound {
		// err might be nil, which is fine
		return errors.WithMessage(err, "filtering old Concluded events")
	}
	// No conclude event found in the past, send transaction.
	tx, err := s.sendConcludeFinalTx(ctx, req)
	if err != nil {
		return err
	}
	if err := execSuccessful(ctx, s.ContractBackend, tx); err != nil {
		log.Warnf("transaction failed: %v", err)
	} else {
		log.Debug("Transaction mined successfully")
	}

	select {
	case <-concluded:
		return nil
	case <-ctx.Done():
		return errors.Wrap(ctx.Err(), "Waiting for final concluded event cancelled by context")
	case err = <-sub.Err():
		return errors.Wrap(err, "Error while waiting for events")
	}
}

func (s *Settler) uncooperativeSettle(ctx context.Context, req channel.SettleReq) error {
	panic("Settling with non-final state currently not implemented")
}

func (s *Settler) sendConcludeFinalTx(ctx context.Context, req channel.SettleReq) (*types.Transaction, error) {
	ethParams := channelParamsToEthParams(req.Params)
	ethState := channelStateToEthState(req.Tx.State)
	s.mu.Lock()
	defer s.mu.Unlock()
	trans, err := s.newTransactor(ctx, big.NewInt(0), GasLimit)
	if err != nil {
		return nil, errors.WithMessage(err, "creating transactor")
	}
	tx, err := s.adjInstance.ConcludeFinal(trans, ethParams, ethState, req.Tx.Sigs)
	if err != nil {
		return nil, errors.Wrap(err, "calling concludeFinal")
	}
	log.Debugf("Sending transaction to the blockchain with txHash: %v", tx.Hash().Hex())
	return tx, nil
}

func (s *Settler) filterOldConfirmations(ctx context.Context, channelID channel.ID) error {
	// Filter
	filterOpts := bind.FilterOpts{
		Start:   uint64(1),
		End:     nil,
		Context: ctx}
	iter, err := s.adjInstance.FilterFinalConcluded(&filterOpts, [][32]byte{channelID})
	if err != nil {
		return errors.WithStack(err)
	}
	if !iter.Next() {
		return errConcludedNotFound
	}
	// Event found, return nil
	return nil
}

// checkAdjInstance checks if the adjudicator instance is set.
// if not it connects to the adjudicator at s.adjAddr.
func (s *Settler) checkAdjInstance() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.adjInstance == nil {
		adjInstance, err := adjudicator.NewAdjudicator(s.adjAddr, s)
		if err != nil {
			return errors.Wrap(err, "failed to connect to adjudicator")
		}
		s.adjInstance = adjInstance
	}
	return nil
}

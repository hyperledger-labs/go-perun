// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package channel // import "perun.network/go-perun/backend/ethereum/channel"

import (
	"context"
	"math/big"
	"time"

	"github.com/pkg/errors"
	"perun.network/go-perun/backend/ethereum/bindings/adjudicator"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/log"
)

// Register registers a state on-chain.
// If the state is a final state, register becomes a no-op.
func (a *Adjudicator) Register(ctx context.Context, request channel.AdjudicatorReq) (*channel.Registered, error) {
	stored := make(chan *adjudicator.AdjudicatorStored)
	sub, iter, err := a.waitForStoredEvent(ctx, stored, request.Params)
	if err != nil {
		return nil, errors.WithMessage(err, "waiting for stored event")
	}
	defer sub.Unsubscribe()
	go func() {
		var ev *adjudicator.AdjudicatorStored
		for iter.Next() {
			ev = iter.Event
		}
		if ev != nil {
			stored <- ev
		}
		iter.Close()
	}()
	for i := 0; i < maxRegisteredEvents; i++ {
		select {
		case ev := <-stored:
			if request.Tx.Version > ev.Version {
				_ = a.refute(ctx, request)
			}
			return storedToRegisteredEvent(ev), nil
		case <-ctx.Done():
			return nil, errors.New("did not receive stored event in time")
		default:
		}
		if request.Tx.State.IsFinal {
			// If a request is final and we have no event seen (we don't need to dispute a previous state)
			// Register becomes a no-op.
			return &channel.Registered{
				ID:      request.Params.ID(),
				Idx:     request.Idx,
				Version: request.Tx.Version,
				Timeout: time.Time{},
			}, nil
		}
		if err = a.register(ctx, request); err != nil {
			a.log.Warnf("Registering failed, trying again")
		}
		// After a register, we wait if we receive an event.
	}
	return nil, errors.Errorf("%d events seen, none were our state", maxRegisteredEvents)
}

func (a *Adjudicator) register(ctx context.Context, req channel.AdjudicatorReq) error {
	ethParams := channelParamsToEthParams(req.Params)
	ethState := channelStateToEthState(req.Tx.State)
	if !a.mu.TryLockCtx(ctx) {
		return errors.New("Could not acquire lock in time")
	}
	defer a.mu.Unlock()
	trans, err := a.newTransactor(ctx, big.NewInt(0), GasLimit)
	if err != nil {
		return errors.WithMessage(err, "creating transactor")
	}
	tx, err := a.contract.Register(trans, ethParams, ethState, req.Tx.Sigs)
	if err != nil {
		return errors.Wrap(err, "calling concludeFinal")
	}
	log.Debugf("Sending transaction to the blockchain with txHash: %v", tx.Hash().Hex())
	return confirmTransaction(ctx, a.ContractBackend, tx)
}

func (a *Adjudicator) refute(ctx context.Context, req channel.AdjudicatorReq) error {
	ethParams := channelParamsToEthParams(req.Params)
	ethState := channelStateToEthState(req.Tx.State)
	if !a.mu.TryLockCtx(ctx) {
		return errors.New("Could not acquire lock in time")
	}
	defer a.mu.Unlock()
	trans, err := a.newTransactor(ctx, big.NewInt(0), GasLimit)
	if err != nil {
		return errors.WithMessage(err, "creating transactor")
	}
	tx, err := a.contract.Refute(trans, ethParams, ethState, req.Tx.Sigs)
	if err != nil {
		return errors.Wrap(err, "calling concludeFinal")
	}
	log.Debugf("Sending transaction to the blockchain with txHash: %v", tx.Hash().Hex())
	return confirmTransaction(ctx, a.ContractBackend, tx)
}

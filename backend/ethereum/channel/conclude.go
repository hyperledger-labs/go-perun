// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package channel

import (
	"context"

	"github.com/pkg/errors"
	"perun.network/go-perun/backend/ethereum/bindings/adjudicator"

	"perun.network/go-perun/channel"
)

// ensureConcluded ensures that conclude or concludeFinal (for non-final and
// final states, resp.) is called on the adjudicator.
// - a subscription on Concluded events is established
// - it searches for a past concluded event
//   - if found, channel is already concluded and success is returned
//   - if none found, conclude/concludeFinal is called on the adjudicator
// - it waits for a Concluded event from the blockchain
func (a *Adjudicator) ensureConcluded(ctx context.Context, req channel.AdjudicatorReq) error {
	// Listen for Concluded event.
	watchOpts, err := a.NewWatchOpts(ctx)
	if err != nil {
		return errors.WithMessage(err, "creating watchOpts")
	}
	concluded := make(chan *adjudicator.AdjudicatorConcluded)
	sub, err := a.contract.WatchConcluded(watchOpts, concluded, [][32]byte{req.Params.ID()})
	if err != nil {
		return errors.Wrap(err, "WatchConcluded failed")
	}
	defer sub.Unsubscribe()

	if found, err := a.filterConcluded(ctx, req.Params.ID()); err != nil {
		return errors.WithMessage(err, "filtering old Concluded events")
	} else if found {
		return nil
	}

	// No conclude event found in the past, send transaction.
	if req.Tx.IsFinal {
		err = errors.WithMessage(a.callConcludeFinal(ctx, req), "calling concludeFinal")
	} else {
		err = errors.WithMessage(a.callConclude(ctx, req), "calling conclude")
	}
	if IsTxFailedError(err) {
		a.log.Warn("Calling conclude(Final) failed, waiting for event anyways...")
	} else if err != nil {
		return err
	}

	select {
	case <-concluded:
		return nil
	case <-ctx.Done():
		return errors.Wrap(ctx.Err(), "context cancelled")
	case err = <-sub.Err():
		return errors.Wrap(err, "subscription error")
	}
}

// filterConcluded returns whether there has been a Concluded event in the past.
func (a *Adjudicator) filterConcluded(ctx context.Context, channelID channel.ID) (bool, error) {
	filterOpts, err := a.NewFilterOpts(ctx)
	if err != nil {
		return false, err
	}
	iter, err := a.contract.FilterConcluded(filterOpts, [][32]byte{channelID})
	if err != nil {
		return false, errors.Wrap(err, "creating iterator")
	}

	if !iter.Next() {
		return false, errors.Wrap(iter.Error(), "iterating")
	}
	// Event found
	return true, nil
}

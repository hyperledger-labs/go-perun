// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package channel

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"perun.network/go-perun/wallet"
)

type (
	// A Settler is used to settle a closed channel.
	// In the case of ledger channel, the implementation is backend-specific.
	Settler interface {
		// Settle should settle the channel passed by FundingReq on the blockchain.
		// It should return an error if the channel could not be settled with the given state.
		// Settle should return an AlreadySettledError if the channel was already settled.
		Settle(context.Context, SettleReq, wallet.Account) error
	}

	// SettleReq is a request to settle a channel.
	SettleReq struct {
		Params *Params
		Idx    Index
		Tx     Transaction
	}

	// An AlreadySettledError is returned whenever we try to settle a channel that was already settled.
	AlreadySettledError struct {
		PeerIdx Index  // index of the peer who settled the channel.
		Version uint64 // version of the state with which the channel was settled.
	}
)

func (e AlreadySettledError) Error() string {
	return fmt.Sprintf("peer[%d] already settled the channel with version %d", e.PeerIdx, e.Version)
}

// NewAlreadySettledError creates a new AlreadySettledError.
func NewAlreadySettledError(idx Index, version uint64) error {
	return errors.WithStack(&AlreadySettledError{idx, version})
}

// IsAlreadySettledError checks whether an error is an AlreadySettledError.
func IsAlreadySettledError(err error) bool {
	_, ok := errors.Cause(err).(*AlreadySettledError)
	return ok
}

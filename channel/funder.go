// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package channel

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
)

type (
	// The Funder interface needs to be implemented by every
	// blockchain backend. It provides functionality to fund a new channel on-chain.
	Funder interface {
		// Fund should fund the channel in FundingReq on the blockchain.
		// It should return an error if own funding did not succeed, possibly
		// because the peer did not fund the channel in time.
		// Depending on the funding protocol, if we fund first and then the peer does
		// not fund in time, a dispute process needs to be initiated to get back the
		// funds from the partially funded channel. In this case, the user should
		// return a PeerTimedOutFundingError containing the index of the peer who
		// did not fund in time. The framework will then initiate the dispute
		// process.
		Fund(context.Context, FundingReq) error
	}

	// A FundingReq bundles all data needed to fund a channel.
	FundingReq struct {
		Params     *Params
		Allocation *Allocation
		Idx        Index // our index
	}

	// A PeerTimedOutFundingError is a special error that is returned whenever
	// a participant does not fund the channel in time.
	PeerTimedOutFundingError struct {
		TimedOutPeerIdx Index // index of the peer who timed-out funding
	}
)

func (e PeerTimedOutFundingError) Error() string {
	return fmt.Sprintf("peer[%d] did not fund channel in time", e.TimedOutPeerIdx)
}

// NewPeerTimedOutFundingError creates a new PeerTimedOutFundingError.
func NewPeerTimedOutFundingError(idx Index) error {
	return errors.WithStack(&PeerTimedOutFundingError{idx})
}

// IsPeerTimedOutFundingError checks whether an error is a PeerTimedOutFundingError.
func IsPeerTimedOutFundingError(err error) bool {
	_, ok := errors.Cause(err).(*PeerTimedOutFundingError)
	return ok
}

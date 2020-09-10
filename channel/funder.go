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
		Params *Params
		State  *State
		Idx    Index // our index
	}

	// A FundingTimeoutError indicates that some peers failed funding some assets in time.
	FundingTimeoutError struct {
		Errors []*AssetFundingError
	}

	// An AssetFundingError indicates the peers who timed-out funding a specific asset.
	AssetFundingError struct {
		Asset         Index   // The asset for which the timeouts occurred
		TimedOutPeers []Index // Indices of the peers who failed to fund in time
	}
)

// NewFundingTimeoutError creates a new FundingTimeoutError.
func NewFundingTimeoutError(fundingErrs []*AssetFundingError) error {
	if len(fundingErrs) == 0 {
		return nil
	}
	return errors.WithStack(&FundingTimeoutError{fundingErrs})
}

func (e FundingTimeoutError) Error() string {
	msg := ""
	for _, assetErr := range e.Errors {
		msg += assetErr.Error() + "; "
	}
	return msg
}

// IsFundingTimeoutError checks whether an error is a FundingTimeoutError.
func IsFundingTimeoutError(err error) bool {
	_, ok := errors.Cause(err).(*FundingTimeoutError)
	return ok
}

func (e AssetFundingError) Error() string {
	msg := fmt.Sprintf("Funding Error on asset [%d] peers: ", e.Asset)
	for _, peerIdx := range e.TimedOutPeers {
		msg += fmt.Sprintf("[%d], ", peerIdx)
	}
	msg += "did not fund channel in time"
	return msg
}

// IsAssetFundingError checks whether an error is a AssetFundingError.
func IsAssetFundingError(err error) bool {
	_, ok := errors.Cause(err).(*AssetFundingError)
	return ok
}

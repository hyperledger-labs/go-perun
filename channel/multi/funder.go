// Copyright 2025 - See NOTICE file for copyright holders.
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

package multi

import (
	"context"
	"fmt"
	"math"
	"time"

	"perun.network/go-perun/channel"
)

// LedgerBackendKey is a representation of LedgerBackendID that kan be used in map lookups.
type LedgerBackendKey struct {
	BackendID uint32
	LedgerID  string
}

// Funder is a multi-ledger funder.
// funders is a map of LedgerIDs corresponding to a funder on some chain.
// egoistic controls whether the funder uses the egoisticIndex to control the funding order.
// egoisticIndex controls which participant index will fund last.
type Funder struct {
	funders       map[LedgerBackendKey]channel.Funder
	egoistic      bool
	egoisticIndex int
}

// NewFunder creates a new funder.
func NewFunder() *Funder {
	return &Funder{
		funders:  make(map[LedgerBackendKey]channel.Funder),
		egoistic: false,
	}
}

// RegisterFunder registers a funder for a given ledger.
func (f *Funder) RegisterFunder(l LedgerBackendID, lf channel.Funder) {
	key := LedgerBackendKey{BackendID: l.BackendID(), LedgerID: string(l.LedgerID().MapKey())}
	f.funders[key] = lf
}

// SetEgoisticPart sets the egoistic chain flag for a given ledger.
func (f *Funder) SetEgoisticPart(index int) {
	f.egoisticIndex = index
	f.egoistic = true
}

// Fund funds a multi-ledger channel. It dispatches funding calls to all
// relevant registered funders. It waits until all participants have funded the
// channel. If any of the funder calls fails, the method returns an error.
func (f *Funder) Fund(ctx context.Context, request channel.FundingReq) error {
	// Define funding timeout.
	duration := request.Params.ChallengeDuration
	if duration > math.MaxInt64 {
		return fmt.Errorf("challenge duration %d is too large", duration)
	}
	d := time.Duration(duration) * time.Second
	ctx, cancel := context.WithTimeout(ctx, d)
	defer cancel()

	ledgerIDs, err := assets(request.State.Assets).LedgerIDs()
	if err != nil {
		return err
	}

	var egoisticLedgers []LedgerBackendID
	var nonEgoisticLedgers []LedgerBackendID

	for i, l := range ledgerIDs {
		if f.egoistic && f.egoisticIndex == i {
			egoisticLedgers = append(egoisticLedgers, l)
		} else {
			nonEgoisticLedgers = append(nonEgoisticLedgers, l)
		}
	}

	// First fund with Funders that are not egoistic.
	err = fundLedgers(ctx, request, nonEgoisticLedgers, f.funders)
	if err != nil {
		return err
	}

	// Then fund with egoistic Funders.
	err = fundLedgers(ctx, request, egoisticLedgers, f.funders)
	if err != nil {
		return err
	}

	return nil
}

func fundLedgers(ctx context.Context, request channel.FundingReq, assetIDs []LedgerBackendID, funders map[LedgerBackendKey]channel.Funder) error {
	// Calculate the total number of funders
	n := len(assetIDs)

	errs := make(chan error, n)

	// Iterate over blockchains to get the LedgerIDs
	for _, assetID := range assetIDs {
		go func(assetID LedgerBackendID) {
			key := LedgerBackendKey{BackendID: assetID.BackendID(), LedgerID: string(assetID.LedgerID().MapKey())}
			// Get the Funder from the funders map
			funder, ok := funders[key]
			if !ok {
				errs <- fmt.Errorf("funder map not found for blockchain %d and ledger %d", assetID.BackendID(), assetID.LedgerID())
				return
			}

			// Call the Fund method
			err := funder.Fund(ctx, request)
			errs <- err
		}(assetID)
	}

	// Collect errors
	for range n {
		err := <-errs
		if err != nil {
			return err
		}
	}
	return nil
}

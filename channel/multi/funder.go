// Copyright 2022 - See NOTICE file for copyright holders.
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
	"time"

	"perun.network/go-perun/channel"
)

// Funder is a multi-ledger funder.
type Funder struct {
	funders map[LedgerIDMapKey]channel.Funder
}

// NewFunder creates a new funder.
func NewFunder() *Funder {
	return &Funder{
		funders: make(map[LedgerIDMapKey]channel.Funder),
	}
}

// RegisterFunder registers a funder for a given ledger.
func (f *Funder) RegisterFunder(l LedgerID, lf channel.Funder) {
	f.funders[l.MapKey()] = lf
}

// Fund funds a multi-ledger channel. It dispatches funding calls to all
// relevant registered funders. It waits until all participants have funded the
// channel. If any of the funder calls fails, the method returns an error.
func (f *Funder) Fund(ctx context.Context, request channel.FundingReq) error {
	// Define funding timeout.
	d := time.Duration(request.Params.ChallengeDuration) * time.Second
	ctx, cancel := context.WithTimeout(ctx, d)
	defer cancel()

	ledgers, err := assets(request.State.Assets).LedgerIDs()
	if err != nil {
		return err
	}

	n := len(ledgers)
	errs := make(chan error, n)
	for _, l := range ledgers {
		go func(l LedgerID) {
			errs <- func() error {
				id := l.MapKey()
				lf, ok := f.funders[id]
				if !ok {
					return fmt.Errorf("Funder not found for ledger %v", id)
				}

				err := lf.Fund(ctx, request)
				return err
			}()
		}(l)
	}

	for i := 0; i < n; i++ {
		err := <-errs
		if err != nil {
			return err
		}
	}

	return nil
}

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
	"perun.network/go-perun/channel"
)

// Adjudicator is a multi-ledger adjudicator.
type Adjudicator struct {
	adjudicators map[AssetIDKey]channel.Adjudicator
}

// NewAdjudicator creates a new adjudicator.
func NewAdjudicator() *Adjudicator {
	return &Adjudicator{
		adjudicators: make(map[AssetIDKey]channel.Adjudicator),
	}
}

// RegisterAdjudicator registers an adjudicator for a given ledger.
func (a *Adjudicator) RegisterAdjudicator(l AssetID, la channel.Adjudicator) {
	key := AssetIDKey{BackendID: l.BackendID(), LedgerID: string(l.LedgerId().MapKey())}
	a.adjudicators[key] = la
}

// LedgerAdjudicator returns the adjudicator for a given ledger.
func (a *Adjudicator) LedgerAdjudicator(l AssetID) (channel.Adjudicator, bool) {
	key := AssetIDKey{BackendID: l.BackendID(), LedgerID: string(l.LedgerId().MapKey())}
	adj, ok := a.adjudicators[key]
	return adj, ok
}

// Register registers a multi-ledger channel. It dispatches Register calls to
// all relevant adjudicators. If any of the calls fails, the method returns an
// error.
func (a *Adjudicator) Register(ctx context.Context, req channel.AdjudicatorReq, subStates []channel.SignedState) error {
	assetIds, err := assets(req.Tx.Assets).LedgerIDs()
	if err != nil {
		return err
	}

	err = a.dispatch(assetIds, func(la channel.Adjudicator) error {
		return la.Register(ctx, req, subStates)
	})
	return err
}

// Progress progresses the state of a multi-ledger channel. It dispatches
// Progress calls to all relevant adjudicators. If any of the calls fails, the
// method returns an error.
func (a *Adjudicator) Progress(ctx context.Context, req channel.ProgressReq) error {
	assetIds, err := assets(req.Tx.Assets).LedgerIDs()
	if err != nil {
		return err
	}

	err = a.dispatch(assetIds, func(la channel.Adjudicator) error {
		return la.Progress(ctx, req)
	})
	return err
}

// Withdraw withdraws the funds from a multi-ledger channel. It dispatches
// Withdraw calls to all relevant adjudicators. If any of the calls fails, the
// method returns an error.
func (a *Adjudicator) Withdraw(ctx context.Context, req channel.AdjudicatorReq, subStates channel.StateMap) error {
	assetIds, err := assets(req.Tx.Assets).LedgerIDs()
	if err != nil {
		return err
	}

	err = a.dispatch(assetIds, func(la channel.Adjudicator) error {
		return la.Withdraw(ctx, req, subStates)
	})
	return err
}

// dispatch dispatches an adjudicator call on all given ledgers.
func (a *Adjudicator) dispatch(assetIds []AssetID, f func(channel.Adjudicator) error) error {
	n := len(assetIds)
	errs := make(chan error, n)
	for _, l := range assetIds {
		go func(l AssetID) {
			err := func() error {
				key := AssetIDKey{BackendID: l.BackendID(), LedgerID: string(l.LedgerId().MapKey())}
				adjs, ok := a.adjudicators[key]
				if !ok {
					return fmt.Errorf("adjudicator not found for id %v", l)
				}

				// Call the provided function f with the Adjudicator
				err := f(adjs)
				return err
			}()
			errs <- err
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

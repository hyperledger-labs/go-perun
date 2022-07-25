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
	adjudicators map[LedgerIDMapKey]channel.Adjudicator
}

// NewAdjudicator creates a new adjudicator.
func NewAdjudicator() *Adjudicator {
	return &Adjudicator{
		adjudicators: make(map[LedgerIDMapKey]channel.Adjudicator),
	}
}

// RegisterAdjudicator registers an adjudicator for a given ledger.
func (a *Adjudicator) RegisterAdjudicator(l LedgerID, la channel.Adjudicator) {
	a.adjudicators[l.MapKey()] = la
}

// LedgerAdjudicator returns the adjudicator for a given ledger.
func (a *Adjudicator) LedgerAdjudicator(l LedgerID) (channel.Adjudicator, bool) {
	adj, ok := a.adjudicators[l.MapKey()]
	return adj, ok
}

// Register registers a multi-ledger channel. It dispatches Register calls to
// all relevant adjudicators. If any of the calls fails, the method returns an
// error.
func (a *Adjudicator) Register(ctx context.Context, req channel.AdjudicatorReq, subStates []channel.SignedState) error {
	ledgers, err := assets(req.Tx.Assets).LedgerIDs()
	if err != nil {
		return err
	}

	err = a.dispatch(ledgers, func(la channel.Adjudicator) error {
		return la.Register(ctx, req, subStates)
	})
	return err
}

// Progress progresses the state of a multi-ledger channel. It dispatches
// Progress calls to all relevant adjudicators. If any of the calls fails, the
// method returns an error.
func (a *Adjudicator) Progress(ctx context.Context, req channel.ProgressReq) error {
	ledgers, err := assets(req.Tx.Assets).LedgerIDs()
	if err != nil {
		return err
	}

	err = a.dispatch(ledgers, func(la channel.Adjudicator) error {
		return la.Progress(ctx, req)
	})
	return err
}

// Withdraw withdraws the funds from a multi-ledger channel. It dispatches
// Withdraw calls to all relevant adjudicators. If any of the calls fails, the
// method returns an error.
func (a *Adjudicator) Withdraw(ctx context.Context, req channel.AdjudicatorReq, subStates channel.StateMap) error {
	ledgers, err := assets(req.Tx.Assets).LedgerIDs()
	if err != nil {
		return err
	}

	err = a.dispatch(ledgers, func(la channel.Adjudicator) error {
		return la.Withdraw(ctx, req, subStates)
	})
	return err
}

// dispatch dispatches an adjudicator call on all given ledgers.
func (a *Adjudicator) dispatch(ledgers []LedgerID, f func(channel.Adjudicator) error) error {
	n := len(ledgers)
	errs := make(chan error, n)
	for _, l := range ledgers {
		go func(l LedgerID) {
			err := func() error {
				id := l.MapKey()
				la, ok := a.adjudicators[id]
				if !ok {
					return fmt.Errorf("Adjudicator not found for ledger %v", id)
				}

				err := f(la)
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

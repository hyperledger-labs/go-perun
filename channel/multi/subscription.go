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
	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"
)

// Subscribe creates a new multi-ledger AdjudicatorSubscription.
func (a *Adjudicator) Subscribe(ctx context.Context, chID map[wallet.BackendID]channel.ID) (channel.AdjudicatorSubscription, error) {
	asub := &AdjudicatorSubscription{
		events: make(chan channel.AdjudicatorEvent),
		errors: make(chan error),
		subs:   []channel.AdjudicatorSubscription{},
		done:   make(chan struct{}),
	}

	for _, la := range a.adjudicators {
		sub, err := la.Subscribe(ctx, chID)
		if err != nil {
			asub.Close()
			return nil, err
		}
		asub.subs = append(asub.subs, sub)

		go func() {
			for {
				select {
				case asub.events <- sub.Next():
				case <-asub.done:
					return
				}
			}
		}()

		go func() {
			asub.errors <- sub.Err()
		}()
	}

	return asub, nil
}

// AdjudicatorSubscription is a multi-ledger adjudicator subscription.
type AdjudicatorSubscription struct {
	subs   []channel.AdjudicatorSubscription
	events chan channel.AdjudicatorEvent
	errors chan error
	done   chan struct{}
}

// Next returns the next event.
func (s *AdjudicatorSubscription) Next() channel.AdjudicatorEvent {
	select {
	case e := <-s.events:
		return e
	case <-s.done:
		return nil
	}
}

// Err blocks until an error occurred and returns it.
func (s *AdjudicatorSubscription) Err() error {
	for i := 0; i < len(s.subs); i++ {
		err := <-s.errors
		if err != nil {
			return err
		}
	}
	return nil
}

// Close closes the subscription.
func (s *AdjudicatorSubscription) Close() error {
	for _, sub := range s.subs {
		sub.Close()
	}
	close(s.done)
	return nil
}

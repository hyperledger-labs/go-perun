// Copyright 2021 - See NOTICE file for copyright holders.
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

package test

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/log"
)

type (
	// MockBackend is a mocked backend useful for testing.
	MockBackend struct {
		log          log.Logger
		rng          rng
		mu           sync.Mutex
		latestEvents map[channel.ID]channel.AdjudicatorEvent
		eventSubs    map[channel.ID][]chan channel.AdjudicatorEvent
	}

	rng interface {
		Intn(n int) int
	}

	threadSafeRng struct {
		mu sync.Mutex
		r  *rand.Rand
	}
)

// NewMockBackend creates a new backend object.
func NewMockBackend(rng *rand.Rand) *MockBackend {
	return &MockBackend{
		log:          log.Get(),
		rng:          newThreadSafePrng(rng),
		latestEvents: make(map[channel.ID]channel.AdjudicatorEvent),
		eventSubs:    make(map[channel.ID][]chan channel.AdjudicatorEvent),
	}
}

func newThreadSafePrng(r *rand.Rand) *threadSafeRng {
	return &threadSafeRng{
		mu: sync.Mutex{},
		r:  r,
	}
}

func (g *threadSafeRng) Intn(n int) int {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.r.Intn(n)
}

// Fund funds the channel.
func (b *MockBackend) Fund(_ context.Context, req channel.FundingReq) error {
	time.Sleep(time.Duration(b.rng.Intn(100)) * time.Millisecond)
	b.log.Infof("Funding: %+v", req)
	return nil
}

// Register registers the channel.
func (b *MockBackend) Register(_ context.Context, req channel.AdjudicatorReq, subChannels []channel.SignedState) error {
	b.log.Infof("Register: %+v", req)

	b.mu.Lock()
	defer b.mu.Unlock()

	channels := append([]channel.SignedState{
		{
			Params: req.Params,
			State:  req.Tx.State,
			Sigs:   req.Tx.Sigs,
		},
	}, subChannels...)

	for _, ch := range channels {
		b.setLatestEvent(
			ch.Params.ID(),
			channel.NewRegisteredEvent(
				ch.Params.ID(),
				&channel.ElapsedTimeout{},
				ch.State.Version,
				ch.State,
				ch.Sigs,
			),
		)
	}
	return nil
}

func (b *MockBackend) setLatestEvent(ch channel.ID, e channel.AdjudicatorEvent) {
	b.latestEvents[ch] = e
	// Update subscriptions.
	if channelSubs, ok := b.eventSubs[ch]; ok {
		for _, events := range channelSubs {
			// Remove previous latest event.
			select {
			case <-events:
			default:
			}
			// Add latest event.
			events <- e
		}
	}
}

// Progress progresses the channel state.
func (b *MockBackend) Progress(_ context.Context, req channel.ProgressReq) error {
	b.log.Infof("Progress: %+v", req)

	b.mu.Lock()
	defer b.mu.Unlock()

	b.setLatestEvent(
		req.Params.ID(),
		channel.NewProgressedEvent(
			req.Params.ID(),
			&channel.ElapsedTimeout{},
			req.NewState.Clone(),
			req.Idx,
		),
	)
	return nil
}

// outcomeRecursive returns the accumulated outcome of the channel and its sub-channels.
func outcomeRecursive(state *channel.State, subStates channel.StateMap) (outcome channel.Balances) {
	outcome = state.Balances.Clone()
	for _, subAlloc := range state.Locked {
		subOutcome := outcomeRecursive(subStates[subAlloc.ID], subStates)
		for a, bals := range subOutcome {
			for p, bal := range bals {
				_p := p
				if len(subAlloc.IndexMap) > 0 {
					_p = int(subAlloc.IndexMap[p])
				}
				outcome[a][_p].Add(outcome[a][_p], bal)
			}
		}
	}
	return
}

// Withdraw withdraws the channel funds.
func (b *MockBackend) Withdraw(_ context.Context, req channel.AdjudicatorReq, subStates channel.StateMap) error {
	outcome := outcomeRecursive(req.Tx.State, subStates)
	b.log.Infof("Withdraw: %+v, %+v, %+v", req, subStates, outcome)

	return nil
}

// Subscribe creates an event subscription.
func (b *MockBackend) Subscribe(_ context.Context, params *channel.Params) (channel.AdjudicatorSubscription, error) {
	b.log.Infof("SubscribeRegistered: %+v", params)

	b.mu.Lock()
	defer b.mu.Unlock()

	sub := &mockSubscription{
		events: make(chan channel.AdjudicatorEvent, 1),
	}
	b.eventSubs[params.ID()] = append(b.eventSubs[params.ID()], sub.events)

	// Feed latest event if any.
	if e, ok := b.latestEvents[params.ID()]; ok {
		sub.events <- e
	}

	return sub, nil
}

type mockSubscription struct {
	events chan channel.AdjudicatorEvent
}

func (s *mockSubscription) Next() channel.AdjudicatorEvent {
	return <-s.events
}

func (s *mockSubscription) Close() error {
	close(s.events)
	return nil
}

func (s *mockSubscription) Err() error {
	return nil
}

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
	"bytes"
	"context"
	"math/big"
	"math/rand"
	"sync"
	"time"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/log"
	"perun.network/go-perun/pkg/io"
	"perun.network/go-perun/wallet"
)

type (
	// MockBackend is a mocked backend useful for testing.
	MockBackend struct {
		log          log.Logger
		rng          rng
		mu           sync.Mutex
		latestEvents map[channel.ID]channel.AdjudicatorEvent
		eventSubs    map[channel.ID][]chan channel.AdjudicatorEvent
		balances     map[addressMapKey]map[assetMapKey]*big.Int
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
		balances:     make(map[string]map[string]*big.Int),
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

	// Check concluded.
	ch := req.Params.ID()
	if b.isConcluded(ch) {
		log.Debug("register: already concluded:", ch)
		return nil
	}

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
	b.mu.Lock()
	defer b.mu.Unlock()

	// Check concluded.
	ch := req.Params.ID()
	if b.isConcluded(ch) {
		log.Debug("withdraw: already concluded:", ch)
		return nil
	}

	outcome := outcomeRecursive(req.Tx.State, subStates)
	b.log.Infof("Withdraw: %+v, %+v, %+v", req, subStates, outcome)
	for a, assetOutcome := range outcome {
		asset := req.Tx.Allocation.Assets[a]
		for p, amount := range assetOutcome {
			participant := req.Params.Parts[p]
			b.addBalance(participant, asset, amount)
		}
	}

	b.setLatestEvent(ch, channel.NewConcludedEvent(ch, &channel.ElapsedTimeout{}, req.Tx.Version))
	return nil
}

func (b *MockBackend) isConcluded(ch channel.ID) bool {
	e, ok := b.latestEvents[ch]
	if !ok {
		return false
	}
	if _, ok := e.(*channel.ConcludedEvent); !ok {
		return false
	}
	return true
}

func (b *MockBackend) addBalance(p wallet.Address, a channel.Asset, v *big.Int) {
	bal := b.balance(p, a)
	bal = new(big.Int).Add(bal, v)
	b.setBalance(p, a, bal)
}

func (b *MockBackend) balance(p wallet.Address, a channel.Asset) *big.Int {
	partBals, ok := b.balances[newAddressMapKey(p)]
	if !ok {
		return big.NewInt(0)
	}
	bal, ok := partBals[newAssetMapKey(a)]
	if !ok {
		return big.NewInt(0)
	}
	return new(big.Int).Set(bal)
}

type (
	addressMapKey = string
	assetMapKey   = string
)

func newAddressMapKey(a wallet.Address) addressMapKey {
	return encodableAsString(a)
}

func newAssetMapKey(a channel.Asset) assetMapKey {
	return encodableAsString(a)
}

func encodableAsString(e io.Encoder) string {
	var buf bytes.Buffer
	if err := e.Encode(&buf); err != nil {
		panic(err)
	}
	return buf.String()
}

// Balance returns the balance for the participant and asset.
func (b *MockBackend) Balance(p wallet.Address, a channel.Asset) *big.Int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.balance(p, a)
}

func (b *MockBackend) setBalance(p wallet.Address, a channel.Asset, v *big.Int) {
	partKey := newAddressMapKey(p)
	partBals, ok := b.balances[partKey]
	if !ok {
		log.Debug("part not found", p)
		partBals = make(map[string]*big.Int)
		b.balances[partKey] = partBals
	}
	log.Debug("set balance:", p, v)
	partBals[newAssetMapKey(a)] = new(big.Int).Set(v)
}

// Subscribe creates an event subscription.
func (b *MockBackend) Subscribe(ctx context.Context, chID channel.ID) (channel.AdjudicatorSubscription, error) {
	b.log.Infof("SubscribeRegistered: %+v", chID)

	b.mu.Lock()
	defer b.mu.Unlock()

	sub := &mockSubscription{
		ctx:    ctx,
		events: make(chan channel.AdjudicatorEvent, 1),
		err:    make(chan error, 1),
	}
	b.eventSubs[chID] = append(b.eventSubs[chID], sub.events)

	// Feed latest event if any.
	if e, ok := b.latestEvents[chID]; ok {
		sub.events <- e
	}

	return sub, nil
}

type mockSubscription struct {
	ctx    context.Context
	events chan channel.AdjudicatorEvent
	err    chan error
}

func (s *mockSubscription) Next() channel.AdjudicatorEvent {
	select {
	case e := <-s.events:
		return e
	case <-s.ctx.Done():
		s.err <- s.ctx.Err()
		return nil
	}
}

func (s *mockSubscription) Close() error {
	close(s.events)
	return nil
}

func (s *mockSubscription) Err() error {
	return <-s.err
}

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
	"encoding"
	"fmt"
	"math/big"
	"math/rand"
	"time"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/multi"
	"perun.network/go-perun/log"
	"perun.network/go-perun/wallet"
	"polycry.pt/poly-go/sync"
)

type (
	// MockBackend is a mocked backend useful for testing.
	MockBackend struct {
		log          log.Logger
		rng          rng
		mu           sync.Mutex
		assetHolder  *assetHolder
		latestEvents map[channel.ID]channel.AdjudicatorEvent
		eventSubs    map[channel.ID][]*MockSubscription
		balances     map[addressMapKey]map[assetMapKey]*big.Int
		id           multi.AssetID
	}

	rng interface {
		Intn(n int) int
	}

	threadSafeRng struct {
		mu sync.Mutex
		r  *rand.Rand
	}

	// LedgerID is the type of the ledger identifier.
	LedgerID string
)

// maximal amount of milliseconds that the Fund method waits before returning.
const fundMaxSleepMs = 100

// NewMockBackend creates a new backend object.
func NewMockBackend(rng *rand.Rand, id string) *MockBackend {
	rndSource := rand.NewSource(rng.Int63())
	//nolint:gosec // OK to use weak RNG for testing.
	backendRng := rand.New(rndSource)
	return &MockBackend{
		log:          log.Default(),
		rng:          newThreadSafePrng(backendRng),
		assetHolder:  newAssetHolder(newThreadSafePrng(backendRng)),
		latestEvents: make(map[channel.ID]channel.AdjudicatorEvent),
		eventSubs:    make(map[channel.ID][]*MockSubscription),
		balances:     make(map[string]map[string]*big.Int),
		id:           multi.AssetID{0, LedgerID(id)},
	}
}

// ID returns the ledger's identifier.
func (b *MockBackend) ID() multi.AssetID {
	return b.id
}

// MapKey returns the map key representation of the ledger identifier.
func (id LedgerID) MapKey() multi.LedgerIDMapKey {
	return multi.LedgerIDMapKey(id)
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

// MockFunder is a funder used for testing.
type MockFunder struct {
	b   *MockBackend
	acc wallet.Address
}

// Fund funds a given channel.
func (f *MockFunder) Fund(ctx context.Context, req channel.FundingReq) error {
	return f.b.Fund(ctx, req, f.acc)
}

// NewFunder returns a new MockFunder.
func (b *MockBackend) NewFunder(acc wallet.Address) *MockFunder {
	return &MockFunder{
		b:   b,
		acc: acc,
	}
}

// Fund funds the channel.
func (b *MockBackend) Fund(ctx context.Context, req channel.FundingReq, acc wallet.Address) error {
	b.log.Infof("Funding: %+v", req)
	b.assetHolder.Fund(req, b, acc)
	return b.assetHolder.WaitForFunding(ctx, req)
}

// MockAdjudicator is an adjudicator used for testing.
type MockAdjudicator struct {
	*MockBackend
	acc wallet.Address
}

// Withdraw withdraws the balances of the given channel and its sub-channels.
func (a *MockAdjudicator) Withdraw(ctx context.Context, req channel.AdjudicatorReq, subStates channel.StateMap) error {
	return a.MockBackend.Withdraw(ctx, req, subStates, a.acc)
}

// NewAdjudicator creates a new MockAdjudicator.
func (b *MockBackend) NewAdjudicator(acc wallet.Address) *MockAdjudicator {
	return &MockAdjudicator{
		MockBackend: b,
		acc:         acc,
	}
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

	// Check register requirements.
	states := make([]*channel.State, 1+len(subChannels))
	states[0] = req.Tx.State
	for i, subCh := range subChannels {
		states[1+i] = subCh.State
	}
	if err := b.checkStates(states, checkRegister); err != nil {
		return err
	}

	channels := append([]channel.SignedState{
		{
			Params: req.Params,
			State:  req.Tx.State,
			Sigs:   req.Tx.Sigs,
		},
	}, subChannels...)

	timeout := time.Now().Add(time.Duration(req.Params.ChallengeDuration) * time.Millisecond)
	for _, ch := range channels {
		b.setLatestEvent(
			ch.Params.ID(),
			channel.NewRegisteredEvent(
				ch.Params.ID(),
				&channel.TimeTimeout{Time: timeout},
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
		for _, sub := range channelSubs {
			// Remove previous latest event.
			select {
			case <-sub.events:
			default:
			}
			// Add latest event.
			sub.events <- e
		}
	}
}

// Progress progresses the channel state.
func (b *MockBackend) Progress(_ context.Context, req channel.ProgressReq) error {
	b.log.Infof("Progress: %+v", req)

	b.mu.Lock()
	defer b.mu.Unlock()

	timeout := time.Now().Add(time.Duration(req.Params.ChallengeDuration) * time.Millisecond)
	b.setLatestEvent(
		req.Params.ID(),
		channel.NewProgressedEvent(
			req.Params.ID(),
			&channel.TimeTimeout{Time: timeout},
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

type checkStateFunc func(e channel.AdjudicatorEvent, ok bool, s *channel.State) error

// checkRegister checks the following for the given channels:
// - If the channel is already registered, the given version must be greater or equal to the registered version.
func checkRegister(e channel.AdjudicatorEvent, ok bool, s *channel.State) error {
	v := s.Version
	if ok && e.Version() > v {
		return fmt.Errorf("invalid version: expected >=%v, got %v", e.Version(), v)
	}
	return nil
}

// checkWithdraw checks the following for the given channels:
// - If the channel is not registered, the given state must be final.
// - If the channel is already registered, the given version must be equal the registered version.
func checkWithdraw(e channel.AdjudicatorEvent, ok bool, s *channel.State) error {
	v := s.Version
	if !ok && !s.IsFinal {
		return fmt.Errorf("invalid version: expected %v, got not registered", v)
	} else if ok && e.Version() != v {
		return fmt.Errorf("invalid version: expected %v, got %v", e.Version(), v)
	}
	return nil
}

func (b *MockBackend) checkStates(states []*channel.State, op checkStateFunc) error {
	for _, s := range states {
		if err := b.checkState(s, op); err != nil {
			return err
		}
	}
	return nil
}

func (b *MockBackend) checkState(s *channel.State, op checkStateFunc) error {
	e, ok := b.latestEvents[s.ID]
	if err := op(e, ok, s); err != nil {
		return err
	}
	return nil
}

// Withdraw withdraws the channel funds.
func (b *MockBackend) Withdraw(_ context.Context, req channel.AdjudicatorReq, subStates channel.StateMap, acc wallet.Address) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Check withdraw requirements.
	states := make([]*channel.State, 1+len(subStates))
	states[0] = req.Tx.State
	i := 1
	for _, s := range subStates {
		states[i] = s
		i++
	}
	if err := b.checkStates(states, checkWithdraw); err != nil {
		return err
	}

	// Redistribute balances if not done already.
	b.assetHolder.mtx.Lock()
	defer b.assetHolder.mtx.Unlock()
	ch := req.Params.ID()
	if !b.isConcluded(ch) {
		outcome := outcomeRecursive(req.Tx.State, subStates)
		b.log.Infof("Withdraw: %+v, %+v, %+v", req, subStates, outcome)
		outcomeSum := outcome.Sum()

		funding := b.assetHolder.balances[ch]
		if funding == nil {
			funding = channel.MakeBalances(len(req.Tx.Assets), req.Tx.NumParts())
		}
		fundingSum := funding.Sum()

		for a, assetOutcome := range outcome {
			// If underfunded, don't redistribute balances.
			if fundingSum[a].Cmp(outcomeSum[a]) < 0 {
				continue
			}
			for p, amount := range assetOutcome {
				funding[a][p].Set(amount)
			}
		}
	}

	// Payout balances.
	balances := b.assetHolder.balances[ch]
	for a, assetBalances := range balances {
		asset := req.Tx.Allocation.Assets[a]
		p := req.Idx
		amount := assetBalances[p]
		b.addBalance(acc, asset, amount)
		amount.Set(big.NewInt(0))
	}

	if !b.isConcluded(ch) {
		b.setLatestEvent(ch, channel.NewConcludedEvent(ch, &channel.ElapsedTimeout{}, req.Tx.Version))
	}
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

func (b *MockBackend) subBalance(p wallet.Address, a channel.Asset, v *big.Int) {
	bal := b.balance(p, a)
	bal = new(big.Int).Sub(bal, v)
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

func encodableAsString(e encoding.BinaryMarshaler) string {
	buff, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return string(buff)
}

// MockBalanceReader is a balance reader used for testing. At initialization, it
// is associated with a given account.
type MockBalanceReader struct {
	b   *MockBackend
	acc wallet.Address
}

// Balance returns the balance of the associated account for the given asset.
func (br *MockBalanceReader) Balance(asset channel.Asset) channel.Bal {
	return br.b.Balance(br.acc, asset)
}

// NewBalanceReader creates balance for the given account.
func (b *MockBackend) NewBalanceReader(acc wallet.Address) *MockBalanceReader {
	return &MockBalanceReader{
		b:   b,
		acc: acc,
	}
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

	sub := NewMockSubscription(ctx)
	sub.onClose = func() { b.removeSubscription(chID, sub) }
	b.eventSubs[chID] = append(b.eventSubs[chID], sub)

	// Feed latest event if any.
	if e, ok := b.latestEvents[chID]; ok {
		sub.events <- e
	}

	return sub, nil
}

func (b *MockBackend) removeSubscription(ch channel.ID, sub *MockSubscription) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Find subscription index.
	i, ok := func() (int, bool) {
		for i, s := range b.eventSubs[ch] {
			if sub == s {
				return i, true
			}
		}
		return 0, false
	}()

	if !ok {
		log.Warnf("Remove subscription: Not found: %v", ch)
		return
	}

	subs := b.eventSubs[ch]
	b.eventSubs[ch] = append(subs[:i], subs[i+1:]...)
}

// assetHolder mocks an assetHolder for the MockBackend.
type assetHolder struct {
	rng       rng
	mtx       sync.Mutex
	balances  map[channel.ID]channel.Balances
	fundedWgs map[channel.ID]*sync.WaitGroup
}

// newAssetHolder returns a new funder.
func newAssetHolder(rng rng) *assetHolder {
	return &assetHolder{
		rng:       rng,
		balances:  make(map[channel.ID]channel.Balances),
		fundedWgs: make(map[channel.ID]*sync.WaitGroup),
	}
}

// initFund initializes the funded WaitGroups for a channel if not already
// initialized.
//
// Must be called before using the Funder for a funding request.
func (f *assetHolder) initFund(req channel.FundingReq) {
	f.mtx.Lock()
	defer f.mtx.Unlock()

	if f.fundedWgs[req.Params.ID()] == nil {
		f.fundedWgs[req.Params.ID()] = &sync.WaitGroup{}
		f.fundedWgs[req.Params.ID()].Add(len(req.Params.Parts))
	}
	if f.balances[req.Params.ID()] == nil {
		f.balances[req.Params.ID()] = channel.MakeBalances(len(req.State.Assets), req.State.NumParts())
	}
}

// Fund simulates funding the channel.
func (f *assetHolder) Fund(req channel.FundingReq, b *MockBackend, acc wallet.Address) {
	f.initFund(req)

	// Simulates a random delay during funding.
	time.Sleep(time.Duration(f.rng.Intn(fundMaxSleepMs+1)) * time.Millisecond)

	for i, asset := range req.State.Assets {
		ma, ok := asset.(*MultiLedgerAsset)
		if ok && ma.AssetID() != b.ID() {
			continue
		}

		bal := req.Agreement[i][req.Idx]
		b.mu.Lock()
		b.subBalance(acc, asset, bal)
		b.mu.Unlock()
		f.mtx.Lock()
		fundingBal := f.balances[req.Params.ID()][i][req.Idx]
		fundingBal.Add(fundingBal, bal)
		f.mtx.Unlock()
	}

	f.fundedWgs[req.Params.ID()].Done()
}

// WaitForFunding waits until all participants have funded the channel.
func (f *assetHolder) WaitForFunding(ctx context.Context, req channel.FundingReq) error {
	challengeDuration := time.Duration(req.Params.ChallengeDuration) * time.Second
	fundCtx, cancel := context.WithTimeout(ctx, challengeDuration)
	defer cancel()

	select {
	case <-f.fundedWgs[req.Params.ID()].WaitCh():
		log.Infof("Funded: %+v", req)
		return nil
	case <-fundCtx.Done():
		return channel.FundingTimeoutError{}
	}
}

// MockSubscription is a subscription for MockBackend.
type MockSubscription struct {
	events  chan channel.AdjudicatorEvent
	err     chan error
	onClose func()
}

// NewMockSubscription creates a new MockSubscription.
func NewMockSubscription(ctx context.Context) *MockSubscription {
	return &MockSubscription{
		events: make(chan channel.AdjudicatorEvent, 1),
		err:    make(chan error, 1),
	}
}

// Next returns the next event.
func (s *MockSubscription) Next() channel.AdjudicatorEvent {
	return <-s.events
}

// Close closes the subscription.
func (s *MockSubscription) Close() error {
	s.onClose()
	close(s.events)
	close(s.err)
	return nil
}

// Err returns the subscription's error after it has failed.
func (s *MockSubscription) Err() error {
	return <-s.err
}

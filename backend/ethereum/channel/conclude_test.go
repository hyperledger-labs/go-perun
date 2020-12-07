// Copyright 2020 - See NOTICE file for copyright holders.
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

package channel_test

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ethchannel "perun.network/go-perun/backend/ethereum/channel"
	"perun.network/go-perun/backend/ethereum/channel/test"
	"perun.network/go-perun/backend/ethereum/wallet/keystore"
	"perun.network/go-perun/channel"
	channeltest "perun.network/go-perun/channel/test"
	"perun.network/go-perun/pkg/errors"
	pkgtest "perun.network/go-perun/pkg/test"
	"perun.network/go-perun/wallet"
)

const defaultTestTimeout = 10 * time.Second

func TestAdjudicator_ConcludeFinal(t *testing.T) {
	t.Run("ConcludeFinal 1 party", func(t *testing.T) { testConcludeFinal(t, 1) })
	t.Run("ConcludeFinal 2 party", func(t *testing.T) { testConcludeFinal(t, 2) })
	t.Run("ConcludeFinal 5 party", func(t *testing.T) { testConcludeFinal(t, 5) })
	t.Run("ConcludeFinal 10 party", func(t *testing.T) { testConcludeFinal(t, 10) })
}

func testConcludeFinal(t *testing.T, numParts int) {
	rng := pkgtest.Prng(t)
	// create test setup
	s := test.NewSetup(t, rng, numParts)
	// create valid state and params
	params, state := channeltest.NewRandomParamsAndState(rng, channeltest.WithParts(s.Parts...), channeltest.WithAssets((*ethchannel.Asset)(&s.Asset)), channeltest.WithIsFinal(true))
	// we need to properly fund the channel
	fundingCtx, funCancel := context.WithTimeout(context.Background(), defaultTxTimeout)
	defer funCancel()
	// fund the contract
	ct := pkgtest.NewConcurrent(t)
	for i, funder := range s.Funders {
		i, funder := i, funder
		go ct.StageN("funding loop", numParts, func(rt pkgtest.ConcT) {
			req := channel.NewFundingReq(params, state, channel.Index(i), state.Balances)
			require.NoError(rt, funder.Fund(fundingCtx, *req), "funding should succeed")
		})
	}
	ct.Wait("funding loop")
	tx := testSignState(t, s.Accs, params, state)

	ctx, cancel := context.WithTimeout(context.Background(), defaultTxTimeout)
	defer cancel()
	ct = pkgtest.NewConcurrent(t)
	initiator := int(rng.Int31n(int32(numParts))) // pick a random initiator
	for i := 0; i < numParts; i++ {
		i := i
		go ct.StageN("register", numParts, func(t pkgtest.ConcT) {
			req := channel.AdjudicatorReq{
				Params:    params,
				Acc:       s.Accs[i],
				Idx:       channel.Index(i),
				Tx:        tx,
				Secondary: (i != initiator),
			}
			diff, err := test.NonceDiff(s.Accs[i].Address(), s.Adjs[i], func() error {
				_, err := s.Adjs[i].Register(ctx, req)
				return err
			})
			require.NoError(t, err, "Withdrawing should succeed")
			if !req.Secondary {
				// The Initiator must send a TX.
				require.Equal(t, diff, 1)
			} else {
				// Everyone else must NOT send a TX.
				require.Equal(t, diff, 0)
			}
		})
	}
	ct.Wait("register")
}

func TestAdjudicator_ConcludeWithSubChannels(t *testing.T) {
	// 0. setup

	const (
		numParts               = 2
		maxCountSubChannels    = 5
		maxCountSubSubChannels = 5
		maxChallengeDuration   = 3600
	)
	ctx, cancel := newDefaultTestContext()
	defer cancel()
	var (
		assert  = assert.New(t)
		require = require.New(t)
		rng     = pkgtest.Prng(t)
	)
	// create backend
	var (
		s                 = test.NewSetup(t, rng, numParts)
		adj               = s.Adjs[0]
		accounts          = s.Accs
		participants      = s.Parts
		asset             = (*ethchannel.Asset)(&s.Asset)
		challengeDuration = uint64(rng.Intn(maxChallengeDuration))
		makeRandomChannel = func(rng *rand.Rand) paramsAndState {
			return makeRandomChannel(rng, participants, asset, challengeDuration)
		}
	)
	// create channels
	var (
		ledgerChannel  = makeRandomChannel(rng)
		subChannels    = makeRandomChannels(rng, maxCountSubChannels, makeRandomChannel)
		subSubChannels = makeRandomChannels(rng, maxCountSubSubChannels, makeRandomChannel)
	)
	// update sub-channel locked funds
	parentChannel := randomChannel(rng, subChannels)
	for _, c := range subSubChannels {
		parentChannel.state.AddSubAlloc(*c.state.ToSubAlloc())
	}
	// update ledger channel locked funds
	for _, c := range subChannels {
		ledgerChannel.state.AddSubAlloc(*c.state.ToSubAlloc())
	}
	// fund
	require.NoError(fund(ctx, s.Funders, ledgerChannel))

	// 1. register channels

	require.NoError(register(ctx, adj, accounts, subSubChannels...))
	require.NoError(register(ctx, adj, accounts, subChannels...))
	require.NoError(register(ctx, adj, accounts, ledgerChannel))

	// 2. wait until ready to conclude

	sub, err := adj.Subscribe(ctx, ledgerChannel.params)
	require.NoError(err)
	require.NoError(sub.Next().Timeout().Wait(ctx))
	sub.Close()

	// 3. withdraw channel with sub-channels

	subStates := channel.MakeStateMap()
	addSubStates(subStates, ledgerChannel)
	addSubStates(subStates, subChannels...)
	addSubStates(subStates, subSubChannels...)

	assert.NoError(withdraw(ctx, adj, accounts, ledgerChannel, subStates))
}

type paramsAndState struct {
	params *channel.Params
	state  *channel.State
}

func makeRandomChannel(rng *rand.Rand, participants []wallet.Address, asset channel.Asset, challengeDuration uint64) paramsAndState {
	params, state := channeltest.NewRandomParamsAndState(
		rng,
		channeltest.WithParts(participants...),
		channeltest.WithAssets(asset),
		channeltest.WithIsFinal(false),
		channeltest.WithNumLocked(0),
		channeltest.WithoutApp(),
		channeltest.WithChallengeDuration(challengeDuration),
	)
	return paramsAndState{params, state}
}

func makeRandomChannels(rng *rand.Rand, maxCount int, makeRandomChannel func(rng *rand.Rand) paramsAndState) []paramsAndState {
	channels := make([]paramsAndState, 1+rng.Intn(maxCount))
	for i := range channels {
		channels[i] = makeRandomChannel(rng)
	}
	return channels
}

func randomChannel(rng *rand.Rand, channels []paramsAndState) *paramsAndState {
	return &channels[rng.Intn(len(channels))]
}

func fund(ctx context.Context, funders []*ethchannel.Funder, c paramsAndState) error {
	errg := errors.NewGatherer()
	for i, funder := range funders {
		i, funder := i, funder
		errg.Go(func() error {
			req := channel.NewFundingReq(c.params, c.state, channel.Index(i), c.state.Balances)
			return funder.Fund(ctx, *req)
		})
	}
	return errg.Wait()
}

func register(ctx context.Context, adj *test.SimAdjudicator, accounts []*keystore.Account, channels ...paramsAndState) error {
	for _, c := range channels {
		tx, err := signState(accounts, c.params, c.state)
		if err != nil {
			return err
		}

		req := channel.AdjudicatorReq{
			Params:    c.params,
			Acc:       accounts[0],
			Idx:       0,
			Tx:        tx,
			Secondary: false,
		}

		if _, err := adj.Register(ctx, req); err != nil {
			return err
		}
	}
	return nil
}

func addSubStates(subStates channel.StateMap, channels ...paramsAndState) {
	for _, c := range channels {
		subStates.Add(c.state.Clone())
	}
}

func withdraw(ctx context.Context, adj *test.SimAdjudicator, accounts []*keystore.Account, c paramsAndState, subStates channel.StateMap) error {
	tx, err := signState(accounts, c.params, c.state)
	if err != nil {
		return err
	}

	for i, a := range accounts {
		req := channel.AdjudicatorReq{
			Params:    c.params,
			Acc:       a,
			Idx:       channel.Index(i),
			Tx:        tx,
			Secondary: i != 0,
		}

		if err := adj.Withdraw(ctx, req, subStates); err != nil {
			return err
		}
	}

	return nil
}

func newDefaultTestContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), defaultTestTimeout)
}

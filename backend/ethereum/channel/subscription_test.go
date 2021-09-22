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

package channel_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"perun.network/go-perun/backend/ethereum/channel/test"
	"perun.network/go-perun/channel"
	channeltest "perun.network/go-perun/channel/test"
	pkgtest "perun.network/go-perun/pkg/test"
)

func TestSubscription(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTestTimeout)
	defer cancel()

	rng := pkgtest.Prng(t)
	numParts := 2 + rng.Intn(maxNumParts-2)
	setup := test.NewSetup(t, rng, numParts, blockInterval)
	adjSetup := newAdjudicatorSetup(setup)
	adj := adjSetup.Adjudicator()

	opts := []channeltest.RandomOpt{channeltest.WithoutApp(), channeltest.WithIsFinal(false)}
	req, subChannels := adjSetup.newAdjudicatorReq(ctx, rng, opts...), []channel.SignedState{}
	sub, err := adj.Subscribe(ctx, req.Params.ID())
	require.NoError(t, err, "subscribing")

	subSetup := newSubscriptionSetup(sub, adj, req, subChannels)
	channeltest.TestSubscription(ctx, t, subSetup)
}

type subscriptionSetup struct {
	adj         channel.Adjudicator
	sub         *testSubscription
	req         *channel.AdjudicatorReq
	subChannels []channel.SignedState
}

func newSubscriptionSetup(
	sub channel.AdjudicatorSubscription,
	adj channel.Adjudicator,
	req *channel.AdjudicatorReq,
	subChannels []channel.SignedState,
) *subscriptionSetup {
	return &subscriptionSetup{
		sub: &testSubscription{
			emitProgressed:          false,
			req:                     req,
			AdjudicatorSubscription: sub,
		},
		adj:         adj,
		req:         req,
		subChannels: subChannels,
	}
}

type testSubscription struct {
	channel.AdjudicatorSubscription
	emitProgressed bool
	req            *channel.AdjudicatorReq
}

// Next emulates app channel functionality because app channels are not supported yet.
func (s *testSubscription) Next() channel.AdjudicatorEvent {
	if s.emitProgressed {
		s.emitProgressed = false
		return channel.NewProgressedEvent(s.req.Tx.ID, &channel.ElapsedTimeout{}, s.req.Tx.State, s.req.Idx)
	}
	return s.AdjudicatorSubscription.Next()
}

func (s *subscriptionSetup) Subscription() channel.AdjudicatorSubscription {
	return s.sub
}

// EmitRegistered operates the adjudicator so that a registered event should be emitted.
func (s *subscriptionSetup) EmitRegistered(ctx context.Context) (channel.Params, channel.State) {
	err := s.adj.Register(ctx, *s.req, s.subChannels)
	if err != nil {
		panic(err)
	}
	return *s.req.Params, *s.req.Tx.State
}

// EmitProgressed emulates app channel functionality because app channels are not supported yet.
func (s *subscriptionSetup) EmitProgressed(ctx context.Context) (channel.Params, channel.State) {
	s.sub.emitProgressed = true
	return *s.req.Params, *s.req.Tx.State
}

// EmitRegistered operates the adjudicator so that a concluded event should be emitted.
func (s *subscriptionSetup) EmitConcluded(ctx context.Context) channel.Params {
	err := s.adj.Withdraw(ctx, *s.req, channeltest.MakeStateMapFromSignedStates(s.subChannels...))
	if err != nil {
		panic(err)
	}
	return *s.req.Params
}

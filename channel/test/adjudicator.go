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
	"testing"

	"github.com/stretchr/testify/require"
	"perun.network/go-perun/channel"
	pkgtest "perun.network/go-perun/pkg/test"
)

// Adjudicator represents an adjudicator test setup.
//
// Adjudicator should return the adjudicator instance used for testing.
// NewRegisterReq should return the inputs to a successful register call.
// NewProgressReq should return the inputs to a successful progress call.
// NewWithdrawReq should return the inputs to a successful withdraw call.
type Adjudicator interface {
	Adjudicator() channel.Adjudicator
	NewRegisterReq(context.Context, *rand.Rand) (*channel.AdjudicatorReq, []channel.SignedState)
	NewProgressReq(context.Context, *rand.Rand) *channel.ProgressReq
	NewWithdrawReq(context.Context, *rand.Rand) (*channel.AdjudicatorReq, channel.StateMap)
}

// TestAdjudicator tests the given adjudicator.
func TestAdjudicator(ctx context.Context, t *testing.T, a Adjudicator) {
	rng := pkgtest.Prng(t)

	adj := a.Adjudicator()

	t.Run("register and subscribe", func(t *testing.T) {
		req, subChannels := a.NewRegisterReq(ctx, rng)
		err := adj.Register(ctx, *req, subChannels)
		require.NoError(t, err, "registering")

		_, err = adj.Subscribe(ctx, req.Params.ID())
		require.NoError(t, err, "subscribing")
	})

	t.Run("progress", func(t *testing.T) {
		req := a.NewProgressReq(ctx, rng)
		err := adj.Progress(ctx, *req)
		require.NoError(t, err, "progressing")
	})

	t.Run("withdraw", func(t *testing.T) {
		req, subChannels := a.NewWithdrawReq(ctx, rng)
		err := adj.Withdraw(ctx, *req, subChannels)
		require.NoError(t, err, "withdrawing")
	})
}

func MakeStateMapFromSignedStates(channels ...channel.SignedState) channel.StateMap {
	m := channel.MakeStateMap()
	for _, c := range channels {
		m.Add(c.State)
	}
	return m
}

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
	"testing"

	"github.com/stretchr/testify/assert"
	"perun.network/go-perun/channel"
)

// SubscriptionTestSetup represents a setup for testing a subscription implementation.
//
// Subscription should return an instance of the subscription implementation.
// EmitRegistered should make the subscription emit a registered event.
// EmitProgressed should make the subscription emit a progressed event.
// EmitConcluded should make the subscription emit a concluded event.
type SubscriptionTestSetup interface {
	Subscription() channel.AdjudicatorSubscription
	EmitRegistered(context.Context) (channel.Params, channel.State)
	EmitProgressed(context.Context) (channel.Params, channel.State)
	EmitConcluded(context.Context) channel.Params
}

func TestSubscription(ctx context.Context, t *testing.T, s SubscriptionTestSetup) {
	sub := s.Subscription()

	{
		params, state := s.EmitRegistered(ctx)
		e, ok := sub.Next().(*channel.RegisteredEvent)
		assert.True(t, ok, "registered")
		assert.True(t, e.ID() == params.ID(), "equal ID")
		assert.True(t, e.State.Equal(&state) == nil, "equal state")
		err := e.Timeout().Wait(ctx)
		assert.NoError(t, err, "registered: waiting")
	}

	{
		params, state := s.EmitProgressed(ctx)
		e, ok := sub.Next().(*channel.ProgressedEvent)
		assert.True(t, ok, "progressed")
		assert.True(t, e.ID() == params.ID(), "equal ID")
		assert.True(t, e.State.Equal(&state) == nil, "equal state")
		err := e.Timeout().Wait(ctx)
		assert.NoError(t, err, "progressed: waiting")
	}

	{
		params := s.EmitConcluded(ctx)
		e, ok := sub.Next().(*channel.ConcludedEvent)
		assert.True(t, ok, "concluded")
		assert.True(t, e.ID() == params.ID(), "equal ID")
		err := e.Timeout().Wait(ctx)
		assert.NoError(t, err, "concluded: waiting")
	}

	{
		err := sub.Close()
		assert.NoError(t, err, "close")
		err = sub.Err()
		assert.NoError(t, err, "err")
	}
}

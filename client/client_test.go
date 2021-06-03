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

package client

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/pkg/test"
	wtest "perun.network/go-perun/wallet/test"
	"perun.network/go-perun/wire"
)

type DummyBus struct {
	t *testing.T
}

func (d DummyBus) Publish(context.Context, *wire.Envelope) error {
	d.t.Error("DummyBus.Publish called")
	return errors.New("DummyBus.Publish called")
}

func (d DummyBus) SubscribeClient(wire.Consumer, wire.Address) error {
	return nil
}

type DummyFunder struct {
	t *testing.T
}

func (d *DummyFunder) Fund(context.Context, channel.FundingReq) error {
	d.t.Error("DummyFunder.Fund called")
	return errors.New("DummyFunder.Fund called")
}

type DummyAdjudicator struct {
	t *testing.T
}

func (d *DummyAdjudicator) Register(context.Context, channel.AdjudicatorReq, []channel.SignedState) error {
	d.t.Error("DummyAdjudicator.Register called")
	return errors.New("DummyAdjudicator.Register called")
}

func (d *DummyAdjudicator) Progress(context.Context, channel.ProgressReq) error {
	d.t.Error("DummyAdjudicator.Register called")
	return errors.New("DummyAdjudicator.Progress called")
}

func (d *DummyAdjudicator) Withdraw(context.Context, channel.AdjudicatorReq, channel.StateMap) error {
	d.t.Error("DummyAdjudicator.Withdraw called")
	return errors.New("DummyAdjudicator.Withdraw called")
}

func (d *DummyAdjudicator) Subscribe(context.Context, *channel.Params) (channel.AdjudicatorSubscription, error) {
	d.t.Error("DummyAdjudicator.SubscribeRegistered called")
	return nil, errors.New("DummyAdjudicator.SubscribeRegistered called")
}

func TestClient_New_NilArgs(t *testing.T) {
	rng := test.Prng(t)
	id := wtest.NewRandomAddress(rng)
	b, f, a, w := &DummyBus{t}, &DummyFunder{t}, &DummyAdjudicator{t}, wtest.RandomWallet()
	assert.Panics(t, func() { New(nil, b, f, a, w) })
	assert.Panics(t, func() { New(id, nil, f, a, w) })
	assert.Panics(t, func() { New(id, b, nil, a, w) })
	assert.Panics(t, func() { New(id, b, f, nil, w) })
	assert.Panics(t, func() { New(id, b, f, a, nil) })
}

func TestClient_Handle_NilArgs(t *testing.T) {
	rng := test.Prng(t)
	c, err := New(wtest.NewRandomAddress(rng), &DummyBus{t}, &DummyFunder{t}, &DummyAdjudicator{t}, wtest.RandomWallet())
	require.NoError(t, err)

	dummyUH := UpdateHandlerFunc(func(*channel.State, ChannelUpdate, *UpdateResponder) {})
	assert.Panics(t, func() { c.Handle(nil, dummyUH) })
	dummyPH := ProposalHandlerFunc(func(ChannelProposal, *ProposalResponder) {})
	assert.Panics(t, func() { c.Handle(dummyPH, nil) })
}

func TestClient_New(t *testing.T) {
	rng := test.Prng(t)
	c, err := New(wtest.NewRandomAddress(rng), &DummyBus{t}, &DummyFunder{t}, &DummyAdjudicator{t}, wtest.RandomWallet())
	assert.NoError(t, err)
	require.NotNil(t, c)
}

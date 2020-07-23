// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package client

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/channel"
	wtest "perun.network/go-perun/wallet/test"
	"perun.network/go-perun/wire"
)

const timeout = 5 * time.Second

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

func (d *DummyAdjudicator) Register(context.Context, channel.AdjudicatorReq) (*channel.RegisteredEvent, error) {
	d.t.Error("DummyAdjudicator.Register called")
	return nil, errors.New("DummyAdjudicator.Register called")
}

func (d *DummyAdjudicator) Withdraw(context.Context, channel.AdjudicatorReq) error {
	d.t.Error("DummyAdjudicator.Withdraw called")
	return errors.New("DummyAdjudicator.Withdraw called")
}

func (d *DummyAdjudicator) SubscribeRegistered(context.Context, *channel.Params) (channel.RegisteredSubscription, error) {
	d.t.Error("DummyAdjudicator.SubscribeRegistered called")
	return nil, errors.New("DummyAdjudicator.SubscribeRegistered called")
}

func TestClient_New_NilArgs(t *testing.T) {
	rng := rand.New(rand.NewSource(0x1111))
	id := wtest.NewRandomAddress(rng)
	b, f, a, w := &DummyBus{t}, &DummyFunder{t}, &DummyAdjudicator{t}, wtest.RandomWallet()
	assert.Panics(t, func() { New(nil, b, f, a, w) })
	assert.Panics(t, func() { New(id, nil, f, a, w) })
	assert.Panics(t, func() { New(id, b, nil, a, w) })
	assert.Panics(t, func() { New(id, b, f, nil, w) })
	assert.Panics(t, func() { New(id, b, f, a, nil) })
}

func TestClient_Handle_NilArgs(t *testing.T) {
	rng := rand.New(rand.NewSource(20200524))
	c, err := New(wtest.NewRandomAddress(rng), &DummyBus{t}, &DummyFunder{t}, &DummyAdjudicator{t}, wtest.RandomWallet())
	require.NoError(t, err)

	dummyUH := UpdateHandlerFunc(func(ChannelUpdate, *UpdateResponder) {})
	assert.Panics(t, func() { c.Handle(nil, dummyUH) })
	dummyPH := ProposalHandlerFunc(func(*ChannelProposal, *ProposalResponder) {})
	assert.Panics(t, func() { c.Handle(dummyPH, nil) })
}

func TestClient_New(t *testing.T) {
	rng := rand.New(rand.NewSource(0x1a2b3c))
	c, err := New(wtest.NewRandomAddress(rng), &DummyBus{t}, &DummyFunder{t}, &DummyAdjudicator{t}, wtest.RandomWallet())
	assert.NoError(t, err)
	require.NotNil(t, c)
}

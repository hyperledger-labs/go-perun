// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package wire

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/pkg/sync"
	"perun.network/go-perun/pkg/test"
	wallettest "perun.network/go-perun/wallet/test"
)

func TestProducer(t *testing.T) {
	r0 := NewReceiver()
	r1 := NewReceiver()
	r2 := NewReceiver()
	p := makeProducer()

	pred := func(*Envelope) bool { return true }

	assert.True(t, p.isEmpty())
	assert.NoError(t, p.Subscribe(r0, pred))
	assert.False(t, p.isEmpty())
	assert.NoError(t, p.Subscribe(r1, pred))
	assert.NoError(t, p.Subscribe(r2, pred))
	assert.Equal(t, len(p.consumers), 3)
	p.delete(r0)
	assert.Equal(t, len(p.consumers), 2)
	assert.False(t, p.isEmpty())
	assert.Panics(t, func() { p.delete(r0) })
}

func TestProducer_produce_DefaultMsgHandler(t *testing.T) {
	missedMsg := make(chan *Envelope, 1)
	recv, send := newPipeConnPair()
	p := newEndpoint(nil, recv, nil)
	go p.recvLoop()

	rng := test.Prng(t)

	p.SetDefaultMsgHandler(func(e *Envelope) {
		missedMsg <- e
	})

	test.AssertTerminates(t, timeout, func() {
		r := NewReceiver()
		p.Subscribe(r, func(e *Envelope) bool { return e.Msg.Type() == ChannelProposal })
		assert.NoError(t, send.Send(NewRandomEnvelope(rng, NewPingMsg())))
		assert.IsType(t, &PingMsg{}, (<-missedMsg).Msg)
	})

	test.AssertNotTerminates(t, timeout, func() {
		r := NewReceiver()
		p.Subscribe(r, func(e *Envelope) bool { return e.Msg.Type() == Ping })
		assert.NoError(t, send.Send(NewRandomEnvelope(rng, NewPingMsg())))
		<-missedMsg
	})
}

func TestProducer_produce_closed(t *testing.T) {
	var missed *Envelope
	p := makeProducer()
	p.SetDefaultMsgHandler(func(e *Envelope) { missed = e })
	assert.NoError(t, p.Close())
	rng := test.Prng(t)
	a := wallettest.NewRandomAddress(rng)
	b := wallettest.NewRandomAddress(rng)
	p.produce(&Envelope{a, b, NewPingMsg()})
	assert.Nil(t, missed, "produce() on closed producer shouldn't do anything")
}

func TestProducer_SetDefaultMsgHandler(t *testing.T) {
	fn := func(*Envelope) {}
	p := makeProducer()

	logUnhandledMsgPtr := reflect.ValueOf(logUnhandledMsg).Pointer()

	assert.Equal(t, logUnhandledMsgPtr, reflect.ValueOf(p.defaultMsgHandler).Pointer())
	p.SetDefaultMsgHandler(fn)
	assert.Equal(t, reflect.ValueOf(fn).Pointer(), reflect.ValueOf(p.defaultMsgHandler).Pointer())
	p.SetDefaultMsgHandler(nil)
	assert.Equal(t, logUnhandledMsgPtr, reflect.ValueOf(p.defaultMsgHandler).Pointer())
}

func TestProducer_Close(t *testing.T) {
	p := makeProducer()
	assert.NoError(t, p.Close())

	err := p.Close()
	assert.Error(t, err)
	assert.True(t, sync.IsAlreadyClosedError(err))

	assert.NotPanics(t, func() { p.delete(nil) },
		"delete() on closed producer shouldn't do anything")
}

func TestProducer_Subscribe(t *testing.T) {
	fn := func(*Envelope) bool { return true }
	t.Run("closed", func(t *testing.T) {
		p := makeProducer()
		p.Close()
		assert.Error(t, p.Subscribe(NewReceiver(), fn))
	})

	t.Run("duplicate", func(t *testing.T) {
		p := makeProducer()
		r := NewReceiver()
		assert.NoError(t, p.Subscribe(r, fn))
		assert.Panics(t, func() { p.Subscribe(r, fn) })
	})

	t.Run("closed consumer", func(t *testing.T) {
		p := makeProducer()
		r := NewReceiver()
		r.Close()
		test.AssertTerminates(t, timeout, func() {
			assert.Error(t, p.Subscribe(r, fn))
		})
		time.Sleep(timeout)
		assert.NotPanics(t, func() {
			assert.Error(t, p.Subscribe(r, fn))
		})
	})
}

// Testproducer_caching tests that the producer correctly writes unhandled
// messages to the cache.
func TestProducer_caching(t *testing.T) {
	assert := assert.New(t)
	isPing := func(e *Envelope) bool { return e.Msg.Type() == Ping }
	isPong := func(e *Envelope) bool { return e.Msg.Type() == Pong }
	prod := makeProducer()
	unhandlesMsg := make([]*Envelope, 0, 2)
	prod.SetDefaultMsgHandler(func(e *Envelope) { unhandlesMsg = append(unhandlesMsg, e) })

	ctx := context.Background()
	prod.Cache(ctx, isPing)

	rng := test.Prng(t)
	ping0 := NewRandomEnvelope(rng, NewPingMsg())
	pong1 := NewRandomEnvelope(rng, NewPongMsg())
	pong2 := NewRandomEnvelope(rng, NewPongMsg())

	prod.produce(ping0)
	assert.Equal(1, prod.cache.Size())
	assert.Len(unhandlesMsg, 0)

	prod.produce(pong1)
	assert.Equal(1, prod.cache.Size())
	assert.Len(unhandlesMsg, 1)

	prod.Cache(ctx, isPong)
	prod.produce(pong2)
	assert.Equal(2, prod.cache.Size())
	assert.Len(unhandlesMsg, 1)

	rec := NewReceiver()
	prod.Subscribe(rec, isPing)
	test.AssertTerminates(t, timeout, func() {
		e, err := rec.Next(ctx)
		assert.NoError(err)
		assert.Same(e, ping0)
	})
	assert.Equal(1, prod.cache.Size())
	assert.Len(unhandlesMsg, 1)

	err := prod.Close()
	require.Error(t, err)
	assert.Contains(err.Error(), "cache")
	assert.Zero(prod.cache.Size(), "producer.Close should flush the cache")

	prod.Cache(ctx, func(*Envelope) bool { return true })
	prod.cache.Put(ping0, nil)
	assert.Zero(prod.cache.Size(), "Cache on closed producer should not enable caching")
}

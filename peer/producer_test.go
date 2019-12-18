// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package peer

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/pkg/sync"
	"perun.network/go-perun/pkg/test"
	wire "perun.network/go-perun/wire/msg"
)

func TestProducer(t *testing.T) {
	r0 := NewReceiver()
	r1 := NewReceiver()
	r2 := NewReceiver()
	p := makeProducer()

	pred := func(wire.Msg) bool { return true }

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
	missedMsg := make(chan wire.Msg, 1)
	recv, send := newPipeConnPair()
	p := newPeer(nil, recv, nil)
	go p.recvLoop()

	p.SetDefaultMsgHandler(func(m wire.Msg) {
		missedMsg <- m
	})

	test.AssertTerminates(t, timeout, func() {
		r := NewReceiver()
		p.Subscribe(r, func(m wire.Msg) bool { return m.Type() == wire.ChannelProposal })
		assert.NoError(t, send.Send(wire.NewPingMsg()))
		assert.IsType(t, &wire.PingMsg{}, <-missedMsg)
	})

	test.AssertNotTerminates(t, timeout, func() {
		r := NewReceiver()
		p.Subscribe(r, func(m wire.Msg) bool { return m.Type() == wire.Ping })
		assert.NoError(t, send.Send(wire.NewPingMsg()))
		<-missedMsg
	})
}

func TestProducer_produce_closed(t *testing.T) {
	var missed wire.Msg
	p := makeProducer()
	p.SetDefaultMsgHandler(func(m wire.Msg) { missed = m })
	assert.NoError(t, p.Close())
	p.produce(wire.NewPingMsg(), &Peer{})
	assert.Nil(t, missed, "produce() on closed producer shouldn't do anything")
}

func TestProducer_SetDefaultMsgHandler(t *testing.T) {
	fn := func(wire.Msg) {}
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
	fn := func(wire.Msg) bool { return true }
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
}

// TestProducer_caching tests that the producer correctly writes unhandled
// messages to the cache.
func TestProducer_caching(t *testing.T) {
	assert := assert.New(t)
	isPing := func(m wire.Msg) bool { return m.Type() == wire.Ping }
	isPong := func(m wire.Msg) bool { return m.Type() == wire.Pong }
	prod := makeProducer()
	unhandlesMsg := make([]wire.Msg, 0, 2)
	prod.SetDefaultMsgHandler(func(m wire.Msg) { unhandlesMsg = append(unhandlesMsg, m) })

	ctx := context.Background()
	prod.Cache(ctx, isPing)
	ping0, peer0 := wire.NewPingMsg(), &Peer{} // dummy peer
	pong1 := wire.NewPongMsg()
	pong2 := wire.NewPongMsg()

	prod.produce(ping0, peer0)
	assert.Equal(1, prod.cache.Size())
	assert.Len(unhandlesMsg, 0)

	prod.produce(pong1, &Peer{})
	assert.Equal(1, prod.cache.Size())
	assert.Len(unhandlesMsg, 1)

	prod.Cache(ctx, isPong)
	prod.produce(pong2, &Peer{})
	assert.Equal(2, prod.cache.Size())
	assert.Len(unhandlesMsg, 1)

	rec := NewReceiver()
	prod.Subscribe(rec, isPing)
	test.AssertTerminates(t, timeout, func() {
		recpeer, recmsg := rec.Next(ctx)
		assert.Same(recpeer, peer0)
		assert.Same(recmsg, ping0)
	})
	assert.Equal(1, prod.cache.Size())
	assert.Len(unhandlesMsg, 1)

	err := prod.Close()
	require.Error(t, err)
	assert.Contains(err.Error(), "cache")
	assert.Zero(prod.cache.Size(), "producer.Close should flush the cache")

	prod.Cache(ctx, func(wire.Msg) bool { return true })
	prod.cache.Put(ping0, nil)
	assert.Zero(prod.cache.Size(), "Cache on closed producer should not enable caching")
}

// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package peer

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

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

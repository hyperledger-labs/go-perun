// Copyright 2019 - See NOTICE file for copyright holders.
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

package wire

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	wallettest "perun.network/go-perun/wallet/test"
	ctxtest "polycry.pt/poly-go/context/test"
	"polycry.pt/poly-go/sync"
	"polycry.pt/poly-go/test"
)

func TestProducer(t *testing.T) {
	r0 := NewReceiver()
	r1 := NewReceiver()
	r2 := NewReceiver()
	p := NewRelay()

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

func TestProducer_produce_closed(t *testing.T) {
	var missed *Envelope
	p := NewRelay()
	p.SetDefaultMsgHandler(func(e *Envelope) { missed = e })
	assert.NoError(t, p.Close())
	rng := test.Prng(t)
	a := wallettest.NewRandomAddress(rng)
	b := wallettest.NewRandomAddress(rng)
	p.Put(&Envelope{a, b, NewPingMsg()})
	assert.Nil(t, missed, "produce() on closed producer shouldn't do anything")
}

func TestProducer_SetDefaultMsgHandler(t *testing.T) {
	fn := func(*Envelope) {}
	p := NewRelay()

	logUnhandledMsgPtr := reflect.ValueOf(logUnhandledMsg).Pointer()

	assert.Equal(t, logUnhandledMsgPtr, reflect.ValueOf(p.defaultMsgHandler).Pointer())
	p.SetDefaultMsgHandler(fn)
	assert.Equal(t, reflect.ValueOf(fn).Pointer(), reflect.ValueOf(p.defaultMsgHandler).Pointer())
	p.SetDefaultMsgHandler(nil)
	assert.Equal(t, logUnhandledMsgPtr, reflect.ValueOf(p.defaultMsgHandler).Pointer())
}

func TestProducer_Close(t *testing.T) {
	p := NewRelay()
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
		p := NewRelay()
		p.Close()
		assert.Error(t, p.Subscribe(NewReceiver(), fn))
	})

	t.Run("duplicate", func(t *testing.T) {
		p := NewRelay()
		r := NewReceiver()
		assert.NoError(t, p.Subscribe(r, fn))
		assert.Panics(t, func() { p.Subscribe(r, fn) }) //nolint:errcheck
	})

	t.Run("closed consumer", func(t *testing.T) {
		p := NewRelay()
		r := NewReceiver()
		r.Close()
		ctxtest.AssertTerminates(t, timeout, func() {
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
	prod := NewRelay()
	unhandlesMsg := make([]*Envelope, 0, 2)
	prod.SetDefaultMsgHandler(func(e *Envelope) { unhandlesMsg = append(unhandlesMsg, e) })

	ctx := context.Background()
	prod.Cache(ctx, isPing)

	rng := test.Prng(t)
	ping0 := NewRandomEnvelope(rng, NewPingMsg())
	pong1 := NewRandomEnvelope(rng, NewPongMsg())
	pong2 := NewRandomEnvelope(rng, NewPongMsg())

	prod.Put(ping0)
	assert.Equal(1, prod.cache.Size())
	assert.Len(unhandlesMsg, 0)

	prod.Put(pong1)
	assert.Equal(1, prod.cache.Size())
	assert.Len(unhandlesMsg, 1)

	prod.Cache(ctx, isPong)
	prod.Put(pong2)
	assert.Equal(2, prod.cache.Size())
	assert.Len(unhandlesMsg, 1)

	rec := NewReceiver()
	prod.Subscribe(rec, isPing) //nolint:errcheck
	ctxtest.AssertTerminates(t, timeout, func() {
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
	prod.cache.Put(ping0)
	assert.Zero(prod.cache.Size(), "Cache on closed producer should not enable caching")
}

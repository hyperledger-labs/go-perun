// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package net

import (
	"context"
	"math/rand"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/pkg/sync/atomic"
	"perun.network/go-perun/pkg/test"
	"perun.network/go-perun/wallet"
	wallettest "perun.network/go-perun/wallet/test"
	"perun.network/go-perun/wire"
	wiretest "perun.network/go-perun/wire/test"
)

var _ Dialer = (*mockDialer)(nil)

const timeout = 100 * time.Millisecond

type mockDialer struct {
	dial   chan Conn
	mutex  sync.RWMutex
	closed atomic.Bool
}

func (d *mockDialer) Close() error {
	if !d.closed.TrySet() {
		return errors.New("dialer already closed")
	}
	close(d.dial)
	return nil
}

func (d *mockDialer) Dial(ctx context.Context, addr wire.Address) (Conn, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	select {
	case <-ctx.Done():
		return nil, errors.New("aborted manually")
	case conn := <-d.dial:
		if conn != nil {
			return conn, nil
		}
		return nil, errors.New("dialer closed")
	}
}

func (d *mockDialer) isClosed() bool {
	return d.closed.IsSet()
}

func (d *mockDialer) put(conn Conn) {
	d.dial <- conn
}

func newMockDialer() *mockDialer {
	return &mockDialer{dial: make(chan Conn)}
}

var _ Listener = (*mockListener)(nil)

type mockListener struct {
	dialer mockDialer
}

func (l *mockListener) Accept() (Conn, error) {
	return l.dialer.Dial(context.Background(), nil)
}

func (l *mockListener) Close() error {
	return l.dialer.Close()
}

func (l *mockListener) put(conn Conn) {
	l.dialer.put(conn)
}

func (l *mockListener) isClosed() bool {
	return l.dialer.isClosed()
}

func newMockListener() *mockListener {
	return &mockListener{dialer: mockDialer{dial: make(chan Conn)}}
}

func nilConsumer(wire.Address) wire.Consumer { return nil }

// TestRegistry_Get tests that when calling Get(), existing peers are returned,
// and when unknown peers are requested, a temporary peer is create that is
// dialed in the background. It also tests that the dialing process combines
// with the Listener, so that if a connection to a peer that is still being
// dialed comes in, the peer is assigned that connection.
func TestRegistry_Get(t *testing.T) {
	t.Parallel()
	rng := rand.New(rand.NewSource(0xDDDDDeDe))
	id := wallettest.NewRandomAccount(rng)
	peerID := wallettest.NewRandomAccount(rng)
	peerAddr := peerID.Address()

	t.Run("peer already in progress (existing)", func(t *testing.T) {
		t.Parallel()

		dialer := newMockDialer()
		r := NewEndpointRegistry(id, nilConsumer, dialer)
		existing := newEndpoint(peerAddr, newMockConn(nil))

		r.endpoints[wallet.Key(peerAddr)] = newFullEndpoint(existing)
		test.AssertTerminates(t, timeout, func() {
			p, err := r.Get(context.Background(), peerAddr)
			assert.NoError(t, err)
			assert.Same(t, p, existing)
		})
	})

	t.Run("new peer (failed dial)", func(t *testing.T) {
		t.Parallel()

		dialer := newMockDialer()
		r := NewEndpointRegistry(id, nilConsumer, dialer)

		dialer.Close()
		test.AssertTerminates(t, timeout, func() {
			p, err := r.Get(context.Background(), peerAddr)
			assert.Error(t, err)
			assert.Nil(t, p)
		})

		<-time.After(timeout)

	})

	t.Run("new peer (successful dial)", func(t *testing.T) {
		t.Parallel()

		dialer := newMockDialer()
		r := NewEndpointRegistry(id, nilConsumer, dialer)

		ct := test.NewConcurrent(t)
		a, b := newPipeConnPair()
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		go ct.Stage("receiver", func(t require.TestingT) {
			dialer.put(a)
			ExchangeAddrsPassive(ctx, peerID, b)
			_, err := b.Recv()
			require.NoError(t, err)
		})
		p, err := r.Get(ctx, peerAddr)
		require.NoError(t, err)
		require.NotNil(t, p)
		require.NoError(t, p.Send(ctx, wiretest.NewRandomEnvelope(rng, wire.NewPingMsg())))
		require.NoError(t, p.Close())

		ct.Wait("receiver")
	})
}

func TestRegistry_authenticatedDial(t *testing.T) {
	t.Parallel()
	rng := rand.New(rand.NewSource(0xb0baFEDD))
	id := wallettest.NewRandomAccount(rng)
	d := &mockDialer{dial: make(chan Conn)}
	r := NewEndpointRegistry(id, nilConsumer, d)

	remoteID := wallettest.NewRandomAccount(rng)
	remoteAddr := remoteID.Address()

	t.Run("dial fail", func(t *testing.T) {
		addr := wallettest.NewRandomAddress(rng)
		de, created := r.getOrCreateDialingEndpoint(addr)
		go d.put(nil)
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		e, err := r.authenticatedDial(ctx, addr, de, created)
		assert.Error(t, err)
		assert.Nil(t, e)
	})

	t.Run("dial success, ExchangeAddrs fail", func(t *testing.T) {
		a, b := newPipeConnPair()
		go func() {
			d.put(a)
			b.Recv()
			b.Send(&wire.Envelope{
				Sender:    remoteAddr,
				Recipient: id.Address(),
				Msg:       wire.NewPingMsg()})
		}()
		de, created := r.getOrCreateDialingEndpoint(remoteAddr)
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		e, err := r.authenticatedDial(ctx, remoteAddr, de, created)
		assert.Error(t, err)
		assert.Nil(t, e)
	})

	t.Run("dial success, ExchangeAddrs imposter", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		a, b := newPipeConnPair()
		go func() {
			d.put(a)
			ExchangeAddrsPassive(ctx, wallettest.NewRandomAccount(rng), b)
		}()
		de, created := r.getOrCreateDialingEndpoint(remoteAddr)
		e, err := r.authenticatedDial(ctx, remoteAddr, de, created)
		assert.Error(t, err)
		assert.Nil(t, e)
	})

	t.Run("dial success, ExchangeAddrs success", func(t *testing.T) {
		a, b := newPipeConnPair()
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		go func() {
			d.put(a)
			ExchangeAddrsPassive(ctx, remoteID, b)
		}()
		de, created := r.getOrCreateDialingEndpoint(remoteAddr)
		e, err := r.authenticatedDial(ctx, remoteAddr, de, created)
		assert.NoError(t, err)
		assert.NotNil(t, e)
	})
}

func TestRegistry_setupConn(t *testing.T) {
	t.Parallel()
	rng := rand.New(rand.NewSource(0xb0baFEDD))
	id := wallettest.NewRandomAccount(rng)
	remoteID := wallettest.NewRandomAccount(rng)

	t.Run("ExchangeAddrs fail", func(t *testing.T) {
		d := &mockDialer{dial: make(chan Conn)}
		r := NewEndpointRegistry(id, nilConsumer, d)
		a, b := newPipeConnPair()
		go b.Send(&wire.Envelope{
			Sender:    id.Address(),
			Recipient: remoteID.Address(),
			Msg:       wire.NewPingMsg()})
		test.AssertTerminates(t, timeout, func() {
			assert.Error(t, r.setupConn(a))
		})
	})

	t.Run("ExchangeAddrs success (peer already exists)", func(t *testing.T) {
		d := &mockDialer{dial: make(chan Conn)}
		r := NewEndpointRegistry(id, nilConsumer, d)
		a, b := newPipeConnPair()
		go ExchangeAddrsActive(context.Background(), remoteID, id.Address(), b)

		r.addEndpoint(remoteID.Address(), newMockConn(nil), false)
		test.AssertTerminates(t, timeout, func() {
			assert.NoError(t, r.setupConn(a))
		})
	})

	t.Run("ExchangeAddrs success (peer did not exist)", func(t *testing.T) {
		d := &mockDialer{dial: make(chan Conn)}
		r := NewEndpointRegistry(id, nilConsumer, d)
		a, b := newPipeConnPair()
		go ExchangeAddrsActive(context.Background(), remoteID, id.Address(), b)

		test.AssertTerminates(t, timeout, func() {
			assert.NoError(t, r.setupConn(a))
		})
	})
}

func TestRegistry_Listen(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	rng := rand.New(rand.NewSource(0xDDDDDeDe))

	id := wallettest.NewRandomAccount(rng)
	addr := id.Address()
	remoteID := wallettest.NewRandomAccount(rng)
	remoteAddr := remoteID.Address()

	d := newMockDialer()
	l := newMockListener()
	r := NewEndpointRegistry(id, nilConsumer, d)

	go func() {
		// Listen() will only terminate if the listener is closed.
		test.AssertTerminates(t, 2*timeout, func() { r.Listen(l) })
	}()

	a, b := newPipeConnPair()
	l.put(a)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	err := ExchangeAddrsActive(ctx, remoteID, addr, b)
	require.NoError(t, err)

	<-time.After(timeout)
	assert.True(r.Has(remoteAddr))

	assert.NoError(r.Close())
	assert.True(l.isClosed(), "closing the registry should close the listener")

	l2 := newMockListener()
	test.AssertTerminates(t, timeout, func() {
		r.Listen(l2)
		assert.True(l2.isClosed(),
			"Listen on closed registry should close the listener immediately")
	})
}

// TestRegistry_addEndpoint tests that addEndpoint() calls the Registry's subscription
// function. Other aspects of the function are already tested in other tests.
func TestRegistry_addEndpoint_Subscribe(t *testing.T) {
	t.Parallel()
	rng := rand.New(rand.NewSource(0xDDDDDeDe))
	called := false
	r := NewEndpointRegistry(wallettest.NewRandomAccount(rng), func(wire.Address) wire.Consumer { called = true; return nil }, nil)

	assert.False(t, called, "onNewEndpoint must not have been called yet")
	r.addEndpoint(wallettest.NewRandomAddress(rng), newMockConn(nil), false)
	assert.True(t, called, "onNewEndpoint must have been called")
}

func TestRegistry_Close(t *testing.T) {
	t.Parallel()

	rng := rand.New(rand.NewSource(0xb0baFEDD))

	t.Run("double close error", func(t *testing.T) {
		r := NewEndpointRegistry(wallettest.NewRandomAccount(rng), nilConsumer, nil)
		r.Close()
		assert.Error(t, r.Close())
	})

	t.Run("dialer close error", func(t *testing.T) {
		d := &mockDialer{dial: make(chan Conn)}
		d.Close()
		r := NewEndpointRegistry(wallettest.NewRandomAccount(rng), nilConsumer, d)

		assert.Error(t, r.Close())
	})
}

// newPipeConnPair creates endpoints that are connected via pipes.
func newPipeConnPair() (a Conn, b Conn) {
	c0, c1 := net.Pipe()
	return NewIoConn(c0), NewIoConn(c1)
}

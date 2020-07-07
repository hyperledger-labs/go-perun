// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package wire

import (
	"context"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/pkg/sync/atomic"
	"perun.network/go-perun/pkg/test"
	wallettest "perun.network/go-perun/wallet/test"
)

var _ Dialer = (*mockDialer)(nil)

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

func (d *mockDialer) Dial(ctx context.Context, addr Address) (Conn, error) {
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

	t.Run("peer already in progress (nonexisting)", func(t *testing.T) {
		t.Parallel()

		dialer := newMockDialer()
		r := NewEndpointRegistry(id, func(*Endpoint) {}, dialer)
		closed := newEndpoint(peerAddr, nil, nil)

		r.peers = []*Endpoint{closed}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		p, err := r.Get(ctx, peerAddr)
		assert.Error(t, err)
		assert.Nil(t, p)
	})

	t.Run("peer already in progress (existing)", func(t *testing.T) {
		t.Parallel()

		dialer := newMockDialer()
		r := NewEndpointRegistry(id, func(*Endpoint) {}, dialer)
		existing := newEndpoint(peerAddr, newMockConn(nil), nil)

		r.peers = []*Endpoint{existing}
		test.AssertTerminates(t, timeout, func() {
			p, err := r.Get(context.Background(), peerAddr)
			assert.NoError(t, err)
			assert.Same(t, p, existing)
		})
	})

	t.Run("new peer (failed dial)", func(t *testing.T) {
		t.Parallel()

		dialer := newMockDialer()
		r := NewEndpointRegistry(id, func(*Endpoint) {}, dialer)

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
		r := NewEndpointRegistry(id, func(*Endpoint) {}, dialer)

		a, b := newPipeConnPair()
		go func() {
			dialer.put(a)
			ExchangeAddrs(context.Background(), peerID, b)
		}()
		ct := test.NewConcurrent(t)
		test.AssertTerminates(t, timeout, func() {
			ct.Stage("terminates", func(t require.TestingT) {
				p, err := r.Get(context.Background(), peerAddr)
				require.NoError(t, err)
				require.NotNil(t, p)
				require.True(t, p.exists())
				require.False(t, p.IsClosed())
			})
		})

		ct.Wait("terminates")
	})
}

func TestRegistry_authenticatedDial(t *testing.T) {
	t.Parallel()
	rng := rand.New(rand.NewSource(0xb0baFEDD))
	id := wallettest.NewRandomAccount(rng)
	d := &mockDialer{dial: make(chan Conn)}
	r := NewEndpointRegistry(id, func(*Endpoint) {}, d)

	remoteID := wallettest.NewRandomAccount(rng)
	remoteAddr := remoteID.Address()

	t.Run("dial fail, existing peer", func(t *testing.T) {
		p := newEndpoint(nil, nil, nil)
		p.create(newMockConn(nil))
		go d.put(nil)
		test.AssertTerminates(t, timeout, func() {
			err := r.authenticatedDial(context.Background(), p, wallettest.NewRandomAddress(rng))
			assert.NoError(t, err)
		})
	})

	t.Run("dial fail, nonexisting peer", func(t *testing.T) {
		p := newEndpoint(nil, nil, nil)
		go d.put(nil)
		test.AssertTerminates(t, timeout, func() {
			err := r.authenticatedDial(context.Background(), p, wallettest.NewRandomAddress(rng))
			assert.Error(t, err)
		})
	})

	t.Run("dial success, ExchangeAddrs fail, nonexisting peer", func(t *testing.T) {
		p := newEndpoint(nil, nil, nil)
		a, b := newPipeConnPair()
		go d.put(a)
		go b.Send(NewPingMsg())
		test.AssertTerminates(t, timeout, func() {
			err := r.authenticatedDial(context.Background(), p, remoteAddr)
			assert.Error(t, err)
		})
	})

	t.Run("dial success, ExchangeAddrs fail, existing peer", func(t *testing.T) {
		p := newEndpoint(nil, newMockConn(nil), nil)
		a, b := newPipeConnPair()
		go d.put(a)
		go b.Send(NewPingMsg())
		test.AssertTerminates(t, timeout, func() {
			err := r.authenticatedDial(context.Background(), p, remoteAddr)
			assert.Nil(t, err)
		})
	})

	t.Run("dial success, ExchangeAddrs imposter, nonexisting peer", func(t *testing.T) {
		p := newEndpoint(nil, nil, nil)
		a, b := newPipeConnPair()
		go d.put(a)
		go ExchangeAddrs(context.Background(), wallettest.NewRandomAccount(rng), b)
		test.AssertTerminates(t, timeout, func() {
			err := r.authenticatedDial(context.Background(), p, remoteAddr)
			assert.Error(t, err)
		})
	})

	t.Run("dial success, ExchangeAddrs imposter, existing peer", func(t *testing.T) {
		p := newEndpoint(nil, newMockConn(nil), nil)
		a, b := newPipeConnPair()
		go d.put(a)
		go ExchangeAddrs(context.Background(), wallettest.NewRandomAccount(rng), b)
		test.AssertTerminates(t, timeout, func() {
			err := r.authenticatedDial(context.Background(), p, remoteAddr)
			assert.NoError(t, err)
		})
	})

	t.Run("dial success, ExchangeAddrs success", func(t *testing.T) {
		p := newEndpoint(nil, nil, nil)
		a, b := newPipeConnPair()
		go d.put(a)
		go ExchangeAddrs(context.Background(), remoteID, b)
		test.AssertTerminates(t, timeout, func() {
			err := r.authenticatedDial(context.Background(), p, remoteAddr)
			assert.NoError(t, err)
		})
	})
}

func TestRegistry_setupConn(t *testing.T) {
	t.Parallel()
	rng := rand.New(rand.NewSource(0xb0baFEDD))
	id := wallettest.NewRandomAccount(rng)
	remoteID := wallettest.NewRandomAccount(rng)

	t.Run("ExchangeAddrs fail", func(t *testing.T) {
		d := &mockDialer{dial: make(chan Conn)}
		r := NewEndpointRegistry(id, func(*Endpoint) {}, d)
		a, b := newPipeConnPair()
		go b.Send(NewPingMsg())
		test.AssertTerminates(t, timeout, func() {
			assert.Error(t, r.setupConn(a))
		})
	})

	t.Run("ExchangeAddrs success (peer already exists)", func(t *testing.T) {
		d := &mockDialer{dial: make(chan Conn)}
		r := NewEndpointRegistry(id, func(*Endpoint) {}, d)
		a, b := newPipeConnPair()
		go ExchangeAddrs(context.Background(), remoteID, b)

		r.addPeer(remoteID.Address(), nil)
		test.AssertTerminates(t, timeout, func() {
			assert.NoError(t, r.setupConn(a))
		})
	})

	t.Run("ExchangeAddrs success (peer did not exist)", func(t *testing.T) {
		d := &mockDialer{dial: make(chan Conn)}
		r := NewEndpointRegistry(id, func(*Endpoint) {}, d)
		a, b := newPipeConnPair()
		go ExchangeAddrs(context.Background(), remoteID, b)

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
	r := NewEndpointRegistry(id, func(*Endpoint) {}, d)

	go func() {
		// Listen() will only terminate if the listener is closed.
		test.AssertTerminates(t, 2*timeout, func() { r.Listen(l) })
	}()

	a, b := newPipeConnPair()
	l.put(a)
	ct := test.NewConcurrent(t)
	test.AssertTerminates(t, timeout, func() {
		ct.Stage("terminates", func(t require.TestingT) {
			address, err := ExchangeAddrs(context.Background(), remoteID, b)
			require.NoError(t, err)
			assert.True(address.Equals(addr))
		})
	})
	ct.Wait("terminates")

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

// TestRegistry_addPeer tests that addPeer() calls the Registry's subscription
// function. Other aspects of the function are already tested in other tests.
func TestRegistry_addPeer_Subscribe(t *testing.T) {
	t.Parallel()
	rng := rand.New(rand.NewSource(0xDDDDDeDe))
	called := false
	r := NewEndpointRegistry(wallettest.NewRandomAccount(rng), func(*Endpoint) { called = true }, nil)

	assert.False(t, called, "subscription must not have been called yet")
	r.addPeer(nil, nil)
	assert.True(t, called, "subscription must have been called")
}

func TestRegistry_delete(t *testing.T) {
	t.Parallel()
	rng := rand.New(rand.NewSource(0xb0baFEDD))
	d := &mockDialer{dial: make(chan Conn)}
	r := NewEndpointRegistry(wallettest.NewRandomAccount(rng), func(*Endpoint) {}, d)

	id := wallettest.NewRandomAccount(rng)
	addr := id.Address()
	assert.Equal(t, 0, r.NumPeers())
	p := newEndpoint(addr, nil, nil)
	r.peers = []*Endpoint{p}
	assert.True(t, r.Has(addr))
	p2, _ := r.find(addr)
	assert.Equal(t, p, p2)

	r.delete(p2)
	assert.Equal(t, 0, r.NumPeers())
	assert.False(t, r.Has(addr))
	p2, _ = r.find(addr)
	assert.Nil(t, p2)

	assert.Panics(t, func() { r.delete(p) }, "double delete must panic")
}

func TestRegistry_Close(t *testing.T) {
	t.Parallel()

	rng := rand.New(rand.NewSource(0xb0baFEDD))

	t.Run("double close error", func(t *testing.T) {
		r := NewEndpointRegistry(wallettest.NewRandomAccount(rng), func(*Endpoint) {}, nil)
		r.Close()
		assert.Error(t, r.Close())
	})

	t.Run("peer close error", func(t *testing.T) {
		d := &mockDialer{dial: make(chan Conn)}
		r := NewEndpointRegistry(wallettest.NewRandomAccount(rng), func(*Endpoint) {}, d)

		mc := newMockConn(nil)
		p := newEndpoint(nil, mc, nil)
		// we close the MockConn so that a second Close() call returns an error.
		// Note that the MockConn doesn't return an AlreadyClosedError on a
		// double-close.
		mc.Close()
		r.peers = append(r.peers, p)
		assert.Error(t, r.Close(),
			"a close error from a peer should be propagated to Registry.Close()")
	})

	t.Run("dialer close error", func(t *testing.T) {
		d := &mockDialer{dial: make(chan Conn)}
		d.Close()
		r := NewEndpointRegistry(wallettest.NewRandomAccount(rng), func(*Endpoint) {}, d)

		assert.Error(t, r.Close())
	})
}

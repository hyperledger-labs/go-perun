// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package peer

import (
	"context"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/backend/sim/wallet"
	"perun.network/go-perun/pkg/sync/atomic"
	"perun.network/go-perun/pkg/test"
	"perun.network/go-perun/wire/msg"
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
		} else {
			return nil, errors.New("dialer closed")
		}
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
	id := wallet.NewRandomAccount(rng)
	peerId := wallet.NewRandomAccount(rng)
	peerAddr := peerId.Address()

	t.Run("peer already in progress (closed)", func(t *testing.T) {
		t.Parallel()

		dialer := newMockDialer()
		r := NewRegistry(id, func(*Peer) {}, dialer)
		closed := newPeer(peerAddr, nil, nil)
		closed.Close()

		r.peers = []*Peer{closed}
		test.AssertTerminates(t, timeout, func() {
			assert.Nil(t, r.Get(context.Background(), peerAddr))
		})
	})

	t.Run("peer already in progress (existing)", func(t *testing.T) {
		t.Parallel()

		dialer := newMockDialer()
		r := NewRegistry(id, func(*Peer) {}, dialer)
		existing := newPeer(peerAddr, newMockConn(nil), nil)

		r.peers = []*Peer{existing}
		test.AssertTerminates(t, timeout, func() {
			assert.Same(t, r.Get(context.Background(), peerAddr), existing)
		})
	})

	t.Run("new peer (failed dial)", func(t *testing.T) {
		t.Parallel()

		dialer := newMockDialer()
		r := NewRegistry(id, func(*Peer) {}, dialer)

		dialer.Close()
		test.AssertTerminates(t, timeout, func() {
			assert.Nil(t, r.Get(context.Background(), peerAddr))
		})

		<-time.After(timeout)

	})

	t.Run("new peer (successful dial)", func(t *testing.T) {
		t.Parallel()

		dialer := newMockDialer()
		r := NewRegistry(id, func(*Peer) {}, dialer)

		a, b := newPipeConnPair()
		go func() {
			dialer.put(a)
			ExchangeAddrs(context.Background(), peerId, b)
		}()
		test.AssertTerminates(t, timeout, func() {
			p := r.Get(context.Background(), peerAddr)
			require.NotNil(t, p)
			require.True(t, p.exists())
			require.False(t, p.IsClosed())
		})
	})
}

func TestRegistry_authenticatedDial(t *testing.T) {
	t.Parallel()
	rng := rand.New(rand.NewSource(0xb0baFEDD))
	id := wallet.NewRandomAccount(rng)
	d := &mockDialer{dial: make(chan Conn)}
	r := NewRegistry(id, func(*Peer) {}, d)

	remoteId := wallet.NewRandomAccount(rng)
	remoteAddr := remoteId.Address()

	t.Run("dial fail, existing peer", func(t *testing.T) {
		p := newPeer(nil, nil, nil)
		p.create(newMockConn(nil))
		go d.put(nil)
		test.AssertTerminates(t, timeout, func() {
			err := r.authenticatedDial(context.Background(), p, wallet.NewRandomAddress(rng))
			assert.NoError(t, err)
		})
	})

	t.Run("dial fail, nonexisting peer", func(t *testing.T) {
		p := newPeer(nil, nil, nil)
		go d.put(nil)
		test.AssertTerminates(t, timeout, func() {
			err := r.authenticatedDial(context.Background(), p, wallet.NewRandomAddress(rng))
			assert.Error(t, err)
		})
	})

	t.Run("dial success, ExchangeAddrs fail, nonexisting peer", func(t *testing.T) {
		p := newPeer(nil, nil, nil)
		a, b := newPipeConnPair()
		go d.put(a)
		go b.Send(msg.NewPingMsg())
		test.AssertTerminates(t, timeout, func() {
			err := r.authenticatedDial(context.Background(), p, remoteAddr)
			assert.Error(t, err)
		})
	})

	t.Run("dial success, ExchangeAddrs fail, existing peer", func(t *testing.T) {
		p := newPeer(nil, newMockConn(nil), nil)
		a, b := newPipeConnPair()
		go d.put(a)
		go b.Send(msg.NewPingMsg())
		test.AssertTerminates(t, timeout, func() {
			err := r.authenticatedDial(context.Background(), p, remoteAddr)
			assert.Nil(t, err)
		})
	})

	t.Run("dial success, ExchangeAddrs imposter, nonexisting peer", func(t *testing.T) {
		p := newPeer(nil, nil, nil)
		a, b := newPipeConnPair()
		go d.put(a)
		go ExchangeAddrs(context.Background(), wallet.NewRandomAccount(rng), b)
		test.AssertTerminates(t, timeout, func() {
			err := r.authenticatedDial(context.Background(), p, remoteAddr)
			assert.Error(t, err)
		})
	})

	t.Run("dial success, ExchangeAddrs imposter, existing peer", func(t *testing.T) {
		p := newPeer(nil, newMockConn(nil), nil)
		a, b := newPipeConnPair()
		go d.put(a)
		go ExchangeAddrs(context.Background(), wallet.NewRandomAccount(rng), b)
		test.AssertTerminates(t, timeout, func() {
			err := r.authenticatedDial(context.Background(), p, remoteAddr)
			assert.NoError(t, err)
		})
	})

	t.Run("dial success, ExchangeAddrs success", func(t *testing.T) {
		p := newPeer(nil, nil, nil)
		a, b := newPipeConnPair()
		go d.put(a)
		go ExchangeAddrs(context.Background(), remoteId, b)
		test.AssertTerminates(t, timeout, func() {
			err := r.authenticatedDial(context.Background(), p, remoteAddr)
			assert.NoError(t, err)
		})
	})
}

func TestRegistry_setupConn(t *testing.T) {
	t.Parallel()
	rng := rand.New(rand.NewSource(0xb0baFEDD))
	id := wallet.NewRandomAccount(rng)
	remoteId := wallet.NewRandomAccount(rng)

	t.Run("ExchangeAddrs fail", func(t *testing.T) {
		d := &mockDialer{dial: make(chan Conn)}
		r := NewRegistry(id, func(*Peer) {}, d)
		a, b := newPipeConnPair()
		go b.Send(msg.NewPingMsg())
		test.AssertTerminates(t, timeout, func() {
			assert.Error(t, r.setupConn(a))
		})
	})

	t.Run("ExchangeAddrs success (peer already exists)", func(t *testing.T) {
		d := &mockDialer{dial: make(chan Conn)}
		r := NewRegistry(id, func(*Peer) {}, d)
		a, b := newPipeConnPair()
		go ExchangeAddrs(context.Background(), remoteId, b)

		r.addPeer(remoteId.Address(), nil)
		test.AssertTerminates(t, timeout, func() {
			assert.NoError(t, r.setupConn(a))
		})
	})

	t.Run("ExchangeAddrs success (peer did not exist)", func(t *testing.T) {
		d := &mockDialer{dial: make(chan Conn)}
		r := NewRegistry(id, func(*Peer) {}, d)
		a, b := newPipeConnPair()
		go ExchangeAddrs(context.Background(), remoteId, b)

		test.AssertTerminates(t, timeout, func() {
			assert.NoError(t, r.setupConn(a))
		})
	})
}

func TestRegistry_Listen(t *testing.T) {
	t.Parallel()
	assert, require := assert.New(t), require.New(t)

	rng := rand.New(rand.NewSource(0xDDDDDeDe))

	id := wallet.NewRandomAccount(rng)
	addr := id.Address()
	remoteId := wallet.NewRandomAccount(rng)
	remoteAddr := remoteId.Address()

	d := newMockDialer()
	l := newMockListener()
	r := NewRegistry(id, func(*Peer) {}, d)

	go t.Run("listen", func(t *testing.T) {
		// Listen() will only terminate if the listener is closed.
		test.AssertTerminates(t, timeout, func() { r.Listen(l) })
	})

	a, b := newPipeConnPair()
	l.put(a)
	test.AssertTerminates(t, timeout, func() {
		address, err := ExchangeAddrs(context.Background(), remoteId, b)
		require.NoError(err)
		assert.True(address.Equals(addr))
	})

	l.Close()

	<-time.After(timeout)

	assert.True(r.Has(remoteAddr))
}

// TestRegistry_addPeer tests that addPeer() calls the Registry's subscription
// function. Other aspects of the function are already tested in other tests.
func TestRegistry_addPeer_Subscribe(t *testing.T) {
	rng := rand.New(rand.NewSource(0xDDDDDeDe))
	called := false
	r := NewRegistry(wallet.NewRandomAccount(rng), func(*Peer) { called = true }, nil)

	assert.False(t, called, "subscription must not have been called yet")
	r.addPeer(nil, nil)
	assert.True(t, called, "subscription must have been called")
}

func TestRegistry_delete(t *testing.T) {
	t.Parallel()
	rng := rand.New(rand.NewSource(0xb0baFEDD))
	d := &mockDialer{dial: make(chan Conn)}
	r := NewRegistry(wallet.NewRandomAccount(rng), func(*Peer) {}, d)

	id := wallet.NewRandomAccount(rng)
	addr := id.Address()
	assert.Equal(t, 0, r.NumPeers())
	p := newPeer(addr, nil, nil)
	r.peers = []*Peer{p}
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
		r := NewRegistry(wallet.NewRandomAccount(rng), func(*Peer) {}, nil)
		r.Close()
		assert.Error(t, r.Close())
	})

	t.Run("peer close error", func(t *testing.T) {
		d := &mockDialer{dial: make(chan Conn)}
		r := NewRegistry(wallet.NewRandomAccount(rng), func(*Peer) {}, d)

		p := newPeer(nil, nil, nil)
		p.Close()
		r.peers = append(r.peers, p)
		assert.Error(t, r.Close())
	})

	t.Run("dialer close error", func(t *testing.T) {
		d := &mockDialer{dial: make(chan Conn)}
		d.Close()
		r := NewRegistry(wallet.NewRandomAccount(rng), func(*Peer) {}, d)

		assert.Error(t, r.Close())
	})
}

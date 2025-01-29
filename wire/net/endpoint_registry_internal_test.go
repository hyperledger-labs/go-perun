// Copyright 2025 - See NOTICE file for copyright holders.
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

package net

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"perun.network/go-perun/channel"

	"perun.network/go-perun/wallet"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/wire"
	perunio "perun.network/go-perun/wire/perunio/serializer"
	wiretest "perun.network/go-perun/wire/test"
	ctxtest "polycry.pt/poly-go/context/test"
	"polycry.pt/poly-go/sync/atomic"
	"polycry.pt/poly-go/test"
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

func (d *mockDialer) Dial(ctx context.Context, addr map[wallet.BackendID]wire.Address, _ wire.EnvelopeSerializer) (Conn, error) {
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

func (l *mockListener) Accept(ser wire.EnvelopeSerializer) (Conn, error) {
	return l.dialer.Dial(context.Background(), nil, ser)
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

func nilConsumer(map[wallet.BackendID]wire.Address) wire.Consumer { return nil }

// TestRegistry_Get tests that when calling Get(), existing peers are returned,
// and when unknown peers are requested, a temporary peer is create that is
// dialed in the background. It also tests that the dialing process combines
// with the Listener, so that if a connection to a peer that is still being
// dialed comes in, the peer is assigned that connection.
func TestRegistry_Get(t *testing.T) {
	t.Parallel()
	rng := test.Prng(t)
	id := wiretest.NewRandomAccountMap(rng, channel.TestBackendID)
	peerID := wiretest.NewRandomAccountMap(rng, channel.TestBackendID)
	peerAddr := wire.AddressMapfromAccountMap(peerID)

	t.Run("peer already in progress (existing)", func(t *testing.T) {
		t.Parallel()

		dialer := newMockDialer()
		r := NewEndpointRegistry(id, nilConsumer, dialer, perunio.Serializer())
		existing := newEndpoint(peerAddr, newMockConn())

		r.endpoints[wire.Keys(peerAddr)] = newFullEndpoint(existing)
		ctxtest.AssertTerminates(t, timeout, func() {
			p, err := r.Endpoint(context.Background(), peerAddr)
			assert.NoError(t, err)
			assert.Same(t, p, existing)
		})
	})

	t.Run("new peer (failed dial)", func(t *testing.T) {
		t.Parallel()

		dialer := newMockDialer()
		r := NewEndpointRegistry(id, nilConsumer, dialer, perunio.Serializer())

		dialer.Close()
		ctxtest.AssertTerminates(t, timeout, func() {
			p, err := r.Endpoint(context.Background(), peerAddr)
			assert.Error(t, err)
			assert.Nil(t, p)
		})

		<-time.After(timeout)
	})

	t.Run("new peer (successful dial)", func(t *testing.T) {
		t.Parallel()

		dialer := newMockDialer()
		r := NewEndpointRegistry(id, nilConsumer, dialer, perunio.Serializer())

		ct := test.NewConcurrent(t)
		a, b := newPipeConnPair()
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		go ct.Stage("receiver", func(t test.ConcT) {
			dialer.put(a)
			_, err := ExchangeAddrsPassive(ctx, peerID, b)
			require.NoError(t, err)
			_, err = b.Recv()
			require.NoError(t, err)
		})
		p, err := r.Endpoint(ctx, peerAddr)
		require.NoError(t, err)
		require.NotNil(t, p)
		require.NoError(t, p.Send(ctx, wiretest.NewRandomEnvelope(rng, wire.NewPingMsg())))
		require.NoError(t, p.Close())

		ct.Wait("receiver")
	})
}

func TestRegistry_authenticatedDial(t *testing.T) {
	t.Parallel()
	rng := test.Prng(t)
	id := wiretest.NewRandomAccountMap(rng, channel.TestBackendID)
	d := &mockDialer{dial: make(chan Conn)}
	r := NewEndpointRegistry(id, nilConsumer, d, perunio.Serializer())

	remoteID := wiretest.NewRandomAccountMap(rng, channel.TestBackendID)
	remoteAddr := wire.AddressMapfromAccountMap(remoteID)

	t.Run("dial fail", func(t *testing.T) {
		addr := wiretest.NewRandomAddress(rng)
		de, created := r.dialingEndpoint(addr)
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
			if _, err := b.Recv(); err != nil {
				panic(err)
			}
			err := b.Send(&wire.Envelope{
				Sender:    remoteAddr,
				Recipient: wire.AddressMapfromAccountMap(id),
				Msg:       wire.NewPingMsg(),
			})
			if err != nil {
				panic(err)
			}
		}()
		de, created := r.dialingEndpoint(remoteAddr)
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		e, err := r.authenticatedDial(ctx, remoteAddr, de, created)
		assert.Error(t, err)
		assert.Nil(t, e)
	})

	t.Run("dial success, ExchangeAddrs imposter", func(t *testing.T) {
		ct := test.NewConcurrent(t)
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		a, b := newPipeConnPair()
		go ct.Stage("passive", func(rt test.ConcT) {
			d.put(a)
			_, err := ExchangeAddrsPassive(ctx, wiretest.NewRandomAccountMap(rng, channel.TestBackendID), b)
			require.True(rt, IsAuthenticationError(err))
		})
		de, created := r.dialingEndpoint(remoteAddr)
		e, err := r.authenticatedDial(ctx, remoteAddr, de, created)
		assert.Error(t, err)
		assert.Nil(t, e)
		ct.Wait("passive")
	})

	t.Run("dial success, ExchangeAddrs success", func(t *testing.T) {
		a, b := newPipeConnPair()
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			d.put(a)
			_, err := ExchangeAddrsPassive(ctx, remoteID, b)
			if err != nil {
				panic(err)
			}
		}()
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		de, created := r.dialingEndpoint(remoteAddr)
		e, err := r.authenticatedDial(ctx, remoteAddr, de, created)
		assert.NoError(t, err)
		assert.NotNil(t, e)
	})
}

func TestRegistry_setupConn(t *testing.T) {
	t.Parallel()
	rng := test.Prng(t)
	id := wiretest.NewRandomAccountMap(rng, channel.TestBackendID)
	remoteID := wiretest.NewRandomAccountMap(rng, channel.TestBackendID)

	t.Run("ExchangeAddrs fail", func(t *testing.T) {
		d := &mockDialer{dial: make(chan Conn)}
		r := NewEndpointRegistry(id, nilConsumer, d, perunio.Serializer())
		a, b := newPipeConnPair()
		go func() {
			err := b.Send(&wire.Envelope{
				Sender:    wire.AddressMapfromAccountMap(id),
				Recipient: wire.AddressMapfromAccountMap(remoteID),
				Msg:       wire.NewPingMsg(),
			})
			if err != nil {
				panic(err)
			}
		}()
		ctxtest.AssertTerminates(t, timeout, func() {
			assert.Error(t, r.setupConn(a))
		})
	})

	t.Run("ExchangeAddrs success (peer already exists)", func(t *testing.T) {
		d := &mockDialer{dial: make(chan Conn)}
		r := NewEndpointRegistry(id, nilConsumer, d, perunio.Serializer())
		a, b := newPipeConnPair()
		go func() {
			err := ExchangeAddrsActive(context.Background(), remoteID, wire.AddressMapfromAccountMap(id), b)
			if err != nil {
				panic(err)
			}
		}()

		r.addEndpoint(wire.AddressMapfromAccountMap(remoteID), newMockConn(), false)
		ctxtest.AssertTerminates(t, timeout, func() {
			assert.NoError(t, r.setupConn(a))
		})
	})

	t.Run("ExchangeAddrs success (peer did not exist)", func(t *testing.T) {
		d := &mockDialer{dial: make(chan Conn)}
		r := NewEndpointRegistry(id, nilConsumer, d, perunio.Serializer())
		a, b := newPipeConnPair()
		go func() {
			err := ExchangeAddrsActive(context.Background(), remoteID, wire.AddressMapfromAccountMap(id), b)
			if err != nil {
				panic(err)
			}
		}()

		ctxtest.AssertTerminates(t, timeout, func() {
			assert.NoError(t, r.setupConn(a))
		})
	})
}

func TestRegistry_Listen(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	rng := test.Prng(t)

	id := wiretest.NewRandomAccountMap(rng, channel.TestBackendID)
	addr := wire.AddressMapfromAccountMap(id)
	remoteID := wiretest.NewRandomAccountMap(rng, channel.TestBackendID)
	remoteAddr := wire.AddressMapfromAccountMap(remoteID)

	d := newMockDialer()
	l := newMockListener()
	r := NewEndpointRegistry(id, nilConsumer, d, perunio.Serializer())

	go func() {
		// Listen() will only terminate if the listener is closed.
		ctxtest.AssertTerminates(t, 2*timeout, func() { r.Listen(l) })
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
	ctxtest.AssertTerminates(t, timeout, func() {
		r.Listen(l2)
		assert.True(l2.isClosed(),
			"Listen on closed registry should close the listener immediately")
	})
}

// TestRegistry_addEndpoint tests that addEndpoint() calls the Registry's subscription
// function. Other aspects of the function are already tested in other tests.
func TestRegistry_addEndpoint_Subscribe(t *testing.T) {
	t.Parallel()
	rng := test.Prng(t)
	called := false
	r := NewEndpointRegistry(
		wiretest.NewRandomAccountMap(rng, channel.TestBackendID),
		func(map[wallet.BackendID]wire.Address) wire.Consumer { called = true; return nil },
		nil,
		perunio.Serializer(),
	)

	assert.False(t, called, "onNewEndpoint must not have been called yet")
	r.addEndpoint(wiretest.NewRandomAddress(rng), newMockConn(), false)
	assert.True(t, called, "onNewEndpoint must have been called")
}

func TestRegistry_Close(t *testing.T) {
	t.Parallel()

	rng := test.Prng(t)

	t.Run("double close error", func(t *testing.T) {
		r := NewEndpointRegistry(
			wiretest.NewRandomAccountMap(rng, channel.TestBackendID),
			nilConsumer,
			nil,
			perunio.Serializer(),
		)
		r.Close()
		assert.Error(t, r.Close())
	})

	t.Run("dialer close error", func(t *testing.T) {
		d := &mockDialer{dial: make(chan Conn)}
		d.Close()
		r := NewEndpointRegistry(
			wiretest.NewRandomAccountMap(rng, channel.TestBackendID),
			nilConsumer,
			d,
			perunio.Serializer(),
		)

		assert.Error(t, r.Close())
	})
}

// newPipeConnPair creates endpoints that are connected via pipes.
func newPipeConnPair() (a Conn, b Conn) {
	c0, c1 := net.Pipe()
	ser := perunio.Serializer()
	return NewIoConn(c0, ser), NewIoConn(c1, ser)
}

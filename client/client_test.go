// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package client

import (
	"context"
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/peer"
	peertest "perun.network/go-perun/peer/test"
	perunsync "perun.network/go-perun/pkg/sync"
	"perun.network/go-perun/pkg/test"
	wtest "perun.network/go-perun/wallet/test"
	wire "perun.network/go-perun/wire/msg"
)

const timeout = 5 * time.Second

type DummyDialer struct {
	t *testing.T
}

func (d DummyDialer) Dial(ctx context.Context, addr peer.Address) (peer.Conn, error) {
	d.t.Fatal("BUG: DummyDialer.Dial called")
	return nil, errors.New("BUG")
}

func (d DummyDialer) Close() error {
	return nil
}

type DummyListener struct {
	perunsync.Closer
	t *testing.T
}

func NewDummyListener(t *testing.T) *DummyListener {
	return &DummyListener{t: t}
}

func (d *DummyListener) Accept() (peer.Conn, error) {
	<-d.Closed()
	return nil, errors.New("EOF")
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
	id := wtest.NewRandomAccount(rng)
	d, f, a, w := &DummyDialer{t}, &DummyFunder{t}, &DummyAdjudicator{t}, wtest.RandomWallet()
	assert.Panics(t, func() { New(nil, d, f, a, w) })
	assert.Panics(t, func() { New(id, nil, f, a, w) })
	assert.Panics(t, func() { New(id, d, nil, a, w) })
	assert.Panics(t, func() { New(id, d, f, nil, w) })
	assert.Panics(t, func() { New(id, d, f, a, nil) })
}

func TestClient_Listen_NilArgs(t *testing.T) {
	rng := rand.New(rand.NewSource(0x20200108))
	id := wtest.NewRandomAccount(rng)
	c := New(id, &DummyDialer{t}, &DummyFunder{t}, &DummyAdjudicator{t}, wtest.RandomWallet())

	assert.Panics(t, func() { c.Listen(nil) })
}

func TestClient_New(t *testing.T) {
	rng := rand.New(rand.NewSource(0x1a2b3c))
	id := wtest.NewRandomAccount(rng)
	c := New(id, &DummyDialer{t}, &DummyFunder{t}, &DummyAdjudicator{t}, wtest.RandomWallet())

	require.NotNil(t, c)
	assert.NotNil(t, c.peers)
}

func TestClient_NewAndListen_ListenerClose(t *testing.T) {
	require := require.New(t)
	ass := assert.New(t)

	rng := rand.New(rand.NewSource(0x1a2b3c))
	id := wtest.NewRandomAccount(rng)
	c := New(id, &DummyDialer{t}, &DummyFunder{t}, &DummyAdjudicator{t}, wtest.RandomWallet())

	require.NotNil(c)

	listener := NewDummyListener(t)
	numGoroutines := runtime.NumGoroutine()

	go c.Listen(listener)

	ass.NoError(c.Close())

	select {
	case <-listener.Closed():
	case <-time.After(1 * time.Second):
		t.Error("Listener was not closed within 1s")
	}

	// wait for listener goroutine to terminate (it may be put to sleep by the
	// scheduler after closing the channel)
	test.Within100ms.Eventually(t, func(t test.T) {
		assert.Equal(t, numGoroutines, runtime.NumGoroutine())
	})
}

func TestClient_NewAndListen(t *testing.T) {
	require := require.New(t)
	ass := assert.New(t)

	rng := rand.New(rand.NewSource(0x1))
	connHub := new(peertest.ConnHub)
	c := New(wtest.NewRandomAccount(rng), &DummyDialer{t}, &DummyFunder{t}, &DummyAdjudicator{t}, wtest.RandomWallet())
	// initialize the listener instance in the main goroutine
	// if it is initialized in a goroutine, the goroutine may be put to sleep
	// and the dialer may complain about a nonexistent listener
	listener := connHub.NewListener(c.id.Address())
	go c.Listen(listener)

	require.Zero(c.peers.NumPeers())
	dialerDone := make(chan struct{})

	go func() {
		defer close(dialerDone)

		dialer := connHub.NewDialer()
		defer dialer.Close()
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		conn, err := dialer.Dial(ctx, c.id.Address())
		ass.NoError(err, "Dialing the Client instance failed")
		ass.NoError(conn.Send(wire.NewPongMsg()))

		for {
			msg, err := conn.Recv()
			if err != nil {
				break
			}
			ass.Equal(wire.AuthResponse, msg.Type())
			authMsg, ok := msg.(*peer.AuthResponseMsg)
			ass.True(ok, "Have a message with type AuthResponse but cast failed")
			ass.Equal(c.id.Address(), authMsg.Address)
		}
	}()

	select {
	case <-time.After(timeout):
		t.Fatal("Authentication exchange timed out")
	case <-dialerDone:
	}

	ass.Zero(c.peers.NumPeers())

	// make a successful connection
	peerID := wtest.NewRandomAccount(rng)
	dialer := connHub.NewDialer()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		conn, err := dialer.Dial(ctx, c.id.Address())
		ass.NoError(err, "Dialing the Client instance failed")
		ass.NoError(conn.Send(peer.NewAuthResponseMsg(peerID)))

		msg, err := conn.Recv()
		ass.NoError(err)
		ass.Equal(wire.AuthResponse, msg.Type())
		authMsg, ok := msg.(*peer.AuthResponseMsg)
		ass.True(ok, "Have a message with type AuthResponse but cast failed")
		ass.Equal(c.id.Address(), authMsg.Address)

		ass.NoError(dialer.Close())
	}()

	select {
	case <-time.After(timeout):
		t.Fatal("Second authentication exchange timed out")
	case <-dialer.Closed():
	}

	// Wait for listener go routine to insert new peer into registry.
	test.Within100ms.Eventually(t, func(t test.T) {
		assert.Equal(t, 1, c.peers.NumPeers())
		assert.True(t, c.peers.Has(peerID.Address()))
	})

	ass.NoError(c.Close())

	select {
	case <-time.After(timeout):
		t.Fatal("Listener not closed within timeout")
	case <-listener.Closed():
	}
}

func TestClient_Multiplexing(t *testing.T) {
	t.Run("1/1", func(t *testing.T) { testClientMultiplexing(t, 1, 1) })
	t.Run("1/1024", func(t *testing.T) { testClientMultiplexing(t, 1, 1024) })
	t.Run("1024/1", func(t *testing.T) { testClientMultiplexing(t, 1024, 1) })
	t.Run("32/32", func(t *testing.T) { testClientMultiplexing(t, 32, 32) })
}

func testClientMultiplexing(
	t *testing.T, numListeners, numDialers int) {
	if !test.Race {
		// Only run tests in parallel when not running a race test - it only
		// supports 8192 concurrent go routines.
		t.Parallel()
	}
	ass := assert.New(t)

	require.Less(t, 0, numListeners)
	require.Less(t, 0, numDialers)

	// the random sleep times are needed to make concurrency-related issues
	// appear more frequently
	// Consequently, the RNG must be seeded externally.

	seed := time.Now().UnixNano()
	rng := rand.New(rand.NewSource(seed))
	connHub := new(peertest.ConnHub)
	listeners := make([]*Client, numListeners)
	dialers := make([]*Client, numDialers)

	t.Logf(
		"testClient_Multiplexing(numListeners=%v, numDialers=%v) seed=%v",
		numListeners, numDialers, seed,
	)

	for i := range listeners {
		i := i
		id := wtest.NewRandomAccount(rng)
		listeners[i] = New(
			id, connHub.NewDialer(), &DummyFunder{t}, &DummyAdjudicator{t}, wtest.RandomWallet())
		go listeners[i].Listen(connHub.NewListener(listeners[i].id.Address()))
	}
	for i := range dialers {
		id := wtest.NewRandomAccount(rng)
		dialers[i] = New(id, connHub.NewDialer(), &DummyFunder{t}, &DummyAdjudicator{t}, wtest.RandomWallet())
	}

	hostBarrier := new(sync.WaitGroup)
	peerBarrier := make(chan struct{})
	// every dialing client connects to every listening client
	numConnections := numListeners * numDialers
	hostBarrier.Add(numConnections)

	// create connections
	for _, d := range dialers {
		for _, l := range listeners {
			sleepTime := time.Duration(rng.Int63n(10) + 1)

			go func(d, l *Client) {
				defer hostBarrier.Done()

				<-peerBarrier
				time.Sleep(sleepTime * time.Millisecond)

				// trigger dialing
				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				defer cancel()
				p, err := d.peers.Get(ctx, l.id.Address())
				ass.NoError(err)
				if err == nil {
					ass.NotNil(p)
				}
			}(d, l)
		}
	}

	close(peerBarrier)
	hostBarrier.Wait()

	// race tests fail with lower eventual success duration because not all
	// Client.Listen routines have added the peer to their registry yet.
	test.Eventually(t, func(t test.T) {
		for _, l := range listeners {
			assert.Equal(t, numDialers, l.peers.NumPeers())
		}
	}, 500*time.Millisecond, 10*time.Millisecond)

	for _, d := range dialers {
		assert.Equal(t, numListeners, d.peers.NumPeers())
	}

	for _, d := range dialers {
		for _, l := range listeners {
			ass.True(d.peers.Has(l.id.Address()))
			ass.True(l.peers.Has(d.id.Address()))
		}
	}

	// close connections
	peerBarrier = make(chan struct{})
	hostBarrier.Add(numConnections)

	for i := 0; i < numDialers; i++ {
		// disconnect numListeners/2 connections from dialer side
		// disconnect numListeners/2 connections from listener side
		xs := rng.Perm(numListeners)

		for k := 0; k < numListeners; k++ {
			sleepTime := time.Duration(rng.Int63n(10) + 1)

			go func(i, k int) {
				defer hostBarrier.Done()

				<-peerBarrier
				time.Sleep(sleepTime * time.Millisecond)

				j := xs[k]
				var peers *peer.Registry
				var addr peer.Address
				if k < numListeners/2 {
					peers = dialers[i].peers
					addr = listeners[j].id.Address()
				} else {
					peers = listeners[j].peers
					addr = dialers[i].id.Address()
				}

				ass.True(peers.Has(addr))
				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				defer cancel()
				p, err := peers.Get(ctx, addr)
				ass.NoError(err)
				if err == nil {
					ass.NotNil(p)
					ass.NoError(p.Close())
				}
			}(i, k)
		}
	}

	close(peerBarrier)
	hostBarrier.Wait()

	test.Within100ms.Eventually(t, func(t test.T) {
		for i, l := range listeners {
			assert.Zerof(t, l.peers.NumPeers(),
				"listener[%d] has unexpected number of peers", i)
		}
	})
	for i, l := range listeners {
		assert.NoErrorf(t, l.Close(), "closing listener[%d]", i)
	}

	test.Within100ms.Eventually(t, func(t test.T) {
		for i, d := range dialers {
			assert.Zero(t, d.peers.NumPeers(),
				"Listener %d has unexpected number of peers", i)
		}
	})
	for i, d := range dialers {
		assert.NoErrorf(t, d.Close(), "closing dialer[%d]", i)
	}
}

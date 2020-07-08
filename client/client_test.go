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
	id := wtest.NewRandomAccount(rng)
	b, f, a, w := &DummyBus{t}, &DummyFunder{t}, &DummyAdjudicator{t}, wtest.RandomWallet()
	assert.Panics(t, func() { New(nil, b, f, a, w) })
	assert.Panics(t, func() { New(id, nil, f, a, w) })
	assert.Panics(t, func() { New(id, b, nil, a, w) })
	assert.Panics(t, func() { New(id, b, f, nil, w) })
	assert.Panics(t, func() { New(id, b, f, a, nil) })
}

func TestClient_Handle_NilArgs(t *testing.T) {
	rng := rand.New(rand.NewSource(20200524))
	id := wtest.NewRandomAccount(rng)
	c, err := New(id, &DummyBus{t}, &DummyFunder{t}, &DummyAdjudicator{t}, wtest.RandomWallet())
	require.NoError(t, err)

	dummyUH := UpdateHandlerFunc(func(ChannelUpdate, *UpdateResponder) {})
	assert.Panics(t, func() { c.Handle(nil, dummyUH) })
	dummyPH := ProposalHandlerFunc(func(*ChannelProposal, *ProposalResponder) {})
	assert.Panics(t, func() { c.Handle(dummyPH, nil) })
}

func TestClient_New(t *testing.T) {
	rng := rand.New(rand.NewSource(0x1a2b3c))
	id := wtest.NewRandomAccount(rng)
	c, err := New(id, &DummyBus{t}, &DummyFunder{t}, &DummyAdjudicator{t}, wtest.RandomWallet())
	assert.NoError(t, err)
	require.NotNil(t, c)
}

/*
// TODO: move into wire/net as EndpointRegistry tests, if test doesn't exist yet.
func _TestClient_NewAndListen(t *testing.T) {
	require := require.New(t)
	ass := assert.New(t)

	rng := rand.New(rand.NewSource(0x1))
	connHub := new(wiretest.ConnHub)
	c := New(wtest.NewRandomAccount(rng), &DummyBus{t}, &DummyFunder{t}, &DummyAdjudicator{t}, wtest.RandomWallet())
	// initialize the listener instance in the main goroutine
	// if it is initialized in a goroutine, the goroutine may be put to sleep
	// and the dialer may complain about a nonexistent listener
	listener := connHub.NewNetListener(c.id.Address())
	go c.Listen(listener)

	require.Zero(c.peers.NumPeers())
	dialerDone := make(chan struct{})

	go func() {
		defer close(dialerDone)

		dialer := connHub.NewNetDialer()
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
			authMsg, ok := msg.(*wire.AuthResponseMsg)
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
	dialer := connHub.NewNetDialer()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		conn, err := dialer.Dial(ctx, c.id.Address())
		ass.NoError(err, "Dialing the Client instance failed")
		ass.NoError(conn.Send(wire.NewAuthResponseMsg(peerID)))

		msg, err := conn.Recv()
		ass.NoError(err)
		ass.Equal(wire.AuthResponse, msg.Type())
		authMsg, ok := msg.(*wire.AuthResponseMsg)
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

FIXME temporarily commented out since it fails the pipeline too often.
func TestClient_Multiplexing(t *testing.T) {
	t.Run("1/1", func(t *testing.T) { testClientMultiplexing(t, 1, 1) })
	t.Run("1/1024", func(t *testing.T) { testClientMultiplexing(t, 1, 1024) })
	t.Run("1024/1", func(t *testing.T) { testClientMultiplexing(t, 1024, 1) })
	t.Run("32/32", func(t *testing.T) { testClientMultiplexing(t, 32, 32) })
}

// TODO: move into wire/net as EndpointRegistry tests, if test doesn't exist yet.
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
	connHub := new(wiretest.ConnHub)
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
			id, connHub.NewNetDialer(), &DummyFunder{t}, &DummyAdjudicator{t}, wtest.RandomWallet())
		go listeners[i].Listen(connHub.NewNetListener(listeners[i].id.Address()))
	}
	for i := range dialers {
		id := wtest.NewRandomAccount(rng)
		dialers[i] = New(id, connHub.NewNetDialer(), &DummyFunder{t}, &DummyAdjudicator{t}, wtest.RandomWallet())
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
				var peers *wire.EndpointRegistry
				var addr wire.Address
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

*/

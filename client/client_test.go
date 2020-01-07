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

	simwallet "perun.network/go-perun/backend/sim/wallet"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/peer"
	peertest "perun.network/go-perun/peer/test"
	perunsync "perun.network/go-perun/pkg/sync"
	"perun.network/go-perun/wallet"
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

type DummyProposalHandler struct {
	t *testing.T
}

func (d DummyProposalHandler) Handle(_ *ChannelProposalReq, _ *ProposalResponder) {
	d.t.Fatal("BUG: DummyProposalHandler called")
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

type DummySettler struct {
	t *testing.T
}

func (d *DummySettler) Settle(context.Context, channel.SettleReq, wallet.Account) error {
	d.t.Error("DummySettler.Settle called")
	return errors.New("DummySettler.Settle called")
}

func TestClient_New_NilHandlerPanic(t *testing.T) {
	rng := rand.New(rand.NewSource(0x1111))
	id := simwallet.NewRandomAccount(rng)
	assert.Panics(t, func() { New(id, nil, nil, nil, nil) })
}

func TestClient_New(t *testing.T) {
	rng := rand.New(rand.NewSource(0x1a2b3c))
	id := simwallet.NewRandomAccount(rng)
	dialer := new(DummyDialer)
	proposalHandler := new(DummyProposalHandler)
	c := New(id, dialer, proposalHandler, &DummyFunder{t}, &DummySettler{t})

	require.NotNil(t, c)
	assert.NotNil(t, c.peers)
	assert.Same(t, c.propHandler, proposalHandler)
}

func TestClient_NewAndListen_ListenerClose(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	rng := rand.New(rand.NewSource(0x1a2b3c))
	id := simwallet.NewRandomAccount(rng)
	dialer := &DummyDialer{t}
	proposalHandler := &DummyProposalHandler{t}
	c := New(id, dialer, proposalHandler, &DummyFunder{t}, &DummySettler{t})

	require.NotNil(c)

	done := make(chan struct{})
	listener := NewDummyListener(t)
	numGoroutines := runtime.NumGoroutine()

	go func() {
		defer close(done)
		c.Listen(listener)
	}()

	assert.NoError(c.Close())

	select {
	case <-listener.Closed():
		break
	case <-time.After(1 * time.Second):
		t.Error("Listener apparently not stopped")
	}

	// yield processor to give the goroutine above time to terminate itself
	// (it may be put to sleep by the scheduler after closing the channel)
	time.Sleep(100 * time.Millisecond)

	assert.Equal(numGoroutines, runtime.NumGoroutine())
}

func TestClient_NewAndListen(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	rng := rand.New(rand.NewSource(0x1))
	connHub := new(peertest.ConnHub)
	c := New(simwallet.NewRandomAccount(rng), &DummyDialer{t}, &DummyProposalHandler{t}, &DummyFunder{t}, &DummySettler{t})
	// initialize the listener instance in the main goroutine
	// if it is initialized in a goroutine, the goroutine may be put to sleep
	// and the dialer may complain about a nonexistent listener
	listener := connHub.NewListener(c.id.Address())
	listenerDone := make(chan struct{})

	go func() {
		defer close(listenerDone)
		c.Listen(listener)
	}()

	require.Zero(c.peers.NumPeers())
	dialerDone := make(chan struct{})

	go func() {
		defer close(dialerDone)

		dialer := connHub.NewDialer()
		defer dialer.Close()
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		conn, err := dialer.Dial(ctx, c.id.Address())
		assert.NoError(err, "Dialing the Client instance failed")
		assert.NoError(conn.Send(wire.NewPongMsg()))

		for {
			msg, err := conn.Recv()
			if err != nil {
				break
			}
			assert.Equal(wire.AuthResponse, msg.Type())
			authMsg, ok := msg.(*peer.AuthResponseMsg)
			assert.True(ok, "Have a message with type AuthResponse but cast failed")
			assert.Equal(c.id.Address(), authMsg.Address)
		}
	}()

	select {
	case <-time.After(timeout):
		t.Fatal("Authentication exchange timed out")
	case <-dialerDone:
	}

	assert.Zero(c.peers.NumPeers())

	// make a successful connection
	dialerDone = make(chan struct{})
	peerID := simwallet.NewRandomAccount(rng)

	go func() {
		defer close(dialerDone)

		dialer := connHub.NewDialer()
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		conn, err := dialer.Dial(ctx, c.id.Address())
		assert.NoError(err, "Dialing the Client instance failed")
		assert.NoError(conn.Send(peer.NewAuthResponseMsg(peerID)))

		msg, err := conn.Recv()
		assert.NoError(err)
		assert.Equal(wire.AuthResponse, msg.Type())
		authMsg, ok := msg.(*peer.AuthResponseMsg)
		assert.True(ok, "Have a message with type AuthResponse but cast failed")
		assert.Equal(c.id.Address(), authMsg.Address)
	}()

	select {
	case <-time.After(timeout):
		t.Fatal("Second authentication exchange timed out")
	case <-dialerDone:
	}

	time.Sleep(timeout)
	assert.Equal(1, c.peers.NumPeers())
	assert.True(c.peers.Has(peerID.Address()))

	assert.NoError(c.Close())

	select {
	case <-time.After(timeout):
		t.Fatal("Listener close timed out")
	case <-listenerDone:
	}
}

func TestClient_Multiplexing(t *testing.T) {
	testClientMultiplexing(t, 1, 1)
	testClientMultiplexing(t, 1, 1024)
	testClientMultiplexing(t, 1024, 1)
	testClientMultiplexing(t, 32, 32)
}

func testClientMultiplexing(
	t *testing.T, numListeners, numDialers int) {
	assert := assert.New(t)

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

	for i := 0; i < numListeners; i++ {
		i := i
		id := simwallet.NewRandomAccount(rng)
		listeners[i] = New(
			id, connHub.NewDialer(), &DummyProposalHandler{t}, &DummyFunder{t}, &DummySettler{t})
		go listeners[i].Listen(connHub.NewListener(listeners[i].id.Address()))
	}
	for i := 0; i < numDialers; i++ {
		id := simwallet.NewRandomAccount(rng)
		dialers[i] = New(id, connHub.NewDialer(), &DummyProposalHandler{t}, &DummyFunder{t}, &DummySettler{t})
	}

	hostBarrier := new(sync.WaitGroup)
	peerBarrier := make(chan struct{})
	// every dialing client connects to every listening client
	numConnections := numListeners * numDialers
	hostBarrier.Add(numConnections)

	// create connections
	for i := 0; i < numDialers; i++ {
		for j := 0; j < numListeners; j++ {
			sleepTime := time.Duration(rng.Int63n(10) + 1)

			go func(i, j int) {
				defer hostBarrier.Done()

				<-peerBarrier
				time.Sleep(sleepTime * time.Millisecond)

				// trigger dialing
				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				defer cancel()
				p, err := dialers[i].peers.Get(ctx, listeners[j].id.Address())
				assert.NoError(err)
				if err == nil {
					assert.NotNil(p)
				}
			}(i, j)
		}
	}

	close(peerBarrier)
	hostBarrier.Wait()
	// race tests fail with lower sleep because not all Client.Listen routines
	// have added the peer to their registry yet.
	time.Sleep(200 * time.Millisecond)

	for i := 0; i < numDialers; i++ {
		assert.Equal(numListeners, dialers[i].peers.NumPeers())
	}

	for i := 0; i < numListeners; i++ {
		assert.Equal(numDialers, listeners[i].peers.NumPeers())
	}

	for i := 0; i < numDialers; i++ {
		for j := 0; j < numListeners; j++ {
			assert.True(dialers[i].peers.Has(listeners[j].id.Address()))
			assert.True(listeners[j].peers.Has(dialers[i].id.Address()))
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

				assert.True(peers.Has(addr))
				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				defer cancel()
				p, err := peers.Get(ctx, addr)
				assert.NoError(err)
				if err == nil {
					assert.NotNil(p)
					assert.NoError(p.Close())
				}
			}(i, k)
		}
	}

	close(peerBarrier)
	hostBarrier.Wait()
	time.Sleep(100 * time.Millisecond)

	for i, l := range listeners {
		np := l.peers.NumPeers()
		assert.Zerof(np, "Listener %d has an unexpected number of peers", i)
		assert.NoErrorf(l.Close(), "closing listener[%d]", i)
	}

	for i, d := range dialers {
		np := d.peers.NumPeers()
		assert.Zero(np, "Listener %d has an unexpected number of peers: %d", i, np)
		assert.NoErrorf(d.Close(), "closing dialer[%d]", i)
	}
}

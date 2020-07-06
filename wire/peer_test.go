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

	_ "perun.network/go-perun/backend/sim" // backend init
	wallettest "perun.network/go-perun/wallet/test"
)

// setup is a test setup consisting of two connected peers.
// It is also a mock dialer.
type setup struct {
	mutex  sync.RWMutex
	closed bool
	alice  *client
	bob    *client
}

// makeSetup creates a test setup.
func makeSetup(t *testing.T) *setup {
	a, b := newPipeConnPair()
	rng := rand.New(rand.NewSource(0xb0baFEDD))
	// We need the setup address when constructing the clients.
	s := new(setup)
	*s = setup{
		alice: makeClient(t, a, rng, s),
		bob:   makeClient(t, b, rng, s),
	}
	return s
}

// Dial simulates creating a connection to a
func (s *setup) Dial(ctx context.Context, addr Address) (Conn, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.closed {
		return nil, errors.New("dialer closed")
	}

	// a: Alice's end, b: Bob's end.
	a, b := newPipeConnPair()

	if addr.Equals(s.alice.peer.PerunAddress) { // Dialing Bob?
		s.bob.Registry.addPeer(s.bob.peer.PerunAddress, b) // Bob accepts connection.
		return a, nil
	} else if addr.Equals(s.bob.peer.PerunAddress) { // Dialing Alice?
		s.alice.Registry.addPeer(s.alice.peer.PerunAddress, a) // Alice accepts connection.
		return b, nil
	} else {
		return nil, errors.New("unknown peer")
	}
}

func (s *setup) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.closed {
		return errors.New("dialer closed")
	}
	s.closed = true
	return nil
}

// client is a simulated client in the test setup.
// All of the client's incoming messages can be read from its receiver.
type client struct {
	peer *Peer
	*Registry
	*Receiver
}

// makeClient creates a simulated test client.
func makeClient(t *testing.T, conn Conn, rng *rand.Rand, dialer Dialer) *client {
	var receiver = NewReceiver()
	var registry = NewRegistry(wallettest.NewRandomAccount(rng), func(p *Peer) {
		assert.NoError(
			t,
			p.Subscribe(receiver, func(Msg) bool { return true }),
			"failed to subscribe a new peer")
	}, dialer)

	return &client{
		peer:     registry.addPeer(wallettest.NewRandomAddress(rng), conn),
		Registry: registry,
		Receiver: receiver,
	}
}

// TestPeer_Close tests that closing a peer will make the peer object unusable.
func TestPeer_Close(t *testing.T) {
	t.Parallel()
	s := makeSetup(t)
	// Remember bob's address for later, we will need it for a registry lookup.
	bobAddress := s.alice.peer.PerunAddress
	// The lookup needs to work because the test relies on it.
	found, _ := s.alice.Registry.find(bobAddress)
	assert.Equal(t, s.alice.peer, found)
	// Close Alice's connection to Bob.
	assert.NoError(t, s.alice.peer.Close(), "closing a peer once must succeed")
	assert.Error(t, s.alice.peer.Close(), "closing peers twice must fail")

	// Sending over closed peers (not connections) must fail.
	err := s.alice.peer.Send(context.Background(), NewPingMsg())
	assert.Error(t, err, "sending to bob must fail", err)
}

func TestPeer_Send_ImmediateAbort(t *testing.T) {
	t.Parallel()
	s := makeSetup(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// This operation should abort immediately.
	assert.Error(t, s.alice.peer.Send(ctx, NewPingMsg()))

	assert.True(t, s.alice.peer.IsClosed(), "peer must be closed after failed sending")
}

func TestPeer_Send_Timeout(t *testing.T) {
	t.Parallel()
	conn, _ := newPipeConnPair()
	p := newPeer(nil, conn, nil)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	assert.Error(t, p.Send(ctx, NewPingMsg()),
		"Send() must timeout on blocked connection")
	assert.True(t, p.IsClosed(), "peer must be closed after failed Send()")
}

func TestPeer_Send_Timeout_Mutex_TryLockCtx(t *testing.T) {
	t.Parallel()
	conn, remote := newPipeConnPair()
	p := newPeer(nil, conn, nil)

	go remote.Recv()
	p.sending.Lock()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	assert.Error(t, p.Send(ctx, NewPingMsg()),
		"Send() must timeout on locked mutex")
	assert.True(t, p.IsClosed(), "peer must be closed after failed Send()")
}

func TestPeer_Send_Close(t *testing.T) {
	t.Parallel()
	conn, _ := newPipeConnPair()
	p := newPeer(nil, conn, nil)

	go func() {
		<-time.NewTimer(timeout).C
		p.Close()
	}()
	assert.Error(t, p.Send(context.Background(), NewPingMsg()),
		"Send() must be aborted by Close()")
}

func TestPeer_IsClosed(t *testing.T) {
	t.Parallel()
	s := makeSetup(t)
	assert.False(t, s.alice.peer.IsClosed(), "fresh peer must be open")
	assert.NoError(t, s.alice.peer.Close(), "closing must succeed")
	assert.True(t, s.alice.peer.IsClosed(), "closed peer must be closed")
}

func TestPeer_create(t *testing.T) {
	t.Parallel()
	p := newPeer(nil, nil, nil)
	assert.False(t, p.exists(), "peer must not yet exist")

	conn := newMockConn(nil)
	p.create(conn)

	assert.True(t, p.exists(), "peer must exist")

	assert.False(t, conn.closed.IsSet(),
		"Peer.create() on nonexisting peers must not close the new connection")

	conn2 := newMockConn(nil)
	p.create(conn2)
	assert.True(t, conn2.closed.IsSet(),
		"Peer.create() on existing peers must close the new connection")
}

// TestPeer_ClosedByRecvLoopOnConnClose is a regression test for
// #181 `peer.Peer` does not handle connection termination properly
func TestPeer_ClosedByRecvLoopOnConnClose(t *testing.T) {
	t.Parallel()
	eofReceived := make(chan struct{})
	onCloseCalled := make(chan struct{})

	rng := rand.New(rand.NewSource(0xcaffe2))
	addr := wallettest.NewRandomAddress(rng)
	conn0, conn1 := newPipeConnPair()
	peer := newPeer(addr, conn0, nil)
	assert.True(t, peer.OnClose(func() {
		close(onCloseCalled)
	}))

	go func() {
		peer.recvLoop()
		close(eofReceived)
	}()

	conn1.Close()
	<-eofReceived

	select {
	case <-onCloseCalled:
	case <-time.After(10 * time.Millisecond):
		t.Error("on close callback time-out")
	}

	assert.True(t, peer.IsClosed())
}

func TestPeer_WaitExists_Timeout(t *testing.T) {
	t.Parallel()
	p := newPeer(nil, nil, nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	assert.False(t, p.waitExists(ctx))
}

func TestPeer_WaitExists_2nd_Close(t *testing.T) {
	t.Parallel()
	p := newPeer(nil, nil, nil)
	go func() {
		<-time.After(timeout)
		p.Close()
	}()
	assert.False(t, p.waitExists(context.Background()))
}

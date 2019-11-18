// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

// Package peer contains the peer connection related code.
package peer

import (
	"context"
	"io"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"perun.network/go-perun/backend/sim/wallet"
	wire "perun.network/go-perun/wire/msg"
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
	// We need the setup adress when constructing the clients.
	s := new(setup)
	*s = setup{
		alice: makeClient(t, a, rng, s),
		bob:   makeClient(t, b, rng, s),
	}
	return s
}

// Dial simulates creating a connection to a peer.
func (s *setup) Dial(ctx context.Context, addr Address) (Conn, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.closed {
		return nil, errors.New("dialer closed")
	}

	// a: Alice's end, b: Bob's end.
	a, b := newPipeConnPair()

	if addr.Equals(s.alice.peer.PerunAddress) { // Dialing Bob?
		s.bob.Registry.Register(s.bob.peer.PerunAddress, b) // Bob accepts connection.
		return a, nil
	} else if addr.Equals(s.bob.peer.PerunAddress) { // Dialing Alice?
		s.alice.Registry.Register(s.alice.peer.PerunAddress, a) // Alice accepts connection.
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
func makeClient(t *testing.T, conn Conn, rng io.Reader, dialer Dialer) *client {
	var receiver = NewReceiver()
	var registry = NewRegistry(func(p *Peer) {
		assert.NoError(
			t,
			receiver.Subscribe(p, func(wire.Msg) bool { return true }),
			"failed to subscribe a new peer")
	}, dialer)

	return &client{
		peer:     registry.Register(wallet.NewRandomAddress(rng), conn),
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
	err := s.alice.peer.Send(context.Background(), wire.NewPingMsg())
	assert.Error(t, err, "sending to bob must fail", err)
}

func TestPeer_Send_ImmediateAbort(t *testing.T) {
	t.Parallel()
	s := makeSetup(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// This operation should abort immediately.
	assert.Error(t, s.alice.peer.Send(ctx, wire.NewPingMsg()))

	assert.True(t, s.alice.peer.isClosed(), "peer must be closed after failed sending")
}

func TestPeer_Send_Timeout(t *testing.T) {
	conn, _ := newPipeConnPair()
	p := newPeer(nil, conn, nil, nil)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	assert.Error(t, p.Send(ctx, wire.NewPingMsg()),
		"Send() must timeout on blocked connection")
	assert.True(t, p.isClosed(), "peer must be closed after failed Send()")
}

func TestPeer_Send_Close(t *testing.T) {
	conn, _ := newPipeConnPair()
	p := newPeer(nil, conn, nil, nil)

	go func() {
		<-time.NewTimer(timeout).C
		p.Close()
	}()
	assert.Error(t, p.Send(context.Background(), wire.NewPingMsg()),
		"Send() must be aborted by Close()")
}

func TestPeer_isClosed(t *testing.T) {
	s := makeSetup(t)
	assert.False(t, s.alice.peer.isClosed(), "fresh peer must be open")
	assert.NoError(t, s.alice.peer.Close(), "closing must succeed")
	assert.True(t, s.alice.peer.isClosed(), "closed peer must be closed")
}

func TestPeer_create(t *testing.T) {
	p := newPeer(nil, nil, nil, nil)
	select {
	case <-p.exists:
		t.Fatal("peer must not yet exist")
	case <-time.NewTimer(timeout).C:
	}

	conn := newMockConn(nil)
	p.create(conn)

	select {
	case <-p.exists:
	default:
		t.Fatal("peer must exist")
	}

	assert.False(t, conn.closed.IsSet(),
		"Peer.create() on nonexisting peers must not close the new connection")

	conn2 := newMockConn(nil)
	p.create(conn2)
	assert.True(t, conn2.closed.IsSet(),
		"Peer.create() on existing peers must close the new connection")
}

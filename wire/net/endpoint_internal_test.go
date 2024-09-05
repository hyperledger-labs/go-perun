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

package net

import (
	"context"
	"math/rand"
	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"
	"sync"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	_ "perun.network/go-perun/backend/sim" // backend init
	"perun.network/go-perun/wire"
	perunio "perun.network/go-perun/wire/perunio/serializer"
	wiretest "perun.network/go-perun/wire/test"
	"polycry.pt/poly-go/test"
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
func makeSetup(rng *rand.Rand) *setup {
	a, b := newPipeConnPair()
	// We need the setup address when constructing the clients.
	s := new(setup)
	*s = setup{
		alice: makeClient(a, rng, s),
		bob:   makeClient(b, rng, s),
	}
	return s
}

// Dial simulates creating a connection to a.
func (s *setup) Dial(ctx context.Context, addr map[wallet.BackendID]wire.Address, _ wire.EnvelopeSerializer) (Conn, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.closed {
		return nil, errors.New("dialer closed")
	}

	// a: Alice's end, b: Bob's end.
	a, b := newPipeConnPair()

	//nolint:gocritic
	if channel.EqualWireMaps(addr, s.alice.endpoint.Address) { // Dialing Bob?
		s.bob.Registry.addEndpoint(s.bob.endpoint.Address, b, true) // Bob accepts connection.
		return a, nil
	} else if channel.EqualWireMaps(addr, s.bob.endpoint.Address) { // Dialing Alice?
		s.alice.Registry.addEndpoint(s.alice.endpoint.Address, a, true) // Alice accepts connection.
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
	endpoint *Endpoint
	Registry *EndpointRegistry
	*wire.Receiver
}

// makeClient creates a simulated test client.
func makeClient(conn Conn, rng *rand.Rand, dialer Dialer) *client {
	receiver := wire.NewReceiver()
	registry := NewEndpointRegistry(wiretest.NewRandomAccountMap(rng), func(map[wallet.BackendID]wire.Address) wire.Consumer {
		return receiver
	}, dialer, perunio.Serializer())

	return &client{
		endpoint: registry.addEndpoint(wiretest.NewRandomAddress(rng), conn, true),
		Registry: registry,
		Receiver: receiver,
	}
}

// TestEndpoint_Close tests that closing a peer will make the peer object unusable.
func TestEndpoint_Close(t *testing.T) {
	t.Parallel()
	rng := test.Prng(t)
	s := makeSetup(rng)
	// Remember bob's address for later, we will need it for a registry lookup.
	bobAddr := s.alice.endpoint.Address
	// The lookup needs to work because the test relies on it.
	found := s.alice.Registry.find(bobAddr)
	assert.Equal(t, s.alice.endpoint, found)
	// Close Alice's connection to Bob.
	assert.NoError(t, s.alice.endpoint.Close(), "closing a peer once must succeed")
	assert.Error(t, s.alice.endpoint.Close(), "closing peers twice must fail")

	// Sending over closed peers (not connections) must fail.
	err := s.alice.endpoint.Send(
		context.Background(),
		wiretest.NewRandomEnvelope(test.Prng(t), wire.NewPingMsg()))
	assert.Error(t, err, "sending to bob must fail", err)
}

func TestEndpoint_Send_ImmediateAbort(t *testing.T) {
	t.Parallel()
	rng := test.Prng(t)
	s := makeSetup(rng)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// This operation should abort immediately.
	assert.Error(t, s.alice.endpoint.Send(ctx,
		wiretest.NewRandomEnvelope(test.Prng(t), wire.NewPingMsg())))

	assert.Error(t, s.alice.endpoint.Close(),
		"peer must be closed after failed sending")
}

func TestEndpoint_Send_Timeout(t *testing.T) {
	t.Parallel()
	rng := test.Prng(t)
	conn, _ := newPipeConnPair()
	p := newEndpoint(nil, conn)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	assert.Error(t, p.Send(ctx, wiretest.NewRandomEnvelope(rng, wire.NewPingMsg())),
		"Send() must timeout on blocked connection")
	assert.Error(t, p.Close(),
		"peer must be closed after failed Send()")
}

func TestEndpoint_Send_Timeout_Mutex_TryLockCtx(t *testing.T) {
	t.Parallel()
	rng := test.Prng(t)
	conn, remote := newPipeConnPair()
	p := newEndpoint(nil, conn)

	go func() {
		remote.Recv() //nolint:errcheck
	}()
	p.sending.Lock()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	assert.Error(t, p.Send(ctx, wiretest.NewRandomEnvelope(rng, wire.NewPingMsg())),
		"Send() must timeout on locked mutex")
	assert.Error(t, p.Close(),
		"peer must be closed after failed Send()")
}

func TestEndpoint_Send_Close(t *testing.T) {
	t.Parallel()
	rng := test.Prng(t)
	conn, _ := newPipeConnPair()
	p := newEndpoint(nil, conn)

	go func() {
		<-time.NewTimer(timeout).C
		p.Close()
	}()

	assert.Error(t, p.Send(context.Background(), wiretest.NewRandomEnvelope(rng, wire.NewPingMsg())),
		"Send() must be aborted by Close()")
}

// TestEndpoint_ClosedByRecvLoopOnConnClose is a regression test for
// #181 `peer.Peer` does not handle connection termination properly.
func TestEndpoint_ClosedByRecvLoopOnConnClose(t *testing.T) {
	t.Parallel()
	eofReceived := make(chan struct{})

	rng := test.Prng(t)
	addr := wiretest.NewRandomAddress(rng)
	conn0, conn1 := newPipeConnPair()
	peer := newEndpoint(addr, conn0)

	go func() {
		if err := peer.recvLoop(nil); err == nil {
			close(eofReceived)
		}
	}()

	conn1.Close()
	<-eofReceived

	assert.Error(t, peer.Close())
}

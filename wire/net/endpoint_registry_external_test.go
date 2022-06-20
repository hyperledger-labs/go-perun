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

package net_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/wire"
	"perun.network/go-perun/wire/net"
	nettest "perun.network/go-perun/wire/net/test"
	perunio "perun.network/go-perun/wire/perunio/serializer"
	wiretest "perun.network/go-perun/wire/test"
	ctxtest "polycry.pt/poly-go/context/test"
	"polycry.pt/poly-go/sync"
	"polycry.pt/poly-go/test"
)

var timeout = 100 * time.Millisecond

func nilConsumer(wire.Address) wire.Consumer { return nil }

// Two nodes (1 dialer, 1 listener node) .Get() each other.
func TestEndpointRegistry_Get_Pair(t *testing.T) {
	t.Parallel()
	assert, require := assert.New(t), require.New(t)
	rng := test.Prng(t)
	var hub nettest.ConnHub
	dialerID := wiretest.NewRandomAccount(rng)
	listenerID := wiretest.NewRandomAccount(rng)
	dialerReg := net.NewEndpointRegistry(dialerID, nilConsumer, hub.NewNetDialer(), perunio.Serializer())
	listenerReg := net.NewEndpointRegistry(listenerID, nilConsumer, nil, perunio.Serializer())
	listener := hub.NewNetListener(listenerID.Address())

	done := make(chan struct{})
	go func() {
		defer close(done)
		listenerReg.Listen(listener)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*timeout)
	defer cancel()
	p, err := dialerReg.Endpoint(ctx, listenerID.Address())
	assert.NoError(err)
	require.NotNil(p)
	assert.True(p.Address.Equal(listenerID.Address()))

	// should allow the listener routine to add the peer to its registry
	time.Sleep(timeout)
	p, err = listenerReg.Endpoint(ctx, dialerID.Address())
	assert.NoError(err)
	require.NotNil(p)
	assert.True(p.Address.Equal(dialerID.Address()))

	listenerReg.Close()
	dialerReg.Close()
	assert.True(sync.IsAlreadyClosedError(listener.Close()))
	ctxtest.AssertTerminates(t, timeout, func() {
		<-done
	})
}

// Tests that calling .Get() concurrently on the same peer works properly.
func TestEndpointRegistry_Get_Multiple(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	rng := test.Prng(t)
	var hub nettest.ConnHub
	dialerID := wiretest.NewRandomAccount(rng)
	listenerID := wiretest.NewRandomAccount(rng)
	dialer := hub.NewNetDialer()
	logPeer := func(addr wire.Address) wire.Consumer {
		t.Logf("subscribing %s\n", addr)
		return nil
	}
	dialerReg := net.NewEndpointRegistry(dialerID, logPeer, dialer, perunio.Serializer())
	listenerReg := net.NewEndpointRegistry(listenerID, logPeer, nil, perunio.Serializer())
	listener := hub.NewNetListener(listenerID.Address())

	done := make(chan struct{})
	go func() {
		defer close(done)
		listenerReg.Listen(listener)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*timeout)
	defer cancel()

	const N = 4
	peers := make(chan *net.Endpoint, N)
	for i := 0; i < N; i++ {
		go func() {
			p, err := dialerReg.Endpoint(ctx, listenerID.Address())
			assert.NoError(err)
			if p != nil {
				assert.True(p.Address.Equal(listenerID.Address()))
			}
			peers <- p
		}()
	}

	ct := test.NewConcurrent(t)
	ctxtest.AssertTerminates(t, timeout, func() {
		ct.Stage("terminates", func(t test.ConcT) {
			require := require.New(t)
			p := <-peers
			require.NotNil(p)
			for i := 0; i < N-1; i++ {
				p0 := <-peers
				require.NotNil(p0)
				assert.Same(p, p0)
			}
		})
	})

	ct.Wait("terminates")

	assert.Equal(1, dialer.NumDialed())

	// should allow the listener routine to add the peer to its registry
	time.Sleep(timeout)
	p, err := listenerReg.Endpoint(ctx, dialerID.Address())
	assert.NoError(err)
	assert.NotNil(p)
	assert.True(p.Address.Equal(dialerID.Address()))
	assert.Equal(1, listener.NumAccepted())

	listenerReg.Close()
	dialerReg.Close()
	assert.True(sync.IsAlreadyClosedError(listener.Close()))
	ctxtest.AssertTerminates(t, timeout, func() { <-done })
}

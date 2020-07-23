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
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/pkg/sync"
	"perun.network/go-perun/pkg/test"
	wallettest "perun.network/go-perun/wallet/test"
	"perun.network/go-perun/wire"
	"perun.network/go-perun/wire/net"
	nettest "perun.network/go-perun/wire/net/test"
)

var timeout = 100 * time.Millisecond

func nilConsumer(wire.Address) wire.Consumer { return nil }

// Two nodes (1 dialer, 1 listener node) .Get() each other.
func TestEndpointRegistry_Get_Pair(t *testing.T) {
	t.Parallel()
	assert, require := assert.New(t), require.New(t)
	rng := rand.New(rand.NewSource(3))
	var hub nettest.ConnHub
	dialerId := wallettest.NewRandomAccount(rng)
	listenerId := wallettest.NewRandomAccount(rng)
	dialerReg := net.NewEndpointRegistry(dialerId, nilConsumer, hub.NewNetDialer())
	listenerReg := net.NewEndpointRegistry(listenerId, nilConsumer, nil)
	listener := hub.NewNetListener(listenerId.Address())

	done := make(chan struct{})
	go func() {
		defer close(done)
		listenerReg.Listen(listener)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*timeout)
	defer cancel()
	p, err := dialerReg.Get(ctx, listenerId.Address())
	assert.NoError(err)
	require.NotNil(p)
	assert.True(p.Address.Equals(listenerId.Address()))

	// should allow the listener routine to add the peer to its registry
	time.Sleep(timeout)
	p, err = listenerReg.Get(ctx, dialerId.Address())
	assert.NoError(err)
	require.NotNil(p)
	assert.True(p.Address.Equals(dialerId.Address()))

	listenerReg.Close()
	dialerReg.Close()
	assert.True(sync.IsAlreadyClosedError(listener.Close()))
	test.AssertTerminates(t, timeout, func() {
		<-done
	})
}

// Tests that calling .Get() concurrently on the same peer works properly.
func TestEndpointRegistry_Get_Multiple(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	rng := rand.New(rand.NewSource(3))
	var hub nettest.ConnHub
	dialerId := wallettest.NewRandomAccount(rng)
	listenerId := wallettest.NewRandomAccount(rng)
	dialer := hub.NewNetDialer()
	logPeer := func(addr wire.Address) wire.Consumer {
		t.Logf("subscribing %s\n", addr)
		return nil
	}
	dialerReg := net.NewEndpointRegistry(dialerId, logPeer, dialer)
	listenerReg := net.NewEndpointRegistry(listenerId, logPeer, nil)
	listener := hub.NewNetListener(listenerId.Address())

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
			p, err := dialerReg.Get(ctx, listenerId.Address())
			assert.NoError(err)
			if p != nil {
				assert.True(p.Address.Equals(listenerId.Address()))
			}
			peers <- p
		}()
	}

	ct := test.NewConcurrent(t)
	test.AssertTerminates(t, timeout, func() {
		ct.Stage("terminates", func(t require.TestingT) {
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
	p, err := listenerReg.Get(ctx, dialerId.Address())
	assert.NoError(err)
	assert.NotNil(p)
	assert.True(p.Address.Equals(dialerId.Address()))
	assert.Equal(1, listener.NumAccepted())

	listenerReg.Close()
	dialerReg.Close()
	assert.True(sync.IsAlreadyClosedError(listener.Close()))
	test.AssertTerminates(t, timeout, func() { <-done })
}

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

package net_test

import (
	"context"
	"testing"
	"time"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/wallet"

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

func nilConsumer(map[wallet.BackendID]wire.Address) wire.Consumer { return nil }

// Two nodes (1 dialer, 1 listener node) .Get() each other.
func TestEndpointRegistry_Get_Pair(t *testing.T) {
	t.Parallel()
	assert, require := assert.New(t), require.New(t)
	rng := test.Prng(t)
	var hub nettest.ConnHub
	dialerID := wiretest.NewRandomAccountMap(rng, channel.TestBackendID)
	listenerID := wiretest.NewRandomAccountMap(rng, channel.TestBackendID)
	dialerReg := net.NewEndpointRegistry(dialerID, nilConsumer, hub.NewNetDialer(), perunio.Serializer())
	listenerReg := net.NewEndpointRegistry(listenerID, nilConsumer, nil, perunio.Serializer())
	listener := hub.NewNetListener(wire.AddressMapfromAccountMap(listenerID))

	done := make(chan struct{})
	go func() {
		defer close(done)
		listenerReg.Listen(listener)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*timeout)
	defer cancel()
	p, err := dialerReg.Endpoint(ctx, wire.AddressMapfromAccountMap(listenerID))
	require.NoError(err)
	require.NotNil(p)
	assert.True(channel.EqualWireMaps(p.Address, wire.AddressMapfromAccountMap(listenerID)))

	// should allow the listener routine to add the peer to its registry
	time.Sleep(timeout)
	p, err = listenerReg.Endpoint(ctx, wire.AddressMapfromAccountMap(dialerID))
	require.NoError(err)
	require.NotNil(p)
	assert.True(channel.EqualWireMaps(p.Address, wire.AddressMapfromAccountMap(dialerID)))

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
	dialerID := wiretest.NewRandomAccountMap(rng, channel.TestBackendID)
	listenerID := wiretest.NewRandomAccountMap(rng, channel.TestBackendID)
	dialer := hub.NewNetDialer()
	logPeer := func(addr map[wallet.BackendID]wire.Address) wire.Consumer {
		t.Logf("subscribing %s\n", wire.Keys(addr))
		return nil
	}
	dialerReg := net.NewEndpointRegistry(dialerID, logPeer, dialer, perunio.Serializer())
	listenerReg := net.NewEndpointRegistry(listenerID, logPeer, nil, perunio.Serializer())
	listener := hub.NewNetListener(wire.AddressMapfromAccountMap(listenerID))

	done := make(chan struct{})
	go func() {
		defer close(done)
		listenerReg.Listen(listener)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*timeout)
	defer cancel()

	const N = 4
	peers := make(chan *net.Endpoint, N)
	for range N {
		go func() {
			p, err := dialerReg.Endpoint(ctx, wire.AddressMapfromAccountMap(listenerID))
			assert.NoError(err)
			if p != nil {
				assert.True(channel.EqualWireMaps(p.Address, wire.AddressMapfromAccountMap(listenerID)))
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
			for range N - 1 {
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
	p, err := listenerReg.Endpoint(ctx, wire.AddressMapfromAccountMap(dialerID))
	assert.NoError(err)
	assert.NotNil(p)
	assert.True(channel.EqualWireMaps(p.Address, wire.AddressMapfromAccountMap(dialerID)))
	assert.Equal(1, listener.NumAccepted())

	listenerReg.Close()
	dialerReg.Close()
	assert.True(sync.IsAlreadyClosedError(listener.Close()))
	ctxtest.AssertTerminates(t, timeout, func() { <-done })
}

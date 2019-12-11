// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package peer_test

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/peer"
	peertest "perun.network/go-perun/peer/test"
	"perun.network/go-perun/pkg/test"
	wallettest "perun.network/go-perun/wallet/test"
)

var timeout = 100 * time.Millisecond

// Two nodes (1 dialer, 1 listener node) .Get() each other.
func TestRegistry_Get_Pair(t *testing.T) {
	assert, require := assert.New(t), require.New(t)
	rng := rand.New(rand.NewSource(3))
	var hub peertest.ConnHub
	dialerId := wallettest.NewRandomAccount(rng)
	listenerId := wallettest.NewRandomAccount(rng)
	dialerReg := peer.NewRegistry(dialerId, func(*peer.Peer) {}, hub.NewDialer())
	listenerReg := peer.NewRegistry(listenerId, func(*peer.Peer) {}, nil)
	listener := hub.NewListener(listenerId.Address())

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
	assert.True(p.PerunAddress.Equals(listenerId.Address()))

	// should allow the listener routine to add the peer to its registry
	time.Sleep(timeout)
	p, err = listenerReg.Get(ctx, dialerId.Address())
	assert.NoError(err)
	require.NotNil(p)
	assert.True(p.PerunAddress.Equals(dialerId.Address()))

	assert.NoError(listenerReg.Close())
	assert.NoError(dialerReg.Close())
	assert.NoError(listener.Close())
	test.AssertTerminates(t, timeout, func() {
		<-done
	})
}

// Tests that calling .Get() concurrently on the same peer works properly.
func TestRegistry_Get_Multiple(t *testing.T) {
	assert, require := assert.New(t), require.New(t)
	rng := rand.New(rand.NewSource(3))
	var hub peertest.ConnHub
	dialerId := wallettest.NewRandomAccount(rng)
	listenerId := wallettest.NewRandomAccount(rng)
	dialer := hub.NewDialer()
	logPeer := func(p *peer.Peer) { t.Logf("subscribing %x\n", p.PerunAddress.Bytes()[:4]) }
	dialerReg := peer.NewRegistry(dialerId, logPeer, dialer)
	listenerReg := peer.NewRegistry(listenerId, logPeer, nil)
	listener := hub.NewListener(listenerId.Address())

	done := make(chan struct{})
	go func() {
		defer close(done)
		listenerReg.Listen(listener)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*timeout)
	defer cancel()

	const N = 4
	peers := make(chan *peer.Peer, N)
	for i := 0; i < N; i++ {
		go func() {
			p, err := dialerReg.Get(ctx, listenerId.Address())
			assert.NoError(err)
			if p != nil {
				assert.True(p.PerunAddress.Equals(listenerId.Address()))
			}
			peers <- p
		}()
	}

	test.AssertTerminates(t, timeout, func() {
		p := <-peers
		require.NotNil(p)
		for i := 0; i < N-1; i++ {
			p0 := <-peers
			require.NotNil(p0)
			assert.Same(p, p0)
		}
	})

	assert.Equal(1, dialer.NumDialed())

	// should allow the listener routine to add the peer to its registry
	time.Sleep(timeout)
	p, err := listenerReg.Get(ctx, dialerId.Address())
	assert.NoError(err)
	assert.NotNil(p)
	assert.True(p.PerunAddress.Equals(dialerId.Address()))
	assert.Equal(1, listener.NumAccepted())

	assert.NoError(listenerReg.Close())
	assert.NoError(dialerReg.Close())
	assert.NoError(listener.Close())
	test.AssertTerminates(t, timeout, func() { <-done })
}

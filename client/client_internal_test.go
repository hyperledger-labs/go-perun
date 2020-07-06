// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package client

import (
	"context"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	channeltest "perun.network/go-perun/channel/test"
	"perun.network/go-perun/wallet"
	wallettest "perun.network/go-perun/wallet/test"
	"perun.network/go-perun/wire"
	wiretest "perun.network/go-perun/wire/test"
)

func TestClient_getPeers(t *testing.T) {
	assert, require := assert.New(t), require.New(t)
	rng := rand.New(rand.NewSource(0xdeadbeef))

	var hub wiretest.ConnHub
	var wg sync.WaitGroup
	addr := make([]wallet.Address, 2)
	wg.Add(len(addr))
	defer wg.Wait()
	for i := range addr {
		id := wallettest.NewRandomAccount(rng)
		addr[i] = id.Address()
		l := hub.NewListener(id.Address())
		defer l.Close()
		reg := wire.NewRegistry(id, func(*wire.Peer) {}, nil)
		go func() {
			defer wg.Done()
			reg.Listen(l)
		}()
	}

	dialer := hub.NewDialer()

	id := wallettest.NewRandomAccount(rng)
	reg := wire.NewRegistry(id, func(*wire.Peer) {}, dialer)
	// dummy client that only has an id and a registry
	c := &Client{
		id:    id,
		peers: reg,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	ps, err := c.getPeers(ctx, nil)
	assert.NoError(err)
	assert.Len(ps, 0, "getPeers on nil list should return empty list")
	ps, err = c.getPeers(ctx, make([]wire.Address, 0))
	assert.NoError(err)
	assert.Len(ps, 0, "getPeers on empty list should return empty list")
	ps, err = c.getPeers(ctx, []wire.Address{c.id.Address()})
	assert.NoError(err)
	assert.Len(ps, 0, "getPeers on list only containing us should return empty list")
	ps, err = c.getPeers(ctx, []wire.Address{addr[0], c.id.Address()})
	assert.NoError(err)
	require.Len(ps, 1, "getPeers on [0, us] should return [0]")
	assert.True(ps[0].PerunAddress.Equals(addr[0]), "getPeers on [0, us] should return [0]")
	ps, err = c.getPeers(ctx, []wire.Address{c.id.Address(), addr[1]})
	assert.NoError(err)
	require.Len(ps, 1, "getPeers on [us, 1] should return [1]")
	assert.True(ps[0].PerunAddress.Equals(addr[1]), "getPeers on [us, 1] should return [1]")
	ps, err = c.getPeers(ctx, []wire.Address{addr[0], addr[1]})
	assert.NoError(err)
	require.Len(ps, 2, "getPeers on [0, 1] should return [0, 1]")
	assert.True(ps[0].PerunAddress.Equals(addr[0]), "getPeers on [0, 1] should return [0, 1]")
	assert.True(ps[1].PerunAddress.Equals(addr[1]), "getPeers on [0, 1] should return [0, 1]")
	ps, err = c.getPeers(ctx, []wire.Address{addr[0], c.id.Address(), addr[1]})
	assert.NoError(err)
	require.Len(ps, 2, "getPeers on [0, us, 1] should return [0, 1]")
	assert.True(ps[0].PerunAddress.Equals(addr[0]), "getPeers on [0, us, 1] should return [0, 1]")
	assert.True(ps[1].PerunAddress.Equals(addr[1]), "getPeers on [0, us, 1] should return [0, 1]")

	_, err = c.getPeers(ctx, []wire.Address{wallettest.NewRandomAddress(rng)})
	assert.Error(err, "getPeers on unknown address should error")
}

func TestClient_Channel(t *testing.T) {
	rng := rand.New(rand.NewSource(0xdeadbeef))
	id := wallettest.NewRandomAccount(rng)
	reg := wire.NewRegistry(id, func(*wire.Peer) {}, nil)
	// dummy client that only has an id and a registry
	c := &Client{
		id:       id,
		peers:    reg,
		channels: makeChanRegistry(),
	}

	cID := channeltest.NewRandomChannelID(rng)

	t.Run("unknown", func(t *testing.T) {
		ch, err := c.Channel(cID)
		assert.Nil(t, ch)
		assert.Error(t, err)
	})

	t.Run("known", func(t *testing.T) {
		ch1 := testCh()
		c.channels.Put(cID, ch1)

		ch2, err := c.Channel(cID)
		assert.Same(t, ch2, ch1)
		assert.NoError(t, err)
	})
}

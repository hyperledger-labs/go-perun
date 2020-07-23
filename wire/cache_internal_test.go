// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package wire

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/pkg/test"
)

func TestCache(t *testing.T) {
	assert, require := assert.New(t), require.New(t)
	rng := test.Prng(t)

	var c Cache
	require.Zero(c.Size())

	ping0 := NewRandomEnvelope(rng, NewPingMsg())
	pong := NewRandomEnvelope(rng, NewPongMsg())
	ping1 := NewRandomEnvelope(rng, NewPingMsg())
	ping2 := NewRandomEnvelope(rng, NewPingMsg())
	// we want to uniquely identify messages by their timestamp
	require.False(ping0.Msg.(*PingMsg).Created.Equal(ping1.Msg.(*PingMsg).Created))

	assert.False(c.Put(ping0), "Put into cache without predicate")
	assert.Zero(c.Size())

	isPing := func(e *Envelope) bool { return e.Msg.Type() == Ping }
	ctx, cancel := context.WithCancel(context.Background())
	c.Cache(ctx, isPing)
	assert.True(c.Put(ping0), "Put into cache with predicate")
	assert.Equal(1, c.Size())
	assert.False(c.Put(pong), "Put into cache with non-matching prediacte")
	assert.Equal(1, c.Size())
	assert.True(c.Put(ping1), "Put into cache with predicate")
	assert.Equal(2, c.Size())

	empty := c.Get(func(*Envelope) bool { return false })
	assert.Len(empty, 0)

	cancel()
	assert.False(c.Put(ping2), "Put into cache with canceled predicate")
	assert.Equal(2, c.Size())
	assert.Len(c.preds, 0, "internal: Put should have removed canceled predicate")

	msgs := c.Get(func(e *Envelope) bool {
		return e.Msg.Type() == Ping &&
			e.Msg.(*PingMsg).Created.Equal(ping0.Msg.(*PingMsg).Created)
	})
	assert.Equal(1, c.Size())
	require.Len(msgs, 1)
	assert.Same(msgs[0], ping0)

	c.Cache(context.Background(), isPing)
	c.Flush()
	assert.Equal(0, c.Size())
	assert.False(c.Put(ping0), "flushed cache should not hold any predicates")
}

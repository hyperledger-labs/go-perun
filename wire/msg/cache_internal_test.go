// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package msg

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	var c Cache
	require.Zero(c.Size())

	type a struct{} // annex dummy type
	a0, ao, a1, a2 := &a{}, &a{}, &a{}, &a{}
	ping0, pong := NewPingMsg(), NewPongMsg()
	ping1, ping2 := NewPingMsg(), NewPingMsg()
	// we want to uniquely identify messages by their timestamp
	require.False(ping0.Created.Equal(ping1.Created))

	assert.False(c.Put(ping0, a0), "Put into cache without predicate")
	assert.Zero(c.Size())

	isPing := func(m Msg) bool { return m.Type() == Ping }
	ctx, cancel := context.WithCancel(context.Background())
	c.Cache(ctx, isPing)
	assert.True(c.Put(ping0, a0), "Put into cache with predicate")
	assert.Equal(1, c.Size())
	assert.False(c.Put(pong, ao), "Put into cache with non-matching prediacte")
	assert.Equal(1, c.Size())
	assert.True(c.Put(ping1, a1), "Put into cache with predicate")
	assert.Equal(2, c.Size())

	empty := c.Get(func(Msg) bool { return false })
	assert.Len(empty, 0)

	cancel()
	assert.False(c.Put(ping2, a2), "Put into cache with canceled predicate")
	assert.Equal(2, c.Size())
	assert.Len(c.preds, 0, "internal: Put should have removed canceled predicate")

	msgs := c.Get(func(m Msg) bool {
		return m.Type() == Ping && m.(*PingMsg).Created.Equal(ping0.Created)
	})
	assert.Equal(1, c.Size())
	require.Len(msgs, 1)
	assert.Same(msgs[0].Msg, ping0)
	assert.Same(msgs[0].Annex, a0)

	c.Cache(context.Background(), isPing)
	c.Flush()
	assert.Equal(0, c.Size())
	assert.False(c.Put(ping0, a0), "flushed cache should not hold any predicates")
}

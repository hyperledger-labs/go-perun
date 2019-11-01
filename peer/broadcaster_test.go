// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package peer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	wire "perun.network/go-perun/wire/msg"
)

// TestBroadcaster_Send broadcasts a message and check that the message is sent
// correctly.
func TestBroadcaster_Send(t *testing.T) {
	N := 5
	peers := make([]*Peer, N)
	msg := wire.NewPingMsg()
	check := func(m wire.Msg) { assert.Equal(t, msg, m) }

	for i := 0; i < N; i++ {
		peers[i] = newPeer(nil, newMockConn(check), nil, nil)
	}

	b := NewBroadcaster(peers)

	assert.Nil(t, b.Send(context.Background(), msg), "broadcast must succeed")
}

// TestBroadcaster_Send_Error tests that when a single transmission fails, the
// whole operation fails.
func TestBroadcaster_Send_Error(t *testing.T) {
	N := 5
	peers := make([]*Peer, N)

	for i := 0; i < N; i++ {
		peers[i] = newPeer(nil, newMockConn(nil), nil, nil)
	}

	peers[1].Close()

	b := NewBroadcaster(peers)

	err := b.Send(context.Background(), wire.NewPingMsg())
	require.Error(t, err, "broadcast must fail")
	assert.Equal(t, len(err.errors), 1)
	assert.Equal(t, err.errors[0].index, 1)
	assert.Equal(t, err.Error(), "failed to send message:\npeer[1]: "+err.errors[0].err.Error())
}
